package wallet

import (
	"centralized-wallet/internal/models"
	"centralized-wallet/internal/utils"
	"centralized-wallet/tests/testutils"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Test Deposit
func TestDepositService(t *testing.T) {

	testCases := []testWalletService{
		{
			BaseHandlerTestCase: testutils.BaseHandlerTestCase{
				Name:          "successful deposit",
				TestType:      "success",
				ExpectedError: nil,
				MockSetup: func() {
					// Mock user existence check
					mockServiceTestHelper.walletRepo.On("UserExists", mock.Anything).Return(true, nil)

					// Mock transaction begin
					mockServiceTestHelper.walletRepo.On("Begin").Return(nil, nil)

					// Mock deposit
					mockWallet := &models.Wallet{
						UserID:       testUserID,
						WalletNumber: testWalletNumber,
						Balance:      150.0, // After deposit
						UpdatedAt:    now,
					}
					mockServiceTestHelper.walletRepo.On("Deposit", mock.AnythingOfType("*sql.Tx"), mock.Anything, mock.Anything).Return(mockWallet, nil)

					// Mock recording the transaction
					mockServiceTestHelper.transactionService.On("RecordTransaction", mock.AnythingOfType("*sql.Tx"), (*string)(nil), mock.Anything, "deposit", 50.0).Return(nil)

					// // Mock commit
					mockServiceTestHelper.walletRepo.On("Commit", mock.AnythingOfType("*sql.Tx")).Return(nil)

					// // Mock rollback
					mockServiceTestHelper.walletRepo.On("Rollback", mock.AnythingOfType("*sql.Tx")).Return(nil)
				},
				MockAssert: func(t *testing.T) {
					mockServiceTestHelper.walletRepo.AssertExpectations(t)
					mockServiceTestHelper.transactionService.AssertExpectations(t)
				},
			},
			userID: testUserID,
		},
		{
			BaseHandlerTestCase: testutils.BaseHandlerTestCase{
				Name:          "error checking user existence",
				TestType:      "error",
				ExpectedError: utils.ErrDatabaseError,
				MockSetup: func() {
					// Mock error when checking if the user exists
					mockServiceTestHelper.walletRepo.On("UserExists", mock.Anything).Return(false, utils.ErrDatabaseError)
				},
				MockAssert: func(t *testing.T) {
					mockServiceTestHelper.walletRepo.AssertExpectations(t)
				},
			},
			userID: testUserID,
		},
		{
			BaseHandlerTestCase: testutils.BaseHandlerTestCase{
				Name:          "error during deposit",
				TestType:      "error",
				ExpectedError: utils.ErrDatabaseError,
				MockSetup: func() {
					// Mock user existence check
					mockServiceTestHelper.walletRepo.On("UserExists", mock.Anything).Return(true, nil)

					// Mock transaction begin
					mockServiceTestHelper.walletRepo.On("Begin").Return(nil, nil)

					// Mock deposit returning an error
					mockServiceTestHelper.walletRepo.On("Deposit", mock.AnythingOfType("*sql.Tx"), mock.Anything, mock.Anything).Return(nil, utils.ErrDatabaseError)

					// Mock rollback
					mockServiceTestHelper.walletRepo.On("Rollback", mock.AnythingOfType("*sql.Tx")).Return(nil)
				},
				MockAssert: func(t *testing.T) {
					mockServiceTestHelper.walletRepo.AssertExpectations(t)
				},
			},
			userID: testUserID,
		},
		{
			BaseHandlerTestCase: testutils.BaseHandlerTestCase{
				Name:          "error recording transaction",
				TestType:      "error",
				ExpectedError: utils.ErrDatabaseError,
				MockSetup: func() {
					// Mock user existence check
					mockServiceTestHelper.walletRepo.On("UserExists", mock.Anything).Return(true, nil)

					// Mock transaction begin
					mockServiceTestHelper.walletRepo.On("Begin").Return(nil, nil)

					// Mock deposit
					mockWallet := &models.Wallet{
						UserID:       testUserID,
						WalletNumber: testWalletNumber,
						Balance:      150.0, // After deposit
						UpdatedAt:    now,
					}
					mockServiceTestHelper.walletRepo.On("Deposit", mock.AnythingOfType("*sql.Tx"), mock.Anything, mock.Anything).Return(mockWallet, nil)

					// Mock recording the transaction returning an error
					mockServiceTestHelper.transactionService.On("RecordTransaction", mock.AnythingOfType("*sql.Tx"), (*string)(nil), &mockWallet.WalletNumber, "deposit", 50.0).Return(utils.ErrDatabaseError)

					// Mock rollback
					mockServiceTestHelper.walletRepo.On("Rollback", mock.AnythingOfType("*sql.Tx")).Return(nil)
				},
				MockAssert: func(t *testing.T) {
					mockServiceTestHelper.walletRepo.AssertExpectations(t)
					mockServiceTestHelper.transactionService.AssertExpectations(t)
				},
			},
			userID: testUserID,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {

			walletService := walletServiceTestInit(tc)
			wallet, err := walletService.Deposit(tc.userID, 50.0)

			if tc.TestType == "success" {
				assert.NoError(t, err)
				assert.NotNil(t, wallet)
			} else {
				assert.EqualError(t, err, tc.ExpectedError.Error())
				assert.Nil(t, wallet)
			}

			tc.MockAssert(t)
		})
	}
}

func TestWithdrawService(t *testing.T) {

	testCases := []testWalletService{
		{
			BaseHandlerTestCase: testutils.BaseHandlerTestCase{
				Name:          "successful withdrawal",
				TestType:      "success",
				ExpectedError: nil,
				MockSetup: func() {
					// Mock getting wallet balance successfully
					mockWallet := createMockWallet(testWalletNumber, testUserID)
					mockServiceTestHelper.walletRepo.On("GetWalletByUserID", mock.Anything).Return(mockWallet, nil)
					// Mock withdrawal of amount

					mockServiceTestHelper.walletRepo.On("Begin").Return(nil, nil)
					mockServiceTestHelper.walletRepo.On("Withdraw", mock.AnythingOfType("*sql.Tx"), mock.Anything, mock.Anything).Return(mockWallet, nil)
					// Mock recording the transaction
					mockServiceTestHelper.transactionService.On("RecordTransaction", mock.AnythingOfType("*sql.Tx"), mock.Anything, (*string)(nil), "withdraw", 50.0).Return(nil)
					mockServiceTestHelper.walletRepo.On("Commit", mock.AnythingOfType("*sql.Tx")).Return(nil)
					mockServiceTestHelper.walletRepo.On("Rollback", mock.AnythingOfType("*sql.Tx")).Return(nil)
				},
				MockAssert: func(t *testing.T) {
					mockServiceTestHelper.walletRepo.AssertExpectations(t)
					mockServiceTestHelper.transactionService.AssertExpectations(t)
				},
			},
			userID: testUserID,
			amount: 50,
		},
		{
			BaseHandlerTestCase: testutils.BaseHandlerTestCase{
				Name:          "insufficient funds",
				TestType:      "error",
				ExpectedError: utils.RepoErrInsufficientFunds,
				MockSetup: func() {
					// Mock balance less than the amount being withdrawn
					mmockWallet := createMockWallet(testWalletNumber, testUserID)
					mockServiceTestHelper.walletRepo.On("GetWalletByUserID", mock.Anything).Return(mmockWallet, nil)
				},
				MockAssert: func(t *testing.T) {
					mockServiceTestHelper.walletRepo.AssertExpectations(t)
				},
			},
			userID: testUserID,
			amount: 150,
		},
		{
			BaseHandlerTestCase: testutils.BaseHandlerTestCase{
				Name:          "error getting wallet balance",
				TestType:      "error",
				ExpectedError: utils.ErrDatabaseError,
				MockSetup: func() {
					// Mock error when getting wallet balance
					mockServiceTestHelper.walletRepo.On("GetWalletByUserID", mock.Anything).Return(nil, utils.ErrDatabaseError)
				},
				MockAssert: func(t *testing.T) {
					mockServiceTestHelper.walletRepo.AssertExpectations(t)
				},
			},
			userID: testUserID,
			amount: 50,
		},
		{
			BaseHandlerTestCase: testutils.BaseHandlerTestCase{
				Name:          "error during withdrawal",
				TestType:      "error",
				ExpectedError: utils.ErrDatabaseError,
				MockSetup: func() {
					mockServiceTestHelper.walletRepo.On("Begin").Return(nil, nil)
					// Mock getting wallet balance successfully
					mockWallet := createMockWallet(testWalletNumber, testUserID)
					mockServiceTestHelper.walletRepo.On("GetWalletByUserID", mock.Anything).Return(mockWallet, nil)
					// Mock withdrawal returning an error
					mockServiceTestHelper.walletRepo.On("Withdraw", mock.AnythingOfType("*sql.Tx"), mock.Anything, mock.Anything).Return(nil, utils.ErrDatabaseError)
					// Mock rollback
					mockServiceTestHelper.walletRepo.On("Rollback", mock.AnythingOfType("*sql.Tx")).Return(nil)
				},
				MockAssert: func(t *testing.T) {
					mockServiceTestHelper.walletRepo.AssertExpectations(t)
				},
			},
			userID: testUserID,
			amount: 50,
		},
		{
			BaseHandlerTestCase: testutils.BaseHandlerTestCase{
				Name:          "error recording transaction",
				TestType:      "error",
				ExpectedError: utils.ErrDatabaseError,
				MockSetup: func() {
					mockWallet := createMockWallet(testWalletNumber, testUserID)

					// Mock getting wallet balance successfully
					mockServiceTestHelper.walletRepo.On("GetWalletByUserID", mock.Anything).Return(mockWallet, nil)

					mockServiceTestHelper.walletRepo.On("Begin").Return(nil, nil)
					// Mock withdrawal of amount
					mockServiceTestHelper.walletRepo.On("Withdraw", mock.AnythingOfType("*sql.Tx"), mock.Anything, mock.Anything).Return(mockWallet, nil)
					// Mock recording the transaction returning an error
					mockServiceTestHelper.transactionService.On("RecordTransaction", mock.AnythingOfType("*sql.Tx"), &mockWallet.WalletNumber, (*string)(nil), "withdraw", 50.0).Return(utils.ErrDatabaseError)

					// Mock rollback
					mockServiceTestHelper.walletRepo.On("Rollback", mock.AnythingOfType("*sql.Tx")).Return(nil)
				},
				MockAssert: func(t *testing.T) {
					mockServiceTestHelper.walletRepo.AssertExpectations(t)
					mockServiceTestHelper.transactionService.AssertExpectations(t)
				},
			},
			userID: testUserID,
			amount: 50,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.Name, func(t *testing.T) {
			walletService := walletServiceTestInit(tt)
			wallet, err := walletService.Withdraw(tt.userID, tt.amount) // Example amount to withdraw

			if tt.TestType == "success" {
				assert.NoError(t, err)
				assert.NotNil(t, wallet)
				assert.Equal(t, tt.userID, wallet.UserID)
			} else {
				assert.EqualError(t, err, tt.ExpectedError.Error())
				assert.Nil(t, wallet)
			}
			tt.MockAssert(t)
		})
	}
}

func TestTransferService(t *testing.T) {

	testCases := []testWalletService{
		{
			BaseHandlerTestCase: testutils.BaseHandlerTestCase{
				Name:          "successful transfer",
				TestType:      "success",
				ExpectedError: nil,
				MockSetup: func() {
					mockWallet := createMockWallet(testWalletNumber, testUserID)
					// Mock getting wallet balance successfully
					mockServiceTestHelper.walletRepo.On("GetWalletByUserID", mock.Anything).Return(mockWallet, nil)

					// Mock withdrawal of amount from the sender
					mockFromWallet := createMockWallet(testWalletNumber, testUserID)

					// Mock the recipient wallet details
					mockToWallet := createMockWallet(testToWalletNumber, testToUserID)

					// Begin transaction
					mockServiceTestHelper.walletRepo.On("Begin").Return(nil, nil)

					// Mock withdrawal from sender's wallet
					mockServiceTestHelper.walletRepo.On("Withdraw", mock.AnythingOfType("*sql.Tx"), mock.Anything, mock.Anything).Return(mockFromWallet, nil)

					// Mock finding recipient wallet
					mockServiceTestHelper.walletRepo.On("FindByWalletNumber", mock.Anything).Return(mockToWallet, nil)

					// Mock deposit into recipient's wallet
					mockServiceTestHelper.walletRepo.On("Deposit", mock.AnythingOfType("*sql.Tx"), mock.Anything, mock.Anything).Return(mockToWallet, nil)

					// Mock recording the transaction
					mockServiceTestHelper.transactionService.On("RecordTransaction", mock.AnythingOfType("*sql.Tx"), &mockFromWallet.WalletNumber, &mockToWallet.WalletNumber, "transfer", 50.0).Return(nil)

					// Mock commit
					mockServiceTestHelper.walletRepo.On("Commit", mock.AnythingOfType("*sql.Tx")).Return(nil)

					// Mock rollback
					mockServiceTestHelper.walletRepo.On("Rollback", mock.AnythingOfType("*sql.Tx")).Return(nil)
				},
				MockAssert: func(t *testing.T) {
					mockServiceTestHelper.walletRepo.AssertExpectations(t)
					mockServiceTestHelper.transactionService.AssertExpectations(t)
				},
			},
			userID: testUserID,
			amount: 50.0,
		},
		{
			BaseHandlerTestCase: testutils.BaseHandlerTestCase{
				Name:          "insufficient funds",
				TestType:      "error",
				ExpectedError: utils.RepoErrInsufficientFunds,
				MockSetup: func() {
					mockWallet := createMockWallet(testWalletNumber, testUserID)
					// Mock balance less than the amount being transferred
					mockServiceTestHelper.walletRepo.On("GetWalletByUserID", mock.Anything).Return(mockWallet, nil)
				},
				MockAssert: func(t *testing.T) {
					mockServiceTestHelper.walletRepo.AssertExpectations(t)
				},
			},
			userID: testUserID,
			amount: 150,
		},
		{
			BaseHandlerTestCase: testutils.BaseHandlerTestCase{
				Name:          "error during withdrawal",
				TestType:      "error",
				ExpectedError: utils.ErrDatabaseError,
				MockSetup: func() {
					// Mock transaction begin
					mockServiceTestHelper.walletRepo.On("Begin").Return(nil, nil)

					mockWallet := createMockWallet(testWalletNumber, testUserID)
					// Mock getting wallet balance successfully
					mockServiceTestHelper.walletRepo.On("GetWalletByUserID", mock.Anything).Return(mockWallet, nil)

					mockToWallet := createMockWallet(testToWalletNumber, testToUserID)

					mockServiceTestHelper.walletRepo.On("FindByWalletNumber", mock.Anything).Return(mockToWallet, nil)
					// Mock withdrawal returning an error
					mockServiceTestHelper.walletRepo.On("Withdraw", mock.AnythingOfType("*sql.Tx"), mock.Anything, mock.Anything).Return(nil, utils.ErrDatabaseError)

					// Mock rollback
					mockServiceTestHelper.walletRepo.On("Rollback", mock.AnythingOfType("*sql.Tx")).Return(nil)
				},
				MockAssert: func(t *testing.T) {
					mockServiceTestHelper.walletRepo.AssertExpectations(t)
				},
			},
			userID: testUserID,
			amount: 50.0,
		},
		{
			BaseHandlerTestCase: testutils.BaseHandlerTestCase{
				Name:          "error recording transaction",
				TestType:      "error",
				ExpectedError: utils.ErrDatabaseError,
				MockSetup: func() {
					// Mock transaction begin
					mockServiceTestHelper.walletRepo.On("Begin").Return(nil, nil)

					mockWallet := createMockWallet(testWalletNumber, testUserID)
					// Mock getting wallet balance successfully
					mockServiceTestHelper.walletRepo.On("GetWalletByUserID", mock.Anything).Return(mockWallet, nil)

					// Mock withdrawal of amount from sender
					mockFromWallet := createMockWallet(testWalletNumber, testUserID)
					mockServiceTestHelper.walletRepo.On("Withdraw", mock.AnythingOfType("*sql.Tx"), mock.Anything, mock.Anything).Return(mockFromWallet, nil)

					// Mock finding recipient wallet
					mockToWallet := &models.Wallet{
						UserID:       testToUserID,
						WalletNumber: testToWalletNumber,
						Balance:      150.0, // After deposit
						UpdatedAt:    now,
					}
					mockServiceTestHelper.walletRepo.On("FindByWalletNumber", mock.Anything).Return(mockToWallet, nil)

					// Mock deposit into recipient's wallet
					mockServiceTestHelper.walletRepo.On("Deposit", mock.AnythingOfType("*sql.Tx"), mock.Anything, mock.Anything).Return(mockToWallet, nil)

					// Mock recording the transaction returning an error
					mockServiceTestHelper.transactionService.On("RecordTransaction", mock.AnythingOfType("*sql.Tx"), &mockFromWallet.WalletNumber, &mockToWallet.WalletNumber, "transfer", 50.0).Return(utils.ErrDatabaseError)

					// Mock rollback
					mockServiceTestHelper.walletRepo.On("Rollback", mock.AnythingOfType("*sql.Tx")).Return(nil)
				},
				MockAssert: func(t *testing.T) {
					mockServiceTestHelper.walletRepo.AssertExpectations(t)
					mockServiceTestHelper.transactionService.AssertExpectations(t)
				},
			},
			userID: testUserID,
			amount: 50.0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			setupServiceMock()
			tc.MockSetup()
			walletService := NewWalletService(mockServiceTestHelper.walletRepo, mockServiceTestHelper.transactionService)
			_, err := walletService.Transfer(tc.userID, tc.walletNumber, tc.amount)
			if tc.TestType == "error" {
				assert.ErrorIs(t, err, tc.ExpectedError)
			} else {
				assert.NoError(t, err)
			}
			tc.MockAssert(t)
		})
	}
}

