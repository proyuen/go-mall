package repository_test

import (
	"log"
	"os"
	"testing"

	"github.com/proyuen/go-mall/internal/model" // Import model for AutoMigrate
	"github.com/proyuen/go-mall/pkg/config"
	"github.com/proyuen/go-mall/pkg/database"
	"github.com/proyuen/go-mall/pkg/snowflake" // Import snowflake package
	"gorm.io/gorm"
)

var testDB *gorm.DB

func TestMain(m *testing.M) {
	// Setup test database configuration
	cfg := &config.DatabaseConfig{
		Host:     "localhost",
		Port:     "5432",
		User:     "postgres",
		Password: "password",
		SSLMode:  "disable",
		TimeZone: "Asia/Shanghai",
	}

	// Prioritize MALL_DATABASE_DBNAME environment variable for DBName
	// Fallback to "go_mall_test" if the environment variable is not set.
	dbNameFromEnv := os.Getenv("MALL_DATABASE_DBNAME")
	if dbNameFromEnv != "" {
		cfg.DBName = dbNameFromEnv
	} else {
		cfg.DBName = "go_mall_test"
	}

	// Initialize database connection
	var err error
	testDB, err = database.NewPostgresDB(cfg)
	if err != nil {
		log.Printf("FATAL: Failed to connect to test database with DBNAME='%s': %v", cfg.DBName, err)
		os.Exit(1) // Exit immediately if DB connection fails, as repository tests depend on it.
	}

	// Explicitly AutoMigrate models for the test database
	// This ensures the schema is up-to-date for tests,
	// especially in environments where NewPostgresDB's internal AutoMigrate might be skipped or insufficient.
	err = testDB.AutoMigrate(
		&model.User{},
		&model.SPU{},
		&model.SKU{},
		&model.Order{},
		&model.OrderItem{},
	)
	if err != nil {
		log.Printf("FATAL: Failed to auto migrate test database: %v", err)
		os.Exit(1) // Exit immediately if migration fails.
	}
	log.Println("Test database migration completed.")

	// Initialize Snowflake ID Generator
	// Use hardcoded NodeID 1 for tests.
	if err := snowflake.Init(1); err != nil {
		log.Printf("FATAL: Failed to initialize snowflake: %v", err)
		os.Exit(1)
	}

	// Run tests
	code := m.Run()

	// Optional: Cleanup test data or drop tables after tests.
	// For simplicity, we are not cleaning up in this example,
	// but in a production CI, you might want to truncate tables
	// or drop the schema to ensure isolation between test runs.

	os.Exit(code)
}
