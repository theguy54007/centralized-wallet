package wallet

import (
	"bytes"
	"centralized-wallet/internal/auth"
	"centralized-wallet/internal/models"
	mockAuth "centralized-wallet/tests/mocks/auth"
	mockRedis "centralized-wallet/tests/mocks/redis"
	mockTransaction "centralized-wallet/tests/mocks/transaction"
	mockWallet "centralized-wallet/tests/mocks/wallet"
	"fmt"

	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var mockHandlerTestHelper struct {
	transactionSerivce *mockTransaction.MockTransactionService
	walletService      *mockWallet.MockWalletService
	blacklistService   *mockAuth.MockBlacklistService
	redisClient        *mockRedis.MockRedisClient
}

// Helper function to setup the router with services

func setupHandlerMock() {
	mockHandlerTestHelper.transactionSerivce = new(mockTransaction.MockTransactionService)
	mockHandlerTestHelper.walletService = new(mockWallet.MockWalletService)
	mockHandlerTestHelper.blacklistService = new(mockAuth.MockBlacklistService)
	mockHandlerTestHelper.redisClient = new(mockRedis.MockRedisClient)

	mockHandlerTestHelper.redisClient.On("Get", mock.Anything, "user:1:wallet_number").Return(testFromWalletNumber, nil)
}

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	setupHandlerMock()

	// should run every time a request is made
	mockHandlerTestHelper.blacklistService.On("IsTokenBlacklisted", generateJWTForTest(testUserID)).Return(false, nil)

	walletRoutes := router.Group("/wallets")
	walletRoutes.Use(auth.JWTMiddleware(mockHandlerTestHelper.blacklistService))
	{
		walletRoutes.GET("/balance", BalanceHandler(mockHandlerTestHelper.walletService))
		walletRoutes.POST("/deposit", DepositHandler(mockHandlerTestHelper.walletService))
		walletRoutes.POST("/withdraw", WithdrawHandler(mockHandlerTestHelper.walletService))
		walletRoutes.POST("/transfer", TransferHandler(mockHandlerTestHelper.walletService))
		walletRoutes.GET("/transactions",
			WalletNumberMiddleware(mockHandlerTestHelper.walletService, mockHandlerTestHelper.redisClient),
			TransactionHistoryHandler(mockHandlerTestHelper.transactionSerivce),
		)
	}
	return router
}

