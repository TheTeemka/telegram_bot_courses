package handlers

import (
	tapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Response struct {
	Messages []string
}

func (r *Response) ToMessages(chatID int64) []tapi.MessageConfig {
	msg := make([]tapi.MessageConfig, len(r.Messages))
	for i, text := range r.Messages {
		msg[i] = tapi.NewMessage(chatID, text)
		msg[i].ParseMode = tapi.ModeMarkdownV2
	}

	return msg
}
func NewResponse(messages ...string) *Response {
	return &Response{
		Messages: messages,
	}
}
