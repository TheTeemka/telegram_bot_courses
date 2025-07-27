package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/TheTeemka/telegram_bot_cources/internal/config"
	"github.com/TheTeemka/telegram_bot_cources/internal/database"
	"github.com/TheTeemka/telegram_bot_cources/internal/repositories"
	"github.com/TheTeemka/telegram_bot_cources/internal/service"
	"github.com/TheTeemka/telegram_bot_cources/internal/telegram"
	"github.com/TheTeemka/telegram_bot_cources/pkg/logging"
	tapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	cfg := config.LoadConfig()

	logging.SetSlog(cfg.EnvStage)

	bot, tracker := setupApp(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	gracefullShutdown(cancel)

	slog.Info("Telegram Bot Started...",
		"BOT Name", bot.BotAPI.Self.FirstName,
		"stage", cfg.EnvStage,
		"cources url", cfg.APIConfig.CourseURL,
		"semester name", bot.CoursesRepo.SemesterName)

	writeChan := make(chan tapi.Chattable, 10)
	go tracker.Start(ctx, writeChan)
	bot.Start(ctx, writeChan)

	slog.Info("Telegram Bot Gracefully shut down")
}

func setupApp(cfg *config.Config) (*telegram.TelegramBot, *service.Tracker) {
	db := database.NewSQLiteDB("./data/db.db")

	courseRepo := repositories.NewCourseRepo(cfg.APIConfig.CourseURL, cfg.APIConfig.IsExampleData)
	subscriptionRepo := repositories.NewSQLiteSubscriptionRepo(db)
	stateRepo := repositories.NewStateRepository(db)

	bot := telegram.NewTelegramBot(cfg.EnvStage, cfg.BotConfig, 5, courseRepo, subscriptionRepo, stateRepo)
	tracker := service.NewTracker(courseRepo, subscriptionRepo, 10*time.Minute)

	return bot, tracker
}

func gracefullShutdown(cancel context.CancelFunc) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGPIPE)
	go func() {
		sig := <-sigCh
		slog.Info("Received shutdown signal", "signal", sig)
		cancel()
	}()
}
