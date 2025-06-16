package handlers

import (
	tapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type MessageFormatter struct {
	chatID   int64
	messages []tapi.MessageConfig
}

func NewMessageFormatter(chatID int64) *MessageFormatter {
	return &MessageFormatter{
		chatID: chatID,
	}
}

func (mf *MessageFormatter) Messages() []tapi.MessageConfig {
	return mf.messages
}

func (mf *MessageFormatter) AddString(text string) {
	msg := tapi.NewMessage(mf.chatID, text)
	msg.ParseMode = tapi.ModeMarkdownV2
	mf.messages = append(mf.messages, msg)
}

func (mf *MessageFormatter) ImmediateMessage(text string) []tapi.MessageConfig {
	mf.AddString(text)
	return mf.messages
}

func (mf *MessageFormatter) AddKeyboardToLastMessage(keyboard [][]tapi.InlineKeyboardButton) {
	if len(mf.messages) == 0 {
		panic("No messages to add inline keyboard markup to")
	}
	mf.messages[len(mf.messages)-1].ReplyMarkup = tapi.NewInlineKeyboardMarkup(keyboard...)
}
