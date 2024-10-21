package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	mockAuth "centralized-wallet/tests/mocks/auth"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

// Helper function to generate a valid JWT token
func generateValidToken() (string, error) {
	return GenerateJWT(123)
}

// Helper function to generate an expired JWT token
func generateExpiredToken() (string, error) {
	exp := -1 * time.Hour
	return GenerateJWT(123, exp)
}

// Helper function to generate a JWT token with invalid signature
func generateInvalidToken() (string, error) {
	secretKey := "invalid-secret-key"
	claims := jwt.MapClaims{
		"user_id": 123,
		"exp":     time.Now().Add(1 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}

// Helper function to set up the router with JWT middleware
func setupRouterWithJWT(mockBlacklistService *mockAuth.MockBlacklistService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(JWTMiddleware(mockBlacklistService))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Success"})
	})
	return router
}

// Table-driven tests for JWT middleware
func TestJWTMiddleware(t *testing.T) {
	testCases := []struct {
		name                 string
		tokenGenerator       func() (string, error)
		mockBlacklistService func(tokenString string, mockBlacklistService *mockAuth.MockBlacklistService)
		expectedStatus       int
		expectedResponseBody string
	}{
		{
			name:           "Missing Authorization header",
			tokenGenerator: func() (string, error) { return "", nil },
			mockBlacklistService: func(tokenString string, mockBlacklistService *mockAuth.MockBlacklistService) {
				// No mock setup needed for this test
			},
			expectedStatus:       http.StatusUnauthorized,
			expectedResponseBody: `{"status":"error","message":"Authorization token is required"}`,
		},
		{
			name: "Token blacklisted",
			tokenGenerator: func() (string, error) {
				return generateValidToken()
			},
			mockBlacklistService: func(tokenString string, mockBlacklistService *mockAuth.MockBlacklistService) {
				mockBlacklistService.On("IsTokenBlacklisted", tokenString).Return(true, nil)
			},
			expectedStatus:       http.StatusUnauthorized,
			expectedResponseBody: `{"status":"error","message":"Invalid token"}`,
		},
		{
			name: "Invalid token",
			tokenGenerator: func() (string, error) {
				return generateInvalidToken()
			},
			mockBlacklistService: func(tokenString string, mockBlacklistService *mockAuth.MockBlacklistService) {
				mockBlacklistService.On("IsTokenBlacklisted", tokenString).Return(false, nil)
			},
			expectedStatus:       http.StatusUnauthorized,
			expectedResponseBody: `{"status":"error","message":"Invalid token"}`,
		},
		{
			name: "Token expired",
			tokenGenerator: func() (string, error) {
				return generateExpiredToken()
			},
			mockBlacklistService: func(tokenString string, mockBlacklistService *mockAuth.MockBlacklistService) {
				mockBlacklistService.On("IsTokenBlacklisted", tokenString).Return(false, nil)
			},
			expectedStatus:       http.StatusUnauthorized,
			expectedResponseBody: `{"status":"error","message":"Token expired"}`,
		},
		{
			name: "Valid token",
			tokenGenerator: func() (string, error) {
				return generateValidToken()
			},
			mockBlacklistService: func(tokenString string, mockBlacklistService *mockAuth.MockBlacklistService) {
				mockBlacklistService.On("IsTokenBlacklisted", tokenString).Return(false, nil)
			},
			expectedStatus:       http.StatusOK,
			expectedResponseBody: `{"message":"Success"}`,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			// Generate token for the test case
			tokenString, err := tt.tokenGenerator()
			assert.NoError(t, err)

			// Mock blacklist service
			mockBlacklistService := new(mockAuth.MockBlacklistService)
			tt.mockBlacklistService(tokenString, mockBlacklistService)

			// Set up router and execute the request
			router := setupRouterWithJWT(mockBlacklistService)
			req, _ := http.NewRequest("GET", "/test", nil)
			if tokenString != "" {
				req.Header.Set("Authorization", "Bearer "+tokenString)
			}
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Assertions
			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.JSONEq(t, tt.expectedResponseBody, w.Body.String())

			// Validate the mock expectations
			mockBlacklistService.AssertExpectations(t)
		})
	}
}
