package seed

import (
	"centralized-wallet/internal/models"
	"database/sql"
	"fmt"
)

func SeedUser(db *sql.DB, user *models.User) error {
	_, err := db.Exec(`
		INSERT INTO users (
			email,
			password
		) VALUES ($1, $2)`,
		user.Email,
		user.Password,
	)

	if err != nil {
		return fmt.Errorf("failed to insert wallet: %v", err)
	}

	return nil
}

func GenerateSampleUsers() []models.User {
	return []models.User{
		{
			ID:       1,
			Email:    "alice@example.com",
			Password: "password1", // In reality, this would be hashed
		},
		{
			ID:       2,
			Email:    "bob@example.com",
			Password: "password2",
		},
		{
			ID:       3,
			Email:    "charlie@example.com",
			Password: "password3",
		},
	}
}
