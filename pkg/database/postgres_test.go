package database

import (
	"testing"

	"github.com/proyuen/go-mall/pkg/config"
	"github.com/stretchr/testify/require"
)

func TestNewPostgresDB(t *testing.T) {
	// 1. Setup Configuration (matching docker-compose.yml)
	cfg := &config.DatabaseConfig{
		Host:     "localhost",
		Port:     "5432",
		User:     "postgres",
		Password: "password",
		DBName:   "mall_db",
		SSLMode:  "disable",
		TimeZone: "Asia/Shanghai",
	}

	// 2. Attempt Connection
	db, err := NewPostgresDB(cfg)
	require.NoError(t, err, "Failed to connect to database")
	require.NotNil(t, db, "Database instance is nil")

	// 3. Verify Connection
	sqlDB, err := db.DB()
	require.NoError(t, err, "Failed to get generic database object")
	require.NotNil(t, sqlDB, "SQL DB instance is nil")

	err = sqlDB.Ping()
	require.NoError(t, err, "Failed to ping database")

	t.Log("Successfully connected and pinged the database!")
}
