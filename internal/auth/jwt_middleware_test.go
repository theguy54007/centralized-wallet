// jwt_middleware_test.go
package auth

import (
	"errors"
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
	// Use a different secret key to invalidate the token
	secretKey := "invalid-secret-key"

	// Create claims
	claims := jwt.MapClaims{
		"user_id": 123,
		"exp":     time.Now().Add(1 * time.Hour).Unix(),
	}

	// Create the token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the invalid secret key
	return token.SignedString([]byte(secretKey))
}

// Test case: Missing Authorization header
func TestJWTMiddleware_MissingAuthorizationHeader(t *testing.T) {
	// Set up Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Mock dependencies
	mockBlacklistService := new(mockAuth.MockBlacklistService)

	// Apply middleware
	router.Use(JWTMiddleware(mockBlacklistService))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Success"})
	})

	// Perform request without Authorization header
	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.JSONEq(t, `{"error": "Authorization token is required"}`, w.Body.String())
}

// Test case: Invalid Authorization header format
func TestJWTMiddleware_InvalidAuthorizationHeader(t *testing.T) {
	// Set up Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Mock dependencies
	mockBlacklistService := new(mockAuth.MockBlacklistService)

	// Apply middleware
	router.Use(JWTMiddleware(mockBlacklistService))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Success"})
	})

	// Perform request with invalid Authorization header
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "InvalidTokenFormat")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.JSONEq(t, `{"error": "Invalid token format"}`, w.Body.String())
}

// Test case: Token is blacklisted
func TestJWTMiddleware_TokenBlacklisted(t *testing.T) {
	// Set up Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Generate a valid token
	tokenString, err := generateValidToken()
	assert.NoError(t, err)

	// Mock dependencies
	mockBlacklistService := new(mockAuth.MockBlacklistService)
	mockBlacklistService.On("IsTokenBlacklisted", tokenString).Return(true, nil)

	// Apply middleware
	router.Use(JWTMiddleware(mockBlacklistService))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Success"})
	})

	// Perform request with blacklisted token
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.JSONEq(t, `{"error": "Token is blacklisted"}`, w.Body.String())
	mockBlacklistService.AssertExpectations(t)
}

// Test case: Error when checking blacklist
func TestJWTMiddleware_BlacklistCheckError(t *testing.T) {
	// Set up Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Generate a valid token
	tokenString, err := generateValidToken()
	assert.NoError(t, err)

	// Mock dependencies
	mockBlacklistService := new(mockAuth.MockBlacklistService)
	mockError := errors.New("blacklist service error")
	mockBlacklistService.On("IsTokenBlacklisted", tokenString).Return(false, mockError)

	// Apply middleware
	router.Use(JWTMiddleware(mockBlacklistService))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Success"})
	})

	// Perform request
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.JSONEq(t, `{"error": "Failed to check blacklist"}`, w.Body.String())
	mockBlacklistService.AssertExpectations(t)
}

// Test case: Invalid token
func TestJWTMiddleware_InvalidToken(t *testing.T) {
	// Set up Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Generate an invalid token
	tokenString, err := generateInvalidToken()
	assert.NoError(t, err)

	// Mock dependencies
	mockBlacklistService := new(mockAuth.MockBlacklistService)
	mockBlacklistService.On("IsTokenBlacklisted", tokenString).Return(false, nil)

	// Apply middleware
	router.Use(JWTMiddleware(mockBlacklistService))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Success"})
	})

	// Perform request with invalid token
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.JSONEq(t, `{"error": "Invalid token"}`, w.Body.String())
	mockBlacklistService.AssertExpectations(t)
}

// Test case: Token expired
func TestJWTMiddleware_TokenExpired(t *testing.T) {
	// Set up Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Generate an expired token
	tokenString, err := generateExpiredToken()
	assert.NoError(t, err)

	// Mock dependencies
	mockBlacklistService := new(mockAuth.MockBlacklistService)
	mockBlacklistService.On("IsTokenBlacklisted", tokenString).Return(false, nil)

	// Apply middleware
	router.Use(JWTMiddleware(mockBlacklistService))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Success"})
	})

	// Perform request with expired token
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.JSONEq(t, `{"error": "Token expired"}`, w.Body.String())
	mockBlacklistService.AssertExpectations(t)
}

// Test case: Valid token
func TestJWTMiddleware_ValidToken(t *testing.T) {
	// Set up Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Generate a valid token
	tokenString, err := generateValidToken()
	assert.NoError(t, err)

	// Mock dependencies
	mockBlacklistService := new(mockAuth.MockBlacklistService)
	mockBlacklistService.On("IsTokenBlacklisted", tokenString).Return(false, nil)

	// Apply middleware
	router.Use(JWTMiddleware(mockBlacklistService))
	router.GET("/test", func(c *gin.Context) {
		// Retrieve user_id from context
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "User ID not found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Success", "user_id": userID})
	})

	// Perform request with valid token
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"message": "Success", "user_id":123}`, w.Body.String())
	mockBlacklistService.AssertExpectations(t)
}
