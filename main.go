package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"

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
	cfg.EnvStage = logging.StageDev //TODO:REMOVE

	logging.SetSlog(cfg.EnvStage)

	slog.Info("Starting Application", "config", cfg)
	db := database.NewSQLiteDB("./data/db.db")

	courseRepo := repositories.NewCourseRepo(cfg.APIConfig)
	subscriptionRepo := repositories.NewSQLiteSubscriptionRepo(db)
	stateRepo := repositories.NewStateRepository(db)
	statisticsRepo := repositories.NewStatisticsRepository(db)

	bot := telegram.NewTelegramBot(cfg.EnvStage, cfg.BotConfig, courseRepo, subscriptionRepo, stateRepo, statisticsRepo)
	tracker := service.NewTracker(courseRepo, subscriptionRepo, cfg.TimeIntervalBetweenParses)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	gracefullShutdown(cancel)

	slog.Info("Telegram Bot Started...",
		"BOT Name", bot.BotAPI.Self.FirstName,
		"stage", cfg.EnvStage,
		"cources url", cfg.APIConfig.CourseURL,
		"semester name", bot.CoursesRepo.SemesterName)

	writeChan := make(chan tapi.Chattable, 10)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		tracker.Start(ctx, writeChan)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		statisticsRepo.Run(ctx)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		bot.Start(ctx, writeChan)
	}()

	wg.Wait()
	slog.Info("Telegram Bot Gracefully shut down")
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
