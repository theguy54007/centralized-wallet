package wallet

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"centralized-wallet/internal/models"
	mockRedis "centralized-wallet/tests/mocks/redis"
	mockWallet "centralized-wallet/tests/mocks/wallet"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	userID            = 1
	walletNumber      = "WAL-123456"
	userIDStr         = "user:1:wallet_number"
	mockWalletService = new(mockWallet.MockWalletService)
	mockRedisService  = new(mockRedis.MockRedisClient)
)

func setupRouterForMiddlewareTest() *gin.Engine {
	router := gin.New()

	// Simulate setting user_id in the context via a middleware
	router.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	})

	// Register the wallet number middleware and the handler
	router.Use(WalletNumberMiddleware(mockWalletService, mockRedisService))

	// Define a simple route for testing
	router.GET("/test", func(c *gin.Context) {
		walletNumber, exists := c.Get("wallet_number")
		if exists {
			c.JSON(http.StatusOK, gin.H{"wallet_number": walletNumber})
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": "Wallet number not found"})
		}
	})

	return router
}

func resetMocks() {
	mockRedisService.ExpectedCalls = nil
	mockWalletService.ExpectedCalls = nil
}

func TestWalletNumberMiddleware_FetchFromRedis(t *testing.T) {
	resetMocks()

	// Set up mock Redis to return a wallet number
	mockRedisService.On("Get", mock.Anything, userIDStr).Return(walletNumber, nil)

	// Set up the request and execute it
	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router := setupRouterForMiddlewareTest()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"wallet_number":"WAL-123456"}`, w.Body.String())
	mockRedisService.AssertCalled(t, "Get", mock.Anything, userIDStr)
}

func TestWalletNumberMiddleware_FetchFromDatabaseAndCacheIt(t *testing.T) {
	resetMocks()

	// Redis returns a nil result, so fallback to DB
	mockRedisService.On("Get", mock.Anything, userIDStr).Return("", redis.Nil)
	// Wallet service returns a valid wallet
	mockWalletService.On("GetWalletByUserID", userID).Return(&models.Wallet{WalletNumber: walletNumber}, nil)
	// Cache the wallet number in Redis
	mockRedisService.On("Set", mock.Anything, userIDStr, walletNumber, 24*time.Hour).Return(nil)

	// Set up the request and execute it
	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router := setupRouterForMiddlewareTest()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"wallet_number":"WAL-123456"}`, w.Body.String())
	mockRedisService.AssertCalled(t, "Get", mock.Anything, userIDStr)
	mockWalletService.AssertCalled(t, "GetWalletByUserID", userID)
	mockRedisService.AssertCalled(t, "Set", mock.Anything, userIDStr, walletNumber, 24*time.Hour)
}

func TestWalletNumberMiddleware_RedisError(t *testing.T) {
	resetMocks()

	// Redis returns an unexpected error
	mockRedisService.On("Get", mock.Anything, userIDStr).Return("", errors.New("redis error"))

	// Set up the request and execute it
	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router := setupRouterForMiddlewareTest()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.JSONEq(t, `{"status":"error","error":"Redis error"}`, w.Body.String())
	mockRedisService.AssertCalled(t, "Get", mock.Anything, userIDStr)
}

func TestWalletNumberMiddleware_DatabaseError(t *testing.T) {
	resetMocks()

	// Redis returns nil, so fallback to DB
	mockRedisService.On("Get", mock.Anything, userIDStr).Return("", redis.Nil)
	// Wallet service returns an error
	mockWalletService.On("GetWalletByUserID", userID).Return(nil, errors.New("database error"))

	// Set up the request and execute it
	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router := setupRouterForMiddlewareTest()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.JSONEq(t, `{"status":"error","error":"Failed to fetch wallet number"}`, w.Body.String())
	mockRedisService.AssertCalled(t, "Get", mock.Anything, userIDStr)
	mockWalletService.AssertCalled(t, "GetWalletByUserID", userID)
}