// Helper function to generate a JWT token for the test
func generateJWTForTest(userID int) string {
	token, _ := auth.TestHelperGenerateJWT(userID)
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
	router := setupRouter()
	// mock the GetBalance method on the wallet service
	mockHandlerTestHelper.walletService.On("GetBalance", testUserID).Return(100.0, nil)

	// Generate JWT for user ID 1
	token := generateJWTForTest(testUserID)
	w := executeRequest(router, "GET", "/wallets/balance", nil, token)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"balance":100}`, w.Body.String())
}

// Deposit Handler Test
func TestDepositHandler(t *testing.T) {
	router := setupRouter()

	// Define the mock wallet with the expected balance and updated_at time
	mockWallet := &models.Wallet{
		UserID:    testUserID,
		Balance:   testAmount + 100.0, // Assume current balance is 100.0
		UpdatedAt: time.Now(),
	}

	// Mock the Deposit method to return the mock wallet
	mockHandlerTestHelper.walletService.On("Deposit", testUserID, testAmount).Return(mockWallet, nil)
	mockHandlerTestHelper.transactionSerivce.On("RecordTransaction", (*int)(nil), &testUserID, "deposit", testAmount).Return(nil)

	// Prepare the request body
	body := map[string]interface{}{"amount": testAmount}
	token := generateJWTForTest(testUserID)

	// Execute the request
	w := executeRequest(router, "POST", "/wallets/deposit", body, token)

	// Assert that the response code is OK
	assert.Equal(t, http.StatusOK, w.Code)

	// Format the expected response with the mock wallet data
	expectedResponse := fmt.Sprintf(`{
		"status": "success",
		"data": {
			"message": "Deposit successful",
			"balance": %.2f,
			"updated_at": "%s"
		}
	}`, mockWallet.Balance, mockWallet.UpdatedAt.Format(time.RFC3339Nano))

	// Assert that the response matches the expected result
	assert.JSONEq(t, expectedResponse, w.Body.String())
}

// Withdraw Handler Test
func TestWithdrawHandler(t *testing.T) {
	router := setupRouter()

	// Define the mock wallet with the expected balance and updated_at time
	mockWallet := &models.Wallet{
		UserID:    testUserID,
		Balance:   50.0, // Assume current balance is 50.0 after withdrawal
		UpdatedAt: time.Now(),
	}

	// Mock the Withdraw method to return the mock wallet
	mockHandlerTestHelper.walletService.On("Withdraw", testUserID, testAmount).Return(mockWallet, nil)
	mockHandlerTestHelper.transactionSerivce.On("RecordTransaction", &testUserID, (*int)(nil), "withdraw", testAmount).Return(nil)

	// Prepare the request body
	body := map[string]interface{}{"amount": testAmount}
	token := generateJWTForTest(testUserID)

	// Execute the request
	w := executeRequest(router, "POST", "/wallets/withdraw", body, token)

	// Assert that the response code is OK
	assert.Equal(t, http.StatusOK, w.Code)

	// Format the expected response with the mock wallet data
	expectedResponse := fmt.Sprintf(`{
		"status": "success",
		"data": {
			"message": "Withdrawal successful",
			"balance": %.2f,
			"updated_at": "%s"
		}
	}`, mockWallet.Balance, mockWallet.UpdatedAt.Format(time.RFC3339Nano))

	// Assert that the response matches the expected result
	assert.JSONEq(t, expectedResponse, w.Body.String())
}

// Transfer Handler Tests
func TestTransferHandler(t *testing.T) {
	// Define the test cases
	testCases := []struct {
		name             string
		userID           int
		fromWalletNumber string
		toWalletNumber   string
		amount           float64
		mockSetup        func()
		expectedStatus   int
		expectedResponse string
	}{
		{
			name:             "ToUserNotExist",
			userID:           testUserID,
			fromWalletNumber: testFromWalletNumber,
			toWalletNumber:   testToWalletNumber,
			amount:           50.0,
			mockSetup: func() {
				// Mock Transfer with user existence failure, returning the correct typed nil
				mockHandlerTestHelper.walletService.On("Transfer", testUserID, testToWalletNumber, 50.0).
					Return((*models.Wallet)(nil), fmt.Errorf("to_user_id does not exist"))
			},
			expectedStatus:   http.StatusBadRequest,
			expectedResponse: `{"status":"error","error":"to_user_id does not exist"}`,
		},
		{
			name:             "FromUserNotExist",
			userID:           testUserID,
			fromWalletNumber: testFromWalletNumber,
			toWalletNumber:   testToWalletNumber,
			amount:           50.0,
			mockSetup: func() {
				// Mock Transfer with user existence failure, returning the correct typed nil
				mockHandlerTestHelper.walletService.On("Transfer", testUserID, testToWalletNumber, 50.0).
					Return((*models.Wallet)(nil), fmt.Errorf("from_user_id does not exist"))
			},
			expectedStatus:   http.StatusBadRequest,
			expectedResponse: `{"status":"error","error":"from_user_id does not exist"}`,
		},
		{
			name:             "Success",
			userID:           testUserID,
			fromWalletNumber: testFromWalletNumber,
			toWalletNumber:   testToWalletNumber,
			amount:           50.0,
			mockSetup: func() {
				// Mock successful transfer, returning a valid wallet
				mockHandlerTestHelper.walletService.On("Transfer", testUserID, testToWalletNumber, 50.0).
					Return(&models.Wallet{
						UserID:    testUserID,
						Balance:   100.0,
						UpdatedAt: now,
					}, nil)

				// Mock recording the transaction
				mockHandlerTestHelper.transactionSerivce.On("RecordTransaction", &testUserID, &testToWalletNumber, "transfer", 50.0).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedResponse: fmt.Sprintf(`{
				"status": "success",
				"data": {
					"message": "Transfer successful",
					"balance": 100.0,
					"updated_at": "%s"
				}
			}`, now.Format(time.RFC3339Nano)),
		},
	}

	// Iterate over the test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			router := setupRouter()

			// Setup mocks for the specific test case
			tc.mockSetup()

			// Prepare the request body
			body := map[string]interface{}{"to_wallet_number": tc.toWalletNumber, "amount": tc.amount}

			// Generate JWT for the sender (from_user_id)
			token := generateJWTForTest(tc.userID)

			// Execute the request using the reusable function
			w := executeRequest(router, "POST", "/wallets/transfer", body, token)

			// Assert status code and response body
			assert.Equal(t, tc.expectedStatus, w.Code)
			assert.JSONEq(t, tc.expectedResponse, w.Body.String())
		})
	}
}

// Transaction History Test
func TestTransactionHistoryHandler(t *testing.T) {
	router := setupRouter()

	transactions := []models.TransactionWithEmails{
		{
			Transaction: models.Transaction{
				ID:               1,
				FromWalletNumber: nil,
				ToWalletNumber:   &testToWalletNumber,
				Type:             "deposit",
				Amount:           100.0,
				CreatedAt:        now,
			},
			FromEmail: nil,
			ToEmail:   &testEmail,
		},
		{
			Transaction: models.Transaction{
				ID:               2,
				FromWalletNumber: &testFromWalletNumber,
				ToWalletNumber:   nil,
				Type:             "withdraw",
				Amount:           testAmount,
				CreatedAt:        now,
			},
			FromEmail: &testEmail,
			ToEmail:   nil,
		},
	}

	mockHandlerTestHelper.transactionSerivce.On("GetTransactionHistory", testFromWalletNumber, "desc", 10).Return(transactions, nil)

	token := generateJWTForTest(testUserID)

	w := executeRequest(router, "GET", "/wallets/transactions?order=desc&limit=10", nil, token)

	expectedResponse := fmt.Sprintf(`{
		"transactions": [
			{
				"id":1,
				"from_wallet_number":null,
				"to_wallet_number":"%s",
				"from_email":null,
				"to_email":"%s",
				"type":"deposit",
				"amount":100.0,
				"created_at":"%s"
			},
			{
				"id":2,
				"from_wallet_number":"%s",
				"to_wallet_number":null,
				"from_email":"%s",
				"to_email":null,
				"type":"withdraw",
				"amount":50.0,
				"created_at":"%s"
			}
		]
	}`, testToWalletNumber, testEmail, now.Format(time.RFC3339Nano), testFromWalletNumber, testEmail, now.Format(time.RFC3339Nano))
	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, expectedResponse, w.Body.String())
}
