package telegram

import (
	"context"
	"log/slog"
	"os"
	"sync"

	"github.com/TheTeemka/telegram_bot_cources/internal/config"
	"github.com/TheTeemka/telegram_bot_cources/internal/handlers"
	"github.com/TheTeemka/telegram_bot_cources/internal/repositories"
	tapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TelegramBot struct {
	BotAPI *tapi.BotAPI
	*handlers.MessageHandler
	workerNum int
}

func NewTelegramBot(stage string, cfg config.BotConfig, workerNum int, coursesRepo *repositories.CourseRepository, subscriptionRepo repositories.CourseSubscriptionRepository) *TelegramBot {
	bot, err := tapi.NewBotAPI(cfg.Token)
	if err != nil {
		slog.Error("Failed to create Telegram Bot", "error", err)
		os.Exit(1)
	}

	// if stage == "dev" {
	// 	bot.Debug = true
	// }

	return &TelegramBot{
		BotAPI:         bot,
		MessageHandler: handlers.NewMessageHandler(cfg.AdminID, coursesRepo, subscriptionRepo),
		workerNum:      workerNum,
	}
}

func (bot *TelegramBot) Start(ctx context.Context, writeChan <-chan tapi.Chattable) {
	updateConfig := tapi.NewUpdate(0)
	updateConfig.Timeout = 69
	updateChan := bot.BotAPI.GetUpdatesChan(updateConfig)

	var wg sync.WaitGroup
	wg.Add(bot.workerNum)
	for range bot.workerNum {
		go func() {
			defer wg.Done()
			bot.Worker(ctx, updateChan)
		}()
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		bot.Sender(ctx, writeChan)
	}()

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
					slog.Error("Failed to send message in worker", "error", err, "username", update.Message.From.UserName, "msg", msg)
					continue
				}
			}
		}
	}
}

func (bot *TelegramBot) Sender(ctx context.Context, writeChan <-chan tapi.Chattable) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-writeChan:
			if !ok {
				slog.Info("Update channel closed")
				return
			}
			_, err := bot.BotAPI.Send(msg)
			if err != nil {
				slog.Error("Failed to send message in sender", "error", err)
				continue
			}
		}
	}
}
