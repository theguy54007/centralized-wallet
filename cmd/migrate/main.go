package main

import (
	"centralized-wallet/internal/database"
	"database/sql"
	"flag"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	// Parse command-line flags
	migrateDirection, steps, forceVersion, showVersion := parseFlags()

	// Initialize the database connection
	dbService := initDB()
	defer closeDB(dbService)

	// Create the migration instance
	m := initMigrate(dbService.GetDB())

	// Perform migration actions based on flags
	handleMigration(m, migrateDirection, steps, forceVersion, showVersion)
}

// parseFlags handles the parsing of command-line flags
func parseFlags() (migrateDirection *string, steps *int, forceVersion *int, showVersion *bool) {
	migrateDirection = flag.String("direction", "up", "Specify migration direction: 'up' or 'down'")
	steps = flag.Int("steps", 1, "Specify the number of steps to roll back (applies only to 'down')")
	forceVersion = flag.Int("force", -1, "Force the database to a specific version without running the migration")
	showVersion = flag.Bool("version", false, "Show the current migration version and dirty state")
	flag.Parse()
	return
}

// initDB initializes the database service and connection
func initDB() database.Service {
	dbService := database.New()

	if dbService.GetDB() == nil {
		log.Fatal("Failed to connect to the database")
	}

	return dbService
}

// closeDB closes the database connection
func closeDB(dbService database.Service) {
	if err := dbService.Close(); err != nil {
		log.Fatalf("Could not close database connection: %v", err)
	}
}

// initMigrate creates and returns a new migration instance
func initMigrate(dbConn *sql.DB) *migrate.Migrate {
	driver, err := postgres.WithInstance(dbConn, &postgres.Config{})
	if err != nil {
		log.Fatalf("Could not create database driver: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations", // Directory for migration files
		"postgres",          // Database name
		driver,
	)
	if err != nil {
		log.Fatalf("Could not initialize migration: %v", err)
	}

	return m
}

// handleMigration handles the migration logic based on flags
func handleMigration(m *migrate.Migrate, migrateDirection *string, steps, forceVersion *int, showVersion *bool) {
	if *forceVersion >= 0 {
		forceMigrationVersion(m, *forceVersion)
		return
	}

	if *showVersion {
		showMigrationVersion(m)
		return
	}

	switch *migrateDirection {
	case "up":
		applyMigrationsUp(m)
	case "down":
		rollbackMigrationsDown(m, *steps)
	default:
		log.Fatalf("Invalid migration direction: %s. Use 'up' or 'down'.", *migrateDirection)
	}
}

// forceMigrationVersion forces the database to a specific version
func forceMigrationVersion(m *migrate.Migrate, version int) {
	if err := m.Force(version); err != nil {
		log.Fatalf("Could not force the migration version: %v", err)
	}
	log.Printf("Forced database to version: %d", version)
}

// showMigrationVersion shows the current migration version and dirty state
func showMigrationVersion(m *migrate.Migrate) {
	version, dirty, err := m.Version()
	if err != nil {
		log.Fatalf("Could not get migration version: %v", err)
	}

	dirtyState := "clean"
	if dirty {
		dirtyState = "dirty"
	}
	log.Printf("Current migration version: %d, state: %s", version, dirtyState)
}

// applyMigrationsUp applies migrations upwards
func applyMigrationsUp(m *migrate.Migrate) {
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Migration up failed: %v", err)
	}
	log.Println("Migrations applied successfully")
}

// rollbackMigrationsDown rolls back migrations by a specific number of steps
func rollbackMigrationsDown(m *migrate.Migrate, steps int) {
	for i := 0; i < steps; i++ {
		if err := m.Steps(-1); err != nil {
			if err == migrate.ErrNoChange {
				log.Println("No more migrations to roll back.")
				break
			}
			log.Fatalf("Migration down failed: %v", err)
		}
		log.Printf("Step %d rolled back successfully", i+1)
	}
}
