package transaction_test

import (
	"centralized-wallet/internal/models"
	"centralized-wallet/internal/transaction"
	"centralized-wallet/internal/utils"
	mockTransaction "centralized-wallet/tests/mocks/transaction"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	testToWalletNumber   = "1234567890"
	testFromWalletNumber = "0987654321"
	testAmount           = 50.0
	testEmail            = "user1@example.com"
	toTestEmail          = "user1@example.com"
	now                  = time.Now()
)

// Mock service test helper struct
var mockTransactionTestHelper struct {
	repo *mockTransaction.MockTransactionRepository
}

// Setup function to initialize mocks
func setupTransactionServiceMock() {
	mockTransactionTestHelper.repo = new(mockTransaction.MockTransactionRepository)
}

func TestRecordTransactionService_Error(t *testing.T) {
	// Step 1: Initialize mocks
	setupTransactionServiceMock()

	// Step 2: Create a new TransactionService
	ts := transaction.NewTransactionService(mockTransactionTestHelper.repo, nil)

	// Test when both fromWalletNumber and toWalletNumber are nil or empty
	var fromWalletNumber *string = nil
	var toWalletNumber *string = nil
	amount := 100.0
	transactionType := "transfer"

	mockTx := new(sql.Tx)
	// Act: Call the RecordTransaction method
	err := ts.RecordTransaction(mockTx, fromWalletNumber, toWalletNumber, transactionType, amount)

	// Assert: Check the expected results
	assert.Error(t, err)
	assert.Equal(t, utils.ServiceErrWalletNumberNil.Error(), err.Error())
}

// New test function for FormatTransactionResponse
func TestFormatTransactionResponse(t *testing.T) {
	// Step 1: Initialize service (no mocks needed for this specific function)
	ts := transaction.NewTransactionService(nil, nil) // Repo not required for this test

	// Step 2: Create mock transaction data
	transactions := []models.TransactionWithEmails{
		// Deposit transaction (Incoming)
		{
			Transaction: models.Transaction{
				ID:               1,
				FromWalletNumber: nil,
				ToWalletNumber:   &testFromWalletNumber,
				TransactionType:  "deposit",
				Amount:           100.0,
				CreatedAt:        now,
			},
			FromEmail: nil,
			ToEmail:   &testEmail,
		},
		// Withdraw transaction (Outgoing)
		{
			Transaction: models.Transaction{
				ID:               2,
				FromWalletNumber: &testFromWalletNumber,
				ToWalletNumber:   nil,
				TransactionType:  "withdraw",
				Amount:           testAmount,
				CreatedAt:        now,
			},
			FromEmail: &testEmail,
			ToEmail:   nil,
		},
		// Transfer In transaction
		{
			Transaction: models.Transaction{
				ID:               3,
				FromWalletNumber: &testToWalletNumber,
				ToWalletNumber:   &testFromWalletNumber,
				TransactionType:  "transfer",
				Amount:           150.0,
				CreatedAt:        now,
			},
			FromEmail: &testEmail,
			ToEmail:   &testEmail, // Can simulate different emails for this if needed
		},
		// Transfer Out transaction
		{
			Transaction: models.Transaction{
				ID:               4,
				FromWalletNumber: &testFromWalletNumber,
				ToWalletNumber:   &testToWalletNumber,
				TransactionType:  "transfer",
				Amount:           200.0,
				CreatedAt:        now,
			},
			FromEmail: &testEmail,
			ToEmail:   &testEmail, // Can simulate different emails for this if needed
		},
	}

	// Step 3: Call the FormatTransactionResponse method
	formattedTransactions := ts.FormatTransactionResponse(testFromWalletNumber, transactions)

	// Step 4: Assert the results
	assert.Len(t, formattedTransactions, 4)

	// Check first transaction (deposit - incoming)
	assert.Equal(t, "incoming", formattedTransactions[0].Direction)
	assert.Equal(t, "deposit", formattedTransactions[0].TransactionType)
	assert.Equal(t, 100.0, formattedTransactions[0].Amount)
	assert.Equal(t, "", formattedTransactions[0].FromWalletNumber) // FromWalletNumber should be empty
	assert.Equal(t, "", formattedTransactions[0].ToEmail)

	// Check second transaction (withdraw - outgoing)
	assert.Equal(t, "outgoing", formattedTransactions[1].Direction)
	assert.Equal(t, "withdraw", formattedTransactions[1].TransactionType)
	assert.Equal(t, testAmount, formattedTransactions[1].Amount)
	assert.Equal(t, "", formattedTransactions[1].FromWalletNumber)
	assert.Equal(t, "", formattedTransactions[1].ToWalletNumber) // ToWalletNumber should be empty

	// Check third transaction (transfer in - incoming)
	assert.Equal(t, "incoming", formattedTransactions[2].Direction)
	assert.Equal(t, "transfer", formattedTransactions[2].TransactionType)
	assert.Equal(t, 150.0, formattedTransactions[2].Amount)
	assert.Equal(t, testToWalletNumber, formattedTransactions[2].FromWalletNumber) // Should be set to sender's wallet number
	assert.Equal(t, testEmail, formattedTransactions[2].FromEmail)                 // Should match sender's email

	// Check fourth transaction (transfer out - outgoing)
	assert.Equal(t, "outgoing", formattedTransactions[3].Direction)
	assert.Equal(t, "transfer", formattedTransactions[3].TransactionType)
	assert.Equal(t, 200.0, formattedTransactions[3].Amount)
	assert.Equal(t, testToWalletNumber, formattedTransactions[3].ToWalletNumber) // Should be set to sender's wallet number
	assert.Equal(t, testEmail, formattedTransactions[3].ToEmail)                 // Should match recipient's email
}
