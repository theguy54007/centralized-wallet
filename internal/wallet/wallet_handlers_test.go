package wallet

import (
	"bytes"
	"centralized-wallet/internal/middleware"
	"centralized-wallet/internal/models"
	"centralized-wallet/internal/repository"
	"centralized-wallet/internal/transaction"
	"centralized-wallet/internal/user"
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
	mockWalletRepo := new(repository.MockWalletRepository)
	mockTransactionService := new(transaction.MockTransactionService)

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
	mockWalletRepo := new(repository.MockWalletRepository)
	mockTransactionService := new(transaction.MockTransactionService)

	// Mock the deposit and transaction recording
	mockWalletRepo.On("Deposit", 1, 50.0).Return(nil)
	mockTransactionService.On("RecordTransaction", 1, "deposit", 50.0).Return(nil)

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

// Test WithdrawHandler with JWT authentication
// Test WithdrawHandler with JWT authentication
func TestWithdrawHandler(t *testing.T) {
	mockWalletRepo := new(repository.MockWalletRepository)
	mockTransactionService := new(transaction.MockTransactionService)

	// Mock the withdraw and transaction recording
	mockWalletRepo.On("Withdraw", 1, 50.0).Return(nil)
	mockTransactionService.On("RecordTransaction", 1, "withdraw", 50.0).Return(nil)

	// Create wallet service and router
	walletService := NewWalletService(mockWalletRepo, mockTransactionService)
	router := setupRouterWithMiddleware(walletService, mockTransactionService)

	// Prepare the request body
	body := map[string]interface{}{"amount": 50.0}
	bodyJSON, _ := json.Marshal(body)

	// Generate JWT for user ID 1
	token := generateJWTForTest(1)

	// Make the request
	req, _ := http.NewRequest("POST", "/wallets/withdraw", bytes.NewBuffer(bodyJSON))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check the response
	assert.Equal(t, http.StatusOK, w.Code)
	expectedResponse := `{"message":"Withdrawal successful"}`
	assert.JSONEq(t, expectedResponse, w.Body.String())
}

// Test TransferHandler with non-existent to_user_id
func TestTransferHandler_ToUserNotExist(t *testing.T) {
	mockWalletRepo := new(repository.MockWalletRepository)
	mockTransactionService := new(transaction.MockTransactionService)

	// Mock user existence and transfer failure
	mockWalletRepo.On("UserExists", 1).Return(true, nil)  // from_user_id exists
	mockWalletRepo.On("UserExists", 2).Return(false, nil) // to_user_id does not exist
	mockWalletRepo.On("Transfer", 1, 2, 50.0).Return(fmt.Errorf("to_user_id does not exist"))

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
	mockWalletRepo := new(repository.MockWalletRepository)
	mockTransactionService := new(transaction.MockTransactionService)

	// Mock user existence for sender (from_user_id does not exist)
	mockWalletRepo.On("UserExists", 1).Return(false, nil) // from_user_id does not exist

	// Mock user existence for receiver (to_user_id exists)
	mockWalletRepo.On("UserExists", 2).Return(true, nil) // to_user_id exists

	mockWalletRepo.On("Transfer", 1, 2, 50.0).Return(fmt.Errorf("from_user_id does not exist"))

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
	mockWalletRepo := new(repository.MockWalletRepository)
	mockTransactionService := new(transaction.MockTransactionService)

	// Mock the Withdraw method (from sender)
	mockWalletRepo.On("UserExists", 1).Return(true, nil)
	mockWalletRepo.On("UserExists", 2).Return(true, nil)
	mockWalletRepo.On("Withdraw", 1, 50.0).Return(nil)

	// Mock the Deposit method (to receiver)
	mockWalletRepo.On("Deposit", 2, 50.0).Return(nil)

	// Mock transaction recording for both users
	mockTransactionService.On("RecordTransaction", 1, "transfer out", 50.0).Return(nil)
	mockTransactionService.On("RecordTransaction", 2, "transfer in", 50.0).Return(nil)

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
	// Mock the transaction service
	mockTransactionService := new(transaction.MockTransactionService)
	mockTransactionService.On("GetTransactionHistory", 1).Return([]models.Transaction{
		{ID: 1, Type: "deposit", Amount: 100.00, CreatedAt: time.Now()},
		{ID: 2, Type: "withdraw", Amount: 50.00, CreatedAt: time.Now()},
	}, nil)

	// Mock the wallet repository (if needed)
	mockWalletRepo := new(repository.MockWalletRepository)

	// Directly use the mockTransactionService in the router setup
	walletService := NewWalletService(mockWalletRepo, mockTransactionService)

	// Setup router with the middleware and transaction handler
	router := setupRouterWithMiddleware(walletService, mockTransactionService)

	// Generate JWT for user ID 1
	token := generateJWTForTest(1)

	// Perform the request to fetch transaction history
	req, _ := http.NewRequest("GET", "/wallets/transactions", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check the status code
	assert.Equal(t, http.StatusOK, w.Code)

	// Parse the response body
	var response struct {
		Transactions []models.Transaction `json:"transactions"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Check the response data
	assert.Len(t, response.Transactions, 2)
	assert.Equal(t, "deposit", response.Transactions[0].Type)
	assert.Equal(t, 100.00, response.Transactions[0].Amount)
	assert.Equal(t, "withdraw", response.Transactions[1].Type)
	assert.Equal(t, 50.00, response.Transactions[1].Amount)
}
