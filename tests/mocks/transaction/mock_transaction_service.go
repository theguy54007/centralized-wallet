package transaction

import (
	"centralized-wallet/internal/models"

	"github.com/stretchr/testify/mock"
)

type MockTransactionService struct {
	mock.Mock
}

func (m *MockTransactionService) RecordTransaction(fromID, ToUserID *int, transactionType string, amount float64) error {
	args := m.Called(fromID, ToUserID, transactionType, amount)
	return args.Error(0)
}

func (m *MockTransactionService) GetTransactionHistory(userID int) ([]models.Transaction, error) {
	args := m.Called(userID)
	return args.Get(0).([]models.Transaction), args.Error(1)
}
