package mock_redis

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"
)

// MockRedisClient is a mock implementation of the Redis client
type MockRedisClient struct {
	mock.Mock
}

// Get mocks the Redis GET command
func (m *MockRedisClient) Get(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

// Set mocks the Redis SET command
func (m *MockRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	args := m.Called(ctx, key, value, expiration)
	return args.Error(0)
}

// Other Redis functions can be added if needed for your tests
