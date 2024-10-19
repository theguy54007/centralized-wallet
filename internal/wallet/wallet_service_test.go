package wallet

import (
	"centralized-wallet/tests/mocks/transaction"
	"centralized-wallet/tests/mocks/wallet"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Test GetBalance
func TestGetBalance(t *testing.T) {
	mockRepo := new(wallet.MockWalletRepository)
	mockTransactionService := new(transaction.MockTransactionService)
	mockRepo.On("GetWalletBalance", 1).Return(100.0, nil)

	walletService := NewWalletService(mockRepo, mockTransactionService)
	balance, err := walletService.GetBalance(1)

	assert.NoError(t, err)
	assert.Equal(t, 100.0, balance)
	mockRepo.AssertExpectations(t)
}

// Test Deposit
func TestDeposit(t *testing.T) {
	mockRepo := new(wallet.MockWalletRepository)
	mockTransactionService := new(transaction.MockTransactionService)

	userID := 1
	// Mock the Deposit method on the repository
	mockRepo.On("Deposit", userID, 50.0).Return(nil)

	// Mock the RecordTransaction method on the transaction service
	mockTransactionService.On("RecordTransaction", (*int)(nil), &userID, "deposit", 50.0).Return(nil)

	// Create the wallet service
	walletService := NewWalletService(mockRepo, mockTransactionService)

	// Call the Deposit method
	err := walletService.Deposit(1, 50.0)

	// Check no errors
	assert.NoError(t, err)

	// Verify that expectations are met
	mockRepo.AssertExpectations(t)
	mockTransactionService.AssertExpectations(t)
}

// Test Withdraw (with sufficient funds)
func TestWithdraw_SufficientFunds(t *testing.T) {
	mockRepo := new(wallet.MockWalletRepository)
	mockTransactionService := new(transaction.MockTransactionService)

	userID := 1
	// Mock the Withdraw method on the repository
	mockRepo.On("Withdraw", userID, 50.0).Return(nil)

	// Mock the RecordTransaction method on the transaction service
	mockTransactionService.On("RecordTransaction", &userID, (*int)(nil), "withdraw", 50.0).Return(nil)

	// Create the wallet service
	walletService := NewWalletService(mockRepo, mockTransactionService)

	// Call the Withdraw method
	err := walletService.Withdraw(userID, 50.0)

	// Check no errors
	assert.NoError(t, err)

	// Verify that expectations are met
	mockRepo.AssertExpectations(t)
	mockTransactionService.AssertExpectations(t)
}

// Test Withdraw (with insufficient funds)
func TestWithdraw_InsufficientFunds(t *testing.T) {
	mockRepo := new(wallet.MockWalletRepository)
	mockTransactionService := new(transaction.MockTransactionService)

	userID := 1
	// Mock the Withdraw method to simulate insufficient funds
	mockRepo.On("Withdraw", userID, 100.0).Return(errors.New("insufficient funds"))

	// Create the wallet service
	walletService := NewWalletService(mockRepo, mockTransactionService)

	// Call the Withdraw method with insufficient funds
	err := walletService.Withdraw(userID, 100.0)

	// Check that an error is returned
	assert.Error(t, err)
	assert.Equal(t, "insufficient funds", err.Error())

	// Verify that expectations are met (no transaction recording should happen)
	mockRepo.AssertExpectations(t)
	mockTransactionService.AssertNotCalled(t, "RecordTransaction", mock.Anything, mock.Anything, mock.Anything)
}

func TestTransfer(t *testing.T) {
	mockRepo := new(wallet.MockWalletRepository)
	mockTransactionService := new(transaction.MockTransactionService)
	fromUserId, toUserId := 1, 2
	// Mock the Withdraw method for the sender (user 1)
	mockRepo.On("Withdraw", fromUserId, 50.0).Return(nil) // Ensure the method matches the expected call

	// Mock the Deposit method for the recipient (user 2)
	mockRepo.On("Deposit", toUserId, 50.0).Return(nil)

	// Mock the RecordTransaction method for both users
	mockTransactionService.On("RecordTransaction", &fromUserId, &toUserId, "transfer", 50.0).Return(nil)
	// mockTransactionService.On("RecordTransaction", 2, "transfer in", 50.0).Return(nil)

	// Create the wallet service
	walletService := NewWalletService(mockRepo, mockTransactionService)

	// Call the Transfer method
	err := walletService.Transfer(1, 2, 50.0)

	// Check no errors
	assert.NoError(t, err)

	// Verify that expectations are met
	mockRepo.AssertExpectations(t)
	mockTransactionService.AssertExpectations(t)
}
