package models

import "time"

type CourseSubscription struct {
	UserID  int64
	Course  string
	Section string
	AddedAt time.Time
}
