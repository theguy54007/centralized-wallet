package transaction

import (
	"centralized-wallet/internal/models"
	"centralized-wallet/internal/repository"
)

type TransactionServiceInterface interface {
	RecordTransaction(userID int, transactionType string, amount float64) error
	GetTransactionHistory(userID int) ([]models.Transaction, error)
}

type TransactionService struct {
	repo *repository.TransactionRepository
}

func NewTransactionService(repo *repository.TransactionRepository) *TransactionService {
	return &TransactionService{repo: repo}
}

// RecordTransaction records a transaction
func (ts *TransactionService) RecordTransaction(userID int, transactionType string, amount float64) error {
	return ts.repo.RecordTransaction(userID, transactionType, amount)
}

// GetTransactionHistory retrieves the transaction history for a specific user
func (ts *TransactionService) GetTransactionHistory(userID int) ([]models.Transaction, error) {
	return ts.repo.GetTransactionHistory(userID)
}
