package mock_wallet

import (
	"centralized-wallet/internal/models"

	"github.com/stretchr/testify/mock"
)

type MockWalletRepository struct {
	mock.Mock
}

func (m *MockWalletRepository) GetWalletBalance(userID int) (float64, error) {
	args := m.Called(userID)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockWalletRepository) Deposit(userID int, amount float64) (*models.Wallet, error) {
	args := m.Called(userID, amount)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*models.Wallet), args.Error(1)
}

func (m *MockWalletRepository) Withdraw(userID int, amount float64) (*models.Wallet, error) {
	args := m.Called(userID, amount)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*models.Wallet), args.Error(1)
}

func (m *MockWalletRepository) Transfer(fromUserID, toUserID int, amount float64) (*models.Wallet, error) {
	args := m.Called(fromUserID, toUserID, amount)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*models.Wallet), args.Error(1)
}

func (m *MockWalletRepository) UserExists(userID int) (bool, error) {
	args := m.Called(userID)
	return args.Bool(0), args.Error(1)
}
