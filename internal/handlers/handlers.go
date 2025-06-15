package handlers

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/TheTeemka/telegram_bot_cources/internal/repositories"
	tapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type MessageHandler struct {
	CoursesRepo                  *repositories.CourseRepository
	CourseSubscriptionRepository repositories.CourseSubscriptionRepository
	AdminID                      int64

	welcomeText string
}

func NewMessageHandler(adminID int64, coursesRepo *repositories.CourseRepository, subscriptionRepo repositories.CourseSubscriptionRepository) *MessageHandler {

	welcomeText := fmt.Sprintf(
		"*Welcome to the Course Bot\\.* 🎓\n\n"+
			"I provide real\\-time insights about class enrollments for *%s*\n\n"+
			"Simply send me a course code \\(e\\.g\\. *CSCI 151*\\) to get:\n"+
			"• Current enrollment numbers\n"+
			"• Available seats\n"+
			"• Section details\n\n"+
			"_Updates every 10 minutes_",
		coursesRepo.SemesterName)

	return &MessageHandler{
		CoursesRepo: coursesRepo,
		AdminID:     adminID,
		welcomeText: welcomeText,

		CourseSubscriptionRepository: subscriptionRepo,
	}
}

func (h *MessageHandler) HandleUpdate(update tapi.Update) *Response {
	if update.Message == nil {
		return nil
	}

	if update.Message.From.ID != h.AdminID {
		return nil
	}

	if !update.Message.IsCommand() {
		return h.HandleCourseCode(update.Message)
	}

	switch update.Message.Command() {
	case "start":
		return h.HandleCommandStart(update.Message)
	case "subscribe":
		return h.HandleSubscribe(update.Message)
	case "unsubscribe":
		return h.HandleUnsubscribe(update.Message)
	case "list":
		return h.ListSubscriptions(update.Message)
	case "show":
		return h.ShowSubscriptions(update.Message)
	default:
		return h.HandleCommandUnknown(update.Message)
	}
}

func (h *MessageHandler) HandleSubscribe(cmd *tapi.Message) *Response {
	courseName := Standartize(cmd.CommandArguments())
	if courseName == "" {
		return NewResponse("Please provide a course code\\. Example: `/subscribe CSCI 151`")
	}

	if _, exists := h.CoursesRepo.GetCourse(courseName); !exists {
		return NewResponse(fmt.Sprintf("Course *%s* not found", courseName))
	}

	err := h.CourseSubscriptionRepository.Subscribe(cmd.From.ID, courseName)
	if err != nil {
		slog.Error("Failed to subscribe",
			"error", err,
			"user_id", cmd.From.ID,
			"course", courseName)
		return NewResponse("Failed to subscribe to the course\\. Please try again\\.")
	}

	return NewResponse(fmt.Sprintf("✅ Successfully subscribed to *%s*", courseName))
}

func (h *MessageHandler) HandleUnsubscribe(cmd *tapi.Message) *Response {
	courseName := Standartize(cmd.CommandArguments())
	if courseName == "" {
		return NewResponse("Please provide a course code\\. Example: `/unsubscribe CSCI 151`")
	}

	if _, exists := h.CoursesRepo.GetCourse(courseName); !exists {
		return NewResponse(fmt.Sprintf("Course *%s* not found", courseName))
	}

	err := h.CourseSubscriptionRepository.UnSubscribe(cmd.From.ID, courseName)
	if err != nil {
		slog.Error("Failed to unsubscribe",
			"error", err,
			"user_id", cmd.From.ID,
			"course", courseName)
		return NewResponse("Failed to unsubscribe from the course\\. Please try again\\.")
	}

	return NewResponse(fmt.Sprintf("✅ Successfully unsubscribed from *%s*", courseName))
}

func (h *MessageHandler) ListSubscriptions(msg *tapi.Message) *Response {
	subs := h.CourseSubscriptionRepository.GetSubscription(msg.From.ID)
	if len(subs) == 0 {
		return NewResponse("You haven't subscribed to any courses yet\\.")
	}

	var sb strings.Builder
	sb.WriteString("*Your subscriptions:*\n")
	for _, sub := range subs {
		sb.WriteString(fmt.Sprintf("•  *%s*", sub.Course))
	}

	return NewResponse(sb.String())
}

func (h *MessageHandler) ShowSubscriptions(msg *tapi.Message) *Response {
	subs := h.CourseSubscriptionRepository.GetSubscription(msg.From.ID)
	if len(subs) == 0 {
		return NewResponse("You haven't subscribed to any courses yet\\.")
	}

	var messages []string
	messages = append(messages, "*Your subscriptions:*")
	for _, sub := range subs {
		sections, exists := h.CoursesRepo.GetCourse(sub.Course)
		if !exists {
			messages = append(messages, fmt.Sprintf("Course '*%s*' not found", sub.Course))
		} else {
			messages = append(messages, h.beatify(sub.Course, sections))
		}
	}

	return NewResponse(messages...)
}

func (h *MessageHandler) HandleCourseCode(updateMsg *tapi.Message) *Response {
	courseName := Standartize(updateMsg.Text)
	sections, exists := h.CoursesRepo.GetCourse(courseName)
	slog.Debug("Received course code", "courseName", courseName, "exists", exists)

	if !exists {
		return NewResponse(fmt.Sprintf("Course '*%s*' not found", courseName))
	}

	return NewResponse(h.beatify(courseName, sections))
}

func (h *MessageHandler) HandleCommandUnknown(cmd *tapi.Message) *Response {
	return NewResponse(fmt.Sprintf("⚠️ Invalid command \\(/%s\\)", cmd.Command()))
}

func (h *MessageHandler) HandleCommandStart(cmd *tapi.Message) *Response {
	return NewResponse(h.welcomeText)
}
