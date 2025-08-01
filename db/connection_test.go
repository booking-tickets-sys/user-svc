package database

import (
	"context"
	"testing"
	"time"

	"user-svc/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConnection(t *testing.T) {
	// Load config
	cfg, err := config.LoadConfig("../config.yaml")
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Test connection creation
	conn, err := NewConnection(cfg.Database)
	require.NoError(t, err, "Failed to create database connection")
	defer conn.Close()

	// Verify connection is not nil
	assert.NotNil(t, conn)
	assert.NotNil(t, conn.DB)

	// Test that we can ping the database
	err = conn.Ping()
	assert.NoError(t, err)
}

func TestConnectionPoolConfiguration(t *testing.T) {
	// Test with custom configuration
	testConfig := config.DatabaseConfig{
		Host:            "localhost",
		Port:            5432,
		User:            "user",
		Password:        "password",
		DBName:          "users",
		SSLMode:         "disable",
		MaxOpenConns:    15,
		MaxIdleConns:    8,
		ConnMaxLifetime: 10 * time.Minute,
	}

	conn, err := NewConnection(testConfig)
	require.NoError(t, err, "Failed to create database connection with custom pool configuration")
	defer conn.Close()

	// Verify connection is not nil
	assert.NotNil(t, conn)
	assert.NotNil(t, conn.DB)

	// Test that we can ping the database
	err = conn.Ping()
	assert.NoError(t, err)
}

func TestTransaction(t *testing.T) {
	// Load config
	cfg, err := config.LoadConfig("../config.yaml")
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Test connection creation
	conn, err := NewConnection(cfg.Database)
	require.NoError(t, err, "Failed to create database connection for transaction test")
	defer conn.Close()

	// Test transaction
	ctx := context.Background()
	tx, err := conn.BeginTx(ctx)
	require.NoError(t, err)
	require.NotNil(t, tx)

	// Commit the transaction
	err = tx.Commit()
	assert.NoError(t, err)
}
