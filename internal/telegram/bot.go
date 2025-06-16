package telegram

import (
	"context"
	"log/slog"
	"os"
	"sync"

	"github.com/TheTeemka/telegram_bot_cources/internal/handlers"
	"github.com/TheTeemka/telegram_bot_cources/internal/repositories"
	tapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TelegramBot struct {
	BotAPI *tapi.BotAPI
	*handlers.MessageHandler
}

func NewTelegramBot(stage string, token string, adminID int64, coursesRepo *repositories.CourseRepository, subscriptionRepo repositories.CourseSubscriptionRepository) *TelegramBot {
	bot, err := tapi.NewBotAPI(token)
	if err != nil {
		slog.Error("Failed to create Telegram Bot", "error", err)
		os.Exit(1)
	}
	if stage == "dev" {
		bot.Debug = true
	}
	return &TelegramBot{
		BotAPI:         bot,
		MessageHandler: handlers.NewMessageHandler(adminID, coursesRepo, subscriptionRepo),
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

			msgs := bot.HandleUpdate(update)
			if msgs == nil {
				continue
			}

			for _, msg := range msgs {
				_, err := bot.BotAPI.Send(msg)
				if err != nil {
					slog.Error("Failed to send message", "error", err, "username", update.Message.From.UserName, "msg", msg)
					continue
				}
			}
		}
	}
}
