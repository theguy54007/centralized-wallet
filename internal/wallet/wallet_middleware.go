package wallet

import (
	redisService "centralized-wallet/internal/redis"
	"centralized-wallet/internal/utils"
	"context"
	"fmt"
	"log"
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
			utils.ErrorResponse(c, utils.ErrUnauthorized)
			c.Abort()
			return
		}

		// Fetch the wallet number using the helper function
		walletNumber, err := getWalletNumber(walletService, redisClient, userID.(int))
		if err != nil {
			switch err {
			case utils.RepoErrWalletNotFound:
				utils.ErrorResponse(c, utils.ErrWalletNotFound)
			default:
				utils.ErrorResponse(c, utils.ErrInternalServerError)
			}
			c.Abort()
			return
		}

		// Store the wallet number in the context for further usage
		c.Set("wallet_number", walletNumber)

		// Proceed with the next middleware or handler
		c.Next()
	}
}

func getWalletNumber(walletService WalletServiceInterface, redisClient redisService.RedisServiceInterface, userID int) (string, error) {
	userIDStr := fmt.Sprintf("user:%d:wallet_number", userID)

	// Try to get the wallet number from Redis
	walletNumber, err := redisClient.Get(context.Background(), userIDStr)
	if err == redis.Nil {
		// Fetch wallet number from the database if not found in Redis
		wallet, err := walletService.GetWalletByUserID(userID)
		if err != nil {
			return "", err
		}
		walletNumber = wallet.WalletNumber

		// Cache the wallet number in Redis with an expiration time (e.g., 24 hours)
		if err := redisClient.Set(context.Background(), userIDStr, walletNumber, 24*time.Hour); err != nil {
			log.Printf("Warning: Failed to cache wallet number in Redis: %v", err)
		}
	} else if err != nil {
		// If there's a Redis error, attempt to fetch the wallet number from the DB
		log.Printf("Warning: Redis error, fetching wallet number from DB: %v", err)
		wallet, err := walletService.GetWalletByUserID(userID)
		if err != nil {
			return "", err
		}
		walletNumber = wallet.WalletNumber

		// Optionally try to cache it in Redis again
		if err := redisClient.Set(context.Background(), userIDStr, walletNumber, 24*time.Hour); err != nil {
			log.Printf("Warning: Failed to cache wallet number in Redis after fallback: %v", err)
		}
	}

	return walletNumber, nil
}
