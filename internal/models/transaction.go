package models

import "time"

// Transaction represents a financial transaction
type Transaction struct {
	ID               int       `json:"id"`
	FromWalletNumber *string   `json:"from_wallet_number"`
	ToWalletNumber   *string   `json:"to_wallet_number"`
	Type             string    `json:"type"`
	Amount           float64   `json:"amount"`
	CreatedAt        time.Time `json:"created_at"`
}

type TransactionWithEmails struct {
	Transaction         // Embedding the existing Transaction struct
	FromEmail   *string `json:"from_email"`
	ToEmail     *string `json:"to_email"`
}
