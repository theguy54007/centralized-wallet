package redis

import (
	"centralized-wallet/tests/testutils"
	"context"
	"log"
	"testing"
)

var redisService *RedisService

// Start Redis container
// func mustStartRedisContainer() (func(context.Context) error, error) {
// 	dbContainer, err := testRedis.Run(
// 		context.Background(),
// 		"docker.io/redis:7.2.4",
// 		testRedis.WithSnapshotting(10, 1),
// 		testRedis.WithLogLevel(testRedis.LogLevelVerbose),
// 	)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Get the container host and port for Redis
// 	dbHost, err := dbContainer.Host(context.Background())
// 	if err != nil {
// 		return dbContainer.Terminate, err
// 	}

// 	dbPort, err := dbContainer.MappedPort(context.Background(), "6379/tcp")
// 	if err != nil {
// 		return dbContainer.Terminate, err
// 	}

// 	// Set the address for Redis connection
// 	address = dbHost
// 	port = dbPort.Port()

// 	// Prepare Redis service configuration
// 	redisService = NewRedisService()
// 	return dbContainer.Terminate, err
// }

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
