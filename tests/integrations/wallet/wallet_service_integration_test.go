package wallet_test

import (
	"centralized-wallet/internal/database"
	"centralized-wallet/internal/models"
	"centralized-wallet/internal/seed"
	"centralized-wallet/internal/transaction"
	"centralized-wallet/internal/utils"
	"centralized-wallet/internal/wallet"
	"centralized-wallet/tests/testutils"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var dbService database.Service

type testWalletService struct {
	name                string
	userId              int
	fromUserId          int
	toWalletNumber      string
	amount              float64
	expectedError       error
	expectedFromBalance float64
	expectedToBalance   float64
	expectedBalance     float64
	expectWallet        bool
	shouldRecordTx      bool
}

func TestMain(m *testing.M) {
	// Start the Postgres container
	teardown, err := testutils.StartPostgresContainer(true)
	if err != nil {
		log.Fatalf("Could not start Postgres container for testing: %v", err)
	}

	// Set up environment variables for the test database
	testutils.InitEnv()

	// Initialize the database service only once
	dbService = database.New()

	// Run the tests
	code := m.Run()

	// Teardown the container after the tests
	testutils.TeardownContainer(teardown)

	// Exit with the appropriate exit code
	os.Exit(code)
}

func setupUserFixtures() {
	db := dbService.GetDB()

	// Seed users
	for _, user := range seed.GenerateSampleUsers() {
		err := seed.SeedUser(db, &user)

		if err != nil {
			log.Fatalf("Failed to seed user: %v", err)
		}
	}
}

func setupWalletFixtures() {
	db := dbService.GetDB()

	// Seed wallets for some users (this will simulate users that already have wallets)
	for _, wallet := range seed.GenerateSampleWallets() {
		err := seed.SeedWallets(db, &wallet)
		if err != nil {
			log.Fatalf("Failed to seed wallet: %v", err)
		}
	}
}

func TestGetWalletByUserIDService(t *testing.T) {
	setupUserFixtures()
	setupWalletFixtures()

	// Initialize the wallet repository and service
	walletRepo := wallet.NewWalletRepository(dbService.GetDB())
	walletService := wallet.NewWalletService(walletRepo, nil)

	// Define the test cases
	testCases := []struct {
		name           string
		userId         int
		expectedError  error
		expectedWallet *models.Wallet
	}{
		{
			name:          "Successful wallet retrieval",
			userId:        1, // Assuming user 1 has a wallet
			expectedError: nil,
			expectedWallet: &models.Wallet{
				ID:           1,
				UserID:       1,
				WalletNumber: "wallet123",
				Balance:      100.0, // Expected balance for this user
			},
		},
		{
			name:           "Wallet not found",
			userId:         9999,                        // Non-existent user
			expectedError:  utils.RepoErrWalletNotFound, // Expecting "wallet not found" error
			expectedWallet: nil,                         // No wallet should be returned
		},
	}

	defer testutils.CleanDatabase(dbService.GetDB())

	// Iterate over the test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Call the service to get the wallet by user ID
			wallet, err := walletService.GetWalletByUserID(tc.userId)

			// Check if the error matches the expected error
			if tc.expectedError != nil {
				assert.ErrorIs(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
			}

			// If we expect a wallet to be returned, check its fields
			if tc.expectedWallet != nil {
				assert.NotNil(t, wallet)
				assert.Equal(t, tc.expectedWallet.UserID, wallet.UserID)
				assert.Equal(t, tc.expectedWallet.WalletNumber, wallet.WalletNumber)
				assert.Equal(t, tc.expectedWallet.Balance, wallet.Balance)
			} else {
				// If no wallet is expected, ensure it is nil
				assert.Nil(t, wallet)
			}
		})
	}
}

func TestCreateWalletService(t *testing.T) {
	setupUserFixtures()

	// Initialize the wallet repository and service
	walletRepo := wallet.NewWalletRepository(dbService.GetDB())
	walletService := wallet.NewWalletService(walletRepo, nil)

	// Define the test cases
	testCases := []testWalletService{
		{
			name:            "Create wallet for new user without existing wallet",
			userId:          3, // Assuming this user does not have a wallet
			expectedError:   nil,
			expectWallet:    true,
			expectedBalance: 0.0,
		},
		{
			name:            "Fail to create wallet for user with existing wallet",
			userId:          3,                            // Assuming this user already has a wallet (pre-seeded in setup)
			expectedError:   utils.ErrWalletAlreadyExists, // Replace with the correct error you're handling
			expectWallet:    false,
			expectedBalance: 0.0,
		},
		{
			name:            "Create wallet for another user without existing wallet",
			userId:          2, // Assuming this user does not have a wallet
			expectedError:   nil,
			expectWallet:    true,
			expectedBalance: 0.0,
		},
	}

	defer testutils.CleanDatabase(dbService.GetDB())

	// Iterate over the test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Call the service to create a wallet
			wallet, err := walletService.CreateWallet(tc.userId)

			// Check if the error matches the expected error
			if tc.expectedError != nil {
				assert.ErrorIs(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
			}

			// If we expect a wallet to be created, check the wallet fields
			if tc.expectWallet {
				assert.NotNil(t, wallet)
				assert.Equal(t, tc.userId, wallet.UserID)
				assert.Equal(t, tc.expectedBalance, wallet.Balance)
			} else {
				// If we don't expect a wallet, ensure it is nil
				assert.Nil(t, wallet)
			}
		})
	}
}