func TestCreateWalletService(t *testing.T) {

	testCases := []testWalletService{
		{
			BaseHandlerTestCase: testutils.BaseHandlerTestCase{
				Name:          "successful wallet creation",
				TestType:      "success",
				ExpectedError: nil,
				MockSetup: func() {
					mockServiceTestHelper.walletRepo.On("UserExists", mock.Anything).Return(false, nil)
					mockServiceTestHelper.walletRepo.On("CreateWallet", mock.Anything).Return(nil)
				},
				MockAssert: func(t *testing.T) {
					mockServiceTestHelper.walletRepo.AssertExpectations(t)
				},
			},
			userID:       testUserID,
			walletNumber: testWalletNumber,
		},
		{
			BaseHandlerTestCase: testutils.BaseHandlerTestCase{
				Name:          "error checking if user exists",
				TestType:      "error",
				ExpectedError: utils.ErrDatabaseError,
				MockSetup: func() {
					mockServiceTestHelper.walletRepo.On("UserExists", mock.Anything).Return(false, utils.ErrDatabaseError)
				},
				MockAssert: func(t *testing.T) {
					mockServiceTestHelper.walletRepo.AssertExpectations(t)
				},
			},
			userID:       testUserID,
			walletNumber: testWalletNumber,
		},
		{
			BaseHandlerTestCase: testutils.BaseHandlerTestCase{
				Name:          "error creating wallet",
				TestType:      "error",
				ExpectedError: utils.ErrDatabaseError,
				MockSetup: func() {
					mockServiceTestHelper.walletRepo.On("UserExists", mock.Anything).Return(false, nil)
					mockServiceTestHelper.walletRepo.On("CreateWallet", mock.Anything).Return(utils.ErrDatabaseError)
				},
				MockAssert: func(t *testing.T) {
					mockServiceTestHelper.walletRepo.AssertExpectations(t)
				},
			},
			userID:       testUserID,
			walletNumber: testWalletNumber,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.Name, func(t *testing.T) {
			setupServiceMock()
			tt.MockSetup()
			walletService := NewWalletService(mockServiceTestHelper.walletRepo, mockServiceTestHelper.transactionService)
			wallet, err := walletService.CreateWallet(tt.userID)

			if tt.TestType == "success" {
				assert.NoError(t, err)
				assert.NotNil(t, *wallet)
				assert.Equal(t, tt.userID, wallet.UserID)
			} else {
				assert.EqualError(t, err, tt.ExpectedError.Error())
				assert.Nil(t, wallet)
			}
			tt.MockAssert(t)
		})
	}
}
