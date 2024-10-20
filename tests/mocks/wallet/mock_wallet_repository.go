package mock_wallet

import (
	"centralized-wallet/internal/models"
	"database/sql"

	"github.com/stretchr/testify/mock"
)

// MockWalletRepository is a mock implementation of WalletRepositoryInterface
type MockWalletRepository struct {
	mock.Mock
}

// Ensure MockWalletRepository implements WalletRepositoryInterface

// CreateWalletWithTx mocks the CreateWalletWithTx function
func (m *MockWalletRepository) CreateWalletWithTx(tx *sql.Tx, wallet *models.Wallet) error {
	args := m.Called(tx, wallet)
	return args.Error(0)
}

// IsWalletNumberExists mocks the IsWalletNumberExists function
func (m *MockWalletRepository) IsWalletNumberExists(walletNumber string) (bool, error) {
	args := m.Called(walletNumber)
	return args.Bool(0), args.Error(1)
}

// GetWalletBalance mocks the GetWalletBalance function
func (m *MockWalletRepository) GetWalletBalance(userID int) (float64, error) {
	args := m.Called(userID)
	return args.Get(0).(float64), args.Error(1)
}

// GetWalletByUserID mocks the GetWalletByUserID function
func (m *MockWalletRepository) GetWalletByUserID(userID int) (*models.Wallet, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Wallet), args.Error(1)
}

// Deposit mocks the Deposit function
func (m *MockWalletRepository) Deposit(userID int, amount float64) (*models.Wallet, error) {
	args := m.Called(userID, amount)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Wallet), args.Error(1)
}

// Withdraw mocks the Withdraw function
func (m *MockWalletRepository) Withdraw(userID int, amount float64) (*models.Wallet, error) {
	args := m.Called(userID, amount)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Wallet), args.Error(1)
}

// Transfer mocks the Transfer function
func (m *MockWalletRepository) Transfer(fromUserID int, toWalletNumber string, amount float64) (*models.Wallet, error) {
	args := m.Called(fromUserID, toWalletNumber, amount)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Wallet), args.Error(1)
}

// UserExists mocks the UserExists function
func (m *MockWalletRepository) UserExists(userID int) (bool, error) {
	args := m.Called(userID)
	return args.Bool(0), args.Error(1)
}

// FindByWalletNumber mocks the FindByWalletNumber function
func (m *MockWalletRepository) FindByWalletNumber(walletNumber string) (*models.Wallet, error) {
	args := m.Called(walletNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Wallet), args.Error(1)
}
