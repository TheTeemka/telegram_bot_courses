package repositories

import "github.com/TheTeemka/telegram_bot_cources/internal/models"

type CourseSubscriptionRepository interface {
	Subscribe(int64, string) error
	GetSubscription(int64) []models.CourseSubscription
	UnSubscribe(int64, string) error
}
