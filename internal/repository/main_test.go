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
	// In a real CI environment, these should come from environment variables
	cfg := &config.DatabaseConfig{
		Host:     "localhost",
		Port:     "5432",
		User:     "postgres",
		Password: "password",
		DBName:   "go_mall_test",
		SSLMode:  "disable",
		TimeZone: "Asia/Shanghai",
	}

	// Initialize database connection
	var err error
	testDB, err = database.NewPostgresDB(cfg)
	if err != nil {
		log.Printf("WARNING: Failed to connect to test database: %v", err)
		log.Println("Skipping database-dependent tests or they will fail.")
		// We don't os.Exit(1) here to allow other independent tests to run if any,
		// but since these are repository tests, they will likely fail/panic.
	}

	// Run tests
	code := m.Run()

	// Cleanup (Optional: Drop tables or clean data)
	// if testDB != nil {
	// 	// cleanData(testDB)
	// }

	os.Exit(code)
}
