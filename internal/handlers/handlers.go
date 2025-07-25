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
			"Simply send me a course code \\(e\\.g\\. *PHYS 161*\\) to get:\n"+
			"• Current enrollment numbers\n"+
			"• Available seats\n"+
			"• Section details\n\n"+
			"Also provides opportunity to track course status by subscription system with notifications\n\n"+
			"_Updates every 60/30/15/5 minutes \n\\[The closer to registration the more frequent updates will be\\]_",
		coursesRepo.SemesterName)

	return &MessageHandler{
		AdminID:     adminID,
		welcomeText: welcomeText,

		CoursesRepo:                  coursesRepo,
		StateRepo:                    stateRepo,
		CourseSubscriptionRepository: subscriptionRepo,
	}
}

func (h *MessageHandler) HandleUpdate(update tapi.Update) []tapi.Chattable {
	if update.CallbackQuery != nil {
		return h.HandleCallback(update.CallbackQuery)
	}

	if update.Message == nil {
		return nil
	}

	if update.Message.IsCommand() {
		return AuthAdmin(h.AdminID, h.HandleCommand)(update.Message)
	}
	return AuthAdmin(h.AdminID, h.HandleMessage)(update.Message)
}

var knownCommands = []string{"start", "subscribe", "unsubscribe", "list", "gatekeep", "donate"}

func (h *MessageHandler) CommandsList() tapi.SetMyCommandsConfig {
	return tapi.NewSetMyCommands(
		tapi.BotCommand{Command: "start", Description: "Start the bot"},
		tapi.BotCommand{Command: "subscribe", Description: "Subscribe to a course"},
		tapi.BotCommand{Command: "unsubscribe", Description: "Unsubscribe from a course"},
		tapi.BotCommand{Command: "list", Description: "List your subscriptions"},
		tapi.BotCommand{Command: "gatekeep", Description: "gatekeep your course and section of choice"},
		tapi.BotCommand{Command: "donate", Description: "Donate to the bot"},
	)
}
func (h *MessageHandler) HandleCommand(cmd *tapi.Message) []tapi.Chattable {
	mf := NewMessageFormatter(cmd.From.ID)
	if !slices.Contains(knownCommands, cmd.Command()) {
		return mf.ImmediateMessage("❌ Unknown command " + cmd.Command())
	}

	switch cmd.Command() {
	case "start":
		return mf.ImmediateMessage(h.welcomeText)
	case "list":
		return h.ListSubscriptions(cmd)
	case "gatekeep":
		return mf.ImmediateMessage("\\[Still in Development\\] \nFor a totally modest fee of 10 doners, you can unleash your inner gatekeeper and accidentally block others from registering for your dream courses\\. Will it work\\? Who knows\\! Do we offer refunds\\? Absolutely not\\.")
	case "donate":
		return mf.ImmediateMessage("\\[Still in Development\\] \n Toss a coin to your humble bot, O student of fate, \nWhen rivals draw near, and the registration deadline won’t wait\\.\nA humble donation, a whisper, a nudge,\nTo tilt odds in your favor in timetable wars")
	}

	h.StateRepo.Upsert(cmd.From.ID, cmd.Command())
	switch cmd.Command() {
	case "subscribe":
		return mf.ImmediateMessage("Please provide a course abbr and section as in docs\\.\nFormat: `[Course Name] [Course Sections]`\\.\nExample: \\'PHYS 161 2L 1PLB 2R 2r 3plb 3L\\' \\.")
	case "unsubscribe":
		return mf.ImmediateMessage("Please provide a course abbr as in docs\\.\nFormat: `[Course Name]`\\.\nExample: \\'PHYS161\\'\\.")
	default:
		return h.HandleCommandUnknown(cmd)
	}
}

