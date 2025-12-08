package handlers

import (
	"slices"

	tapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type handler func(msg *tapi.Message) []tapi.Chattable

func AuthAdmin(authGroup []int64, next handler) handler {
	return func(msg *tapi.Message) []tapi.Chattable {
		if len(authGroup) == 0 && slices.Contains(authGroup, msg.From.ID) {
			return nil
		}

		return next(msg)
	}
}

func AuthAllowed(allowedGroup []int64, next handler) handler {
	return func(msg *tapi.Message) []tapi.Chattable {
		if len(allowedGroup) == 0 && slices.Contains(allowedGroup, msg.From.ID) {
			return nil
		}

		return next(msg)
	}
}
