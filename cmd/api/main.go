package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/TheTeemka/telegram_bot_cources/internal/config"
	"github.com/TheTeemka/telegram_bot_cources/internal/repositories"
	"github.com/TheTeemka/telegram_bot_cources/internal/telegram"
	"github.com/TheTeemka/telegram_bot_cources/pkg/logging"
)

func main() {
	stage := flag.String("stage", "dev", "Environment stage (dev, prod)")
	flag.Parse()
	logging.SetSlog(*stage)

	cfg := config.LoadConfig(*stage)

	courcesRepo := repositories.NewCourseRepo(cfg.APIConfig.CourseURL, 10*time.Minute)
	bot := telegram.NewTelegramBot(cfg.BotConfig.Token, cfg.BotConfig.AdminID, courcesRepo)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGPIPE)

	go func() {
		sig := <-sigCh
		slog.Info("Received shutdown signal", "signal", sig)
		cancel()
	}()

	slog.Info("Telegram Bot Started...",
		"BOT Name", bot.BotAPI.Self.FirstName,
		"stage", stage,
		"cources url", cfg.APIConfig.CourseURL,
		"semester name", courcesRepo.SemesterName)

	bot.Start(ctx)

	slog.Info("Telegram Bot Gracefully shut down")
}
