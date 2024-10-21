package seed

import (
	"centralized-wallet/internal/models"
	"database/sql"
	"fmt"
)

func SeedTransactions(db *sql.DB, transaction *models.Transaction) error {
	_, err := db.Exec(`
		INSERT INTO transactions (
			transaction_type,
			from_wallet_number,
			to_wallet_number,
			amount,
			created_at
		) VALUES ($1, $2, $3, $4, $5)`,
		transaction.TransactionType,
		transaction.FromWalletNumber,
		transaction.ToWalletNumber,
		transaction.Amount,
		transaction.CreatedAt, // Make sure to add the created_at field
	)

	if err != nil {
		return fmt.Errorf("failed to insert transaction: %v", err)
	}

	return nil
}
