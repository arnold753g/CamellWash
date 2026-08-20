package main

import (
	"log"

	"starter-backend/internal/config"
	"starter-backend/internal/database"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	db, err := database.Open(cfg.DatabaseDSN, cfg.AppEnv)
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close(db)
	if err := database.Migrate(db); err != nil {
		log.Fatal(err)
	}
	log.Print("database migrations are up to date")
}
