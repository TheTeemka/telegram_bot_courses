package telegram

import (
	"fmt"
	"log/slog"

	"github.com/TheTeemka/telegram_bot_cources/internal/repositories"

	tapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TelegramHandler struct {
	courseRepo *repositories.CourseRepository
}

func (bot *TelegramBot) HandleUpdate(update tapi.Update) {
	if update.Message == nil {
		return
	}

	if update.Message.From.ID != bot.AdminID {
		return
	}

	if update.Message.IsCommand() {
		bot.HandleCommand(update.Message)
	} else {
		bot.HandleCourseCode(update.Message)
	}
}

func (bot *TelegramBot) HandleCourseCode(updateMsg *tapi.Message) {
	courseName := Standartize(updateMsg.Text)
	sections, exists := bot.CoursesRepo.GetCourse(courseName)
	slog.Debug("Received course code", "courseName", courseName, "exists", exists)

	if !exists {
		msg := tapi.NewMessage(updateMsg.Chat.ID, fmt.Sprintf("Cource '%s' not found", courseName))
		_, err := bot.BotAPI.Send(msg)
		if err != nil {
			slog.Error("Failed to send message", "error", err)
		}
		return
	}

	msg := tapi.NewMessage(updateMsg.Chat.ID,
		bot.beatify(courseName, sections))
	msg.ParseMode = "MarkdownV2"

	_, err := bot.BotAPI.Send(msg)
	if err != nil {
		slog.Error("Failed to send message", "error", err)
	}

}
func (bot *TelegramBot) HandleCommand(commandMsg *tapi.Message) {
	cmd := commandMsg.Command()
	chatID := commandMsg.Chat.ID
	slog.Debug("Received command", "command", cmd)

	var msg tapi.MessageConfig
	switch commandMsg.Command() {
	case "start":
		msg = tapi.NewMessage(chatID, bot.welcomeText)
	default:
		msg = tapi.NewMessage(chatID, fmt.Sprintf("⚠️ Invalid command \\(/%s\\)", cmd))
	}

	msg.ParseMode = "MarkdownV2"
	_, err := bot.BotAPI.Send(msg)
	if err != nil {
		slog.Error("Failed to send command response",
			"command", cmd,
			"error", err,
			"chat_id", chatID)
	}
}
