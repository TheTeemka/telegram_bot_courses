package handlers

import "errors"

var (
	ErrNotEnoughParams = errors.New("not enough params provided")
	InvalidParams      = errors.New("invalid parameters provided for the command")
)
