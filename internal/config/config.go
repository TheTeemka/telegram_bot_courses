package config

import (
	"flag"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	EnvStage string
	BotConfig
	APIConfig
}

type BotConfig struct {
	Token        string
	AdminID      []int64
	WorkerNumber int
	IsPrivate    bool
	KaspiCard    string
}

type APIConfig struct {
	IsExampleData             bool
	CourseURL                 string
	TimeIntervalBetweenParses time.Duration
}

// envStage = ("dev", "prod")
func LoadConfig() *Config {
	stage := flag.String("stage", "dev", "Environment stage (dev, prod)")
	private := flag.Bool("private", false, "Is the bot running in public mode? (default: false)")
	exampleData := flag.Bool("example-data", false, "Load example data for testing (default: false)")
	workerNumTelegram := flag.Int("telegram-workers", 10, "Number of Telegram workers for processing updates")
	timeIntreval := flag.Duration("time-interval", 3*time.Hour, "Time interval between course parses")

	flag.Parse()

	err := godotenv.Load(".env")
	if err != nil {
		panic("Failed to load .env file: " + err.Error())
	}

	cfg := &Config{
		EnvStage: *stage,
		BotConfig: BotConfig{
			Token:        os.Getenv("TELEGRAM_BOT_TOKEN"),
			IsPrivate:    *private,
			WorkerNumber: *workerNumTelegram,
			KaspiCard:    os.Getenv("KASPI_CARD"),
		},
		APIConfig: APIConfig{
			IsExampleData:             *exampleData,
			CourseURL:                 os.Getenv("COURCES_API_URL"),
			TimeIntervalBetweenParses: *timeIntreval,
		},
	}

	if cfg.APIConfig.CourseURL == "" {
		panic("COURCES_API_URL environment variable is not set")
	}

	if *private {
		cfg.BotConfig.AdminID = parseInt64Array(os.Getenv("TELEGRAM_ADMIN_ID"))
		if len(cfg.BotConfig.AdminID) == 0 {
			panic("TELEGRAM_ADMIN_ID environment variable is not set or invalid")
		}
	}

	return cfg
}

func MustInt64(s string) int64 {
	x, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		panic("Failed to parse int64: " + err.Error())
	}
	return x
}

func parseInt64Array(s string) []int64 {
	fields := strings.Split(s, ",")
	arr := make([]int64, len(fields))
	for _, f := range fields {
		arr = append(arr, MustInt64(f))
	}
	return arr
}
