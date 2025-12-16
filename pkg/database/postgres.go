package database

import (
	"fmt"
	"log"
	"time"

	"github.com/proyuen/go-mall/internal/model"
	"github.com/proyuen/go-mall/pkg/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger" // Import GORM logger
)

// NewPostgresDB initializes and returns a new GORM database instance for PostgreSQL.
// It configures connection pooling, GORM performance settings, and performs auto-migration.
func NewPostgresDB(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		cfg.Host, cfg.User, cfg.Password, cfg.DBName, cfg.Port, cfg.SSLMode, cfg.TimeZone)

	// Configure GORM with performance settings
	gormConfig := &gorm.Config{
		PrepareStmt: true,                               // Cache pre-compiled statements for performance
		Logger:      logger.Default.LogMode(logger.Info), // Enable Info level logging for GORM
	}

	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Println("Database connection established")

	// Get the underlying sql.DB to configure connection pooling
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying SQL DB: %w", err)
	}

	// Connection Pooling configuration for high concurrency
	sqlDB.SetMaxOpenConns(100)           // Maximum number of open connections (prevents database overload)
	sqlDB.SetMaxIdleConns(50)            // Maximum number of connections in the idle connection pool (keeps connections warm)
	sqlDB.SetConnMaxLifetime(time.Hour) // Maximum amount of time a connection may be reused

	// Auto Migrate
	// WARNING: In production environments, database migration should be managed
	// separately (e.g., using Goose, Flyway, or a dedicated migration tool)
	// and executed before application startup. Running AutoMigrate directly
	// in the application can lead to unexpected behavior or downtime during upgrades.
	err = db.AutoMigrate(
		&model.User{},
		&model.SPU{},
		&model.SKU{},
		&model.Order{},
		&model.OrderItem{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to auto migrate database: %w", err)
	}
	log.Println("Database migration completed")

	return db, nil
}