func TestDepositService(t *testing.T) {
	// Initialize the wallet repository and service
	setupUserFixtures()
	setupWalletFixtures()

	walletRepo := wallet.NewWalletRepository(dbService.GetDB())
	transactionRepo := transaction.NewTransactionRepository(dbService.GetDB())
	transactionService := transaction.NewTransactionService(transactionRepo)
	walletService := wallet.NewWalletService(walletRepo, transactionService)

	// Define the test cases
	testCases := []testWalletService{
		{
			name:            "Successful deposit for user with wallet",
			userId:          1,     // Assuming this user has an existing wallet (from setup)
			amount:          100.0, // Deposit amount
			expectedError:   nil,   // No error expected
			expectedBalance: 200.0, // Assuming initial balance is 100.00 (will increase by 100.00)
			expectWallet:    true,  // Wallet should exist
			shouldRecordTx:  true,  // Transaction should be recorded
		},
		{
			name:            "Wallet not found for user",
			userId:          9999,                        // Non-existent user ID
			amount:          50.0,                        // Deposit amount
			expectedError:   utils.RepoErrWalletNotFound, // Error expected for non-existent wallet
			expectedBalance: 0.0,                         // No change in balance
			expectWallet:    false,
			shouldRecordTx:  false,
		},
	}

	defer testutils.CleanDatabase(dbService.GetDB())

	// Iterate over the test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Call the service to deposit into the wallet
			wallet, err := walletService.Deposit(tc.userId, tc.amount)

			// Check if the error matches the expected error
			if tc.expectedError != nil {
				assert.ErrorIs(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
			}

			// If we expect a wallet to be returned, check the wallet fields
			if tc.expectWallet {
				assert.NotNil(t, wallet)
				assert.Equal(t, tc.expectedBalance, wallet.Balance)
			} else {
				// If no wallet should be returned, ensure it is nil
				assert.Nil(t, wallet)
			}

			// Check if a transaction should be recorded
			if tc.shouldRecordTx {
				// Verify that the transaction was recorded in the database
				verifyTransactionRecorded(t, wallet.WalletNumber, tc.amount, "deposit", false)
			}
		})
	}
}

func TestWithdrawService(t *testing.T) {
	setupUserFixtures()
	setupWalletFixtures()

	// Initialize the wallet repository and service
	walletRepo := wallet.NewWalletRepository(dbService.GetDB())
	transactionRepo := transaction.NewTransactionRepository(dbService.GetDB())
	transactionService := transaction.NewTransactionService(transactionRepo)
	walletService := wallet.NewWalletService(walletRepo, transactionService)

	// Define the test cases
	testCases := []testWalletService{
		{
			name:            "Successful withdrawal with sufficient funds",
			userId:          1,    // Assuming this user has an existing wallet with enough balance
			amount:          50.0, // Withdraw amount
			expectedError:   nil,  // No error expected
			expectedBalance: 50.0, // Assuming initial balance is 100.00 (will decrease by 50.00)
			expectWallet:    true, // Wallet should exist
			shouldRecordTx:  true, // Transaction should be recorded
		},
		{
			name:            "Withdrawal with insufficient funds",
			userId:          1,                              // Assuming this user has an existing wallet
			amount:          200.0,                          // Withdraw amount greater than balance
			expectedError:   utils.RepoErrInsufficientFunds, // Error expected for insufficient funds
			expectedBalance: 100.0,                          // Balance should not change
			expectWallet:    false,                          // Wallet should exist
			shouldRecordTx:  false,                          // Transaction should not be recorded
		},
		{
			name:            "Wallet not found for user",
			userId:          9999,                        // Non-existent user ID
			amount:          50.0,                        // Withdraw amount
			expectedError:   utils.RepoErrWalletNotFound, // Error expected for non-existent wallet
			expectedBalance: 0.0,                         // No change in balance
			expectWallet:    false,                       // Wallet should not exist
			shouldRecordTx:  false,                       // Transaction should not be recorded
		},
	}

	defer testutils.CleanDatabase(dbService.GetDB())

	// Iterate over the test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Call the service to withdraw from the wallet
			wallet, err := walletService.Withdraw(tc.userId, tc.amount)

			// Check if the error matches the expected error
			if tc.expectedError != nil {
				assert.ErrorIs(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
			}

			// If we expect a wallet to be returned, check the wallet fields
			if tc.expectWallet {
				assert.NotNil(t, wallet)
				assert.Equal(t, tc.expectedBalance, wallet.Balance)
			} else {
				// If no wallet should be returned, ensure it is nil
				assert.Nil(t, wallet)
			}

			// Check if a transaction should be recorded
			if tc.shouldRecordTx {
				// Verify that the transaction was recorded in the database
				verifyTransactionRecorded(t, wallet.WalletNumber, tc.amount, "withdraw", true)
			}
		})
	}
}

