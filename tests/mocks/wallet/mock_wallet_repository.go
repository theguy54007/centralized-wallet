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
func (m *MockWalletRepository) CreateWallet(wallet *models.Wallet) error {
	args := m.Called(wallet)
	return args.Error(0)
}

// mock begin transaction
func (m *MockWalletRepository) Begin() (*sql.Tx, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sql.Tx), args.Error(1)
}

// mock commit transaction
func (m *MockWalletRepository) Commit(tx *sql.Tx) error {
	args := m.Called(tx)
	return args.Error(0)
}

// mock rollback transaction
func (m *MockWalletRepository) Rollback(tx *sql.Tx) error {
	args := m.Called(tx)
	return args.Error(0)
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
func (m *MockWalletRepository) Deposit(tx *sql.Tx, userID int, amount float64) (*models.Wallet, error) {
	args := m.Called(tx, userID, amount)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Wallet), args.Error(1)
}

// Withdraw mocks the Withdraw function
func (m *MockWalletRepository) Withdraw(tx *sql.Tx, userID int, amount float64) (*models.Wallet, error) {
	args := m.Called(tx, userID, amount)
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
