package main

import (
	"centralized-wallet/internal/database"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	// Initialize the database service from database.go
	dbService := database.New()
	dbConn := dbService.GetDB()

	// Ensure database connection is healthy
	if dbConn == nil {
		log.Fatal("Failed to connect to the database")
	}

	// Create a migration driver using the existing DB connection
	driver, err := postgres.WithInstance(dbConn, &postgres.Config{})
	if err != nil {
		log.Fatalf("Could not create database driver: %v", err)
	}

	// Set up the migration instance with the file source and database driver
	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations", // Directory for migration files
		"postgres",          // Database name
		driver,
	)
	if err != nil {
		log.Fatalf("Could not initialize migration: %v", err)
	}

	// Apply the migrations
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Migration failed: %v", err)
	}

	log.Println("Migrations applied successfully")

	// Close the database connection after migration
	if err := dbService.Close(); err != nil {
		log.Fatalf("Could not close database connection: %v", err)
	}
}
