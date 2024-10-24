package mock_transaction

import (
	"centralized-wallet/internal/models"
	"database/sql"

	"github.com/stretchr/testify/mock"
)

type MockTransactionRepository struct {
	mock.Mock
}

// Mock CreateTransaction method
func (m *MockTransactionRepository) CreateTransaction(tx *sql.Tx, transaction *models.Transaction) error {
	args := m.Called(transaction)
	return args.Error(0)
}

// Mock GetTransactionHistory method
func (m *MockTransactionRepository) GetTransactionHistory(walletNumber string, orderBy string, limit, offset int) ([]models.TransactionWithEmails, error) {
	args := m.Called(walletNumber, orderBy, limit)
	return args.Get(0).([]models.TransactionWithEmails), args.Error(1)
}
