package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	// Test loading config from the root directory (where config.yaml is located)
	cfg, err := LoadConfig("../config.yaml")
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Test database configuration
	assert.Equal(t, "localhost", cfg.Database.Host)
	assert.Equal(t, 5432, cfg.Database.Port)
	assert.Equal(t, "user", cfg.Database.User)
	assert.Equal(t, "password", cfg.Database.Password)
	assert.Equal(t, "users", cfg.Database.DBName)
	assert.Equal(t, "disable", cfg.Database.SSLMode)

	// Test connection pool configuration
	assert.Equal(t, 10, cfg.Database.MaxOpenConns)
	assert.Equal(t, 5, cfg.Database.MaxIdleConns)
	assert.Equal(t, 5*time.Minute, cfg.Database.ConnMaxLifetime)
}

func TestLoadConfigWithBase64SecretKey(t *testing.T) {
	// Test loading config with base64 encoded secret key
	config, err := LoadConfig("../config.yaml")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify the secret key was decoded
	expected := "your-super-secret-jwt-key-here-make-it-at-least-32-characters-long"
	if config.Security.JWT.SecretKey != expected {
		t.Errorf("Expected secret key to be decoded to '%s', got '%s'", expected, config.Security.JWT.SecretKey)
	}
}
