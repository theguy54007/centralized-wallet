package transaction

import (
	"centralized-wallet/internal/models"
	"centralized-wallet/internal/repository"
	"time"
)

type TransactionServiceInterface interface {
	RecordTransaction(fromUserID, toUserID *int, transactionType string, amount float64) error
	GetTransactionHistory(userID int) ([]models.Transaction, error)
}

type TransactionService struct {
	repo *repository.TransactionRepository
}

func NewTransactionService(repo *repository.TransactionRepository) *TransactionService {
	return &TransactionService{repo: repo}
}

// RecordTransaction records a transaction
func (ts *TransactionService) RecordTransaction(fromUserID, toUserID *int, transactionType string, amount float64) error {
	transaction := models.Transaction{
		FromUserID: fromUserID,
		ToUserID:   toUserID,
		Type:       transactionType,
		Amount:     amount,
		CreatedAt:  time.Now(),
	}
	return ts.repo.CreateTransaction(&transaction)
}

// GetTransactionHistory retrieves the transaction history for a specific user
func (ts *TransactionService) GetTransactionHistory(userID int) ([]models.Transaction, error) {
	return ts.repo.GetTransactionHistory(userID)
}
