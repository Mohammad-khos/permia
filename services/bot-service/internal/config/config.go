package config

import (
	"github.com/joho/godotenv"
	"os"
)

// Config holds all configuration for the application.
type Config struct {
	TelegramBotToken string
	CoreApiURL       string
}

// LoadConfig reads configuration from file and environment variables.
func LoadConfig(path string) (*Config, error) {
	err := godotenv.Load(path)
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		TelegramBotToken: os.Getenv("TELEGRAM_BOT_TOKEN"),
		CoreApiURL:       os.Getenv("CORE_API_URL"),
	}

	return cfg, nil
}