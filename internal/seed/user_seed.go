package seed

import (
	"centralized-wallet/internal/models"
	"centralized-wallet/internal/user"
	"database/sql"
	"fmt"
)

func SeedUser(db *sql.DB, userData *models.User) error {
	hashedPassword, _ := user.HashPassword(userData.Password)
	_, err := db.Exec(`
		INSERT INTO users (
			email,
			password
		) VALUES ($1, $2)`,
		userData.Email,
		hashedPassword,
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
			Email:    "jack@example.com",
			Password: "password1", // In reality, this would be hashed
		},
		{
			ID:       2,
			Email:    "david@example.com",
			Password: "password2",
		},
		{
			ID:       3,
			Email:    "carole@example.com",
			Password: "password3",
		},
	}
}
