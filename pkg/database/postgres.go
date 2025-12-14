package database

import (
	"fmt"
	"log"

	"github.com/proyuen/go-mall/internal/model"
	"github.com/proyuen/go-mall/pkg/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewPostgresDB(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		cfg.Host, cfg.User, cfg.Password, cfg.DBName, cfg.Port, cfg.SSLMode, cfg.TimeZone)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Println("Database connection established")

	// Auto Migrate
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