package handlers

import (
	"fmt"
	"log/slog"

	"github.com/TheTeemka/telegram_bot_cources/internal/repositories"
	tapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type MessageHandler struct {
	CoursesRepo *repositories.CourseRepository
	AdminID     int64

	welcomeText string
}

func NewMessageHandler(adminID int64, coursesRepo *repositories.CourseRepository) *MessageHandler {

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
	}
}

func (h *MessageHandler) HandleUpdate(update tapi.Update) tapi.Chattable {
	if update.Message == nil {
		return nil
	}

	if update.Message.From.ID != h.AdminID {
		return nil
	}

	if update.Message.IsCommand() {
		return h.HandleCommand(update.Message)
	} else {
		return h.HandleCourseCode(update.Message)
	}
}

func (h *MessageHandler) HandleCourseCode(updateMsg *tapi.Message) tapi.Chattable {
	courseName := Standartize(updateMsg.Text)
	sections, exists := h.CoursesRepo.GetCourse(courseName)
	slog.Debug("Received course code", "courseName", courseName, "exists", exists)

	if !exists {
		msg := tapi.NewMessage(updateMsg.Chat.ID, fmt.Sprintf("Cource '%s' not found", courseName))
		return msg
	}

	msg := tapi.NewMessage(updateMsg.Chat.ID,
		h.beatify(courseName, sections))
	msg.ParseMode = "MarkdownV2"

	return msg

}
func (h *MessageHandler) HandleCommand(commandMsg *tapi.Message) tapi.Chattable {
	cmd := commandMsg.Command()
	chatID := commandMsg.Chat.ID
	slog.Debug("Received command", "command", cmd)

	var msg tapi.MessageConfig
	switch commandMsg.Command() {
	case "start":
		msg = tapi.NewMessage(chatID, h.welcomeText)
	default:
		msg = tapi.NewMessage(chatID, fmt.Sprintf("⚠️ Invalid command \\(/%s\\)", cmd))
	}

	msg.ParseMode = "MarkdownV2"
	return msg
}
