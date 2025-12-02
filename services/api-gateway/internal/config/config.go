package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	// Server
	Port    string
	AppEnv  string
	AppName string

	// Backend Services
	CoreServiceURL string
	BotServiceURL  string

	// Traefik
	TraefikEntrypoint string
	TraefikDashboard  string

	// CORS
	AllowOrigins string

	// Logging
	LogLevel string
}

// LoadConfig loads environment variables from .env file and returns a Config struct
func LoadConfig(envPath string) (*Config, error) {
	// Load .env file (optional - won't fail if file doesn't exist)
	_ = godotenv.Load(envPath)

	cfg := &Config{
		Port:              getEnv("GATEWAY_PORT", "8000"),
		AppEnv:            getEnv("APP_ENV", "development"),
		AppName:           getEnv("APP_NAME", "Permia API Gateway"),
		CoreServiceURL:    getEnv("CORE_API_URL", "http://core-service:8080/api/v1"),
		BotServiceURL:     getEnv("BOT_API_URL", "http://bot-service:8081/api/v1"),
		TraefikEntrypoint: getEnv("TRAEFIK_ENTRYPOINT", "web"),
		TraefikDashboard:  getEnv("TRAEFIK_DASHBOARD", "true"),
		AllowOrigins:      getEnv("ALLOW_ORIGINS", "*"),
		LogLevel:          getEnv("LOG_LEVEL", "info"),
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate checks if required config values are set
func (c *Config) Validate() error {
	if c.Port == "" {
		return fmt.Errorf("GATEWAY_PORT is required")
	}
	if c.CoreServiceURL == "" {
		return fmt.Errorf("CORE_API_URL is required")
	}
	return nil
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
