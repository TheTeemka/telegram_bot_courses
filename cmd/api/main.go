package main

import (
	"flag"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/TheTeemka/telegram_bot_cources/internal/config"
	"github.com/TheTeemka/telegram_bot_cources/internal/repo"
	"github.com/TheTeemka/telegram_bot_cources/internal/telegram"
)

const (
	StageDev  = "dev"
	StageProd = "prod"
)

func main() {
	stage := flag.String("stage", "dev", "Environment stage (dev, prod)")
	flag.Parse()
	setSlog(*stage)

	cfg := config.LoadConfig(*stage)

	courcesRepo := repo.NewCourceRepo(cfg.CourcesAPIURL, 10*time.Minute)
	bot := telegram.NewTelegramBot(cfg.TelegramBotToken, courcesRepo)

	slog.Info("Telegram Bot Started...", "BOT Name", bot.BotAPI.Self.FirstName, "stage", cfg.Stage, "cources url", cfg.CourcesAPIURL, "semester name", courcesRepo.SemesterName)

	go courcesRepo.Watch()
	bot.Start()
}

func setSlog(stage string) {
	var l slog.Level
	switch stage {
	case StageDev:
		l = slog.LevelDebug
	case StageProd:
		l = slog.LevelInfo
	default:
		panic("Unknown stage")
	}

	h := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     l,
		AddSource: stage == StageDev,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.SourceKey {
				source, _ := a.Value.Any().(*slog.Source)
				if source != nil {
					source.File = filepath.Base(source.File)
				}
			}
			return a
		},
	})

	slog.SetDefault(slog.New(h))
}
