package transaction

import (
	"centralized-wallet/internal/models"
	"centralized-wallet/internal/redis"
	"centralized-wallet/internal/utils"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

type TransactionServiceInterface interface {
	RecordTransaction(tx *sql.Tx, fromWalletNumber *string, toWalletNumber *string, transactionType string, amount float64) error
	GetTransactionHistory(walletNumber string, orderBy string, limit, offset int) ([]models.FormattedTransaction, error)
	FormatTransactionResponse(walletNumber string, transactions []models.TransactionWithEmails) []models.FormattedTransaction
}

type TransactionService struct {
	repo         TransactionRepositoryInterface
	redisService redis.RedisServiceInterface
}

func NewTransactionService(repo TransactionRepositoryInterface, redis redis.RedisServiceInterface) *TransactionService {
	return &TransactionService{
		repo:         repo,
		redisService: redis,
	}
}

// RecordTransaction records a transaction
func (ts *TransactionService) RecordTransaction(tx *sql.Tx, fromWalletNumber, toWalletNumber *string, transactionType string, amount float64) error {
	// Check if both fromWalletNumber and toWalletNumber are nil or empty
	if (fromWalletNumber == nil || *fromWalletNumber == "") && (toWalletNumber == nil || *toWalletNumber == "") {
		return utils.ServiceErrWalletNumberNil
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
	if err := ts.repo.CreateTransaction(tx, &transaction); err != nil {
		return err
	}

	if ts.redisService == nil {
		return nil
	}

	if fromWalletNumber != nil && *fromWalletNumber != "" {
		pageKeyPatternFrom := fmt.Sprintf("user:%s:transactions:page:*", *fromWalletNumber)
		err := ts.InvalidateTransactionCache(pageKeyPatternFrom)
		if err != nil {
			return err
		}
	}

	if toWalletNumber != nil && *toWalletNumber != "" {
		pageKeyPatternTo := fmt.Sprintf("user:%s:transactions:page:*", *toWalletNumber)
		err := ts.InvalidateTransactionCache(pageKeyPatternTo)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetTransactionHistory retrieves the transaction history for a specific wallet number.
func (ts *TransactionService) GetTransactionHistory(walletNumber string, orderBy string, limit, offset int) ([]models.FormattedTransaction, error) {
	pageSize := 30
	pageKey := fmt.Sprintf("user:%s:transactions:page:%d%s", walletNumber, offset/pageSize, orderBy)

	// Check Redis cache first if available
	if ts.redisService != nil {
		cachedTransactions, err := ts.redisService.Get(context.Background(), pageKey)
		if err == nil && cachedTransactions != "" {
			var transactions []models.FormattedTransaction
			err = json.Unmarshal([]byte(cachedTransactions), &transactions)
			if err == nil {
				return ts.paginateTransactions(transactions, limit, offset%pageSize), nil
			}
		}
	}

	// Fetch from the database if not cached
	transactions, err := ts.repo.GetTransactionHistory(walletNumber, orderBy, limit, offset)
	if err != nil {
		return nil, err
	}

	// Format transactions
	formattedTransactions := ts.FormatTransactionResponse(walletNumber, transactions)

	// Cache the formatted transactions if redis available
	if ts.redisService != nil {
		// Cache formatted transactions in Redis
		cacheData, err := json.Marshal(formattedTransactions)
		if err == nil {
			ts.redisService.Set(context.Background(), pageKey, cacheData, 10*time.Minute)
		}
	}

	// Return the formatted transactions
	if len(formattedTransactions) > limit {
		return formattedTransactions[:limit], nil
	}
	return formattedTransactions, nil
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

func (ts *TransactionService) paginateTransactions(transactions []models.FormattedTransaction, limit int, offset int) []models.FormattedTransaction {
	// Check if the offset is within bounds
	if offset >= len(transactions) {
		// If offset is beyond the range of available data, return an empty slice
		return []models.FormattedTransaction{}
	}

	// Determine start and end indices
	start := offset
	end := offset + limit
	if end > len(transactions) {
		end = len(transactions)
	}

	return transactions[start:end]
}

func (ts *TransactionService) InvalidateTransactionCache(keyPattern string) error {
	return ts.redisService.DeleteKeysByPattern(context.Background(), keyPattern)
}
