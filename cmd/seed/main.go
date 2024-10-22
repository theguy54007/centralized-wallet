package main

import (
	"centralized-wallet/internal/database"
	"centralized-wallet/internal/seed"
	"database/sql"
	"flag"
	"log"
)

func main() {
	truncate := flag.Bool("truncate", false, "Whether to truncate tables before seeding")
	flag.Parse()

	dbService := database.InitDB()
	if *truncate {
		log.Println("Truncating tables before seeding...")
		truncateTables(dbService.GetDB())
	}

	defer dbService.Close()
	log.Println("Seeding users...")
	SeedUsers(dbService)

	log.Println("Seeding completed successfully.")

}

// SeedUsers seeds the users data
func SeedUsers(dbService database.Service) {
	users := seed.GenerateSampleUsers()
	for _, user := range users {
		err := seed.SeedUser(dbService.GetDB(), &user)
		if err != nil {
			log.Fatalf("Failed to seed user: %v", err)
		}
	}
}

func truncateTables(db *sql.DB) {
	db.Exec("TRUNCATE TABLE users RESTART IDENTITY CASCADE")
}
