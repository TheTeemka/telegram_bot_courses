package telegramfmt

import (
	"fmt"

	tapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type MessageFormatter struct {
	chatID   int64
	messages []tapi.Chattable
}

const ParseMode = tapi.ModeHTML

func NewMessageFormatter(chatID int64) *MessageFormatter {
	return &MessageFormatter{
		chatID: chatID,
	}
}

func (mf *MessageFormatter) Messages() []tapi.Chattable {
	return mf.messages
}

func (mf *MessageFormatter) Add(chattable tapi.Chattable) {
	mf.messages = append(mf.messages, chattable)
}

func (mf *MessageFormatter) AddString(text string) {
	msg := tapi.NewMessage(mf.chatID, text)
	msg.ParseMode = ParseMode
	mf.messages = append(mf.messages, msg)
}

func (mf *MessageFormatter) ImmediateMessage(text string) []tapi.Chattable {
	mf.AddString(text)
	return mf.messages
}

func (mf *MessageFormatter) AddKeyboardToLastMessage(keyboard [][]tapi.InlineKeyboardButton) {
	if len(mf.messages) == 0 {
		panic("No messages to add inline keyboard markup to")
	}

	msgCfg := mf.messages[len(mf.messages)-1].(tapi.MessageConfig)
	msgCfg.ReplyMarkup = tapi.NewInlineKeyboardMarkup(keyboard...)
	mf.messages[len(mf.messages)-1] = msgCfg
}

func (mf *MessageFormatter) ImmediateNotFoundCourse(courseAbbr string, action string) []tapi.Chattable {
	msg := tapi.NewMessage(mf.chatID, fmt.Sprintf("❌ Course <b>%s</b> not found %s",
		tapi.EscapeText(ParseMode, courseAbbr),
		tapi.EscapeText(ParseMode, action),
	))
	msg.ParseMode = ParseMode
	return []tapi.Chattable{msg}
}

func (mf *MessageFormatter) AddNotFoundCourse(courseAbbr string) {
	msg := tapi.NewMessage(mf.chatID, fmt.Sprintf("❌ Course <b>%s</b> not found", tapi.EscapeText(ParseMode, courseAbbr)))
	msg.ParseMode = ParseMode
	mf.messages = append(mf.messages, msg)
}

func (mf *MessageFormatter) ImmediateNotFoundCourseSection(courseAbbr string, section string, action string) []tapi.Chattable {
	msg := tapi.NewMessage(mf.chatID, fmt.Sprintf("❌ Course <b>%s</b> Section <b>%s</b> not found %s",
		tapi.EscapeText(ParseMode, courseAbbr),
		tapi.EscapeText(ParseMode, section),
		tapi.EscapeText(ParseMode, action),
	))
	msg.ParseMode = ParseMode
	return []tapi.Chattable{msg}
}

func (mf *MessageFormatter) AddNotFoundCourseSection(courseAbbr string, section string) {
	msg := tapi.NewMessage(mf.chatID, fmt.Sprintf("❌ Course <b>%s</b> Section <b>%s</b> not found",
		tapi.EscapeText(ParseMode, courseAbbr),
		tapi.EscapeText(ParseMode, section),
	))
	msg.ParseMode = ParseMode
	mf.messages = append(mf.messages, msg)
}

func (mf *MessageFormatter) UnsubscribeOrIgnoreSection(courseAbbr string, section string) {
	ignore := "delete"
	unsubscribe := fmt.Sprintf("unsubscribe_%s_%s;delete", courseAbbr, section)
	mf.AddKeyboardToLastMessage([][]tapi.InlineKeyboardButton{
		{
			{Text: "Ignore", CallbackData: &ignore},
			{Text: "Unsubscribe", CallbackData: &unsubscribe},
		},
	})
}

func (mf *MessageFormatter) UnsubscribeOrIgnoreCourse(courseAbbr string) {
	ignore := "delete"
	unsubscribe := fmt.Sprintf("unsubscribe_%s;delete", courseAbbr)
	mf.AddKeyboardToLastMessage([][]tapi.InlineKeyboardButton{
		{
			{Text: "Ignore", CallbackData: &ignore},
			{Text: "Unsubscribe", CallbackData: &unsubscribe},
		},
	})
}
