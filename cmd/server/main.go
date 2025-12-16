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
	"github.com/proyuen/go-mall/pkg/hasher"
	"github.com/proyuen/go-mall/pkg/snowflake"
	"github.com/proyuen/go-mall/pkg/token"
)

func main() {
	// 1. Load Configuration
	cfg, err := config.LoadConfig("./configs")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2. Initialize Snowflake ID Generator
	// In a distributed deployment, this NodeID (1) must be unique per instance (e.g., from config or env).
	if err := snowflake.Init(1); err != nil {
		log.Fatalf("Failed to initialize snowflake: %v", err)
	}

	// 3. Initialize Database (Connect & Migrate)
	db, err := database.NewPostgresDB(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// 3. Initialize Repositories, Services, Handlers, and Router
	txManager := database.NewTransactionManager(db)

	// User Module
	userRepo := repository.NewUserRepository(db)
	// Initialize password hasher with default cost
	passwordHasher := hasher.NewBcryptHasher(0)
	// Initialize token maker
	tokenMaker, err := token.NewJWTMaker(cfg.JWT.Secret)
	if err != nil {
		log.Fatalf("Failed to create token maker: %v", err)
	}
	userService := service.NewUserService(userRepo, passwordHasher, tokenMaker)
	userHandler := handler.NewUserHandler(userService)

	// Product Module
	productRepo := repository.NewProductRepository(db)
	productService := service.NewProductService(productRepo)
	productHandler := handler.NewProductHandler(productService)

	// Order Module
	orderRepo := repository.NewOrderRepository(db)
	orderService := service.NewOrderService(orderRepo, productRepo, txManager)
	orderHandler := handler.NewOrderHandler(orderService)

	router := router.NewRouter(userHandler, productHandler, orderHandler, tokenMaker)
	engine := router.InitRoutes()

	// 4. Start Server
	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	log.Printf("Server starting on %s in %s mode...\n", addr, cfg.Server.Mode)
	if err := engine.Run(addr); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
