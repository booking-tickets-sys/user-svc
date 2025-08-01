package config

import (
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	App      AppConfig      `mapstructure:"app"`
	Server   ServerConfig   `mapstructure:"server"`
	Security SecurityConfig `mapstructure:"security"`
	Logging  LoggingConfig  `mapstructure:"logging"`
	Database DatabaseConfig `mapstructure:"database"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	DBName          string        `mapstructure:"db_name"`
	SSLMode         string        `mapstructure:"ssl_mode"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

// AppConfig holds application-level configuration
type AppConfig struct {
	Name        string `mapstructure:"name"`
	Version     string `mapstructure:"version"`
	Environment string `mapstructure:"environment"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	GRPC GRPCConfig `mapstructure:"grpc"`
}

// GRPCConfig holds gRPC server configuration
type GRPCConfig struct {
	Port                    int           `mapstructure:"port"`
	Host                    string        `mapstructure:"host"`
	GracefulShutdownTimeout time.Duration `mapstructure:"graceful_shutdown_timeout"`
}

// SecurityConfig holds security-related configuration
type SecurityConfig struct {
	JWT    JWTConfig    `mapstructure:"jwt"`
	Paseto PasetoConfig `mapstructure:"paseto"`
}

// JWTConfig holds JWT-specific configuration
type JWTConfig struct {
	SecretKey       string        `mapstructure:"secret_key"`
	SecretKeyLength int           `mapstructure:"secret_key_length"`
	TokenDuration   time.Duration `mapstructure:"token_duration"`
	Issuer          string        `mapstructure:"issuer"`
}

// UnmarshalText implements encoding.TextUnmarshaler to automatically decode base64 secret key
func (j *JWTConfig) UnmarshalText(text []byte) error {
	// This is a placeholder - the actual unmarshaling will be handled by mapstructure
	return nil
}

// AfterUnmarshal decodes the base64 secret key after unmarshaling
func (j *JWTConfig) AfterUnmarshal() error {
	if j.SecretKey != "" {
		decoded, err := decodeBase64(j.SecretKey)
		if err != nil {
			return fmt.Errorf("failed to decode base64 secret key: %w", err)
		}
		j.SecretKey = decoded
	}
	return nil
}

// decodeBase64 decodes a base64 string
func decodeBase64(encoded string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}

// PasetoConfig holds PASETO-specific configuration
type PasetoConfig struct {
	SecretKey       string        `mapstructure:"secret_key"`
	SecretKeyLength int           `mapstructure:"secret_key_length"`
	TokenDuration   time.Duration `mapstructure:"token_duration"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string     `mapstructure:"level"`
	Format string     `mapstructure:"format"`
	Output string     `mapstructure:"output"`
	File   FileConfig `mapstructure:"file"`
}

// FileConfig holds file logging configuration
type FileConfig struct {
	Enabled    bool   `mapstructure:"enabled"`
	Path       string `mapstructure:"path"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxAge     int    `mapstructure:"max_age"`
	MaxBackups int    `mapstructure:"max_backups"`
}

// LoadConfig reads configuration from file or environment variables
func LoadConfig(path string) (*Config, error) {
	// Set the config file path directly
	viper.SetConfigFile(path)
	viper.SetConfigType("yaml")

	// Read environment variables
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Decode base64 secret keys after unmarshaling
	if err := config.Security.JWT.AfterUnmarshal(); err != nil {
		return nil, fmt.Errorf("failed to process JWT config: %w", err)
	}

	return &config, nil
}
