package wallet

import (
	"github.com/stretchr/testify/mock"
)

type MockWalletRepository struct {
	mock.Mock
}

func (m *MockWalletRepository) GetWalletBalance(userID int) (float64, error) {
	args := m.Called(userID)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockWalletRepository) Deposit(userID int, amount float64) error {
	args := m.Called(userID, amount)
	return args.Error(0)
}

func (m *MockWalletRepository) Withdraw(userID int, amount float64) error {
	args := m.Called(userID, amount)
	return args.Error(0)
}

func (m *MockWalletRepository) Transfer(fromUserID, toUserID int, amount float64) error {
	args := m.Called(fromUserID, toUserID, amount)
	return args.Error(0)
}

func (m *MockWalletRepository) UserExists(userID int) (bool, error) {
	args := m.Called(userID)
	return args.Bool(0), args.Error(1)
}
