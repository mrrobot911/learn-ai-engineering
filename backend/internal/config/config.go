package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
)

// Config holds application configuration loaded from environment variables.
type Config struct {
	Port        string
	Host        string
	Environment string
	DatabaseURL string
}

// Load reads configuration from environment variables with development defaults.
// Returns a clear error message if required configuration is missing.
func Load() (*Config, error) {
	cfg := &Config{
		Host:        getEnv("HOST", "0.0.0.0"),
		Port:        getEnv("PORT", "8080"),
		Environment: getEnv("ENV", "development"),
		DatabaseURL: getEnv("DATABASE_URL", ""),
	}

	// DATABASE_URL is required in non-development environments,
	// but for the foundation we allow it to be empty with a clear message if needed
	// However per requirements: "missing required -> clear message"
	if cfg.Environment == "production" && cfg.DatabaseURL == "" {
		return nil, errors.New("DATABASE_URL is required in production environment")
	}
	if cfg.Environment == "test" && cfg.DatabaseURL == "" {
		return nil, errors.New("DATABASE_URL is required in test environment")
	}

	return cfg, nil
}

// Addr returns the host:port address for the HTTP server.
func (c *Config) Addr() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}
