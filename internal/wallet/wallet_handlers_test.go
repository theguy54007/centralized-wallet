package wallet

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	return router
}

// Test BalanceHandler
func TestBalanceHandler(t *testing.T) {
	mockRepo := new(MockWalletRepository)
	mockRepo.On("GetWalletBalance", 1).Return(100.0, nil)

	walletService := NewWalletService(mockRepo)
	router := setupRouter()
	router.GET("/wallets/:user_id/balance", BalanceHandler(walletService))

	req, _ := http.NewRequest("GET", "/wallets/1/balance", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	expectedResponse := `{"balance":100}`
	assert.JSONEq(t, expectedResponse, w.Body.String())
}

// Test DepositHandler
func TestDepositHandler(t *testing.T) {
	mockRepo := new(MockWalletRepository)
	mockRepo.On("Deposit", 1, 50.0).Return(nil)

	walletService := NewWalletService(mockRepo)
	router := setupRouter()
	router.POST("/wallets/:user_id/deposit", DepositHandler(walletService))

	body := map[string]interface{}{"amount": 50.0}
	bodyJSON, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "/wallets/1/deposit", bytes.NewBuffer(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	expectedResponse := `{"message":"Deposit successful"}`
	assert.JSONEq(t, expectedResponse, w.Body.String())
}

// Test WithdrawHandler
func TestWithdrawHandler(t *testing.T) {
	mockRepo := new(MockWalletRepository)
	mockRepo.On("Withdraw", 1, 50.0).Return(nil)

	walletService := NewWalletService(mockRepo)
	router := setupRouter()
	router.POST("/wallets/:user_id/withdraw", WithdrawHandler(walletService))

	body := map[string]interface{}{"amount": 50.0}
	bodyJSON, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "/wallets/1/withdraw", bytes.NewBuffer(bodyJSON))
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
	router := setupRouter()
	router.POST("/wallets/transfer", TransferHandler(walletService))

	body := map[string]interface{}{"from_user_id": 1, "to_user_id": 2, "amount": 50.0}
	bodyJSON, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "/wallets/transfer", bytes.NewBuffer(bodyJSON))
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
	mockRepo.On("UserExists", 1).Return(false, nil) // from_user_id does not exist
	mockRepo.On("UserExists", 2).Return(true, nil)  // to_user_id exists
	mockRepo.On("Transfer", 1, 2, 50.0).Return(fmt.Errorf("from_user_id does not exist"))

	walletService := NewWalletService(mockRepo)
	router := setupRouter()
	router.POST("/wallets/transfer", TransferHandler(walletService))

	body := map[string]interface{}{"from_user_id": 1, "to_user_id": 2, "amount": 50.0}
	bodyJSON, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "/wallets/transfer", bytes.NewBuffer(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

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
	router := setupRouter()
	router.POST("/wallets/transfer", TransferHandler(walletService))

	body := map[string]interface{}{"from_user_id": 1, "to_user_id": 2, "amount": 50.0}
	bodyJSON, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "/wallets/transfer", bytes.NewBuffer(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	expectedResponse := `{"message":"Transfer successful"}`
	assert.JSONEq(t, expectedResponse, w.Body.String())
}
