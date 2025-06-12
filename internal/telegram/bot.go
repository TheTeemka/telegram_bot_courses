package telegram

import (
	"log/slog"
	"os"

	"github.com/TheTeemka/telegram_bot_cources/internal/courses"
	"github.com/TheTeemka/telegram_bot_cources/internal/repo"
	tapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Section = courses.Section

type TelegramBot struct {
	BotAPI      *tapi.BotAPI
	CourcesRepo *repo.CourceRepo
}

func NewTelegramBot(token string, courcesRepo *repo.CourceRepo) *TelegramBot {
	bot, err := tapi.NewBotAPI(token)
	if err != nil {
		slog.Error("Failed to create Telegram Bot", "error", err)
		os.Exit(1)
	}

	return &TelegramBot{
		BotAPI:      bot,
		CourcesRepo: courcesRepo,
	}
}

func (bot *TelegramBot) Start() error {
	updateConfig := tapi.NewUpdate(0)
	updateConfig.Timeout = 69
	updateChan := bot.BotAPI.GetUpdatesChan(updateConfig)

	for update := range updateChan {
		if update.Message == nil {
			continue
		}

		if update.Message.IsCommand() {
			//TODO: Handle Command
		} else {
			bot.HandleCourceCode(update.Message)
		}
	}
	return nil
}
