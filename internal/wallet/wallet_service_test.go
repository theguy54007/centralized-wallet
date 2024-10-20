package wallet

import (
	"centralized-wallet/tests/mocks/transaction"
	mockWallet "centralized-wallet/tests/mocks/wallet"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var mockServiceTestHelper struct {
	walletRepo         *mockWallet.MockWalletRepository
	transactionService *transaction.MockTransactionService
}

func setupServiceMock() {
	mockServiceTestHelper.walletRepo = new(mockWallet.MockWalletRepository)
	mockServiceTestHelper.transactionService = new(transaction.MockTransactionService)
}

// Test GetBalance
func TestGetBalance(t *testing.T) {
	setupServiceMock()
	mockServiceTestHelper.walletRepo.On("GetWalletBalance", 1).Return(100.0, nil)

	walletService := NewWalletService(mockServiceTestHelper.walletRepo, mockServiceTestHelper.transactionService)
	balance, err := walletService.GetBalance(1)

	assert.NoError(t, err)
	assert.Equal(t, 100.0, balance)
	mockServiceTestHelper.walletRepo.AssertExpectations(t)
}

// Test Deposit
func TestDeposit(t *testing.T) {
	setupServiceMock()
	mockServiceTestHelper.walletRepo.On("Deposit", testUserID, 50.0).Return(nil)
	mockServiceTestHelper.transactionService.On("RecordTransaction", (*int)(nil), &testUserID, "deposit", 50.0).Return(nil)

	// Create the wallet service
	walletService := NewWalletService(mockServiceTestHelper.walletRepo, mockServiceTestHelper.transactionService)

	// Call the Deposit method
	err := walletService.Deposit(1, 50.0)

	// Check no errors
	assert.NoError(t, err)

	// Verify that expectations are met
	mockServiceTestHelper.walletRepo.AssertExpectations(t)
	mockServiceTestHelper.transactionService.AssertExpectations(t)
}

func TestWithdraw(t *testing.T) {
	// Define the test cases
	tests := []struct {
		name                        string
		amount                      float64
		mockWithdrawResult          error
		expectError                 bool
		expectedErrorMessage        string
		mockRecordTransactionCalled bool
	}{
		{
			name:                        "SufficientFunds",
			amount:                      50.0,
			mockWithdrawResult:          nil,
			expectError:                 false,
			expectedErrorMessage:        "",
			mockRecordTransactionCalled: true,
		},
		{
			name:                        "InsufficientFunds",
			amount:                      100.0,
			mockWithdrawResult:          errors.New("insufficient funds"),
			expectError:                 true,
			expectedErrorMessage:        "insufficient funds",
			mockRecordTransactionCalled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock services and helpers
			setupServiceMock()

			// Mock the Withdraw method
			mockServiceTestHelper.walletRepo.On("Withdraw", testUserID, tt.amount).Return(tt.mockWithdrawResult)

			// Only set up the transaction recording mock if it's expected to be called
			if tt.mockRecordTransactionCalled {
				mockServiceTestHelper.transactionService.On("RecordTransaction", &testUserID, (*int)(nil), "withdraw", tt.amount).Return(nil)
			}

			// Create the wallet service using the mocked services
			walletService := NewWalletService(mockServiceTestHelper.walletRepo, mockServiceTestHelper.transactionService)

			// Call the Withdraw method
			err := walletService.Withdraw(testUserID, tt.amount)

			// Check the error expectation
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErrorMessage, err.Error())
			} else {
				assert.NoError(t, err)
			}

			// Verify that the mock expectations are met
			mockServiceTestHelper.walletRepo.AssertExpectations(t)
			if tt.mockRecordTransactionCalled {
				mockServiceTestHelper.transactionService.AssertExpectations(t)
			} else {
				mockServiceTestHelper.transactionService.AssertNotCalled(t, "RecordTransaction", mock.Anything, mock.Anything, mock.Anything)
			}
		})
	}
}

func TestTransfer(t *testing.T) {
	setupServiceMock()
	mockServiceTestHelper.walletRepo.On("Withdraw", testUserID, 50.0).Return(nil)
	mockServiceTestHelper.walletRepo.On("Deposit", testToUserID, 50.0).Return(nil)
	mockServiceTestHelper.transactionService.On("RecordTransaction", &testUserID, &testToUserID, "transfer", 50.0).Return(nil)

	// Create the wallet service
	walletService := NewWalletService(mockServiceTestHelper.walletRepo, mockServiceTestHelper.transactionService)

	// Call the Transfer method
	err := walletService.Transfer(1, 2, 50.0)

	// Check no errors
	assert.NoError(t, err)

	// Verify that expectations are met
	mockServiceTestHelper.walletRepo.AssertExpectations(t)
	mockServiceTestHelper.transactionService.AssertExpectations(t)
}
