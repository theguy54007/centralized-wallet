package mock_transaction

import (
	"centralized-wallet/internal/models"

	"github.com/stretchr/testify/mock"
)

type MockTransactionService struct {
	mock.Mock
}

// RecordTransaction mocks the RecordTransaction function
func (m *MockTransactionService) RecordTransaction(fromWalletNumber, toWalletNumber *string, transactionType string, amount float64) error {
	args := m.Called(fromWalletNumber, toWalletNumber, transactionType, amount)
	return args.Error(0)
}

// GetTransactionHistory mocks the GetTransactionHistory function
func (m *MockTransactionService) GetTransactionHistory(walletNumber string, orderBy string, limit int) ([]models.TransactionWithEmails, error) {
	args := m.Called(walletNumber, orderBy, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.TransactionWithEmails), args.Error(1)
}
