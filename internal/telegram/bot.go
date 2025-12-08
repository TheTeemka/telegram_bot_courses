package telegram

import (
	"context"
	"encoding/json"
	"errors"
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

func NewTelegramBot(stage string, cfg config.BotConfig,
	coursesRepo *repositories.CourseRepository,
	subscriptionRepo repositories.CourseSubscriptionRepository,
	stateRepo repositories.StateRepository,
	statisticsRepo *repositories.StatisticsRepository) *TelegramBot {
	bot, err := tapi.NewBotAPI(cfg.Token)
	if err != nil {
		slog.Error("Failed to create Telegram Bot", "error", err)
		os.Exit(1)
	}

	handler := handlers.NewMessageHandler(bot, cfg, coursesRepo, subscriptionRepo, stateRepo, statisticsRepo)

	res, err := bot.Request(handler.CommandsList())
	if err != nil {
		slog.Error("Failed to set bot commands", "error", err)
		os.Exit(1)
	} else if !res.Ok {
		slog.Error("Failed to set bot commands", "desc", res.Description)
		os.Exit(1)
	}

	return &TelegramBot{
		BotAPI:         bot,
		MessageHandler: handler,
		workerNum:      cfg.WorkerNumber,
	}
}

func (bot *TelegramBot) Start(ctx context.Context, writeChan <-chan tapi.Chattable) {
	updateConfig := tapi.NewUpdate(0)
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
				var unmarshalTypeErr *json.UnmarshalTypeError
				if err != nil && !errors.As(err, &unmarshalTypeErr) {
					var (
						userID   int64
						userName string
					)
					if update.Message != nil {
						userID = update.Message.From.ID
						userName = update.Message.From.UserName
					} else if update.CallbackQuery != nil {
						userID = update.CallbackQuery.From.ID
						userName = update.CallbackQuery.From.UserName
					} else {
						slog.Error("No user information in update")
						continue
					}

					slog.Error("Failed to send message in worker", "error", err,
						"userID", userID, "username", userName, "msg", msg)
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
