package handlers

import (
	"fmt"
	"log/slog"
	"slices"
	"strings"

	"github.com/TheTeemka/telegram_bot_cources/internal/repositories"
	tapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type MessageHandler struct {
	StateRepo                    repositories.StateRepository
	CoursesRepo                  *repositories.CourseRepository
	CourseSubscriptionRepository repositories.CourseSubscriptionRepository
	AdminID                      []int64

	welcomeText string
}

func NewMessageHandler(adminID []int64, coursesRepo *repositories.CourseRepository,
	subscriptionRepo repositories.CourseSubscriptionRepository,
	stateRepo repositories.StateRepository) *MessageHandler {
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
		AdminID:     adminID,
		welcomeText: welcomeText,

		CoursesRepo:                  coursesRepo,
		StateRepo:                    stateRepo,
		CourseSubscriptionRepository: subscriptionRepo,
	}
}

func (h *MessageHandler) HandleUpdate(update tapi.Update) []tapi.MessageConfig {

	if update.CallbackQuery != nil {
		return h.HandleCallback(update.CallbackQuery)
	}

	if update.Message == nil || (len(h.AdminID) == 0 && slices.Contains(h.AdminID, update.Message.From.ID)) {
		return nil
	}

	if update.Message.IsCommand() {
		return h.HandleCommand(update.Message)
	}
	return h.HandleMessage(update.Message)
}

var knownCommands = []string{"start", "subscribe", "unsubscribe", "list"}

func (h *MessageHandler) HandleCommand(cmd *tapi.Message) []tapi.MessageConfig {
	mf := NewMessageFormatter(cmd.From.ID)
	if !slices.Contains(knownCommands, cmd.Command()) {
		return mf.ImmediateMessage("❌ Unknown command " + cmd.Command())
	}

	switch cmd.Command() {
	case "start":
		return mf.ImmediateMessage(h.welcomeText)
	case "list":
		return h.ListSubscriptions(cmd)
	}

	h.StateRepo.Upsert(cmd.From.ID, cmd.Command())
	switch cmd.Command() {
	case "subscribe":
		return mf.ImmediateMessage("Please provide a course code as in docs\\.\nFormat: `[Course Name] [Course Sections]`\\.\nExample: \\'PHYS 161 2L 3L\\' \\| \\'phys 161 2l 1plb 1l\\'\\.")
	case "unsubscribe":
		return mf.ImmediateMessage("Please provide a course code as in docs\\.\nFormat: `[Course Name]`\\.\nExample: \\'PHYS161\\'\\.")
	default:
		return h.HandleCommandUnknown(cmd)
	}
}

func (h *MessageHandler) HandleMessage(msg *tapi.Message) []tapi.MessageConfig {
	mf := NewMessageFormatter(msg.From.ID)

	state, err := h.StateRepo.GetState(msg.From.ID)
	if err != nil {
		slog.Error("Failed to get state for user", "user_id", msg.From.ID, "error", err)
		return mf.ImmediateMessage("⚠️ Failed to retrieve your state\\. Please try again later\\.")
	}

	err = h.StateRepo.Upsert(msg.From.ID, "")
	if err != nil {
		slog.Error("Failed to clear state for user", "user_id", msg.From.ID, "error", err)
		return mf.ImmediateMessage("⚠️ Failed to clear your state\\. Please try again later\\.")
	}

	switch state {
	case "":
		return h.HandleCourseCode(msg)
	case "start":
		return h.HandleCommandStart(msg)
	case "subscribe":
		return h.HandleSubscribe(msg)
	case "unsubscribe":
		return h.HandleUnsubscribe(msg)
	case "list":
		return h.ListSubscriptions(msg)
	default:
		return h.HandleCommandUnknown(msg) //TODO: panic
	}
}

func (h *MessageHandler) HandleSubscribe(cmd *tapi.Message) []tapi.MessageConfig {
	mf := NewMessageFormatter(cmd.From.ID)
	courseName, sectionNames, ok := h.parseCommandArguments(cmd.Text)
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
	slog.Debug("Parsing command arguments", "args", args)
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

	slog.Debug("Parsed course name", "courseName", courseName, "index", ind)
	var section []string
	for i := ind; i < len(fields); i++ {
		if !isDigit(fields[i][0]) {
			section[len(section)-1] += fields[i]
		} else {
			section = append(section, fields[i])
		}
	}
	slog.Debug("Parsing command arguments", "section", section)

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
	courseName := cmd.Text
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
			if section.Size >= section.Cap {
				sb.WriteString(fmt.Sprintf("•   ~%-10s %-7s \\(%d/%d\\)~\n", sub.Course, sub.Section, section.Size, section.Cap))
			} else {
				sb.WriteString(fmt.Sprintf("•   %-10s %-7s \\(%d/%d\\)\n", sub.Course, section.SectionName, section.Size, section.Cap))
			}
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

	return mf.ImmediateMessage(h.beatify(course))
}

func (h *MessageHandler) HandleCommandUnknown(cmd *tapi.Message) []tapi.MessageConfig {
	mf := NewMessageFormatter(cmd.From.ID)

	return mf.ImmediateMessage(fmt.Sprintf("⚠️ Unknown State \\(/%s\\)", cmd.Command()))
}

func (h *MessageHandler) HandleCommandStart(cmd *tapi.Message) []tapi.MessageConfig {
	mf := NewMessageFormatter(cmd.From.ID)

	return mf.ImmediateMessage(h.welcomeText)
}

func (h *MessageHandler) HandleCallback(callback *tapi.CallbackQuery) []tapi.MessageConfig {
	mf := NewMessageFormatter(callback.From.ID)

	args := strings.Split(callback.Data, "_")
	if len(args) != 2 {
		return mf.ImmediateMessage("⚠️ Invalid callback data format")
	}
	action := args[0]
	courseAbbr := args[1]

	switch action {
	case "show":
		course, exists := h.CoursesRepo.GetCourse(courseAbbr)
		if !exists {
			mf.AddString(fmt.Sprintf("Course '*%s*' not found", courseAbbr))
		} else {
			mf.AddString(h.beatify(course))
		}
	case "unsubscribe":
		err := h.CourseSubscriptionRepository.UnSubscribe(callback.From.ID, courseAbbr)
		if err != nil {
			mf.AddString("Failed to unsubscribe from the course\\. Please try again\\.")
		} else {
			mf.AddString(fmt.Sprintf("✅ Successfully unsubscribed from *%s*", courseAbbr))
		}
	default:
		mf.AddString("⚠️ Unknown action in callback data")
	}

	return mf.messages
}

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
