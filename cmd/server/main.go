package main

import (
	"fmt"
	"log"

	"github.com/proyuen/go-mall/internal/handler"
	"github.com/proyuen/go-mall/internal/repository"
	"github.com/proyuen/go-mall/internal/router"
	"github.com/proyuen/go-mall/internal/service"
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

	// 3. Initialize Repositories, Services, Handlers, and Router
	// User Module
	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo, cfg.JWT.Secret)
	userHandler := handler.NewUserHandler(userService)

	// Product Module
	productRepo := repository.NewProductRepository(db)
	productService := service.NewProductService(productRepo)
	productHandler := handler.NewProductHandler(productService)

	router := router.NewRouter(userHandler, productHandler, cfg.JWT.Secret)
	engine := router.InitRoutes()

	// 4. Start Server
	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	log.Printf("Server starting on %s in %s mode...\n", addr, cfg.Server.Mode)
	if err := engine.Run(addr); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
