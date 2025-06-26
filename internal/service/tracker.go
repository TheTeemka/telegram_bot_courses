package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/TheTeemka/telegram_bot_cources/internal/repositories"
	tapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Tracker struct {
	courseRepo       *repositories.CourseRepository
	subscriptionRepo repositories.CourseSubscriptionRepository
	ticker           *time.Ticker
}

func NewTracker(courseRepo *repositories.CourseRepository, subscriptionRepo repositories.CourseSubscriptionRepository, d time.Duration) *Tracker {
	return &Tracker{
		courseRepo:       courseRepo,
		subscriptionRepo: subscriptionRepo,
		ticker:           time.NewTicker(d),
	}
}

func (t *Tracker) Start(ctx context.Context, writeChan chan<- tapi.Chattable) {
	for {
		select {
		case <-ctx.Done():
			slog.Info("Tracker stopped")
			t.ticker.Stop()
			return

		case <-t.ticker.C:
			subs, err := t.subscriptionRepo.GetAll()
			if err != nil {
				slog.Error("Failed to get subscriptions", "error", err)
				continue
			}
			t.courseRepo.Parse()
			for _, sub := range subs {

				sect, exists := t.courseRepo.GetSection(sub.Course, sub.Section)
				if !exists {
					writeChan <- immediateMessage(sub.TelegramID,
						fmt.Sprintf("%s %s available anymore", sub.Course, sub.Section))
				}

				if sub.IsFull && sect.Size < sect.Cap {
					writeChan <- immediateMessage(sub.TelegramID,
						fmt.Sprintf("%s %s is available //(%d/%d//)",
							sub.Course, sub.Section, sect.Size, sect.Cap))

					sub.IsFull = false
					err := t.subscriptionRepo.Update(sub)
					if err != nil {
						slog.Error("Failed to update subscription", "error", err, "subscription", sub)
						continue
					}

				} else if !sub.IsFull && sect.Size >= sect.Cap {
					writeChan <- immediateMessage(sub.TelegramID,
						fmt.Sprintf("%s %s is full //(%d/%d//)",
							sub.Course, sub.Section, sect.Size, sect.Cap))

					sub.IsFull = true
					err := t.subscriptionRepo.Update(sub)
					if err != nil {
						slog.Error("Failed to update subscription", "error", err, "subscription", sub)
						continue
					}
				}
			}
		}
	}
}

func immediateMessage(chatId int64, text string) tapi.Chattable {
	msg := tapi.NewMessage(chatId, text)
	msg.ParseMode = tapi.ModeMarkdownV2
	return msg
}
