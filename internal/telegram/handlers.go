package telegram

import (
	"fmt"
	"log/slog"

	tapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (bot *TelegramBot) HandleCourceCode(updateMsg *tapi.Message) {
	courceName := Standartize(updateMsg.Text)
	sections, exists := bot.CourcesRepo.GetCourse(courceName)
	slog.Debug("Received cource code", "courceName", courceName, "exists", exists)

	if !exists {
		msg := tapi.NewMessage(updateMsg.Chat.ID, fmt.Sprintf("Cource '%s' not found", courceName))
		_, err := bot.BotAPI.Send(msg)
		if err != nil {
			slog.Error("Failed to send message", "error", err)
		}
		return
	}

	msg := tapi.NewMessage(updateMsg.Chat.ID,
		bot.beatify(courceName, sections))
	msg.ParseMode = "MarkdownV2"

	_, err := bot.BotAPI.Send(msg)
	if err != nil {
		slog.Error("Failed to send message", "error", err)
	}

}

func (b *TelegramBot) HandleCommand(commandMsg *tapi.Message) {
	cmd := commandMsg.Command()
	chatID := commandMsg.Chat.ID
	slog.Debug("Received command", "command", cmd)

	var msg tapi.MessageConfig
	switch commandMsg.Command() {
	case "start":
		msg = tapi.NewMessage(chatID, "Welcome to the Cource Bot!\n")
	default:
		msg = tapi.NewMessage(chatID, fmt.Sprintf("invalid command(/%s)", cmd))
	}

	_, err := b.BotAPI.Send(msg)
	if err != nil {
		slog.Error("Failed to send start message", "error", err, "chat_id", chatID)
	}

}
