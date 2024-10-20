package wallet

import (
	"centralized-wallet/internal/apperrors"
	"centralized-wallet/internal/models"
	mockTransaction "centralized-wallet/tests/mocks/transaction"
	mockWallet "centralized-wallet/tests/mocks/wallet"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var mockServiceTestHelper struct {
	walletRepo         *mockWallet.MockWalletRepository
	transactionService *mockTransaction.MockTransactionService
}

func setupServiceMock() {
	mockServiceTestHelper.walletRepo = new(mockWallet.MockWalletRepository)
	mockServiceTestHelper.transactionService = new(mockTransaction.MockTransactionService)
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

	mockWallet := &models.Wallet{
		UserID:    testUserID,
		Balance:   150.0, // Expected balance after the deposit
		UpdatedAt: time.Now(),
	}

	// Set up the mock for Deposit
	mockServiceTestHelper.walletRepo.On("Deposit", testUserID, 50.0).Return(mockWallet, nil)

	// Set up the mock for RecordTransaction
	mockServiceTestHelper.transactionService.On("RecordTransaction", (*string)(nil), mock.AnythingOfType("*string"), "deposit", 50.0).Return(nil)

	// Create the wallet service
	walletService := NewWalletService(mockServiceTestHelper.walletRepo, mockServiceTestHelper.transactionService)

	// Call the Deposit method and check the response
	wallet, err := walletService.Deposit(testUserID, 50.0)
	assert.NoError(t, err)
	assert.Equal(t, 150.0, wallet.Balance) // Check the updated balance

	// Verify the expectations
	mockServiceTestHelper.walletRepo.AssertExpectations(t)
	mockServiceTestHelper.transactionService.AssertExpectations(t)
}

func TestWithdraw(t *testing.T) {
	// Define the test cases
	tests := []struct {
		name                        string
		amount                      float64
		mockWithdrawResult          *models.Wallet
		mockWithdrawError           error
		expectError                 bool
		expectedErrorMessage        string
		mockRecordTransactionCalled bool
	}{
		{
			name:                        "SufficientFunds",
			amount:                      50.0,
			mockWithdrawResult:          &models.Wallet{UserID: testUserID, Balance: 100.0, UpdatedAt: time.Now()},
			mockWithdrawError:           nil,
			expectError:                 false,
			expectedErrorMessage:        "",
			mockRecordTransactionCalled: true,
		},
		{
			name:                        "InsufficientFunds",
			amount:                      100.0,
			mockWithdrawResult:          &models.Wallet{},
			mockWithdrawError:           errors.New("insufficient funds"),
			expectError:                 true,
			expectedErrorMessage:        "insufficient funds",
			mockRecordTransactionCalled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock services and helpers
			setupServiceMock()

			// Mock the Withdraw method to return the mock result and error
			mockServiceTestHelper.walletRepo.On("Withdraw", testUserID, tt.amount).Return(tt.mockWithdrawResult, tt.mockWithdrawError)

			// Only set up the transaction recording mock if it's expected to be called
			if tt.mockRecordTransactionCalled {
				mockServiceTestHelper.transactionService.On("RecordTransaction", mock.AnythingOfType("*string"), (*string)(nil), "withdraw", tt.amount).Return(nil)
			}

			// Create the wallet service using the mocked services
			walletService := NewWalletService(mockServiceTestHelper.walletRepo, mockServiceTestHelper.transactionService)

			// Call the Withdraw method
			result, err := walletService.Withdraw(testUserID, tt.amount)

			// Check the error expectation
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErrorMessage, err.Error())
			} else {
				assert.NoError(t, err)

				// Check the returned wallet result
				assert.Equal(t, tt.mockWithdrawResult.UserID, result.UserID)
				assert.Equal(t, tt.mockWithdrawResult.Balance, result.Balance)
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

	// Define the mock wallets for both users after the transfer
	mockFromWallet := &models.Wallet{
		UserID:    testUserID,
		Balance:   50.0,
		UpdatedAt: time.Now(),
	}

	// Mock the Transfer method in the wallet repository to return the expected wallets
	mockServiceTestHelper.walletRepo.On("Transfer", testUserID, testToWalletNumber, 50.0).Return(mockFromWallet, nil)

	mockServiceTestHelper.walletRepo.On("FindByWalletNumber", testToWalletNumber).Return(mockFromWallet, nil)
	// Mock the transaction recording
	mockServiceTestHelper.transactionService.On("RecordTransaction", mock.AnythingOfType("*string"), mock.AnythingOfType("*string"), "transfer", 50.0).Return(nil)

	// Create the wallet service
	walletService := NewWalletService(mockServiceTestHelper.walletRepo, mockServiceTestHelper.transactionService)

	// Call the Transfer method
	fromWallet, err := walletService.Transfer(testUserID, testToWalletNumber, 50.0)

	// Check no errors
	assert.NoError(t, err)

	// Assert the returned wallet has the expected values
	assert.Equal(t, mockFromWallet.UserID, fromWallet.UserID)
	assert.Equal(t, mockFromWallet.Balance, fromWallet.Balance)

	// Verify that expectations are met
	mockServiceTestHelper.walletRepo.AssertExpectations(t)
	mockServiceTestHelper.transactionService.AssertExpectations(t)
}

func TestCreateWallet(t *testing.T) {

	// Define test cases in a table-driven format
	tests := []struct {
		name                 string
		userExists           bool
		userExistsError      error
		repoError            error
		expectedError        error
		expectedWalletNumber string
	}{
		{
			name:                 "successful wallet creation",
			userExists:           false,
			userExistsError:      nil,
			repoError:            nil,
			expectedError:        nil,
			expectedWalletNumber: "WAL-12345-20211020000000-ABC123",
		},
		{
			name:            "user already has a wallet",
			userExists:      true,
			userExistsError: nil,
			repoError:       nil,
			expectedError:   apperrors.ErrWalletAlreadyExists,
		},
		{
			name:            "error checking if user exists",
			userExists:      false,
			userExistsError: errors.New("user service failure"),
			repoError:       nil,
			expectedError:   errors.New("failed to check if user exists: user service failure"),
		},
		{
			name:            "error creating wallet",
			userExists:      false,
			userExistsError: nil,
			repoError:       errors.New("database error"),
			expectedError:   errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupServiceMock()
			// Mock UserExists behavior based on the test case
			mockServiceTestHelper.walletRepo.On("UserExists", mock.Anything).Return(tt.userExists, tt.userExistsError)

			// Only mock the CreateWallet method if user doesn't already exist
			if !tt.userExists && tt.userExistsError == nil {
				// Mock the CreateWallet method
				mockServiceTestHelper.walletRepo.On("CreateWallet", mock.Anything).Return(tt.repoError)
			}
			walletService := NewWalletService(mockServiceTestHelper.walletRepo, mockServiceTestHelper.transactionService)
			// Call CreateWallet
			wallet, err := walletService.CreateWallet(12345)

			// Assert based on the expected error and wallet number
			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
				assert.Nil(t, wallet)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, wallet)
			}

			// Assert that the mock expectations were met
			mockServiceTestHelper.walletRepo.AssertExpectations(t)
		})
	}
}
