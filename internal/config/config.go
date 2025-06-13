package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Stage            string
	TelegramBotToken string
	CourcesAPIURL    string
}

// envStage = {"dev", "prod"}
func LoadConfig(envStage string) *Config {
	godotenv.Load(".env." + envStage)
	cfg := &Config{
		Stage:            envStage,
		CourcesAPIURL:    os.Getenv("COURCES_API_URL"),
		TelegramBotToken: os.Getenv("TELEGRAM_BOT_TOKEN"),
	}

	if cfg.TelegramBotToken == "" {
		panic("TELEGRAM_BOT_TOKEN environment variable not set")
	}

	return cfg
}
