package repository

import (
	"centralized-wallet/internal/models"
	"database/sql"
)

type TransactionRepository struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

// RecordTransaction records a new transaction in the database
func (r *TransactionRepository) CreateTransaction(transaction *models.Transaction) error {
	query := `INSERT INTO transactions (from_user_id, to_user_id, transaction_type, amount, created_at)
			  VALUES ($1, $2, $3, $4, $5)`

	_, err := r.db.Exec(
		query,
		transaction.FromUserID,
		transaction.ToUserID,
		transaction.Type,
		transaction.Amount,
		transaction.CreatedAt,
	)
	return err
}

// GetTransactionHistory fetches the transaction history for a given user
func (tr *TransactionRepository) GetTransactionHistory(userID int) ([]models.Transaction, error) {
	query := `SELECT id, from_user_id, to_user_id, transaction_type, amount, created_at
	          FROM transactions
	          WHERE from_user_id = $1 OR to_user_id = $1
	          ORDER BY created_at DESC`

	rows, err := tr.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []models.Transaction
	for rows.Next() {
		var transaction models.Transaction
		err := rows.Scan(&transaction.ID, &transaction.FromUserID, &transaction.ToUserID, &transaction.Type, &transaction.Amount, &transaction.CreatedAt)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, transaction)
	}

	return transactions, nil
}
