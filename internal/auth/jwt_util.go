package auth

import (
	"github.com/golang-jwt/jwt/v5"
	// "log"
	"os"
	"time"
)

// var jwtSecret []byte
var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

// GenerateJWT generates a new JWT token for a user
func GenerateJWT(userID int) (string, error) {
	// Set token claims
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 72).Unix(), // Token expiration (e.g., 72 hours)
	}

	// Create the token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the secret key
	return token.SignedString(jwtSecret)
}

// ValidateJWT validates the given token string
// func ValidateJWT(tokenString string) (*jwt.Token, error) {
// 	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
// 		// Ensure the token's signing method is HMAC (HS256)
// 		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
// 			return nil, jwt.ErrSignatureInvalid
// 		}
// 		return jwtSecret, nil
// 	})
// }

func ValidateJWT(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Ensure the token's signing method is HMAC (HS256)
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return jwtSecret, nil
	})
}

// A helper function to check the token's validity and claims.
func CheckTokenClaims(token *jwt.Token) error {
	// Extract claims and verify token validity
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// Validate expiration
		if exp, ok := claims["exp"].(float64); ok {
			if time.Unix(int64(exp), 0).Before(time.Now()) {
				return jwt.ErrTokenExpired
			}
		} else {
			return jwt.ErrTokenExpired
		}
		return nil
	}
	return jwt.ErrSignatureInvalid
}
