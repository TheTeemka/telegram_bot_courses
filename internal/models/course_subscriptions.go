package models

type CourseSubscription struct {
	TelegramID int64
	Course     string
	Section    string
	IsFull     bool
}