func (h *MessageHandler) HandleMessage(msg *tapi.Message) []tapi.Chattable {
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

func (h *MessageHandler) HandleSubscribe(cmd *tapi.Message) []tapi.Chattable {
	mf := NewMessageFormatter(cmd.From.ID)
	courseName, sectionNames, ok := h.parseCommandArguments(cmd.Text)
	if !ok {
		return mf.InvalidCourseCode(cmd.Text)
	}

	if _, exists := h.CoursesRepo.GetCourse(courseName); !exists {
		return mf.NotFoundCourse(courseName)
	}

	err := h.CourseSubscriptionRepository.Subscribe(cmd.From.ID, courseName, sectionNames)
	if err != nil {
		slog.Error("Failed to subscribe",
			"error", err,
			"user_id", cmd.From.ID,
			"course", courseName)
		return mf.ImmediateMessage("⚠️ Failed to subscribe to the course\\. Please try again\\.")
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

	slog.Debug("Parsed course name", "courseName", courseName, "index", ind)
	if ind == len(fields) {
		return "", nil, false
	}

	var section []string
	for i := ind; i < len(fields); i++ {
		if !isDigit(fields[i][0]) {
			if len(section) == 0 {
				return "", nil, false
			}
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
	slog.Debug("Parsing command arguments", "section", section)
	return StandartizeCourseName(courseName), section, true //TODO: Section ToUpper

}

func isDigit(b byte) bool {
	return '0' <= rune(b) && rune(b) <= '9'
}

func (h *MessageHandler) HandleUnsubscribe(cmd *tapi.Message) []tapi.Chattable {
	mf := NewMessageFormatter(cmd.From.ID)
	courseName := cmd.Text
	if courseName == "" {
		return mf.InvalidCourseCode(courseName)
	}

	if _, exists := h.CoursesRepo.GetCourse(courseName); !exists {
		return mf.NotFoundCourse(courseName)
	}

	err := h.CourseSubscriptionRepository.UnSubscribe(cmd.From.ID, courseName)
	if err != nil {
		slog.Error("Failed to subscribe",
			"error", err,
			"user_id", cmd.From.ID,
			"course", courseName)
		return mf.ImmediateMessage("⚠️ Failed to subscribe to the course\\. Please try again\\.")
	}

	return mf.ImmediateMessage(fmt.Sprintf("✅ Successfully unsubscribed from *%s*", courseName))
}

func (h *MessageHandler) Clear(cmd *tapi.Message) []tapi.Chattable {
	mf := NewMessageFormatter(cmd.From.ID)

	err := h.CourseSubscriptionRepository.ClearSubscriptions(cmd.From.ID)
	if err != nil {
		slog.Error("Failed to subscribe",
			"error", err,
			"user_id", cmd.From.ID)
		return mf.ImmediateMessage("⚠️  Failed to subscribe to the course\\. Please try again\\.")
	}

	return mf.ImmediateMessage(("✅ Successfully cleared"))
}

func (h *MessageHandler) ListSubscriptions(cmd *tapi.Message) []tapi.Chattable {
	mf := NewMessageFormatter(cmd.From.ID)
	subs, err := h.CourseSubscriptionRepository.GetSubscriptions(cmd.From.ID)
	if err != nil {
		slog.Error("⚠️ Failed to get subscriptions", "err", err)
		return mf.ImmediateMessage("⚠️ Failed to retrieve your subscriptions\\. Please try again later\\.")
	}
	if len(subs) == 0 {
		return mf.ImmediateMessage("⚠️ You haven't subscribed to any courses yet\\.")
	}

	var sb strings.Builder
	sb.WriteString("Your subscriptions:\n")
	for _, sub := range subs {
		_, exists := h.CoursesRepo.GetCourse(sub.Course)
		if !exists {
			mf.AddString(fmt.Sprintf("❌ Course '*%s*' not found\n", sub.Course))
			ignore := "delete"
			unsubscribe := fmt.Sprintf("unsubscribe_%s;delete", sub.Course)
			mf.AddKeyboardToLastMessage([][]tapi.InlineKeyboardButton{
				{
					{Text: "Ignore", CallbackData: &ignore},
					{Text: "Unsubscribe", CallbackData: &unsubscribe},
				},
			})
			continue
		}

		section, exists := h.CoursesRepo.GetSection(sub.Course, sub.Section)
		if !exists {
			mf.AddString(fmt.Sprintf("❌ Course '*%s*' Section '*%s*' not found\n", sub.Course, sub.Section))
			ignore := "delete"
			unsubscribe := fmt.Sprintf("unsubscribe_%s_%s;delete", sub.Course, sub.Section)
			mf.AddKeyboardToLastMessage([][]tapi.InlineKeyboardButton{
				{
					{Text: "Ignore", CallbackData: &ignore},
					{Text: "Unsubscribe", CallbackData: &unsubscribe},
				},
			})
		} else {
			sb.WriteString(formatCourseSection(sub.Course, sub.Section, section.Size, section.Cap))
		}
	}
	mf.AddString(sb.String())
	return mf.Messages()
}

func (h *MessageHandler) HandleCourseCode(updateMsg *tapi.Message) []tapi.Chattable {
	mf := NewMessageFormatter(updateMsg.From.ID)

	courseAbbr := StandartizeCourseName(updateMsg.Text)
	course, exists := h.CoursesRepo.GetCourse(courseAbbr)
	slog.Debug("Received course code", "courseName", courseAbbr, "exists", exists)

	if !exists {
		return mf.NotFoundCourse(courseAbbr)
	}

	return mf.ImmediateMessage(formatCourseInDetails(course, h.CoursesRepo.SemesterName, h.CoursesRepo.LastTimeParsed))
}

func (h *MessageHandler) HandleCommandUnknown(cmd *tapi.Message) []tapi.Chattable {
	mf := NewMessageFormatter(cmd.From.ID)

	return mf.ImmediateMessage(fmt.Sprintf("⚠️ Unknown State \\(/%s\\)", cmd.Command()))
}

func (h *MessageHandler) HandleCommandStart(cmd *tapi.Message) []tapi.Chattable {
	mf := NewMessageFormatter(cmd.From.ID)

	return mf.ImmediateMessage(h.welcomeText)
}

func (h *MessageHandler) HandleCallback(callback *tapi.CallbackQuery) []tapi.Chattable {
	mf := NewMessageFormatter(callback.From.ID)

	cmds := strings.Split(callback.Data, ";")

	for _, cmd := range cmds {
		args := strings.Split(cmd, "_")

		switch args[0] {
		case "delete":
			deleteCFG := tapi.NewDeleteMessage(callback.From.ID, callback.Message.MessageID)
			mf.Add(deleteCFG)
		case "unsubscribe":
			if len(args) == 2 {
				err := h.CourseSubscriptionRepository.UnSubscribe(
					callback.From.ID, args[1],
				)
				if err != nil {
					slog.Error("Failed to unsubscribe", "error", err, "course", args[1])
				}
			} else if len(args) == 3 {
				err := h.CourseSubscriptionRepository.UnSubscribeSection(
					callback.From.ID, args[1], args[2],
				)
				if err != nil {
					slog.Error("Failed to unsubscribe", "error", err, "course", args[1], "section", args[2])
				}
			} else {
				slog.Error("Invalid ignore command format", "command", cmd)
				continue
			}
		}
	}

	return mf.messages
}
