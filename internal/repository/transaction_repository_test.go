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

	// Test a transfer transaction
	fromUserID := 1
	toUserID := 2
	transaction := &models.Transaction{
		FromUserID: &fromUserID,
		ToUserID:   &toUserID,
		Type:       "transfer",
		Amount:     100.0,
		CreatedAt:  time.Now(),
	}

	err := transactionRepo.CreateTransaction(transaction)
	assert.NoError(t, err)

	// Test a deposit transaction
	toUserID = 1
	transaction = &models.Transaction{
		FromUserID: nil,
		ToUserID:   &toUserID,
		Type:       "deposit",
		Amount:     200.0,
		CreatedAt:  time.Now(),
	}

	err = transactionRepo.CreateTransaction(transaction)
	assert.NoError(t, err)

	cleanDB(t)
}

// TestGetTransactionHistory tests the retrieval of a user's transaction history
func TestGetTransactionHistory(t *testing.T) {
	// Reuse the initialized database connection
	transactionRepo := &TransactionRepository{db: dbService.GetDB()}

	// Insert mock transactions into the DB
	fromUserID := 1
	toUserID := 2
	transaction := &models.Transaction{
		FromUserID: &fromUserID,
		ToUserID:   &toUserID,
		Type:       "transfer",
		Amount:     100.0,
		CreatedAt:  time.Now(),
	}

	err := transactionRepo.CreateTransaction(transaction)
	assert.NoError(t, err)

	// Retrieve transaction history for user 1
	transactions, err := transactionRepo.GetTransactionHistory(1)
	assert.NoError(t, err)
	assert.Len(t, transactions, 1)

	// Verify the transaction details
	assert.Equal(t, fromUserID, *transactions[0].FromUserID)
	assert.Equal(t, toUserID, *transactions[0].ToUserID)
	assert.Equal(t, "transfer", transactions[0].Type)
	assert.Equal(t, 100.0, transactions[0].Amount)
	cleanDB(t)
}

func cleanDB(t *testing.T) {
	if err := testutils.CleanDatabase(dbService.GetDB()); err != nil {
		t.Fatalf("Failed to clean the database: %v", err)
	}
}
