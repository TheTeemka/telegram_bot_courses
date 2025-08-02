package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/TheTeemka/telegram_bot_cources/internal/repositories"
	"github.com/TheTeemka/telegram_bot_cources/internal/telegramfmt"
	"github.com/TheTeemka/telegram_bot_cources/internal/ticker"
	tapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Tracker struct {
	courseRepo       *repositories.CourseRepository
	subscriptionRepo repositories.CourseSubscriptionRepository
	ticker           *ticker.DynamicTicker
}

func NewTracker(courseRepo *repositories.CourseRepository, subscriptionRepo repositories.CourseSubscriptionRepository, timeInterval time.Duration) *Tracker {
	return &Tracker{
		courseRepo:       courseRepo,
		subscriptionRepo: subscriptionRepo,
		ticker:           ticker.NewDynamicTicker(timeInterval),
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
			slog.Info("Tracker ticked, checking subscriptions")
			subs, err := t.subscriptionRepo.GetAll()
			if err != nil {
				slog.Error("Failed to get subscriptions", "error", err)
				continue
			}
			t.courseRepo.Parse()
			for _, sub := range subs {
				mf := telegramfmt.NewMessageFormatter(sub.TelegramID)
				_, exists := t.courseRepo.GetCourse(sub.Course)
				if !exists {
					mf.AddString(fmt.Sprintf("%s %s is not existent anymore", sub.Course, sub.Section))
					mf.UnsubscribeOrIgnoreCourse(sub.Course)

					writeChan <- mf.Messages()[0]
					continue
				}

				sect, exists := t.courseRepo.GetSection(sub.Course, sub.Section)
				if !exists {
					mf.AddString(fmt.Sprintf("%s %s is not existent anymore", sub.Course, sub.Section))
					mf.UnsubscribeOrIgnoreSection(sub.Course, sub.Section)

					writeChan <- mf.Messages()[0]
					continue
				}

				if sub.IsFull && sect.Size < sect.Cap {
					writeChan <- immediateMessage(sub.TelegramID,
						fmt.Sprintf("🔆 %s %s now has free places \\(%d/%d\\)",
							sub.Course, sub.Section, sect.Size, sect.Cap))

					sub.IsFull = false
					err := t.subscriptionRepo.Update(sub)
					if err != nil {
						slog.Error("Failed to update subscription", "error", err, "subscription", sub)
						continue
					}

				} else if !sub.IsFull && sect.Size >= sect.Cap {
					writeChan <- immediateMessage(sub.TelegramID,
						fmt.Sprintf("🚫 %s %s is full \\(%d/%d\\)",
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
