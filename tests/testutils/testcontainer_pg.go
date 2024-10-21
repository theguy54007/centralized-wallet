package testutils

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/golang-migrate/migrate/v4"
	pgM "github.com/golang-migrate/migrate/v4/database/postgres"
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
	MigrationPath = "../../../migrations" // Path to your migrations
	Schema        = "public"
)

func StartPostgresContainer(applyMigration bool, migrationPath string) (func(context.Context) error, error) {
	// Start the Postgres container with or without init scripts
	dbContainer, err := postgres.Run(
		context.Background(),
		"postgres:latest",
		// postgres.WithInitScripts(initScripts...),
		postgres.WithDatabase(Database),
		postgres.WithUsername(Username),
		postgres.WithPassword(Password),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(10*time.Second),
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

	// Apply migrations if required
	if applyMigration {
		if err := applyMigrations(migrationPath); err != nil {
			return dbContainer.Terminate, fmt.Errorf("failed to apply migrations: %v", err)
		}
	}

	return dbContainer.Terminate, nil
}

// Teardown the container after tests
func TeardownContainer(teardown func(context.Context) error) {
	if teardown != nil && teardown(context.Background()) != nil {
		log.Fatalf("could not teardown postgres container")
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

func applyMigrations(migrationPath string) error {
	// Set up connection string
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable&search_path=%s", Username, Password, Host, Port, Database, Schema)

	// Connect to the database
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return fmt.Errorf("could not connect to database: %v", err)
	}
	defer db.Close()

	// Create migration driver using the DB connection
	driver, err := pgM.WithInstance(db, &pgM.Config{})
	if err != nil {
		return fmt.Errorf("could not create database driver: %v", err)
	}

	// Initialize the migration instance
	m, err := migrate.NewWithDatabaseInstance(
		"file://"+migrationPath, // Migration files location
		"postgres",              // Database name
		driver,
	)
	if err != nil {
		return fmt.Errorf("could not initialize migration: %v", err)
	}

	// Run migrations
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migration failed: %v", err)
	}

	// log.Println("Migrations applied successfully")
	return nil
}

func CleanDatabase(db *sql.DB) error {
	// List all the tables to truncate
	tables := []string{"transactions", "wallets", "users"} // Add your tables here

	// Disable constraints to allow truncation in the right order
	if _, err := db.Exec("SET session_replication_role = 'replica';"); err != nil {
		return fmt.Errorf("could not disable constraints: %v", err)
	}

	// Truncate all tables
	for _, table := range tables {
		if _, err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE;", table)); err != nil {
			return fmt.Errorf("could not truncate table %s: %v", table, err)
		}
	}

	// Re-enable constraints
	if _, err := db.Exec("SET session_replication_role = 'origin';"); err != nil {
		return fmt.Errorf("could not re-enable constraints: %v", err)
	}

	log.Println("Database cleaned successfully")
	return nil
}
