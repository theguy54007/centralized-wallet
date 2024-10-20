package mock_user

import (
	"centralized-wallet/internal/models"

	"github.com/stretchr/testify/mock"
)

// Mock UserRepository
type MockUserRepository struct {
	mock.Mock
}

// Ensure MockUserRepository implements UserRepositoryInterface

// CreateUserWithTx mocks the CreateUserWithTx function
func (m *MockUserRepository) CreateUser(email, password string) (*models.User, error) {
	args := m.Called(email, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

// IsEmailInUse mocks the IsEmailInUse function
func (m *MockUserRepository) IsEmailInUse(email string) (bool, error) {
	args := m.Called(email)
	return args.Bool(0), args.Error(1)
}

// GetUserByEmail mocks the GetUserByEmail function
func (m *MockUserRepository) GetUserByEmail(email string) (*models.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}
