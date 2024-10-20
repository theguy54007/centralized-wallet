package redis

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisService struct {
	Client *redis.Client
}

type RedisServiceInterface interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
}

var (
	address  string
	port     string
	password string
	database string
)

// NewRedisService initializes the Redis client with configuration from the environment.
func NewRedisService() *RedisService {
	initEnv()
	// Parse the Redis database number.
	dbNum, err := strconv.Atoi(database)
	if err != nil {
		log.Fatalf("Invalid Redis database number: %v", err)
	}

	// Construct the full address.
	fullAddress := fmt.Sprintf("%s:%s", address, port)

	// Initialize Redis client.
	rdb := redis.NewClient(&redis.Options{
		Addr:     fullAddress,
		Password: password,
		DB:       dbNum,
	})

	// Test Redis connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := rdb.Ping(ctx).Result(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	return &RedisService{
		Client: rdb,
	}
}

// Health checks the health status of Redis.
func (r *RedisService) Health(ctx context.Context) map[string]string {
	stats := make(map[string]string)

	// Ping the Redis server.
	pong, err := r.Client.Ping(ctx).Result()
	if err != nil {
		log.Printf("Redis is down: %v", err)
		stats["redis_status"] = "down"
		stats["redis_error"] = err.Error()
	} else {
		stats["redis_status"] = "up"
		stats["redis_message"] = "It's healthy"
		stats["redis_ping_response"] = pong
	}

	// Get Redis info and pool stats.
	info, err := r.Client.Info(ctx).Result()
	if err != nil {
		stats["redis_message"] = fmt.Sprintf("Failed to retrieve Redis info: %v", err)
		return stats
	}

	poolStats := r.Client.PoolStats()

	// Parse Redis info.
	redisInfo := parseRedisInfo(info)
	stats["redis_version"] = redisInfo["redis_version"]
	stats["redis_connected_clients"] = redisInfo["connected_clients"]
	stats["redis_used_memory"] = redisInfo["used_memory"]
	stats["redis_pool_size_percentage"] = fmt.Sprintf("%.2f%%", calculatePoolUtilization(poolStats))

	return stats
}

// Implement the Get method
func (r *RedisService) Get(ctx context.Context, key string) (string, error) {
	return r.Client.Get(ctx, key).Result()
}

// Implement the Set method
func (r *RedisService) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.Client.Set(ctx, key, value, expiration).Err()
}

// Utility function to calculate pool utilization as a percentage.
func calculatePoolUtilization(poolStats *redis.PoolStats) float64 {
	if poolStats.TotalConns == 0 {
		return 0
	}
	return float64(poolStats.TotalConns-poolStats.IdleConns) / float64(poolStats.TotalConns) * 100
}

// parseRedisInfo parses the Redis info response into a key-value map.
func parseRedisInfo(info string) map[string]string {
	result := make(map[string]string)
	lines := strings.Split(info, "\r\n")
	for _, line := range lines {
		if strings.Contains(line, ":") {
			parts := strings.Split(line, ":")
			result[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	return result
}

func initEnv() {
	address = os.Getenv("REDIS_ADDRESS")
	port = os.Getenv("REDIS_PORT")
	password = os.Getenv("REDIS_PASSWORD")
	database = os.Getenv("REDIS_DATABASE")
}
