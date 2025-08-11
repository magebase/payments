package db

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config holds database configuration
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// YugabyteConfig holds Yugabyte DB specific configuration
type YugabyteConfig struct {
	Config
	MaxConnections int
	MinConnections int
	MaxIdleTime    int // seconds
}

// ClickHouseConfig holds ClickHouse specific configuration
type ClickHouseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	Protocol string
}

// LoadYugabyteConfig loads Yugabyte configuration from environment variables
func LoadYugabyteConfig() *YugabyteConfig {
	return &YugabyteConfig{
		Config: Config{
			Host:     getEnv("YB_HOST", "localhost"),
			Port:     getEnvAsInt("YB_PORT", 5433),
			User:     getEnv("YB_USER", "yugabyte"),
			Password: getEnv("YB_PASSWORD", "yugabyte"),
			DBName:   getEnv("YB_DBNAME", "payments"),
			SSLMode:  getEnv("YB_SSLMODE", "disable"),
		},
		MaxConnections: getEnvAsInt("YB_MAX_CONNECTIONS", 10),
		MinConnections: getEnvAsInt("YB_MIN_CONNECTIONS", 2),
		MaxIdleTime:    getEnvAsInt("YB_MAX_IDLE_TIME", 300),
	}
}

// LoadClickHouseConfig loads ClickHouse configuration from environment variables
func LoadClickHouseConfig() *ClickHouseConfig {
	return &ClickHouseConfig{
		Host:     getEnv("CH_HOST", "localhost"),
		Port:     getEnvAsInt("CH_PORT", 9000),
		User:     getEnv("CH_USER", "default"),
		Password: getEnv("CH_PASSWORD", ""),
		DBName:   getEnv("CH_DBNAME", "payments"),
		Protocol: getEnv("CH_PROTOCOL", "native"),
	}
}

// GetDSN returns the database connection string for Yugabyte
func (c *YugabyteConfig) GetDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode)
}

// GetClickHouseDSN returns the ClickHouse connection string
func (c *ClickHouseConfig) GetClickHouseDSN() string {
	if c.Password != "" {
		return fmt.Sprintf("%s://%s:%s@%s:%d/%s",
			c.Protocol, c.User, c.Password, c.Host, c.Port, c.DBName)
	}
	return fmt.Sprintf("%s://%s@%s:%d/%s",
		c.Protocol, c.User, c.Host, c.Port, c.DBName)
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt gets an environment variable as an integer or returns a default value
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvAsSlice gets an environment variable as a slice or returns a default value
func getEnvAsSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}
