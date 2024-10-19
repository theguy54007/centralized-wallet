package wallet

import (
	"bytes"
	"centralized-wallet/internal/middleware"
	"centralized-wallet/internal/models"
	"centralized-wallet/internal/transaction"
	"centralized-wallet/internal/user"
	mockTransaction "centralized-wallet/tests/mocks/transaction"
	"centralized-wallet/tests/mocks/wallet"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

var (
	testUserID   = 1
	testToUserID = 2
	testAmount   = 50.0
	testEmail    = "user1@example.com"
)

var now = time.Now()

// Helper function to setup the router with services
func setupRouter(walletService *WalletService, transactionService transaction.TransactionServiceInterface) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	walletRoutes := router.Group("/wallets")
	walletRoutes.Use(middleware.JWTMiddleware())
	{
		walletRoutes.GET("/balance", BalanceHandler(walletService))
		walletRoutes.POST("/deposit", DepositHandler(walletService))
		walletRoutes.POST("/withdraw", WithdrawHandler(walletService))
		walletRoutes.POST("/transfer", TransferHandler(walletService))
		walletRoutes.GET("/transactions", TransactionHistoryHandler(transactionService))
	}
	return router
}

// Helper function to generate a JWT token for the test
func generateJWTForTest(userID int) string {
	token, _ := user.TestHelperGenerateJWT(userID)
	return token
}

// Helper function to prepare and execute a request
func executeRequest(router *gin.Engine, method, url string, body interface{}, token string) *httptest.ResponseRecorder {
	var reqBody []byte
	if body != nil {
		reqBody, _ = json.Marshal(body)
	}
	req, _ := http.NewRequest(method, url, bytes.NewBuffer(reqBody))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

// Balance Handler Test
func TestBalanceHandler(t *testing.T) {
	mockWalletRepo := new(wallet.MockWalletRepository)
	mockTransactionService := new(mockTransaction.MockTransactionService)

	mockWalletRepo.On("GetWalletBalance", testUserID).Return(100.0, nil)

	walletService := NewWalletService(mockWalletRepo, mockTransactionService)
	router := setupRouter(walletService, mockTransactionService)

	// Generate JWT for user ID 1
	token := generateJWTForTest(testUserID)

	w := executeRequest(router, "GET", "/wallets/balance", nil, token)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"balance":100}`, w.Body.String())
}

// Deposit Handler Test
func TestDepositHandler(t *testing.T) {
	mockWalletRepo := new(wallet.MockWalletRepository)
	mockTransactionService := new(mockTransaction.MockTransactionService)

	mockWalletRepo.On("Deposit", testUserID, testAmount).Return(nil)
	mockTransactionService.On("RecordTransaction", (*int)(nil), &testUserID, "deposit", testAmount).Return(nil)

	walletService := NewWalletService(mockWalletRepo, mockTransactionService)
	router := setupRouter(walletService, mockTransactionService)

	body := map[string]interface{}{"amount": testAmount}
	token := generateJWTForTest(testUserID)

	w := executeRequest(router, "POST", "/wallets/deposit", body, token)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"message":"Deposit successful"}`, w.Body.String())
}

