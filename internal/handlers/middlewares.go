package handlers

import (
	"slices"

	tapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type handler func(msg *tapi.Message) []tapi.MessageConfig

func AuthAdmin(authGroup []int64, next handler) handler {
	return func(msg *tapi.Message) []tapi.MessageConfig {
		if len(authGroup) == 0 && slices.Contains(authGroup, msg.From.ID) {
			return nil
		}

		return next(msg)
	}
}
