package repository_test

import (
	"log"
	"os"
	"testing"

	"github.com/proyuen/go-mall/pkg/config"
	"github.com/proyuen/go-mall/pkg/database"
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
		log.Printf("WARNING: Failed to connect to test database with DBNAME='%s': %v", cfg.DBName, err)
		log.Println("Skipping database-dependent tests or they will fail.")
	}

	// Run tests
	code := m.Run()

	// Cleanup (Optional: Drop tables or clean data)
	// For production-grade tests, consider adding a cleanup function here
	// to ensure a clean state between test runs or after all tests complete.

	os.Exit(code)
}