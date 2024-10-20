package wallet

import (
	"centralized-wallet/internal/models"
	redisService "centralized-wallet/internal/redis"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// WalletNumberMiddleware fetches the wallet number for the user and adds it to the context, using Redis for caching.
func WalletNumberMiddleware(walletService WalletServiceInterface, redisClient redisService.RedisServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the user ID from the context (set by JWT middleware)
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"status": "error", "error": "User ID not found"})
			c.Abort()
			return
		}

		userIDStr := fmt.Sprintf("user:%d:wallet_number", userID)

		// Try to get the wallet number from Redis
		walletNumber, err := redisClient.Get(context.Background(), userIDStr)
		if err == redis.Nil { // If the wallet number is not in Redis, fetch it from the database
			var wallet *models.Wallet
			wallet, err = walletService.GetWalletByUserID(userID.(int))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": "Failed to fetch wallet number"})
				c.Abort()
				return
			}
			walletNumber = wallet.WalletNumber

			// Cache the wallet number in Redis with an expiration time (e.g., 24 hours)
			err = redisClient.Set(context.Background(), userIDStr, walletNumber, 24*time.Hour)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": "Failed to cache wallet number"})
				c.Abort()
				return
			}
		} else if err != nil { // Handle any Redis errors
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": "Redis error"})
			c.Abort()
			return
		}

		// Store the wallet number in the context for further usage
		c.Set("wallet_number", walletNumber)

		// Proceed with the next middleware or handler
		c.Next()
	}
}
