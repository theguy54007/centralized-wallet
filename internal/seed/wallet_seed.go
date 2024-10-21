package seed

import (
	"centralized-wallet/internal/models"
	"database/sql"
	"fmt"
)

func SeedWallets(db *sql.DB, wallet *models.Wallet) error {
	_, err := db.Exec(`
		INSERT INTO wallets (
			user_id,
			wallet_number,
			balance
		) VALUES ($1, $2, $3)`,
		wallet.UserID,
		wallet.WalletNumber,
		wallet.Balance,
	)

	if err != nil {
		return fmt.Errorf("failed to insert wallet: %v", err)
	}

	return nil
}

func GenerateSampleWallets() []models.Wallet {
	// now := time.Now()

	return []models.Wallet{
		{
			ID:           1,
			UserID:       1, // Belongs to Alice
			WalletNumber: "wallet123",
			Balance:      100.00, // Balance of 100
		},
		{
			ID:           2,
			UserID:       2, // Belongs to Bob
			WalletNumber: "wallet456",
			Balance:      200.00, // Balance of 200
		},
		{
			ID:           3,
			UserID:       3, // Belongs to Charlie
			WalletNumber: "wallet789",
			Balance:      300.00, // Balance of 300
		},
	}
}
