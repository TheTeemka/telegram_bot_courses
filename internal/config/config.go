package config

import (
	"os"

	_ "github.com/joho/godotenv/autoload"
)

type Config struct {
	TelegramBotToken string
}

func LoadConfig() *Config {
	cfg := &Config{
		TelegramBotToken: os.Getenv("TELEGRAM_BOT_TOKEN"),
	}

	if cfg.TelegramBotToken == "" {
		panic("TELEGRAM_BOT_TOKEN environment variable not set")
	}

	return cfg
}
