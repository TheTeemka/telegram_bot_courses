package telegram

import (
	"fmt"
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
	AdminID     int64

	welcomeText string
}

func NewTelegramBot(token string, adminID int64, courcesRepo *repo.CourceRepo) *TelegramBot {
	bot, err := tapi.NewBotAPI(token)
	if err != nil {
		slog.Error("Failed to create Telegram Bot", "error", err)
		os.Exit(1)
	}

	welcomeText := fmt.Sprintf(
		"*Welcome to the Course Bot\\.* 🎓\n\n"+
			"I provide real\\-time insights about class enrollments for *%s*\n\n"+
			"Simply send me a course code \\(e\\.g\\. *CSCI 151*\\) to get:\n"+
			"• Current enrollment numbers\n"+
			"• Available seats\n"+
			"• Section details\n\n"+
			"_Updates every 10 minutes_",
		courcesRepo.SemesterName)

	return &TelegramBot{
		BotAPI:      bot,
		CourcesRepo: courcesRepo,
		AdminID:     adminID,

		welcomeText: welcomeText,
	}
}

func (bot *TelegramBot) Start() {
	updateConfig := tapi.NewUpdate(0)
	updateConfig.Timeout = 69
	updateChan := bot.BotAPI.GetUpdatesChan(updateConfig)

	const workerNum = 5
	for range workerNum {
		go bot.Worker(updateChan)
	}

	select {}
}

func (bot *TelegramBot) Worker(updateChan tapi.UpdatesChannel) {
	for update := range updateChan {
		if update.Message == nil {
			continue
		}

		if update.Message.From.ID != bot.AdminID {
			return
		}

		if update.Message.IsCommand() {
			bot.HandleCommand(update.Message)
		} else {
			bot.HandleCourceCode(update.Message)
		}
	}
}
