package testutils

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	// "github.com/golang-migrate/migrate/v4/database/pgx"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	Host          string
	Port          string
	Username      = "user" // Make sure these values are set
	Password      = "passwordhaha"
	Database      = "test_db"
	MigrationPath = "../../migrations" // Path to your migrations
	Schema        = "public"
)

func StartPostgresContainer(applyMigration bool) (func(context.Context) error, error) {
	// If migration is enabled, provide the path to the migration files
	var initScripts []string
	if applyMigration {
		// Find migration files in the migrations directory
		migrationFiles, err := filepath.Glob(filepath.Join(MigrationPath, "*.sql"))
		if err != nil {
			log.Fatalf("Error finding migration files: %v", err)
		}
		initScripts = migrationFiles
	}

	// Start the Postgres container with or without init scripts
	dbContainer, err := postgres.Run(
		context.Background(),
		"postgres:latest",
		postgres.WithDatabase(Database),
		postgres.WithUsername(Username),
		postgres.WithPassword(Password),
		// Apply init scripts if migration is enabled
		postgres.WithInitScripts(initScripts...),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second),
		),
	)
	if err != nil {
		return nil, err
	}

	// Get the database host and port
	dbHost, err := dbContainer.Host(context.Background())
	if err != nil {
		return dbContainer.Terminate, err
	}

	dbPort, err := dbContainer.MappedPort(context.Background(), "5432/tcp")
	if err != nil {
		return dbContainer.Terminate, err
	}

	Host = dbHost
	Port = dbPort.Port()

	log.Printf("Postgres container started at %s:%s", Host, Port)

	return dbContainer.Terminate, nil
}

// Teardown the container after tests
func TeardownContainer(teardown func(context.Context) error) {
	if teardown != nil && teardown(context.Background()) != nil {
		log.Fatalf("could not teardown postgres container")
	}
}

func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		// Check if we're in the project root by looking for the "migrations" directory
		if _, err := os.Stat(filepath.Join(dir, "migrations")); !os.IsNotExist(err) {
			return dir, nil
		}

		// Move one level up in the directory tree
		parent := filepath.Dir(dir)
		if parent == dir {
			// We've reached the root of the file system without finding the project root
			return "", fmt.Errorf("could not find project root")
		}

		dir = parent
	}
}

func InitEnv() {
	os.Setenv("DB_HOST", Host)
	os.Setenv("DB_PORT", Port)
	os.Setenv("DB_USERNAME", Username)
	os.Setenv("DB_PASSWORD", Password)
	os.Setenv("DB_DATABASE", Database)
	os.Setenv("DB_SCHEMA", Schema)
}
