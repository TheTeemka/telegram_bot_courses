package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Stage            string
	TelegramBotToken string
	CourcesAPIURL    string
	AdminID          int64
}

// envStage = {"dev", "prod"}
func LoadConfig(envStage string) *Config {
	err := godotenv.Load(".env." + envStage)
	if err != nil {
		panic("Error loading .env file: " + err.Error())
	}

	cfg := &Config{
		Stage:            envStage,
		CourcesAPIURL:    os.Getenv("COURCES_API_URL"),
		TelegramBotToken: os.Getenv("TELEGRAM_BOT_TOKEN"),
	}

	if cfg.TelegramBotToken == "" {
		panic("TELEGRAM_BOT_TOKEN environment variable not set")
	}

	adminIDstr := os.Getenv("ADMIN_ID")
	adminIDint, err := strconv.ParseInt(adminIDstr, 10, 64)
	if err != nil {
		panic("ADMIN_ID environment variable must be an integer")
	}
	cfg.AdminID = adminIDint

	return cfg
}
