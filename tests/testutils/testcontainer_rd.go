package testutils

import (
	"context"
	"os"

	testRedis "github.com/testcontainers/testcontainers-go/modules/redis"
)

var (
	RD_Address string
	RD_Port    string

	RD_Password = "password"
	RD_Database = "0"
)

func StartRedisContainer() (func(context.Context) error, error) {
	dbContainer, err := testRedis.Run(
		context.Background(),
		"docker.io/redis:7.2.4",
		testRedis.WithSnapshotting(10, 1),
		testRedis.WithLogLevel(testRedis.LogLevelVerbose),
	)
	if err != nil {
		return nil, err
	}

	// Get the container host and port for Redis
	dbHost, err := dbContainer.Host(context.Background())
	if err != nil {
		return dbContainer.Terminate, err
	}

	dbPort, err := dbContainer.MappedPort(context.Background(), "6379/tcp")
	if err != nil {
		return dbContainer.Terminate, err
	}

	// Set the address for Redis connection
	RD_Address = dbHost
	RD_Port = dbPort.Port()

	// Prepare Redis service configuration
	return dbContainer.Terminate, err
}

func InitRDEnv() {
	os.Setenv("REDIS_ADDRESS", RD_Address)
	os.Setenv("REDIS_PORT", RD_Port)
	os.Setenv("REDIS_PASSWORD", RD_Password)
	os.Setenv("REDIS_DATABASE", RD_Database)
}
