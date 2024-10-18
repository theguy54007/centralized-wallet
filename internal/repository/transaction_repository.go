package repository

import (
	"centralized-wallet/internal/models"
	"database/sql"
	"time"
)

type TransactionRepository struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

// RecordTransaction records a new transaction in the database
func (r *TransactionRepository) RecordTransaction(userID int, transactionType string, amount float64) error {
	query := `INSERT INTO transactions (user_id, transaction_type, amount, created_at)
			  VALUES ($1, $2, $3, $4)`
	_, err := r.db.Exec(query, userID, transactionType, amount, time.Now())
	return err
}

// GetTransactionHistory fetches the transaction history for a given user
func (r *TransactionRepository) GetTransactionHistory(userID int) ([]models.Transaction, error) {
	query := `SELECT id, transaction_type, amount, created_at FROM transactions WHERE user_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []models.Transaction
	for rows.Next() {
		var transaction models.Transaction
		err := rows.Scan(&transaction.ID, &transaction.Type, &transaction.Amount, &transaction.CreatedAt)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, transaction)
	}

	return transactions, nil
}
