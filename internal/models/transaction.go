package models

import "time"

// Transaction represents a financial transaction
type Transaction struct {
	ID               int       `db:"id" json:"id"`
	FromWalletNumber *string   `db:"from_wallet_number" json:"from_wallet_number"` // Nullable field, so it's a pointer
	ToWalletNumber   *string   `db:"to_wallet_number" json:"to_wallet_number"`     // Nullable field, so it's a pointer
	TransactionType  string    `db:"transaction_type" json:"transaction_type"`
	Amount           float64   `db:"amount" json:"amount"`
	CreatedAt        time.Time `db:"created_at" json:"created_at"`
}

type TransactionWithEmails struct {
	Transaction         // Embedding the existing Transaction struct
	FromEmail   *string `json:"from_email"`
	ToEmail     *string `json:"to_email"`
}

type FormattedTransaction struct {
	TransactionType  string  `json:"transaction_type"`
	Amount           float64 `json:"amount"`
	Direction        string  `json:"direction"`
	FromWalletNumber string  `json:"from_wallet_number,omitempty"`
	FromEmail        string  `json:"from_email,omitempty"`
	ToWalletNumber   string  `json:"to_wallet_number,omitempty"`
	ToEmail          string  `json:"to_email,omitempty"`
}
