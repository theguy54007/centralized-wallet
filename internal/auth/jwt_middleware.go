package auth

import (
	"errors"

	"centralized-wallet/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func JWTMiddleware(blacklistService BlacklistServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			utils.ErrorResponse(c, utils.ErrUnauthorized, nil, "")
			c.Abort()
			return
		}

		// Ensure the token has "Bearer" prefix
		if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
			tokenString = tokenString[7:]
		} else {
			utils.ErrorResponse(c, utils.ErrInvalidAuthorization, nil, "")
			c.Abort()
			return
		}

		// Check if the token is blacklisted
		isBlacklisted, err := blacklistService.IsTokenBlacklisted(tokenString)
		if err != nil {
			utils.ErrorResponse(c, utils.ErrInternalServerError, err, "[JWTMiddleware] Error checking if token is blacklisted")
			c.Abort()
			return
		}
		if isBlacklisted {
			utils.ErrorResponse(c, utils.ErrInvalidToken, nil, "") // Or custom error for blacklisted token
			c.Abort()
			return
		}

		// Validate the JWT token
		token, err := ValidateJWT(tokenString)
		if err != nil {
			if errors.Is(err, jwt.ErrTokenExpired) {
				utils.ErrorResponse(c, utils.ErrTokenExpired, nil, "")
			} else {
				utils.ErrorResponse(c, utils.ErrInvalidToken, nil, "")
			}
			c.Abort()
			return
		}

		// Add token info to context
		c.Set("token", token)
		c.Set("token_string", tokenString)

		// Add the user ID to the context
		claims, ok := token.Claims.(jwt.MapClaims)
		if ok && token.Valid {
			c.Set("user_id", int(claims["user_id"].(float64)))
		} else {
			utils.ErrorResponse(c, utils.ErrInvalidToken, nil, "")
			c.Abort()
			return
		}

		// Continue to next handler
		c.Next()
	}
}
