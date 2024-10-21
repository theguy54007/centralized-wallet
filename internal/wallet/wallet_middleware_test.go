package wallet

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"centralized-wallet/internal/models"
	"centralized-wallet/internal/utils"
	mockRedis "centralized-wallet/tests/mocks/redis"
	mockWallet "centralized-wallet/tests/mocks/wallet"
	"centralized-wallet/tests/testutils"

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

// Helper function to execute the request and return the response
func executeRequest(method, path string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, path, nil)
	w := httptest.NewRecorder()
	router := setupRouterForMiddlewareTest()
	router.ServeHTTP(w, req)
	return w
}

// Table-driven tests for wallet number middleware
func TestWalletNumberMiddleware(t *testing.T) {
	testCases := []struct {
		name                  string
		mockSetup             func()
		expectedStatus        int
		expectedErrorResponse *utils.AppError
		expectedResponseBody  string
	}{
		{
			name: "FetchFromRedis",
			mockSetup: func() {
				mockRedisService.On("Get", mock.Anything, userIDStr).Return(walletNumber, nil)
			},
			expectedStatus:       http.StatusOK,
			expectedResponseBody: `{"wallet_number":"WAL-123456"}`,
		},
		{
			name: "FetchFromDatabaseAndCacheIt",
			mockSetup: func() {
				// Redis returns nil, so fallback to DB
				mockRedisService.On("Get", mock.Anything, userIDStr).Return("", redis.Nil)
				mockWalletService.On("GetWalletByUserID", userID).Return(&models.Wallet{WalletNumber: walletNumber}, nil)
				mockRedisService.On("Set", mock.Anything, userIDStr, walletNumber, 24*time.Hour).Return(nil)
			},
			expectedStatus:       http.StatusOK,
			expectedResponseBody: `{"wallet_number":"WAL-123456"}`,
		},
		{
			name: "DatabaseError",
			mockSetup: func() {
				// Redis returns nil, so fallback to DB
				mockRedisService.On("Get", mock.Anything, userIDStr).Return("", redis.Nil)
				mockWalletService.On("GetWalletByUserID", userID).Return(nil, errors.New("database error"))
			},
			expectedStatus:        http.StatusInternalServerError,
			expectedErrorResponse: utils.ErrInternalServerError,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mocks and set them up for the test case
			resetMocks()
			tt.mockSetup()

			// Execute the request and get the response
			w := executeRequest("GET", "/test")

			if tt.expectedErrorResponse != nil {
				testutils.AssertAPIErrorResponse(t, w, tt.expectedErrorResponse)
			} else {
				assert.Equal(t, tt.expectedStatus, w.Code)
				assert.JSONEq(t, tt.expectedResponseBody, w.Body.String())
			}
			mockRedisService.AssertExpectations(t)
			mockWalletService.AssertExpectations(t)
		})
	}
}
