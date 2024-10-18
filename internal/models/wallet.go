package models

import "time"

type Wallet struct {
	ID        int       `db:"id"`         // Wallet ID
	UserID    int       `db:"user_id"`    // Foreign key to the user
	Balance   float64   `db:"balance"`    // The balance in the wallet
	CreatedAt time.Time `db:"created_at"` // Timestamp when the wallet was created
	UpdatedAt time.Time `db:"updated_at"` // Timestamp when the wallet was last updated
}
