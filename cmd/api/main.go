package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
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
	bot := telegram.NewTelegramBot(cfg.TelegramBotToken, cfg.AdminID, courcesRepo)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGPIPE)

	go func() {
		sig := <-sigCh
		slog.Info("Received shutdown signal", "signal", sig)
		cancel()
	}()

	slog.Info("Telegram Bot Started...",
		"BOT Name", bot.BotAPI.Self.FirstName,
		"stage", cfg.Stage,
		"cources url", cfg.CourcesAPIURL,
		"semester name", courcesRepo.SemesterName)
	bot.Start(ctx)
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
