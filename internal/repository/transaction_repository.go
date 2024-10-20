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
func (repo *TransactionRepository) GetTransactionHistory(userID int, orderBy string, limit int) ([]models.TransactionWithEmails, error) {
	transactions := []models.TransactionWithEmails{}

	query := `
		SELECT
			t.id,
			t.from_user_id,
			f.email as from_email,
			t.to_user_id,
			tu.email as to_email,
			t.transaction_type,
			t.amount,
			t.created_at
		FROM transactions t
		LEFT JOIN users f ON t.from_user_id = f.id
		LEFT JOIN users tu ON t.to_user_id = tu.id
		WHERE t.from_user_id = $1 OR t.to_user_id = $1
		ORDER BY t.created_at ` + orderBy + `
		LIMIT $2`

	// Execute the query
	rows, err := repo.db.Query(query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Iterate over the rows and scan the data into the struct
	for rows.Next() {
		var transaction models.TransactionWithEmails

		// Scan the row into the TransactionWithEmails struct
		err := rows.Scan(
			&transaction.ID,
			&transaction.FromUserID,
			&transaction.FromEmail,
			&transaction.ToUserID,
			&transaction.ToEmail,
			&transaction.Type,
			&transaction.Amount,
			&transaction.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		transactions = append(transactions, transaction)
	}

	// Check for any error that might have occurred during iteration
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return transactions, nil
}
