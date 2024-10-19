package models

import "time"

// Transaction represents a financial transaction
type Transaction struct {
	ID         int       `json:"id"`
	FromUserID *int      `json:"from_user_id"`
	ToUserID   *int      `json:"to_user_id"`
	Type       string    `json:"type"`
	Amount     float64   `json:"amount"`
	CreatedAt  time.Time `json:"created_at"`
}

type TransactionWithEmails struct {
	Transaction         // Embedding the existing Transaction struct
	FromEmail   *string `json:"from_email"`
	ToEmail     *string `json:"to_email"`
}
