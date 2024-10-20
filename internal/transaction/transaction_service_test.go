package transaction_test

import (
	"centralized-wallet/internal/models"
	"centralized-wallet/internal/transaction"
	mockTransaction "centralized-wallet/tests/mocks/transaction"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock service test helper struct
var mockTransactionTestHelper struct {
	repo *mockTransaction.MockTransactionRepository
}

// Setup function to initialize mocks
func setupTransactionServiceMock() {
	mockTransactionTestHelper.repo = new(mockTransaction.MockTransactionRepository)
}

// Test RecordTransaction method
func TestRecordTransaction_Success(t *testing.T) {
	// Step 1: Initialize mocks
	setupTransactionServiceMock()

	// Step 2: Create a new TransactionService
	ts := transaction.NewTransactionService(mockTransactionTestHelper.repo)

	// Define mock data
	fromWalletNumber := "from-wallet"
	toWalletNumber := "to-wallet"
	amount := 100.0
	transactionType := "transfer"

	// Mock the CreateTransaction method
	mockTransactionTestHelper.repo.On("CreateTransaction", mock.AnythingOfType("*models.Transaction")).Return(nil)

	// Act: Call the RecordTransaction method
	err := ts.RecordTransaction(&fromWalletNumber, &toWalletNumber, transactionType, amount)

	// Assert: Check the expected results
	assert.NoError(t, err)
	mockTransactionTestHelper.repo.AssertExpectations(t)
}

func TestRecordTransaction_Error(t *testing.T) {
	// Step 1: Initialize mocks
	setupTransactionServiceMock()

	// Step 2: Create a new TransactionService
	ts := transaction.NewTransactionService(mockTransactionTestHelper.repo)

	// Test when both fromWalletNumber and toWalletNumber are nil or empty
	var fromWalletNumber *string = nil
	var toWalletNumber *string = nil
	amount := 100.0
	transactionType := "transfer"

	// Act: Call the RecordTransaction method
	err := ts.RecordTransaction(fromWalletNumber, toWalletNumber, transactionType, amount)

	// Assert: Check the expected results
	assert.Error(t, err)
	assert.Equal(t, "either fromWalletNumber or toWalletNumber must be provided", err.Error())
}

// Test GetTransactionHistory method
func TestGetTransactionHistory_Success(t *testing.T) {
	// Step 1: Initialize mocks
	setupTransactionServiceMock()

	// Step 2: Create a new TransactionService
	ts := transaction.NewTransactionService(mockTransactionTestHelper.repo)

	// Define mock data
	walletNumber := "test-wallet"
	orderBy := "created_at"
	limit := 10
	fromEmail := "from@example.com"
	toEmail := "to@example.com"
	// Mock the GetTransactionHistory method
	mockTransactionTestHelper.repo.On("GetTransactionHistory", walletNumber, orderBy, limit).Return([]models.TransactionWithEmails{
		{
			Transaction: models.Transaction{
				ID:        1,
				Amount:    100.0,
				CreatedAt: time.Now(),
			},
			FromEmail: &fromEmail,
			ToEmail:   &toEmail,
		},
	}, nil)

	// Act: Call the GetTransactionHistory method
	transactions, err := ts.GetTransactionHistory(walletNumber, orderBy, limit)

	// Assert: Check the expected results
	assert.NoError(t, err)
	assert.Len(t, transactions, 1)
	assert.Equal(t, 1, transactions[0].ID)
	assert.Equal(t, "from@example.com", *transactions[0].FromEmail)
	assert.Equal(t, "to@example.com", *transactions[0].ToEmail)
	mockTransactionTestHelper.repo.AssertExpectations(t)
}

func TestGetTransactionHistory_NoResults(t *testing.T) {
	// Step 1: Initialize mocks
	setupTransactionServiceMock()

	// Step 2: Create a new TransactionService
	ts := transaction.NewTransactionService(mockTransactionTestHelper.repo)

	// Define mock data
	walletNumber := "test-wallet"
	orderBy := "created_at"
	limit := 10

	// Mock the GetTransactionHistory method to return no transactions
	mockTransactionTestHelper.repo.On("GetTransactionHistory", walletNumber, orderBy, limit).Return([]models.TransactionWithEmails{}, nil)

	// Act: Call the GetTransactionHistory method
	transactions, err := ts.GetTransactionHistory(walletNumber, orderBy, limit)

	// Assert: Check the expected results
	assert.NoError(t, err)
	assert.Len(t, transactions, 0)
	mockTransactionTestHelper.repo.AssertExpectations(t)
}
