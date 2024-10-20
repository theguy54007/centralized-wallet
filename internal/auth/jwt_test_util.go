package auth

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TestHelperGenerateJWT creates a JWT token for testing purposes
func TestHelperGenerateJWT(userID int) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 72).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}
