package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	BotConfig
	APIConfig
}

type BotConfig struct {
	Token     string `env:"TELEGRAM_BOT_TOKEN" env-required:"true"`
	AdminID   int64  `env:"TELEGRAM_ADMIN_ID" env-required:"true"`
	WorkerNum int    `env:"TELEGRAM_WORKER_NUM" env-default:"4"`
}

type APIConfig struct {
	CourseURL string `env:"COURCES_API_URL"`
}

// envStage = {"dev", "prod"}
func LoadConfig(envStage string) *Config {
	err := godotenv.Load(".env." + envStage)
	if err != nil {
		panic("Failed to load .env file: " + err.Error())
	}

	var cfg Config
	err = cleanenv.ReadEnv(&cfg)
	if err != nil {
		panic("Failed to load environment variables: " + err.Error())
	}

	return &cfg
}
