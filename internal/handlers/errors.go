package handlers

import "errors"

var (
	ErrNotEnoughParams = errors.New("not enough fields in command arguments")
	InvalidParams      = errors.New("invalid parameters provided for the command")
)
