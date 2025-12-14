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
	// 1. Setup Config (Assumes docker-compose is running)
	cfg := &config.DatabaseConfig{
		Host:     "localhost",
		Port:     "5432",
		User:     "postgres",
		Password: "password",
		DBName:   "mall_db",
		SSLMode:  "disable",
		TimeZone: "Asia/Shanghai",
	}

	// 2. Connect to DB
	var err error
	testDB, err = database.NewPostgresDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to test database: %v", err)
	}

	// 3. Run Tests
	os.Exit(m.Run())
}
