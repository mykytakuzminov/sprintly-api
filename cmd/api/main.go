package main

import (
	"log"

	"github.com/mykytakuzminov/task-manager-api/internal/config"
	"github.com/mykytakuzminov/task-manager-api/internal/repository/postgres"
)

func main() {
	cfg := config.Load()

	pool, err := postgres.NewPool(cfg)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer pool.Close()
	log.Println("database connected successfully")

	if err := postgres.RunMigrations(cfg); err != nil {
		log.Fatalf("migrations failed: %v", err)
	}
	log.Println("migrations applied successfully")
}
