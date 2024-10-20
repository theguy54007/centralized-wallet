package models

type User struct {
	ID        int    `json:"id"`
	Email     string `json:"email"`
	Password  string `json:"-"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}
