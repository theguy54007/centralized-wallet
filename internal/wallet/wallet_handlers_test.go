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
	// "github.com/stretchr/testify/mock"
)

func setupRouterWithMiddleware(walletService *WalletService, transactionService transaction.TransactionServiceInterface) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	// Apply the JWT middleware to all wallet-related routes
	walletRoutes := router.Group("/wallets")
	walletRoutes.Use(middleware.JWTMiddleware()) // Use middleware
	{
		walletRoutes.GET("/balance", BalanceHandler(walletService))
		walletRoutes.POST("/deposit", DepositHandler(walletService))
		walletRoutes.POST("/withdraw", WithdrawHandler(walletService))
		walletRoutes.POST("/transfer", TransferHandler(walletService))
		walletRoutes.GET("/transactions", TransactionHistoryHandler(transactionService)) // Pass TransactionServiceInterface
	}

	return router
}

// Helper function to generate a JWT token for the test
func generateJWTForTest(userID int) string {
	token, _ := user.TestHelperGenerateJWT(userID)
	return token
}

func TestBalanceHandler(t *testing.T) {
	mockWalletRepo := new(wallet.MockWalletRepository)
	mockTransactionService := new(mockTransaction.MockTransactionService)

	mockWalletRepo.On("GetWalletBalance", 1).Return(100.0, nil)

	walletService := NewWalletService(mockWalletRepo, mockTransactionService)
	router := setupRouterWithMiddleware(walletService, mockTransactionService) // Now it works

	// Generate JWT for user ID 1
	token := generateJWTForTest(1)

	req, _ := http.NewRequest("GET", "/wallets/balance", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	expectedResponse := `{"balance":100}`
	assert.JSONEq(t, expectedResponse, w.Body.String())
}

// Test DepositHandler with JWT authentication
func TestDepositHandler(t *testing.T) {
	mockWalletRepo := new(wallet.MockWalletRepository)
	mockTransactionService := new(mockTransaction.MockTransactionService)

	// Mock the deposit and transaction recording
	userID := 1
	mockWalletRepo.On("Deposit", userID, 50.0).Return(nil)
	mockTransactionService.On("RecordTransaction", (*int)(nil), &userID, "deposit", 50.0).Return(nil)

	// Create wallet service and router
	walletService := NewWalletService(mockWalletRepo, mockTransactionService)
	router := setupRouterWithMiddleware(walletService, mockTransactionService)

	// Prepare the request body
	body := map[string]interface{}{"amount": 50.0}
	bodyJSON, _ := json.Marshal(body)

	// Generate JWT for user ID 1
	token := generateJWTForTest(1)

	// Make the request
	req, _ := http.NewRequest("POST", "/wallets/deposit", bytes.NewBuffer(bodyJSON))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check the response
	assert.Equal(t, http.StatusOK, w.Code)
	expectedResponse := `{"message":"Deposit successful"}`
	assert.JSONEq(t, expectedResponse, w.Body.String())
}

func TestWithdrawHandler(t *testing.T) {
	mockWalletRepo := new(wallet.MockWalletRepository)
	mockTransactionService := new(mockTransaction.MockTransactionService)

	// Mock the successful withdrawal
	mockWalletRepo.On("Withdraw", 1, 50.0).Return(nil)

	userID := 1
	mockTransactionService.On("RecordTransaction", &userID, (*int)(nil), "withdraw", 50.0).Return(nil)

	walletService := NewWalletService(mockWalletRepo, mockTransactionService)
	router := setupRouterWithMiddleware(walletService, mockTransactionService)

	body := map[string]interface{}{"amount": 50.0}
	bodyJSON, _ := json.Marshal(body)

	// Generate JWT for user ID 1
	token := generateJWTForTest(1)

	req, _ := http.NewRequest("POST", "/wallets/withdraw", bytes.NewBuffer(bodyJSON))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	expectedResponse := `{"message":"Withdrawal successful"}`
	assert.JSONEq(t, expectedResponse, w.Body.String())
}

// Test TransferHandler with non-existent to_user_id
func TestTransferHandler_ToUserNotExist(t *testing.T) {
	mockWalletRepo := new(wallet.MockWalletRepository)
	mockTransactionService := new(mockTransaction.MockTransactionService)

	// Mock user existence and transfer failure
	fromUserId, toUserId := 1, 2
	mockWalletRepo.On("UserExists", fromUserId).Return(true, nil) // from_user_id exists
	mockWalletRepo.On("UserExists", toUserId).Return(false, nil)  // to_user_id does not exist
	mockWalletRepo.On("Transfer", fromUserId, toUserId, 50.0).Return(fmt.Errorf("to_user_id does not exist"))

	// Create wallet service and router
	walletService := NewWalletService(mockWalletRepo, mockTransactionService)
	router := setupRouterWithMiddleware(walletService, mockTransactionService)

	// Prepare the request body
	body := map[string]interface{}{"to_user_id": 2, "amount": 50.0}
	bodyJSON, _ := json.Marshal(body)

	// Generate JWT for user ID 1
	token := generateJWTForTest(1)

	// Make the request
	req, _ := http.NewRequest("POST", "/wallets/transfer", bytes.NewBuffer(bodyJSON))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	fmt.Println(w.Body.String())
	// Check the response
	assert.Equal(t, http.StatusBadRequest, w.Code)
	expectedResponse := `{"error":"to_user_id does not exist"}`
	assert.JSONEq(t, expectedResponse, w.Body.String())
}

// Test TransferHandler with non-existent from_user_id
func TestTransferHandler_FromUserNotExist(t *testing.T) {
	mockWalletRepo := new(wallet.MockWalletRepository)
	mockTransactionService := new(mockTransaction.MockTransactionService)

	// Mock user existence for sender (from_user_id does not exist)
	fromUserId, toUserId := 1, 2
	mockWalletRepo.On("UserExists", fromUserId).Return(false, nil) // from_user_id does not exist

	// Mock user existence for receiver (to_user_id exists)
	mockWalletRepo.On("UserExists", toUserId).Return(true, nil) // to_user_id exists

	mockWalletRepo.On("Transfer", fromUserId, toUserId, 50.0).Return(fmt.Errorf("from_user_id does not exist"))

	// Create wallet service and router
	walletService := NewWalletService(mockWalletRepo, mockTransactionService)
	router := setupRouterWithMiddleware(walletService, mockTransactionService)

	// Prepare the request body
	body := map[string]interface{}{"to_user_id": 2, "amount": 50.0}
	bodyJSON, _ := json.Marshal(body)

	// Generate JWT for user ID 1 (from_user_id)
	token := generateJWTForTest(1)

	// Make the request
	req, _ := http.NewRequest("POST", "/wallets/transfer", bytes.NewBuffer(bodyJSON))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check the response
	assert.Equal(t, http.StatusBadRequest, w.Code)
	expectedResponse := `{"error":"from_user_id does not exist"}`
	assert.JSONEq(t, expectedResponse, w.Body.String())
}

// Test TransferHandler with valid from_user_id and to_user_id
func TestTransferHandler_Success(t *testing.T) {
	mockWalletRepo := new(wallet.MockWalletRepository)
	mockTransactionService := new(mockTransaction.MockTransactionService)

	fromUserId, toUserId := 1, 2
	// Mock the Withdraw method (from sender)
	mockWalletRepo.On("UserExists", fromUserId).Return(true, nil)
	mockWalletRepo.On("UserExists", toUserId).Return(true, nil)
	mockWalletRepo.On("Withdraw", fromUserId, 50.0).Return(nil)

	// Mock the Deposit method (to receiver)
	mockWalletRepo.On("Deposit", toUserId, 50.0).Return(nil)

	// Mock transaction recording for both users
	mockTransactionService.On("RecordTransaction", &fromUserId, &toUserId, "transfer", 50.0).Return(nil)
	// mockTransactionService.On("RecordTransaction", "transfer in", 50.0).Return(nil)

	// Create wallet service and router
	walletService := NewWalletService(mockWalletRepo, mockTransactionService)
	router := setupRouterWithMiddleware(walletService, mockTransactionService)

	// Prepare the request body
	body := map[string]interface{}{"to_user_id": 2, "amount": 50.0}
	bodyJSON, _ := json.Marshal(body)

	// Generate JWT for user ID 1 (sender)
	token := generateJWTForTest(1)

	// Make the request
	req, _ := http.NewRequest("POST", "/wallets/transfer", bytes.NewBuffer(bodyJSON))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check the response
	assert.Equal(t, http.StatusOK, w.Code)
	expectedResponse := `{"message":"Transfer successful"}`
	assert.JSONEq(t, expectedResponse, w.Body.String())

	// Assert that all mocked methods were called
	mockWalletRepo.AssertExpectations(t)
	mockTransactionService.AssertExpectations(t)
}

func TestTransactionHistoryHandler(t *testing.T) {
	mockTransactionService := new(mockTransaction.MockTransactionService)
	router := setupRouterWithMiddleware(nil, mockTransactionService)

	userId := 1
	email := "user1@example.com"
	now := time.Now()

	// Mock transaction history response
	transactions := []models.TransactionWithEmails{
		{
			Transaction: models.Transaction{
				ID:         1,
				FromUserID: nil,
				ToUserID:   &userId,
				Type:       "deposit",
				Amount:     100.0,
				CreatedAt:  now,
			},
			FromEmail: nil,
			ToEmail:   &email,
		},
		{
			Transaction: models.Transaction{
				ID:         2,
				FromUserID: &userId,
				ToUserID:   nil,
				Type:       "withdraw",
				Amount:     50.0,
				CreatedAt:  now,
			},
			FromEmail: &email,
			ToEmail:   nil,
		},
	}

	// Mocking the service method
	mockTransactionService.On("GetTransactionHistory", 1, "desc", 10).Return(transactions, nil)

	// Generate JWT for user ID 1
	token := generateJWTForTest(1)

	// Make a request with query parameters for sorting and limit
	req, _ := http.NewRequest("GET", "/wallets/transactions?order=desc&limit=10", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check the response
	assert.Equal(t, http.StatusOK, w.Code)

	// Update expected response to handle null values correctly
	expectedResponse := fmt.Sprintf(`{
		"transactions": [
			{"id":1,"from_user_id":null,"to_user_id":1,"from_email":null,"to_email":"%s","type":"deposit","amount":100.0,"created_at":"%s"},
			{"id":2,"from_user_id":1,"to_user_id":null,"from_email":"%s","to_email":null,"type":"withdraw","amount":50.0,"created_at":"%s"}
		]
	}`, email, now.Format(time.RFC3339Nano), email, now.Format(time.RFC3339Nano))

	// Use JSONEq to compare the expected and actual responses
	assert.JSONEq(t, expectedResponse, w.Body.String())
}
