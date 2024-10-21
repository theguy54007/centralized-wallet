package transaction

import (
	"centralized-wallet/internal/models"
	"database/sql"
	"fmt"
	"time"
)

type TransactionServiceInterface interface {
	RecordTransaction(tx *sql.Tx, fromWalletNumber *string, toWalletNumber *string, transactionType string, amount float64) error
	GetTransactionHistory(walletNumber string, orderBy string, limit, offset int) ([]models.FormattedTransaction, error)
	FormatTransactionResponse(walletNumber string, transactions []models.TransactionWithEmails) []models.FormattedTransaction
}

type TransactionService struct {
	repo TransactionRepositoryInterface
}

func NewTransactionService(repo TransactionRepositoryInterface) *TransactionService {
	return &TransactionService{repo: repo}
}

// RecordTransaction records a transaction
func (ts *TransactionService) RecordTransaction(tx *sql.Tx, fromWalletNumber, toWalletNumber *string, transactionType string, amount float64) error {
	// Check if both fromWalletNumber and toWalletNumber are nil or empty
	if (fromWalletNumber == nil || *fromWalletNumber == "") && (toWalletNumber == nil || *toWalletNumber == "") {
		return fmt.Errorf("either fromWalletNumber or toWalletNumber must be provided")
	}

	// Create the transaction struct
	transaction := models.Transaction{
		FromWalletNumber: fromWalletNumber,
		ToWalletNumber:   toWalletNumber,
		TransactionType:  transactionType,
		Amount:           amount,
		CreatedAt:        time.Now(),
	}

	// Save the transaction using the repository
	return ts.repo.CreateTransaction(tx, &transaction)
}

// GetTransactionHistory retrieves the transaction history for a specific wallet number.
func (ts *TransactionService) GetTransactionHistory(walletNumber string, orderBy string, limit, offset int) ([]models.FormattedTransaction, error) {

	transactions, err := ts.repo.GetTransactionHistory(walletNumber, orderBy, limit, offset)
	if err != nil {
		return nil, err
	}

	return ts.FormatTransactionResponse(walletNumber, transactions), nil
}

// transaction_service.go
func (ts *TransactionService) FormatTransactionResponse(walletNumber string, transactions []models.TransactionWithEmails) []models.FormattedTransaction {
	formattedTransactions := []models.FormattedTransaction{}

	for _, tx := range transactions {
		var formattedTx models.FormattedTransaction
		formattedTx.TransactionType = tx.TransactionType
		formattedTx.Amount = tx.Amount

		// Check direction based on the user's wallet number and the presence of from/to wallet numbers
		if tx.FromWalletNumber != nil && *tx.FromWalletNumber == walletNumber {
			// Outgoing transaction
			formattedTx.Direction = "outgoing"
			if tx.ToWalletNumber != nil {
				formattedTx.ToWalletNumber = *tx.ToWalletNumber
			}
			if tx.ToEmail != nil {
				formattedTx.ToEmail = *tx.ToEmail
			}
		} else if tx.ToWalletNumber != nil && *tx.ToWalletNumber == walletNumber {
			// Incoming transaction
			formattedTx.Direction = "incoming"
			if tx.FromWalletNumber != nil {
				formattedTx.FromWalletNumber = *tx.FromWalletNumber
			}
			if tx.FromEmail != nil {
				formattedTx.FromEmail = *tx.FromEmail
			}
		}

		formattedTransactions = append(formattedTransactions, formattedTx)
	}

	return formattedTransactions
}
