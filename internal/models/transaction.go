package models

import "time"

type Transaction struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Type      string    `json:"transaction_type"`
	Amount    float64   `json:"amount"`
	CreatedAt time.Time `json:"created_at"`
}
