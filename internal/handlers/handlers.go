package handlers

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/TheTeemka/telegram_bot_cources/internal/repositories"
	tapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type MessageHandler struct {
	CoursesRepo                  *repositories.CourseRepository
	CourseSubscriptionRepository repositories.CourseSubscriptionRepository
	AdminID                      []int64

	welcomeText string
}

func NewMessageHandler(adminID []int64, coursesRepo *repositories.CourseRepository, subscriptionRepo repositories.CourseSubscriptionRepository) *MessageHandler {
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

		CourseSubscriptionRepository: subscriptionRepo,
	}
}

func (h *MessageHandler) HandleUpdate(update tapi.Update) []tapi.MessageConfig {
	// if update.CallbackQuery != nil {
	// 	return h.HandleCallback(update.CallbackQuery)
	// }

	if update.Message == nil {
		return nil
	}

	if update.Message.IsCommand() {
		return AuthAdmin(h.AdminID, h.HandleCommand)(update.Message)
	}
	return AuthAdmin(h.AdminID, h.HandleCourseCode)(update.Message)
}

func (h *MessageHandler) HandleCommand(cmd *tapi.Message) []tapi.MessageConfig {
	switch cmd.Command() {
	case "start":
		return h.HandleCommandStart(cmd)
	case "subscribe":
		return h.HandleSubscribe(cmd)
	case "unsubscribe":
		return h.HandleUnsubscribe(cmd)
	case "clear":
		return h.Clear(cmd)
	case "list":
		return h.ListSubscriptions(cmd)
	default:
		return h.HandleCommandUnknown(cmd)
	}
}

func (h *MessageHandler) HandleSubscribe(cmd *tapi.Message) []tapi.MessageConfig {
	mf := NewMessageFormatter(cmd.From.ID)
	courseName, sectionNames, ok := h.parseCommandArguments(cmd.CommandArguments())
	if !ok {
		return mf.ImmediateMessage("Please provide a course code\\. Example: `/subscribe [Course Name] [Course Sections].`\\.")
	}

	if _, exists := h.CoursesRepo.GetCourse(courseName); !exists {
		return mf.ImmediateMessage(fmt.Sprintf("Course *%s* not found", courseName))
	}

	err := h.CourseSubscriptionRepository.Subscribe(cmd.From.ID, courseName, sectionNames)
	if err != nil {
		slog.Error("Failed to subscribe",
			"error", err,
			"user_id", cmd.From.ID,
			"course", courseName)
		return mf.ImmediateMessage("Failed to subscribe to the course\\. Please try again\\.")
	}

	return mf.ImmediateMessage(fmt.Sprintf("✅ Successfully subscribed to *%s \\(%s\\)*", courseName, strings.Join(sectionNames, ", ")))
}

func (h *MessageHandler) parseCommandArguments(args string) (string, []string, bool) {
	fields := strings.Fields(args)
	if len(fields) < 2 {
		return "", nil, false
	}
	courseName := fields[0]
	ind := 1
	if !isDigit(courseName[len(courseName)-1]) && isDigit(fields[1][0]) {
		courseName += fields[1]
		ind++
	}

	if ind == len(fields) {
		return "", nil, false
	}

	var section []string
	for i := ind; i < len(fields); i++ {
		if !isDigit(fields[i][0]) {
			section[len(section)-1] += fields[i]
		} else {
			section = append(section, fields[i])
		}
	}

	for i := range section {
		sec, ok := StandartizeSectionName(section[i], h.CoursesRepo.SectionAbbrList)
		if !ok {
			return "", nil, false
		}
		section[i] = sec
	}
	return StandartizeCourseName(courseName), section, true //TODO: Section ToUpper

}

func isDigit(b byte) bool {
	return '0' <= rune(b) && rune(b) <= '9'
}

func (h *MessageHandler) HandleUnsubscribe(cmd *tapi.Message) []tapi.MessageConfig {
	mf := NewMessageFormatter(cmd.From.ID)
	courseName := cmd.CommandArguments()
	if courseName == "" {
		return mf.ImmediateMessage("Please provide a course code\\. Example: `/unsubscribe [Course Name].`\\.")
	}

	if _, exists := h.CoursesRepo.GetCourse(courseName); !exists {
		return mf.ImmediateMessage(fmt.Sprintf("Course *%s* not found", courseName))
	}

	err := h.CourseSubscriptionRepository.UnSubscribe(cmd.From.ID, courseName)
	if err != nil {
		slog.Error("Failed to subscribe",
			"error", err,
			"user_id", cmd.From.ID,
			"course", courseName)
		return mf.ImmediateMessage("Failed to subscribe to the course\\. Please try again\\.")
	}

	return mf.ImmediateMessage(fmt.Sprintf("✅ Successfully unsubscribed from *%s*", courseName))
}

func (h *MessageHandler) Clear(cmd *tapi.Message) []tapi.MessageConfig {
	mf := NewMessageFormatter(cmd.From.ID)

	err := h.CourseSubscriptionRepository.ClearSubscriptions(cmd.From.ID)
	if err != nil {
		slog.Error("Failed to subscribe",
			"error", err,
			"user_id", cmd.From.ID)
		return mf.ImmediateMessage("Failed to subscribe to the course\\. Please try again\\.")
	}

	return mf.ImmediateMessage(("✅ Successfully cleared"))
}

