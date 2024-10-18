package wallet

import (
	"centralized-wallet/internal/repository"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock for WalletRepositoryInterface
type MockWalletRepository struct {
	mock.Mock
}

// Ensure MockWalletRepository implements WalletRepositoryInterface
var _ repository.WalletRepositoryInterface = &MockWalletRepository{}

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

func (m *MockWalletRepository) Transfer(fromUserID int, toUserID int, amount float64) error {
	args := m.Called(fromUserID, toUserID, amount)
	return args.Error(0)
}

// Test GetBalance
func TestGetBalance(t *testing.T) {
	mockRepo := new(MockWalletRepository)
	mockRepo.On("GetWalletBalance", 1).Return(100.0, nil)

	walletService := NewWalletService(mockRepo)
	balance, err := walletService.GetBalance(1)

	assert.NoError(t, err)
	assert.Equal(t, 100.0, balance)
	mockRepo.AssertExpectations(t)
}

// Test Deposit
func TestDeposit(t *testing.T) {
	mockRepo := new(MockWalletRepository)
	mockRepo.On("Deposit", 1, 50.0).Return(nil)

	walletService := NewWalletService(mockRepo)
	err := walletService.Deposit(1, 50.0)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// Test Withdraw (with sufficient funds)
func TestWithdraw_SufficientFunds(t *testing.T) {
	mockRepo := new(MockWalletRepository)
	mockRepo.On("Withdraw", 1, 50.0).Return(nil)

	walletService := NewWalletService(mockRepo)
	err := walletService.Withdraw(1, 50.0)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// Test Withdraw (with insufficient funds)
func TestWithdraw_InsufficientFunds(t *testing.T) {
	mockRepo := new(MockWalletRepository)
	mockRepo.On("Withdraw", 1, 100.0).Return(errors.New("insufficient funds"))

	walletService := NewWalletService(mockRepo)
	err := walletService.Withdraw(1, 100.0)

	assert.Error(t, err)
	assert.Equal(t, "insufficient funds", err.Error())
	mockRepo.AssertExpectations(t)
}

// Test Transfer
func TestTransfer(t *testing.T) {
	mockRepo := new(MockWalletRepository)
	mockRepo.On("Transfer", 1, 2, 50.0).Return(nil)

	walletService := NewWalletService(mockRepo)
	err := walletService.Transfer(1, 2, 50.0)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}
