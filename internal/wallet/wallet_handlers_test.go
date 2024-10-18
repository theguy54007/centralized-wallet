package wallet

import (
	"bytes"
	"centralized-wallet/internal/middleware"
	"centralized-wallet/internal/user"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupRouterWithMiddleware(walletService *WalletService) *gin.Engine {
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
	}

	return router
}

// Helper function to generate a JWT token for the test
func generateJWTForTest(userID int) string {
	token, _ := user.TestHelperGenerateJWT(userID)
	return token
}

func TestBalanceHandler(t *testing.T) {
	mockRepo := new(MockWalletRepository)
	mockRepo.On("GetWalletBalance", 1).Return(100.0, nil)

	walletService := NewWalletService(mockRepo)
	router := setupRouterWithMiddleware(walletService)

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
	mockRepo := new(MockWalletRepository)
	mockRepo.On("Deposit", 1, 50.0).Return(nil)

	walletService := NewWalletService(mockRepo)
	router := setupRouterWithMiddleware(walletService)

	body := map[string]interface{}{"amount": 50.0}
	bodyJSON, _ := json.Marshal(body)

	// Generate JWT for user ID 1
	token := generateJWTForTest(1)

	req, _ := http.NewRequest("POST", "/wallets/deposit", bytes.NewBuffer(bodyJSON))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	expectedResponse := `{"message":"Deposit successful"}`
	assert.JSONEq(t, expectedResponse, w.Body.String())
}

// Test WithdrawHandler with JWT authentication
func TestWithdrawHandler(t *testing.T) {
	mockRepo := new(MockWalletRepository)
	mockRepo.On("Withdraw", 1, 50.0).Return(nil)

	walletService := NewWalletService(mockRepo)
	router := setupRouterWithMiddleware(walletService)

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
	mockRepo := new(MockWalletRepository)
	mockRepo.On("UserExists", 1).Return(true, nil)  // from_user_id exists
	mockRepo.On("UserExists", 2).Return(false, nil) // to_user_id does not exist
	mockRepo.On("Transfer", 1, 2, 50.0).Return(fmt.Errorf("to_user_id does not exist"))

	walletService := NewWalletService(mockRepo)
	router := setupRouterWithMiddleware(walletService)

	body := map[string]interface{}{"to_user_id": 2, "amount": 50.0}
	bodyJSON, _ := json.Marshal(body)

	// Generate JWT for user ID 1
	token := generateJWTForTest(1)

	req, _ := http.NewRequest("POST", "/wallets/transfer", bytes.NewBuffer(bodyJSON))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	expectedResponse := `{"error":"to_user_id does not exist"}`
	assert.JSONEq(t, expectedResponse, w.Body.String())
}

// Test TransferHandler with non-existent from_user_id
func TestTransferHandler_FromUserNotExist(t *testing.T) {
	mockRepo := new(MockWalletRepository)
	// from_user_id does not exist
	mockRepo.On("UserExists", 1).Return(false, nil)
	// to_user_id exists
	mockRepo.On("UserExists", 2).Return(true, nil)
	mockRepo.On("Transfer", 1, 2, 50.0).Return(fmt.Errorf("from_user_id does not exist"))

	walletService := NewWalletService(mockRepo)
	router := setupRouterWithMiddleware(walletService)

	// Set up request body
	body := map[string]interface{}{"to_user_id": 2, "amount": 50.0}
	bodyJSON, _ := json.Marshal(body)

	// Generate JWT for user ID 1 (from_user_id)
	token := generateJWTForTest(1)

	// Create the request with the JWT token
	req, _ := http.NewRequest("POST", "/wallets/transfer", bytes.NewBuffer(bodyJSON))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert the response
	assert.Equal(t, http.StatusBadRequest, w.Code)
	expectedResponse := `{"error":"from_user_id does not exist"}`
	assert.JSONEq(t, expectedResponse, w.Body.String())
}

// Test TransferHandler with valid from_user_id and to_user_id
func TestTransferHandler_Success(t *testing.T) {
	mockRepo := new(MockWalletRepository)
	mockRepo.On("UserExists", 1).Return(true, nil)  // from_user_id exists
	mockRepo.On("UserExists", 2).Return(true, nil)  // to_user_id exists
	mockRepo.On("Transfer", 1, 2, 50.0).Return(nil) // Transfer succeeds

	walletService := NewWalletService(mockRepo)
	router := setupRouterWithMiddleware(walletService)

	body := map[string]interface{}{"to_user_id": 2, "amount": 50.0}
	bodyJSON, _ := json.Marshal(body)

	// Generate JWT for user ID 1
	token := generateJWTForTest(1)

	req, _ := http.NewRequest("POST", "/wallets/transfer", bytes.NewBuffer(bodyJSON))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	expectedResponse := `{"message":"Transfer successful"}`
	assert.JSONEq(t, expectedResponse, w.Body.String())
}
