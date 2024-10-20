package user

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"testing"

	"centralized-wallet/internal/models"
	mockAuth "centralized-wallet/tests/mocks/auth"
	mockUser "centralized-wallet/tests/mocks/user"
	"centralized-wallet/tests/testutils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// Test registration handler

var mockHandlerTestHelper struct {
	userService      *mockUser.MockUserService
	blacklistService *mockAuth.MockBlacklistService
}

// Helper function to setup the router with services

func setupHandlerMock() {
	mockHandlerTestHelper.userService = new(mockUser.MockUserService)
	mockHandlerTestHelper.blacklistService = new(mockAuth.MockBlacklistService)
}

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	setupHandlerMock()
	return router
}

func TestRegistrationHandler(t *testing.T) {
	router := setupRouter()
	router.POST("/register", RegistrationHandler(mockHandlerTestHelper.userService))
	mockHandlerTestHelper.userService.On("RegisterUser", "test@example.com", "password123").Return(&models.User{ID: 1, Email: "test@example.com"}, nil)

	body := map[string]interface{}{"email": "test@example.com", "password": "password123"}
	w := testutils.ExecuteRequest(router, "POST", "/register", body, "")

	assert.Equal(t, http.StatusCreated, w.Code)
	expectedResponse := `{"message":"User registered successfully","user":{"id":1,"email":"test@example.com"}}`
	assert.JSONEq(t, expectedResponse, w.Body.String())
}

// Test registration handler with duplicate email
func TestRegistrationHandler_DuplicateEmail(t *testing.T) {
	router := setupRouter()
	router.POST("/register", RegistrationHandler(mockHandlerTestHelper.userService))
	mockHandlerTestHelper.userService.On("RegisterUser", "test@example.com", "password123").Return(nil, errors.New("email already in use"))

	body := map[string]interface{}{"email": "test@example.com", "password": "password123"}
	w := testutils.ExecuteRequest(router, "POST", "/register", body, "")

	assert.Equal(t, http.StatusBadRequest, w.Code)
	expectedResponse := `{"error":"Email already in use"}`
	assert.JSONEq(t, expectedResponse, w.Body.String())
}

// Test login handler with JWT
func TestLoginHandler(t *testing.T) {
	// Set JWT_SECRET environment variable for testing
	os.Setenv("JWT_SECRET", "test-secret-key")
	// Generate the hash for "password123" directly in the test
	hashedPassword, _ := HashPassword("password123")
	// Mock LoginUser to return a valid user
	user := &models.User{
		ID:       1,
		Email:    "test@example.com",
		Password: hashedPassword,
	}

	router := setupRouter()
	mockHandlerTestHelper.userService.On("LoginUser", "test@example.com", "password123").Return(user, nil)
	router.POST("/login", LoginHandler(mockHandlerTestHelper.userService))

	body := map[string]interface{}{"email": "test@example.com", "password": "password123"}
	w := testutils.ExecuteRequest(router, "POST", "/login", body, "")
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Assert that the response contains a JWT token
	assert.NotEmpty(t, response["token"])
	assert.Equal(t, "Login successful", response["message"])
	assert.Equal(t, map[string]interface{}{"id": float64(1), "email": "test@example.com"}, response["user"])
}

// Test login handler with incorrect password
func TestLoginHandler_IncorrectPassword(t *testing.T) {
	router := setupRouter()
	mockHandlerTestHelper.userService.On("LoginUser", "test@example.com", "wrongpassword").Return(nil, errors.New("invalid password"))
	router.POST("/login", LoginHandler(mockHandlerTestHelper.userService))

	body := map[string]interface{}{"email": "test@example.com", "password": "wrongpassword"}
	w := testutils.ExecuteRequest(router, "POST", "/login", body, "")

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	expectedResponse := `{"error":"Invalid email or password"}`
	assert.JSONEq(t, expectedResponse, w.Body.String())
}
