package transaction_test

import (
	"centralized-wallet/internal/database"
	"centralized-wallet/internal/models"
	"centralized-wallet/internal/seed"
	"centralized-wallet/internal/transaction"
	"centralized-wallet/tests/testutils"
	"database/sql"
	"log"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var dbService database.Service

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

// GenerateSampleTransactions returns 5 sample transactions for testing
func GenerateSampleTransactions() []models.Transaction {
	now := time.Now()

	// Creating dummy wallet numbers
	fromWallet1 := "wallet123"
	fromWallet2 := "wallet456"
	toWallet1 := "wallet789"
	toWallet2 := "wallet101112"

	return []models.Transaction{
		{
			ID:               1,
			FromWalletNumber: &fromWallet1,            // From wallet
			ToWalletNumber:   &toWallet1,              // To wallet
			TransactionType:  "transfer",              // Transaction type: transfer
			Amount:           50.00,                   // Transfer amount
			CreatedAt:        now.Add(-1 * time.Hour), // Transaction created 1 hour ago
		},
		{
			ID:               2,
			FromWalletNumber: nil,                     // Deposit to wallet, no from wallet
			ToWalletNumber:   &toWallet1,              // To wallet
			TransactionType:  "deposit",               // Transaction type: deposit
			Amount:           200.00,                  // Deposit amount
			CreatedAt:        now.Add(-2 * time.Hour), // Transaction created 2 hours ago
		},
		{
			ID:               3,
			FromWalletNumber: nil,                        // Deposit to wallet, no from wallet
			ToWalletNumber:   &toWallet2,                 // To wallet
			TransactionType:  "deposit",                  // Transaction type: deposit
			Amount:           300.00,                     // Deposit amount
			CreatedAt:        now.Add(-30 * time.Minute), // Transaction created 30 minutes ago
		},
		{
			ID:               4,
			FromWalletNumber: &fromWallet2,               // From wallet
			ToWalletNumber:   &toWallet2,                 // To wallet
			TransactionType:  "transfer",                 // Transaction type: transfer
			Amount:           150.00,                     // Transfer amount
			CreatedAt:        now.Add(-15 * time.Minute), // Transaction created 15 minutes ago
		},
		{
			ID:               5,
			FromWalletNumber: &fromWallet1,               // From wallet
			ToWalletNumber:   &toWallet2,                 // To wallet
			TransactionType:  "transfer",                 // Transaction type: transfer
			Amount:           75.00,                      // Transfer amount
			CreatedAt:        now.Add(-10 * time.Minute), // Transaction created 10 minutes ago
		},
	}
}

func setupFixtures() {
	for _, transaction := range GenerateSampleTransactions() {
		seed.SeedTransactions(dbService.GetDB(), &transaction)
	}

	db := dbService.GetDB()
	rows, err := db.Query("SELECT id, from_wallet_number, to_wallet_number, amount FROM transactions")
	if err != nil {
		log.Fatalf("Failed to query seeded transactions: %v", err)
	}
	defer rows.Close()

	log.Println("Seeded transactions:")
	for rows.Next() {
		var id int
		var fromWallet sql.NullString
		var toWallet sql.NullString
		var amount float64

		err := rows.Scan(&id, &fromWallet, &toWallet, &amount)
		if err != nil {
			log.Fatalf("Failed to scan transaction row: %v", err)
		}

		log.Printf("ID: %d, FromWallet: %s, ToWallet: %s, Amount: %.2f\n",
			id, fromWallet.String, toWallet.String, amount)
	}
}

func TestGetTransactionHistoryService(t *testing.T) {
	// Set up test data (fixtures)
	setupFixtures()

	// Reuse the initialized database connection
	transactionRepo := transaction.NewTransactionRepository(dbService.GetDB())
	transactionService := transaction.NewTransactionService(transactionRepo)

	// Define the test cases (table-driven test)
	testCases := []struct {
		name            string
		walletNumber    string
		orderBy         string
		limit           int
		offset          int
		expectedLength  int
		expectedAmounts []float64
	}{
		{
			name:            "First page, descending order",
			walletNumber:    "wallet123",
			orderBy:         "DESC",
			limit:           2,
			offset:          0,
			expectedLength:  2,
			expectedAmounts: []float64{75.00, 50.00},
		},
		{
			name:            "First page, ascending order",
			walletNumber:    "wallet123",
			orderBy:         "ASC",
			limit:           2,
			offset:          0,
			expectedLength:  2,
			expectedAmounts: []float64{50.00, 75.00},
		},
		{
			name:            "Second page, descending order",
			walletNumber:    "wallet123",
			orderBy:         "DESC",
			limit:           1,
			offset:          1,
			expectedLength:  1,
			expectedAmounts: []float64{50.00},
		},
		{
			name:            "Limit more than available transactions",
			walletNumber:    "wallet123",
			orderBy:         "DESC",
			limit:           10,
			offset:          0,
			expectedLength:  2,
			expectedAmounts: []float64{75.00, 50.00},
		},
		{
			name:            "Offset beyond data range",
			walletNumber:    "wallet123",
			orderBy:         "DESC",
			limit:           2,
			offset:          10,
			expectedLength:  0,
			expectedAmounts: []float64{},
		},
	}

	defer testutils.CleanDatabase(dbService.GetDB())

	// Iterate over each test case
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Call the method you're testing
			transactions, err := transactionService.GetTransactionHistory(tc.walletNumber, tc.orderBy, tc.limit, tc.offset)

			// Ensure no error occurred
			assert.NoError(t, err)

			// Assert that the correct number of transactions are returned
			assert.Len(t, transactions, tc.expectedLength)

			// Check the transaction amounts if there are results
			for i, tx := range transactions {
				assert.Equal(t, tc.expectedAmounts[i], tx.Amount)
			}
		})
	}
}
