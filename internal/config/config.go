package config

import (
	"os"
	"strconv"
)

// Config holds all configuration for the application
type Config struct {
	Port int
}

// LoadConfig loads the configuration from environment variables
func LoadConfig() *Config {
	port := 8080
	if portStr := os.Getenv("PORT"); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			port = p
		}
	}

	return &Config{
		Port: port,
	}
}
