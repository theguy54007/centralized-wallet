package auth

import (
	"context"
	"fmt"
	"time"

	redisSerivice "centralized-wallet/internal/redis"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

// BlacklistServiceInterface is an interface for the BlacklistService
type BlacklistServiceInterface interface {
	BlacklistToken(tokenString string, token *jwt.Token) error
	IsTokenBlacklisted(tokenString string) (bool, error)
	RemoveBlacklistedToken(ctx context.Context, tokenString string) error
}

// BlacklistService is a service for blacklisting JWT tokens
type BlacklistService struct {
	redis *redisSerivice.RedisService
}

// NewBlacklistService initializes a new BlacklistService
func NewBlacklistService(redis *redisSerivice.RedisService) *BlacklistService {
	return &BlacklistService{
		redis: redis,
	}
}

// BlacklistToken adds the given JWT token to the blacklist with its expiration time
// BlacklistToken blacklists a token by adding it to Redis with its expiration time
func (b *BlacklistService) BlacklistToken(tokenString string, token *jwt.Token) error {
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return jwt.ErrInvalidKey
	}

	// Extract the expiration time from the JWT token
	exp, ok := claims["exp"].(float64)
	if !ok {
		return jwt.ErrTokenExpired
	}
	expiration := time.Unix(int64(exp), 0)

	// Store the token in Redis with its expiration time as TTL
	ctx := context.Background()
	err := b.redis.Client.Set(ctx, tokenString, "blacklisted", time.Until(expiration)).Err()
	if err != nil {
		return err
	}

	return nil
}

// IsTokenBlacklisted checks if a token is present in the blacklist
func (b *BlacklistService) IsTokenBlacklisted(tokenString string) (bool, error) {
	ctx := context.Background()

	// Check if the token exists in Redis
	_, err := b.redis.Client.Get(ctx, tokenString).Result()
	if err == redis.Nil {
		// Token is not blacklisted
		return false, nil
	} else if err != nil {
		return false, err
	}

	// Token is blacklisted
	return true, nil
}

// RemoveBlacklistedToken removes a token from the blacklist (optional)
func (b *BlacklistService) RemoveBlacklistedToken(ctx context.Context, tokenString string) error {
	err := b.redis.Client.Del(ctx, tokenString).Err()
	if err != nil {
		return fmt.Errorf("could not remove blacklisted token: %w", err)
	}
	return nil
}
