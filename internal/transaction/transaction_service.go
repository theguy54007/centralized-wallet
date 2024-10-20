package transaction

import (
	"centralized-wallet/internal/models"
	"centralized-wallet/internal/repository"
	"fmt"
	"time"
)

type TransactionServiceInterface interface {
	RecordTransaction(fromWalletNumber, toWalletNumber *string, transactionType string, amount float64) error
	GetTransactionHistory(walletNumber string, orderBy string, limit int) ([]models.TransactionWithEmails, error)
}

type TransactionService struct {
	repo *repository.TransactionRepository
}

func NewTransactionService(repo *repository.TransactionRepository) *TransactionService {
	return &TransactionService{repo: repo}
}

// RecordTransaction records a transaction
func (ts *TransactionService) RecordTransaction(fromWalletNumber, toWalletNumber *string, transactionType string, amount float64) error {
	// Check if both fromWalletNumber and toWalletNumber are nil or empty
	if (fromWalletNumber == nil || *fromWalletNumber == "") && (toWalletNumber == nil || *toWalletNumber == "") {
		return fmt.Errorf("either fromWalletNumber or toWalletNumber must be provided")
	}

	// Create the transaction struct
	transaction := models.Transaction{
		FromWalletNumber: fromWalletNumber,
		ToWalletNumber:   toWalletNumber,
		Type:             transactionType,
		Amount:           amount,
		CreatedAt:        time.Now(),
	}

	// Save the transaction using the repository
	return ts.repo.CreateTransaction(&transaction)
}

// GetTransactionHistory retrieves the transaction history for a specific wallet number.
func (ts *TransactionService) GetTransactionHistory(walletNumber string, orderBy string, limit int) ([]models.TransactionWithEmails, error) {
	return ts.repo.GetTransactionHistory(walletNumber, orderBy, limit)
}
