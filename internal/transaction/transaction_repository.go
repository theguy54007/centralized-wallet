package transaction

import (
	"centralized-wallet/internal/models"
	"database/sql"
)

type TransactionRepositoryInterface interface {
	CreateTransaction(transaction *models.Transaction) error
	GetTransactionHistory(walletNumber string, orderBy string, limit, offset int) ([]models.TransactionWithEmails, error)
}
type TransactionRepository struct {
	db *sql.DB
}

// Ensure WalletRepository implements WalletRepositoryInterface
var _ TransactionRepositoryInterface = &TransactionRepository{}

func NewTransactionRepository(db *sql.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

// CreateTransaction inserts a new transaction with wallet numbers.
func (r *TransactionRepository) CreateTransaction(transaction *models.Transaction) error {
	query := `INSERT INTO transactions (from_wallet_number, to_wallet_number, transaction_type, amount, created_at)
			  VALUES ($1, $2, $3, $4, $5)`

	_, err := r.db.Exec(
		query,
		transaction.FromWalletNumber,
		transaction.ToWalletNumber,
		transaction.Type,
		transaction.Amount,
		transaction.CreatedAt,
	)
	return err
}

// GetTransactionHistory fetches the transaction history for a given wallet number.
func (repo *TransactionRepository) GetTransactionHistory(walletNumber string, orderBy string, limit, offset int) ([]models.TransactionWithEmails, error) {
	transactions := []models.TransactionWithEmails{}

	query := `
		SELECT
			t.id,
			uf.email as from_email,
			t.from_wallet_number,
			tu.email as to_email,
			t.to_wallet_number,
			t.transaction_type,
			t.amount,
			t.created_at
		FROM transactions t
		LEFT JOIN wallets wf ON t.from_wallet_number = wf.wallet_number
		LEFT JOIN users uf ON wf.user_id = uf.id
		LEFT JOIN wallets wtu ON t.to_wallet_number = wtu.wallet_number
		LEFT JOIN users tu ON wtu.user_id = tu.id
		WHERE t.from_wallet_number = $1 OR t.to_wallet_number = $1
		ORDER BY t.created_at ` + orderBy + `
		LIMIT $2
		OFFSET $3`

	// Execute the query
	rows, err := repo.db.Query(query, walletNumber, limit, offset)
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
			&transaction.FromEmail,
			&transaction.FromWalletNumber,
			&transaction.ToEmail,
			&transaction.ToWalletNumber,
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
