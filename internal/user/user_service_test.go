package user

import (
	"centralized-wallet/internal/models"
	"centralized-wallet/internal/utils"
	mockUser "centralized-wallet/tests/mocks/user"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Mock service test helper struct
var mockServiceTestHelper struct {
	userRepo *mockUser.MockUserRepository
}

// Setup function to initialize mocks
func setupServiceMock() {
	mockServiceTestHelper.userRepo = new(mockUser.MockUserRepository)
}

// Test RegisterUser method
func TestRegisterUser_Success(t *testing.T) {
	// Step 1: Initialize mocks
	setupServiceMock()

	// Step 2: Create a new UserService
	us := NewUserService(mockServiceTestHelper.userRepo)

	// Mock methods
	mockServiceTestHelper.userRepo.On("IsEmailInUse", "test@example.com").Return(false, nil)
	mockServiceTestHelper.userRepo.On("CreateUser", "test@example.com", "password").Return(&models.User{ID: 1, Email: "test@example.com"}, nil)

	// Act
	user, err := us.RegisterUser("test@example.com", "password")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 1, user.ID)
	assert.Equal(t, "test@example.com", user.Email)
	mockServiceTestHelper.userRepo.AssertExpectations(t)
}

func TestRegisterUser_EmailAlreadyInUse(t *testing.T) {
	// Step 1: Initialize mocks
	setupServiceMock()

	// Step 2: Create a new UserService
	us := NewUserService(mockServiceTestHelper.userRepo)

	// Mock methods
	mockServiceTestHelper.userRepo.On("IsEmailInUse", "test@example.com").Return(true, nil)

	// Act
	user, err := us.RegisterUser("test@example.com", "password")

	// Assert
	assert.Nil(t, user)
	assert.EqualError(t, err, utils.ErrEmailAlreadyInUse.Message)
	mockServiceTestHelper.userRepo.AssertExpectations(t)
}

// Test LoginUser method
func TestLoginUser_Success(t *testing.T) {
	// Step 1: Initialize mocks
	setupServiceMock()

	// Step 2: Create a new UserService
	us := NewUserService(mockServiceTestHelper.userRepo)

	// Mock the GetUserByEmail method
	mockServiceTestHelper.userRepo.On("GetUserByEmail", "test@example.com").Return(&models.User{ID: 1, Email: "test@example.com", Password: "$2a$10$gjS3c/wGiZO4VMHO.bSOsex36CrnGO.lrFhYKltC/FIEPlT49XDNq"}, nil)
	// Note: Since `VerifyPassword` is not being mocked, the real function will be used.

	// Act: Call the LoginUser method
	user, err := us.LoginUser("test@example.com", "password")

	// Assert: Check the expected results
	assert.NoError(t, err)
	assert.Equal(t, 1, user.ID)
	assert.Equal(t, "test@example.com", user.Email)

	// Verify that all expectations were met
	mockServiceTestHelper.userRepo.AssertExpectations(t)
}

func TestLoginUser_InvalidEmailOrPassword(t *testing.T) {
	// Step 1: Initialize mocks
	setupServiceMock()

	// Step 2: Create a new UserService
	us := NewUserService(mockServiceTestHelper.userRepo)

	// Mock methods
	mockServiceTestHelper.userRepo.On("GetUserByEmail", "test@example.com").Return(nil, nil)

	// Act
	user, err := us.LoginUser("test@example.com", "wrongpassword")

	// Assert
	assert.Nil(t, user)
	assert.EqualError(t, err, utils.ErrInvalidCredentials.Error())
	mockServiceTestHelper.userRepo.AssertExpectations(t)
}
