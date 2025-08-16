// Package config provides shared database configuration
// Safe to import - doesn't affect existing implementations
package config

import (
	"fmt"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
}

// DefaultConfig returns default database configuration from environment
func DefaultConfig() *DatabaseConfig {
	return &DatabaseConfig{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "5432"),
		User:     getEnv("DB_USER", "openpenpal_user"),
		Password: getEnv("DB_PASSWORD", ""),
		Database: getEnv("DB_NAME", "openpenpal"),
	}
}

// NewDB creates a new database connection
func NewDB() (*gorm.DB, error) {
	config := DefaultConfig()
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=require",
		config.Host, config.Port, config.User, config.Password, config.Database)
	
	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}

// getEnv helper function
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}