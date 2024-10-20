package repository

import (
	"centralized-wallet/internal/database"
	"centralized-wallet/internal/models"
	"centralized-wallet/tests/testutils"

	"log"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Global variable for the shared database service
var dbService database.Service

// TestMain will be executed before any of the tests and will handle the setup and teardown of the test container
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

// TestCreateTransaction tests the creation of a transaction in the database
func TestCreateTransaction(t *testing.T) {
	// Reuse the initialized database connection
	transactionRepo := &TransactionRepository{db: dbService.GetDB()}

	fromWalletNumber := "WAL-1-20231010-ABC123" // Example wallet number
	toWalletNumber := "WAL-2-20231010-DEF456"
	// Test a transfer transaction
	transaction := &models.Transaction{
		FromWalletNumber: &fromWalletNumber,
		ToWalletNumber:   &toWalletNumber,
		Type:             "transfer",
		Amount:           100.0,
		CreatedAt:        time.Now(),
	}

	err := transactionRepo.CreateTransaction(transaction)
	assert.NoError(t, err)

	// Test a deposit transaction
	toWalletNumber = "WAL-1-20231010-ABC123" // Updating to the same wallet number for deposit
	transaction = &models.Transaction{
		FromWalletNumber: nil,
		ToWalletNumber:   &toWalletNumber,
		Type:             "deposit",
		Amount:           200.0,
		CreatedAt:        time.Now(),
	}

	err = transactionRepo.CreateTransaction(transaction)
	assert.NoError(t, err)

	cleanDB(t)
}

func TestGetTransactionHistory(t *testing.T) {
	// Reuse the initialized database connection
	transactionRepo := &TransactionRepository{db: dbService.GetDB()}

	// Insert mock wallets
	fromWalletNumber := "WAL-1-20231010-ABC123" // Example wallet number for from user
	toWalletNumber := "WAL-2-20231010-DEF456"   // Example wallet number for to user
	fromUserEmail := "from_user@example.com"
	toUserEmail := "to_user@example.com"

	_, err := dbService.GetDB().Exec(`
		INSERT INTO users (id, email, password) VALUES
		($1, $2, 'password1'),
		($3, $4, 'password2')`,
		1, fromUserEmail, 2, toUserEmail,
	)
	assert.NoError(t, err)

	_, err = dbService.GetDB().Exec(`
		INSERT INTO wallets (user_id, wallet_number, balance) VALUES
		(1, $1, 1000),
		(2, $2, 2000)`,
		fromWalletNumber, toWalletNumber,
	)
	assert.NoError(t, err)

	// Insert mock transaction into the DB
	transaction := &models.Transaction{
		FromWalletNumber: &fromWalletNumber,
		ToWalletNumber:   &toWalletNumber,
		Type:             "transfer",
		Amount:           100.0,
		CreatedAt:        time.Now(),
	}

	err = transactionRepo.CreateTransaction(transaction)
	assert.NoError(t, err)

	// Retrieve transaction history for the first wallet
	transactions, err := transactionRepo.GetTransactionHistory(fromWalletNumber, "DESC", 1)
	assert.NoError(t, err)
	assert.Len(t, transactions, 1)

	// Verify the transaction details with nil checks
	if transactions[0].FromWalletNumber != nil {
		assert.Equal(t, fromWalletNumber, *transactions[0].FromWalletNumber)
	} else {
		t.Errorf("Expected FromWalletNumber to be %s, but got nil", fromWalletNumber)
	}

	if transactions[0].ToWalletNumber != nil {
		assert.Equal(t, toWalletNumber, *transactions[0].ToWalletNumber)
	} else {
		t.Errorf("Expected ToWalletNumber to be %s, but got nil", toWalletNumber)
	}

	assert.Equal(t, "transfer", transactions[0].Type)
	assert.Equal(t, 100.0, transactions[0].Amount)

	// Verify the emails
	if transactions[0].FromEmail != nil {
		assert.Equal(t, fromUserEmail, *transactions[0].FromEmail)
	} else {
		t.Errorf("Expected FromEmail to be %s, but got nil", fromUserEmail)
	}

	if transactions[0].ToEmail != nil {
		assert.Equal(t, toUserEmail, *transactions[0].ToEmail)
	} else {
		t.Errorf("Expected ToEmail to be %s, but got nil", toUserEmail)
	}

	cleanDB(t)
}

func cleanDB(t *testing.T) {
	if err := testutils.CleanDatabase(dbService.GetDB()); err != nil {
		t.Fatalf("Failed to clean the database: %v", err)
	}
}