func (h *MessageHandler) ListSubscriptions(cmd *tapi.Message) []tapi.MessageConfig {
	mf := NewMessageFormatter(cmd.From.ID)

	subs, err := h.CourseSubscriptionRepository.GetSubscriptions(cmd.From.ID)
	if err != nil {
		slog.Error("Failed to get subscriptions", "err", err)
		return mf.ImmediateMessage("Failed to retrieve your subscriptions\\. Please try again later\\.")
	}
	if len(subs) == 0 {
		return mf.ImmediateMessage("You haven't subscribed to any courses yet\\.")
	}

	var sb strings.Builder
	sb.WriteString("Your subscriptions:\n")
	for _, sub := range subs {
		_, exists := h.CoursesRepo.GetCourse(sub.Course)
		if !exists {
			sb.WriteString(fmt.Sprintf("❌ Course '*%s*' not found\n", sub.Course))
			continue
		}

		section, exists := h.CoursesRepo.GetSection(sub.Course, sub.Section)
		if !exists {
			sb.WriteString(fmt.Sprintf("❌ Course '*%s*' Section '*%s*' not found\n", sub.Course, sub.Section))
		} else {
			sb.WriteString(formatCourseSection(sub.Course, sub.Section, section.Size, section.Cap))
		}
	}

	return mf.ImmediateMessage(sb.String())
}

func (h *MessageHandler) HandleCourseCode(updateMsg *tapi.Message) []tapi.MessageConfig {
	mf := NewMessageFormatter(updateMsg.From.ID)

	courseAbbr := StandartizeCourseName(updateMsg.Text)
	course, exists := h.CoursesRepo.GetCourse(courseAbbr)
	slog.Debug("Received course code", "courseName", courseAbbr, "exists", exists)

	if !exists {
		return mf.ImmediateMessage(fmt.Sprintf("Course '*%s*' not found", courseAbbr))
	}

	return mf.ImmediateMessage(formatCourseInDetails(course, h.CoursesRepo.SemesterName, h.CoursesRepo.LastTimeParsed))
}

func (h *MessageHandler) HandleCommandUnknown(cmd *tapi.Message) []tapi.MessageConfig {
	mf := NewMessageFormatter(cmd.From.ID)

	return mf.ImmediateMessage(fmt.Sprintf("⚠️ Invalid command \\(/%s\\)", cmd.Command()))
}

func (h *MessageHandler) HandleCommandStart(cmd *tapi.Message) []tapi.MessageConfig {
	mf := NewMessageFormatter(cmd.From.ID)

	return mf.ImmediateMessage(h.welcomeText)
}

// func (h *MessageHandler) HandleCallback(callback *tapi.CallbackQuery) []tapi.MessageConfig {
// 	mf := NewMessageFormatter(callback.From.ID)

// 	args := strings.Split(callback.Data, "_")
// 	if len(args) != 2 {
// 		return mf.ImmediateMessage("⚠️ Invalid callback data format")
// 	}
// 	action := args[0]
// 	courseAbbr := args[1]

// 	switch action {
// 	case "show":
// 		course, exists := h.CoursesRepo.GetCourse(courseAbbr)
// 		if !exists {
// 			mf.AddString(fmt.Sprintf("Course '*%s*' not found", courseAbbr))
// 		} else {
// 			mf.AddString(h.beatify(course))
// 		}
// 	case "unsubscribe":
// 		err := h.CourseSubscriptionRepository.UnSubscribe(callback.From.ID, courseAbbr)
// 		if err != nil {
// 			mf.AddString("Failed to unsubscribe from the course\\. Please try again\\.")
// 		} else {
// 			mf.AddString(fmt.Sprintf("✅ Successfully unsubscribed from *%s*", courseAbbr))
// 		}
// 	default:
// 		mf.AddString("⚠️ Unknown action in callback data")
// 	}

// 	return mf.messages
// }

// func (h *MessageHandler) ListSubscriptions(cmd *tapi.Message) []tapi.MessageConfig {
// 	mf := NewMessageFormatter(cmd.From.ID)
// 	subs := h.CourseSubscriptionRepository.GetSubscriptions(cmd.From.ID)
// 	if len(subs) == 0 {
// 		return mf.ImmediateMessage("You haven't subscribed to any courses yet\\.")
// 	}

// 	mf.AddString("*Your subscriptions:*\n")
// 	for _, sub := range subs {
// 		mf.AddString(fmt.Sprintf("•  *%s %s*", sub.Course, sub.Section))
// 		callbackSub := fmt.Sprintf("show_%s_%s", sub.Course, sub.Section)
// 		callbackUnSub := fmt.Sprintf("unsubscribe_%s_%s", sub.Course, sub.Section)

// 		mf.AddKeyboardToLastMessage([][]tapi.InlineKeyboardButton{
// 			{
// 				{Text: "ℹ️ Show", CallbackData: &callbackSub},
// 				{Text: "❌ Unsubscribe", CallbackData: &callbackUnSub},
// 			},
// 		})
// 	}

// 	return mf.Messages()
// }
