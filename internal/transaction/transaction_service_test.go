package transaction_test

import (
	"centralized-wallet/internal/transaction"
	"centralized-wallet/internal/utils"
	mockTransaction "centralized-wallet/tests/mocks/transaction"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
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
	ts := transaction.NewTransactionService(mockTransactionTestHelper.repo)

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
	assert.Equal(t, utils.ServiceErrWalletNumberNil, err.Error())
}
