package config

import (
	"flag"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	EnvStage string
	BotConfig
	APIConfig
}

type BotConfig struct {
	Token   string `env:"TELEGRAM_BOT_TOKEN" env-required:"true"`
	AdminID int64  `env:"TELEGRAM_ADMIN_ID"`
}

type APIConfig struct {
	CourseURL string `env:"COURCES_API_URL"`
}

// envStage = {"dev", "prod"}
func LoadConfig() *Config {
	stage := flag.String("stage", "dev", "Environment stage (dev, prod)")
	public := flag.Bool("public", false, "Is the bot running in public mode? (default: false)")
	flag.Parse()

	err := godotenv.Load(".env." + *stage)
	if err != nil {
		panic("Failed to load .env file: " + err.Error())
	}

	var cfg Config
	err = cleanenv.ReadEnv(&cfg)
	if err != nil {
		panic("Failed to load environment variables: " + err.Error())
	}

	if *stage != "dev" && cfg.APIConfig.CourseURL == "" {
		panic("COURCES_API_URL is required in production environment")
	}

	if !*public && cfg.BotConfig.AdminID == 0 {
		panic("TELEGRAM_ADMIN_ID is required in non-public mode")
	}
	cfg.EnvStage = *stage

	return &cfg
}
