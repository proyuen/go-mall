package main

import (
	"log"

	"github.com/proyuen/go-mall/pkg/config"
	"github.com/proyuen/go-mall/pkg/database"
)

func main() {
	// 1. Load Configuration
	cfg, err := config.LoadConfig("./configs")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2. Initialize Database (Connect & Migrate)
	db, err := database.NewPostgresDB(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// 3. Start Server (Placeholder)
	// TODO: Inject 'db' into repositories -> services -> handlers
	log.Printf("Server starting on port %s in %s mode...\n", cfg.Server.Port, cfg.Server.Mode)

	// Keep the application running for now (simulating server)
	select {}
}