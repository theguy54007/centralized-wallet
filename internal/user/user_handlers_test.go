package user

import (
	"errors"
	"net/http"
	"os"
	"testing"

	"centralized-wallet/internal/auth"
	"centralized-wallet/internal/models"
	"centralized-wallet/internal/utils"
	mockAuth "centralized-wallet/tests/mocks/auth"
	mockUser "centralized-wallet/tests/mocks/user"
	"centralized-wallet/tests/testutils"

	"github.com/gin-gonic/gin"
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

var (
	email    = "test@example.com"
	password = "password123"
)

func TestRegistrationHandler_Success(t *testing.T) {

	router := setupRouter()
	router.POST("/register", RegistrationHandler(mockHandlerTestHelper.userService))
	mockHandlerTestHelper.userService.On("RegisterUser", email, password).Return(&models.User{ID: 1, Email: email}, nil)

	body := map[string]interface{}{"email": email, "password": password}
	w := testutils.ExecuteRequest(router, "POST", "/register", body, "")
	testutils.AssertAPISuccessResponse(t, w, utils.MsgUserRegistered, models.User{ID: 1, Email: email})
}

// Test registration handler with duplicate email
func TestRegistrationHandler_DuplicateEmail(t *testing.T) {
	router := setupRouter()
	router.POST("/register", RegistrationHandler(mockHandlerTestHelper.userService))
	mockHandlerTestHelper.userService.On("RegisterUser", email, password).Return(nil, utils.ErrEmailAlreadyInUse)

	body := map[string]interface{}{"email": email, "password": password}
	w := testutils.ExecuteRequest(router, "POST", "/register", body, "")
	testutils.AssertAPIErrorResponse(t, w, utils.ErrEmailAlreadyInUse)
}

// Test login handler with JWT
func TestLoginHandler_Success(t *testing.T) {
	// Set JWT_SECRET environment variable for testing
	os.Setenv("JWT_SECRET", "test-secret-key")
	// Generate the hash for "password123" directly in the test
	hashedPassword, _ := HashPassword(password)
	// Mock LoginUser to return a valid user
	user := &models.User{
		ID:       1,
		Email:    email,
		Password: hashedPassword,
	}

	token, _ := auth.GenerateJWT(user.ID)
	router := setupRouter()
	mockHandlerTestHelper.userService.On("LoginUser", user.Email, password).Return(user, nil)
	router.POST("/login", LoginHandler(mockHandlerTestHelper.userService))

	body := map[string]interface{}{"email": user.Email, "password": password}
	w := testutils.ExecuteRequest(router, "POST", "/login", body, "")
	testutils.AssertAPISuccessResponse(t, w, "Login successful",
		gin.H{
			"token": token,
			"user": map[string]any{
				"id":    1,
				"email": user.Email,
			},
		}, http.StatusOK)

}

// Test login handler with incorrect password
func TestLoginHandler_IncorrectPassword(t *testing.T) {
	router := setupRouter()
	wrongpassword := "wrongpassword"
	mockHandlerTestHelper.userService.On("LoginUser", email, wrongpassword).Return(nil, errors.New("invalid password"))
	router.POST("/login", LoginHandler(mockHandlerTestHelper.userService))

	body := map[string]interface{}{"email": email, "password": wrongpassword}
	w := testutils.ExecuteRequest(router, "POST", "/login", body, "")

	testutils.AssertAPIErrorResponse(t, w, utils.ErrInvalidCredentials)
}
