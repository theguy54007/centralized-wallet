package models

import "time"

type Wallet struct {
	ID           int       `db:"id" json:"id"`           // Wallet ID
	UserID       int       `db:"user_id" json:"user_id"` // Foreign key to the user
	WalletNumber string    `db:"wallet_number" json:"wallet_number"`
	Balance      float64   `db:"balance" json:"balance"`       // The balance in the wallet
	CreatedAt    time.Time `db:"created_at" json:"created_at"` // Timestamp when the wallet was created
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"` // Timestamp when the wallet was last updated
}
