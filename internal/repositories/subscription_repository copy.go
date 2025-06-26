package repositories

import "github.com/TheTeemka/telegram_bot_cources/internal/models"

type CourseSubscriptionRepository interface {
	Subscribe(int64, string, []string) error
	GetSubscriptions(int64) ([]*models.CourseSubscription, error)
	GetAll() ([]*models.CourseSubscription, error)
	Update(*models.CourseSubscription) error
	UnSubscribe(int64, string) error
	ClearSubscriptions(int64) error
}
