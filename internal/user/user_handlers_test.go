package user

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"centralized-wallet/internal/models"
	mockUser "centralized-wallet/tests/mocks/user"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// Test registration handler
func TestRegistrationHandler(t *testing.T) {
	mockService := new(mockUser.MockUserService)
	mockService.On("RegisterUser", "test@example.com", "password123").Return(&models.User{ID: 1, Email: "test@example.com"}, nil)

	router := gin.Default()
	router.POST("/register", RegistrationHandler(mockService))

	body := map[string]interface{}{"email": "test@example.com", "password": "password123"}
	bodyJSON, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(bodyJSON))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	expectedResponse := `{"message":"User registered successfully","user":{"id":1,"email":"test@example.com"}}`
	assert.JSONEq(t, expectedResponse, w.Body.String())
}

// Test registration handler with duplicate email
func TestRegistrationHandler_DuplicateEmail(t *testing.T) {
	mockService := new(mockUser.MockUserService)
	mockService.On("RegisterUser", "test@example.com", "password123").Return(nil, errors.New("email already in use"))

	router := gin.Default()
	router.POST("/register", RegistrationHandler(mockService))

	body := map[string]interface{}{"email": "test@example.com", "password": "password123"}
	bodyJSON, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(bodyJSON))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	expectedResponse := `{"error":"Email already in use"}`
	assert.JSONEq(t, expectedResponse, w.Body.String())
}

// Test login handler with JWT
func TestLoginHandler(t *testing.T) {
	// Set JWT_SECRET environment variable for testing
	os.Setenv("JWT_SECRET", "test-secret-key")

	mockService := new(mockUser.MockUserService)

	// Generate the hash for "password123" directly in the test
	hashedPassword, _ := HashPassword("password123")

	// Mock LoginUser to return a valid user
	user := &models.User{
		ID:       1,
		Email:    "test@example.com",
		Password: hashedPassword,
	}
	mockService.On("LoginUser", "test@example.com", "password123").Return(user, nil)

	router := gin.Default()
	router.POST("/login", LoginHandler(mockService))

	body := map[string]interface{}{"email": "test@example.com", "password": "password123"}
	bodyJSON, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(bodyJSON))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

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
	mockService := new(mockUser.MockUserService)
	mockService.On("LoginUser", "test@example.com", "wrongpassword").Return(nil, errors.New("invalid password"))

	router := gin.Default()
	router.POST("/login", LoginHandler(mockService))

	body := map[string]interface{}{"email": "test@example.com", "password": "wrongpassword"}
	bodyJSON, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(bodyJSON))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	expectedResponse := `{"error":"Invalid email or password"}`
	assert.JSONEq(t, expectedResponse, w.Body.String())
}
