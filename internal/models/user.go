package models

type User struct {
	ID        int    `db:"id" json:"id"`
	Email     string `db:"email" json:"email"`
	Password  string `db:"password" json:"-"`                      // Excluded from JSON responses for security
	CreatedAt string `db:"created_at" json:"created_at,omitempty"` // `omitempty` avoids sending empty values
	UpdatedAt string `db:"updated_at" json:"updated_at,omitempty"`
}