func TestTransferService(t *testing.T) {
	setupUserFixtures()
	setupWalletFixtures()

	// Initialize the wallet repository and service
	walletRepo := wallet.NewWalletRepository(dbService.GetDB())
	transactionRepo := transaction.NewTransactionRepository(dbService.GetDB())
	transactionService := transaction.NewTransactionService(transactionRepo)
	walletService := wallet.NewWalletService(walletRepo, transactionService)

	// Define the test cases
	testCases := []testWalletService{
		{
			name:                "Successful transfer with sufficient funds",
			fromUserId:          1,           // Assuming user 1 has sufficient balance
			toWalletNumber:      "wallet456", // Assuming wallet 456 belongs to user 2
			amount:              50.0,        // Transfer amount
			expectedError:       nil,         // No error expected
			expectedFromBalance: 50.0,        // Assuming initial balance is 100.00 (will decrease by 50.00)
			expectedToBalance:   250.0,       // Assuming recipient initial balance is 200.00 (will increase by 50.00)
			expectWallet:        true,        // Wallets should exist
			shouldRecordTx:      true,        // Transaction should be recorded
		},
		{
			name:                "Transfer with insufficient funds",
			fromUserId:          1,                              // Assuming user 1 has insufficient funds
			toWalletNumber:      "wallet456",                    // Valid recipient wallet
			amount:              200.0,                          // Transfer amount greater than balance
			expectedError:       utils.RepoErrInsufficientFunds, // Error expected for insufficient funds
			expectedFromBalance: 100.0,                          // Balance should not change
			expectedToBalance:   200.0,                          // Recipient balance should not change
			expectWallet:        false,                          // Wallet should exist
			shouldRecordTx:      false,                          // Transaction should not be recorded
		},
		{
			name:                "Recipient wallet not found",
			fromUserId:          1,                           // Assuming user 1 has sufficient balance
			toWalletNumber:      "nonexistent_wallet",        // Non-existent wallet
			amount:              50.0,                        // Transfer amount
			expectedError:       utils.RepoErrWalletNotFound, // Error expected for non-existent recipient wallet
			expectedFromBalance: 100.0,                       // Sender balance should not change
			expectedToBalance:   0.0,                         // Recipient balance should not change (invalid wallet)
			expectWallet:        false,                       // Sender wallet should exist
			shouldRecordTx:      false,                       // Transaction should not be recorded
		},
	}

	defer testutils.CleanDatabase(dbService.GetDB())

	// Iterate over the test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Call the service to perform the transfer
			fromWallet, err := walletService.Transfer(tc.fromUserId, tc.toWalletNumber, tc.amount)

			// Check if the error matches the expected error
			if tc.expectedError != nil {
				assert.ErrorIs(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
			}

			// If we expect wallets to exist, check the wallet balances
			if tc.expectWallet {
				assert.NotNil(t, fromWallet)

				// Check the balance of the sender's wallet
				if fromWallet != nil {
					assert.Equal(t, tc.expectedFromBalance, fromWallet.Balance)
				}

				// Verify the recipient's wallet balance
				if tc.shouldRecordTx {
					toWallet, err := walletRepo.FindByWalletNumber(tc.toWalletNumber)
					assert.NoError(t, err)
					assert.Equal(t, tc.expectedToBalance, toWallet.Balance)
				}
			}

			// Check if a transaction should be recorded
			if tc.shouldRecordTx {
				// Verify that the transaction was recorded in the database
				verifyTransactionRecorded(t, fromWallet.WalletNumber, tc.amount, "transfer", true)
				verifyTransactionRecorded(t, tc.toWalletNumber, tc.amount, "transfer", false)
			}
		})
	}
}

func verifyTransactionRecorded(t *testing.T, walletNumber string, amount float64, txType string, isFromWallet bool) {
	db := dbService.GetDB()

	// Query the transactions table for the matching transaction based on wallet number
	var count int
	queryField := "to_wallet_number"
	if isFromWallet {
		queryField = "from_wallet_number"
	}

	err := db.QueryRow(`
		SELECT COUNT(*)
		FROM transactions
		WHERE `+queryField+` = $1
		  AND transaction_type = $2
		  AND amount = $3`, walletNumber, txType, amount).Scan(&count)

	// Assert that the transaction exists
	assert.NoError(t, err)
	assert.Equal(t, 1, count, "Expected 1 transaction to be recorded")
}
