package telegram

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"

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

func (bot *TelegramBot) Start(ctx context.Context) {
	updateConfig := tapi.NewUpdate(0)
	updateConfig.Timeout = 69
	updateChan := bot.BotAPI.GetUpdatesChan(updateConfig)

	const workerNum = 5

	var wg sync.WaitGroup
	wg.Add(workerNum)
	for range workerNum {
		go func() {
			defer wg.Done()
			bot.Worker(ctx, updateChan)
		}()
	}

	wg.Wait()
	slog.Info("Telegram Bot Gracefully shut down")
}

func (bot *TelegramBot) Worker(ctx context.Context, updateChan tapi.UpdatesChannel) {
	for {
		select {
		case <-ctx.Done():
			return
		case update, ok := <-updateChan:
			if !ok {
				slog.Info("Update channel closed")
				return
			}
			bot.HandleUpdate(update)
		}
	}
}
