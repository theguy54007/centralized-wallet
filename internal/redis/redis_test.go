package redis

import (
	"centralized-wallet/tests/testutils"
	"context"
	"log"
	"testing"
)

var redisService *RedisService

// Test setup: Start the Redis container and initialize the Redis service.
func TestMain(m *testing.M) {
	// Start the Redis container
	// teardown, err := mustStartRedisContainer()
	teardown, err := testutils.StartRedisContainer()
	if err != nil {
		log.Fatalf("could not start redis container: %v", err)
	}

	testutils.InitRDEnv()
	// Run the test suite
	m.Run()

	// Teardown the Redis container after tests
	if teardown != nil && teardown(context.Background()) != nil {
		log.Fatalf("could not teardown redis container: %v", err)
	}
}

// TestNew checks if the Redis service is initialized correctly.
func TestNew(t *testing.T) {
	srv := NewRedisService()
	if srv == nil {
		t.Fatal("NewRedisService() returned nil")
	}
}

// TestHealth checks the health of the Redis service.
func TestHealth(t *testing.T) {
	srv := NewRedisService()

	// Check the health of the Redis connection
	stats := srv.Health(context.Background())

	// Ensure the Redis server is running
	if stats["redis_status"] != "up" {
		t.Fatalf("expected redis_status to be 'up', got '%s'", stats["redis_status"])
	}

	// Check if Redis version information is returned
	if _, ok := stats["redis_version"]; !ok {
		t.Fatalf("expected redis_version to be present, got %v", stats["redis_version"])
	}
}