// Withdraw Handler Test
func TestWithdrawHandler(t *testing.T) {
	mockWalletRepo := new(wallet.MockWalletRepository)
	mockTransactionService := new(mockTransaction.MockTransactionService)

	mockWalletRepo.On("Withdraw", testUserID, testAmount).Return(nil)
	mockTransactionService.On("RecordTransaction", &testUserID, (*int)(nil), "withdraw", testAmount).Return(nil)

	walletService := NewWalletService(mockWalletRepo, mockTransactionService)
	router := setupRouter(walletService, mockTransactionService)

	body := map[string]interface{}{"amount": testAmount}
	token := generateJWTForTest(testUserID)

	w := executeRequest(router, "POST", "/wallets/withdraw", body, token)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"message":"Withdrawal successful"}`, w.Body.String())
}

// Transfer Handler Tests
func TestTransferHandler(t *testing.T) {
	mockWalletRepo := new(wallet.MockWalletRepository)
	mockTransactionService := new(mockTransaction.MockTransactionService)

	// Define the test cases
	testCases := []struct {
		name             string
		fromUserId       int
		toUserId         int
		amount           float64
		mockSetup        func()
		expectedStatus   int
		expectedResponse string
	}{
		{
			name:       "ToUserNotExist",
			fromUserId: 1,
			toUserId:   2,
			amount:     50.0,
			mockSetup: func() {
				mockWalletRepo.On("UserExists", 1).Return(true, nil)  // from_user_id exists
				mockWalletRepo.On("UserExists", 2).Return(false, nil) // to_user_id does not exist
			},
			expectedStatus:   http.StatusBadRequest,
			expectedResponse: `{"error":"to_user_id does not exist"}`,
		},
		{
			name:       "FromUserNotExist",
			fromUserId: 1,
			toUserId:   2,
			amount:     50.0,
			mockSetup: func() {
				mockWalletRepo.On("UserExists", 1).Return(false, nil) // from_user_id does not exist
				mockWalletRepo.On("UserExists", 2).Return(true, nil)  // to_user_id exists
			},
			expectedStatus:   http.StatusBadRequest,
			expectedResponse: `{"error":"from_user_id does not exist"}`,
		},
		{
			name:       "Success",
			fromUserId: 1,
			toUserId:   2,
			amount:     50.0,
			mockSetup: func() {
				mockWalletRepo.On("UserExists", 1).Return(true, nil) // from_user_id exists
				mockWalletRepo.On("UserExists", 2).Return(true, nil) // to_user_id exists
				mockWalletRepo.On("Withdraw", 1, 50.0).Return(nil)   // Withdraw from sender
				mockWalletRepo.On("Deposit", 2, 50.0).Return(nil)    // Deposit to receiver
				mockTransactionService.On("RecordTransaction", &testUserID, &testToUserID, "transfer", 50.0).Return(nil)
			},
			expectedStatus:   http.StatusOK,
			expectedResponse: `{"message":"Transfer successful"}`,
		},
	}

	// Iterate over the test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset mocks for the specific test case
			mockWalletRepo.ExpectedCalls = nil
			mockTransactionService.ExpectedCalls = nil

			// Setup mocks for the specific test case
			tc.mockSetup()

			// Create wallet service and router
			walletService := NewWalletService(mockWalletRepo, mockTransactionService)
			router := setupRouter(walletService, mockTransactionService)

			// Prepare the request body
			body := map[string]interface{}{"to_user_id": tc.toUserId, "amount": tc.amount}
			// bodyJSON, _ := json.Marshal(body)

			// Generate JWT for the sender (from_user_id)
			token := generateJWTForTest(tc.fromUserId)

			// Execute the request using the reusable function
			w := executeRequest(router, "POST", "/wallets/transfer", body, token)

			// Assert status code and response body
			assert.Equal(t, tc.expectedStatus, w.Code)
			assert.JSONEq(t, tc.expectedResponse, w.Body.String())

			// Assert that all mocked methods were called
			mockWalletRepo.AssertExpectations(t)
			mockTransactionService.AssertExpectations(t)
		})
	}
}

// Transaction History Test
func TestTransactionHistoryHandler(t *testing.T) {
	mockTransactionService := new(mockTransaction.MockTransactionService)
	router := setupRouter(nil, mockTransactionService)

	transactions := []models.TransactionWithEmails{
		{
			Transaction: models.Transaction{
				ID:         1,
				FromUserID: nil,
				ToUserID:   &testUserID,
				Type:       "deposit",
				Amount:     100.0,
				CreatedAt:  now,
			},
			FromEmail: nil,
			ToEmail:   &testEmail,
		},
		{
			Transaction: models.Transaction{
				ID:         2,
				FromUserID: &testUserID,
				ToUserID:   nil,
				Type:       "withdraw",
				Amount:     testAmount,
				CreatedAt:  now,
			},
			FromEmail: &testEmail,
			ToEmail:   nil,
		},
	}

	mockTransactionService.On("GetTransactionHistory", testUserID, "desc", 10).Return(transactions, nil)

	token := generateJWTForTest(testUserID)

	w := executeRequest(router, "GET", "/wallets/transactions?order=desc&limit=10", nil, token)

	expectedResponse := fmt.Sprintf(`{
		"transactions": [
			{"id":1,"from_user_id":null,"to_user_id":1,"from_email":null,"to_email":"%s","type":"deposit","amount":100.0,"created_at":"%s"},
			{"id":2,"from_user_id":1,"to_user_id":null,"from_email":"%s","to_email":null,"type":"withdraw","amount":50.0,"created_at":"%s"}
		]
	}`, testEmail, now.Format(time.RFC3339Nano), testEmail, now.Format(time.RFC3339Nano))

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, expectedResponse, w.Body.String())
}
