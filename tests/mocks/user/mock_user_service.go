package mock_user

import (
	"centralized-wallet/internal/models"

	"github.com/stretchr/testify/mock"
)

type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) RegisterUser(email, password string) (*models.User, error) {
	args := m.Called(email, password)
	// Check if the first argument is nil before type casting
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) LoginUser(email, password string) (*models.User, error) {
	args := m.Called(email, password)
	// Check if the first argument is nil before type casting
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}
