package mock_wallet

import (
	"centralized-wallet/internal/models"
	"database/sql"

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

// Deposit mocks the Deposit function and returns a wallet struct
func (m *MockWalletService) Deposit(userID int, amount float64) (*models.Wallet, error) {
	args := m.Called(userID, amount)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Wallet), args.Error(1)
}

// Withdraw mocks the Withdraw function and returns a wallet struct
func (m *MockWalletService) Withdraw(userID int, amount float64) (*models.Wallet, error) {
	args := m.Called(userID, amount)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Wallet), args.Error(1)
}

// Transfer mocks the Transfer function and returns a wallet struct
func (m *MockWalletService) Transfer(fromUserID int, toWalletNumber string, amount float64) (*models.Wallet, error) {
	args := m.Called(fromUserID, toWalletNumber, amount)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Wallet), args.Error(1)
}

// GetWalletByUserID mocks the GetWalletByUserID function
func (m *MockWalletService) GetWalletByUserID(userID int) (*models.Wallet, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Wallet), args.Error(1)
}

// CreateWalletWithTx mocks the CreateWalletWithTx function
func (m *MockWalletService) CreateWalletWithTx(tx *sql.Tx, userID int) (*models.Wallet, error) {
	args := m.Called(tx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Wallet), args.Error(1)
}
