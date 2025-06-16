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

func (h *MessageHandler) HandleUpdate(update tapi.Update) []tapi.MessageConfig {

	if update.CallbackQuery != nil {
		return h.HandleCallback(update.CallbackQuery)
	}

	if update.Message == nil || update.Message.From.ID != h.AdminID {
		return nil
	}

	if update.Message.IsCommand() {
		return h.HandleCommand(update.Message)
	}
	return h.HandleCourseCode(update.Message)
}

func (h *MessageHandler) HandleCommand(cmd *tapi.Message) []tapi.MessageConfig {
	switch cmd.Command() {
	case "start":
		return h.HandleCommandStart(cmd)
	case "subscribe":
		return h.HandleSubscribe(cmd)
	case "unsubscribe":
		return h.HandleUnsubscribe(cmd)
	case "list":
		return h.ListSubscriptions(cmd)
	case "showall":
		return h.ShowAllSubscriptions(cmd)
	default:
		return h.HandleCommandUnknown(cmd)
	}
}

func (h *MessageHandler) HandleSubscribe(cmd *tapi.Message) []tapi.MessageConfig {
	mf := NewMessageFormatter(cmd.From.ID)
	courseName := Standartize(cmd.CommandArguments())
	if courseName == "" {
		return mf.ImmediateMessage("Please provide a course code\\. Example: `/subscribe CSCI 151`")
	}

	if _, exists := h.CoursesRepo.GetCourse(courseName); !exists {
		return mf.ImmediateMessage(fmt.Sprintf("Course *%s* not found", courseName))
	}

	err := h.CourseSubscriptionRepository.Subscribe(cmd.From.ID, courseName)
	if err != nil {
		slog.Error("Failed to subscribe",
			"error", err,
			"user_id", cmd.From.ID,
			"course", courseName)
		return mf.ImmediateMessage("Failed to subscribe to the course\\. Please try again\\.")
	}

	return mf.ImmediateMessage(fmt.Sprintf("✅ Successfully subscribed to *%s*", courseName))
}

func (h *MessageHandler) HandleUnsubscribe(cmd *tapi.Message) []tapi.MessageConfig {
	mf := NewMessageFormatter(cmd.From.ID)
	courseName := Standartize(cmd.CommandArguments())
	if courseName == "" {
		return mf.ImmediateMessage("Please provide a course code\\. Example: `/unsubscribe CSCI 151`")
	}

	if _, exists := h.CoursesRepo.GetCourse(courseName); !exists {
		return mf.ImmediateMessage(fmt.Sprintf("Course *%s* not found", courseName))
	}

	err := h.CourseSubscriptionRepository.UnSubscribe(cmd.From.ID, courseName)
	if err != nil {
		slog.Error("Failed to unsubscribe",
			"error", err,
			"user_id", cmd.From.ID,
			"course", courseName)
		return mf.ImmediateMessage("Failed to unsubscribe from the course\\. Please try again\\.")
	}

	return mf.ImmediateMessage(fmt.Sprintf("✅ Successfully unsubscribed from *%s*", courseName))
}

func (h *MessageHandler) ListSubscriptions(cmd *tapi.Message) []tapi.MessageConfig {
	mf := NewMessageFormatter(cmd.From.ID)
	subs := h.CourseSubscriptionRepository.GetSubscription(cmd.From.ID)
	if len(subs) == 0 {
		return mf.ImmediateMessage("You haven't subscribed to any courses yet\\.")
	}

	mf.AddString("*Your subscriptions:*\n")
	for _, sub := range subs {
		mf.AddString(fmt.Sprintf("•  *%s*", sub.Course))
		callbackSub := fmt.Sprintf("show_%s", sub.Course)
		callbackUnSub := fmt.Sprintf("unsubscribe_%s", sub.Course)

		mf.AddKeyboardToLastMessage([][]tapi.InlineKeyboardButton{
			{
				{Text: "ℹ️ Show", CallbackData: &callbackSub},
				{Text: "❌ Unsubscribe", CallbackData: &callbackUnSub},
			},
		})
	}

	return mf.Messages()
}

func (h *MessageHandler) ShowAllSubscriptions(cmd *tapi.Message) []tapi.MessageConfig {
	mf := NewMessageFormatter(cmd.From.ID)

	subs := h.CourseSubscriptionRepository.GetSubscription(cmd.From.ID)
	if len(subs) == 0 {
		return mf.ImmediateMessage("You haven't subscribed to any courses yet\\.")
	}

	mf.AddString("*Your subscriptions:*")
	for _, sub := range subs {
		sections, exists := h.CoursesRepo.GetCourse(sub.Course)
		if !exists {
			mf.AddString(fmt.Sprintf("Course '*%s*' not found", sub.Course))
		} else {
			mf.AddString(h.beatify(sub.Course, sections))
		}
	}

	return mf.Messages()
}

func (h *MessageHandler) HandleCourseCode(updateMsg *tapi.Message) []tapi.MessageConfig {
	mf := NewMessageFormatter(updateMsg.From.ID)

	courseName := Standartize(updateMsg.Text)
	sections, exists := h.CoursesRepo.GetCourse(courseName)
	slog.Debug("Received course code", "courseName", courseName, "exists", exists)

	if !exists {
		return mf.ImmediateMessage(fmt.Sprintf("Course '*%s*' not found", courseName))
	}

	return mf.ImmediateMessage(h.beatify(courseName, sections))
}

func (h *MessageHandler) HandleCommandUnknown(cmd *tapi.Message) []tapi.MessageConfig {
	mf := NewMessageFormatter(cmd.From.ID)

	return mf.ImmediateMessage(fmt.Sprintf("⚠️ Invalid command \\(/%s\\)", cmd.Command()))
}

func (h *MessageHandler) HandleCommandStart(cmd *tapi.Message) []tapi.MessageConfig {
	mf := NewMessageFormatter(cmd.From.ID)

	return mf.ImmediateMessage(h.welcomeText)
}

func (h *MessageHandler) HandleCallback(callback *tapi.CallbackQuery) []tapi.MessageConfig {
	mf := NewMessageFormatter(callback.From.ID)

	args := strings.Split(callback.Data, "_")
	if len(args) != 2 {
		return mf.ImmediateMessage("⚠️ Invalid callback data format")
	}
	action := args[0]
	course := args[1]

	switch action {
	case "show":
		sections, exists := h.CoursesRepo.GetCourse(course)
		if !exists {
			mf.AddString(fmt.Sprintf("Course '*%s*' not found", course))
		} else {
			mf.AddString(h.beatify(course, sections))
		}
	case "unsubscribe":
		err := h.CourseSubscriptionRepository.UnSubscribe(callback.From.ID, course)
		if err != nil {
			mf.AddString("Failed to unsubscribe from the course\\. Please try again\\.")
		} else {
			mf.AddString(fmt.Sprintf("✅ Successfully unsubscribed from *%s*", course))
		}
	default:
		mf.AddString("⚠️ Unknown action in callback data")
	}

	return mf.messages
}
