package mock_wallet

import (
	"centralized-wallet/internal/models"

	"github.com/stretchr/testify/mock"
)

// MockWalletService is a mock implementation of the WalletService
type MockWalletService struct {
	mock.Mock
}

// GetBalance mocks the GetBalance function
func (m *MockWalletService) GetBalance(userID int) (float64, error) {
	args := m.Called(userID)
	return args.Get(0).(float64), args.Error(1)
}

// UserExists mocks the UserExists function
func (m *MockWalletService) UserExists(userID int) (bool, error) {
	args := m.Called(userID)
	return args.Bool(0), args.Error(1)
}

// Deposit mocks the Deposit function
func (m *MockWalletService) Deposit(userID int, amount float64) (*models.Wallet, error) {
	args := m.Called(userID, amount)
	return args.Get(0).(*models.Wallet), args.Error(1)
}

// Withdraw mocks the Withdraw function
func (m *MockWalletService) Withdraw(userID int, amount float64) (*models.Wallet, error) {
	args := m.Called(userID, amount)
	return args.Get(0).(*models.Wallet), args.Error(1)
}

// Transfer mocks the Transfer function
func (m *MockWalletService) Transfer(fromUserID, toUserID int, amount float64) (*models.Wallet, error) {
	args := m.Called(fromUserID, toUserID, amount)
	return args.Get(0).(*models.Wallet), args.Error(1)
}
