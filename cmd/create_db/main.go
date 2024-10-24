package main

import (
	"centralized-wallet/internal/database"
	"database/sql"
	"fmt"
	"log"
	"os"
)

func main() {
	// Initialize the DB connection from database.go, connect to "postgres" to check/create target database
	dbService := database.New(false) // false indicates we are not connecting to the target DB yet
	defer closeDB(dbService)

	// Get the database name from environment variables
	dbName := os.Getenv("DB_DATABASE")

	// Check if the database exists
	if !databaseExists(dbService.GetDB(), dbName) {
		log.Printf("Database %s does not exist. Creating the database...", dbName)
		if err := createDatabase(dbName); err != nil {
			log.Fatalf("Could not create database: %v", err)
		}
		log.Println("Database created successfully.")
	} else {
		log.Printf("Database %s already exists.", dbName)
	}

	// Now connect to the newly created target database
	dbService = database.New(true) // true indicates to connect to the actual target database
	defer closeDB(dbService)
}

// closeDB closes the database connection
func closeDB(dbService database.Service) {
	if err := dbService.Close(); err != nil {
		log.Fatalf("Could not close database connection: %v", err)
	}
}

// databaseExists checks if the database exists
func databaseExists(dbConn *sql.DB, dbName string) bool {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)"
	err := dbConn.QueryRow(query, dbName).Scan(&exists)
	if err != nil {
		log.Fatalf("Could not check if the database exists: %v", err)
	}
	return exists
}

// createDatabase creates the database if it does not exist
func createDatabase(dbName string) error {
	// Get the connection details from environment variables
	dbHost := os.Getenv("DB_HOST")
	dbUser := os.Getenv("DB_USERNAME")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbPort := os.Getenv("DB_PORT")

	// Connect to the 'postgres' database
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/postgres?sslmode=disable", dbUser, dbPassword, dbHost, dbPort)
	conn, err := sql.Open("pgx", connStr)
	if err != nil {
		return err
	}
	defer conn.Close()

	// Execute the create database query
	_, err = conn.Exec("CREATE DATABASE " + dbName)
	if err != nil {
		return err
	}

	return nil
}
