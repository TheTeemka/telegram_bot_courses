package main

import (
	"log/slog"
	"time"

	"github.com/TheTeemka/telegram_bot_cources/internal/config"
	"github.com/TheTeemka/telegram_bot_cources/internal/repo"
	"github.com/TheTeemka/telegram_bot_cources/internal/telegram"
)

func main() {
	cfg := config.LoadConfig()
	courcesRepo := repo.NewCourceRepo(10 * time.Minute)
	bot := telegram.NewTelegramBot(cfg.TelegramBotToken, courcesRepo)

	slog.Info("Telegram Bot Started...", "BOT Name", bot.BotAPI.Self.FirstName)

	go courcesRepo.Watch()
	bot.Start()
}
