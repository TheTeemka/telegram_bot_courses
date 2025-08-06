package handlers

import (
	"bytes"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"strings"

	"github.com/TheTeemka/telegram_bot_cources/internal/repositories"
	"github.com/TheTeemka/telegram_bot_cources/internal/telegramfmt"
	tapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type MessageHandler struct {
	BotAPI           *tapi.BotAPI
	StateRepo        repositories.StateRepository
	CoursesRepo      *repositories.CourseRepository
	SubscriptionRepo repositories.CourseSubscriptionRepository
	StatisticsRepo   *repositories.StatisticsRepository
	Private          bool
	AdminID          []int64

	welcomeText string
	faq         string
	KaspiCard   string
}

func NewMessageHandler(botAPI *tapi.BotAPI, adminID []int64, private bool, kaspiCard string,
	coursesRepo *repositories.CourseRepository,
	subscriptionRepo repositories.CourseSubscriptionRepository,
	stateRepo repositories.StateRepository,
	statisticsRepo *repositories.StatisticsRepository) *MessageHandler {

	return &MessageHandler{
		BotAPI:  botAPI,
		AdminID: adminID,
		Private: private,

		faq:         generateFAQText(),
		welcomeText: generateWelcomeText(coursesRepo.SemesterName),

		KaspiCard:        kaspiCard,
		CoursesRepo:      coursesRepo,
		StateRepo:        stateRepo,
		SubscriptionRepo: subscriptionRepo,
		StatisticsRepo:   statisticsRepo,
	}
}

func (h *MessageHandler) HandleUpdate(update tapi.Update) []tapi.Chattable {
	h.StatisticsRepo.AddOne("Total_Request_Number")
	if update.CallbackQuery != nil {
		return h.HandleCallback(update.CallbackQuery)
	}

	if update.Message == nil {
		return nil
	}

	if h.Private {
		if update.Message.IsCommand() {
			return AuthAdmin(h.AdminID, h.HandleCommand)(update.Message)
		}
		return AuthAdmin(h.AdminID, h.HandleMessage)(update.Message)
	}

	if update.Message.IsCommand() {
		return h.HandleCommand(update.Message)
	}
	return h.HandleMessage(update.Message)

}

var knownCommands = []string{"start", "subscribe", "unsubscribe", "list", "donate", "faq", "parsestat", "nextupdatetime"}

func (h *MessageHandler) CommandsList() tapi.SetMyCommandsConfig {
	return tapi.NewSetMyCommands(
		tapi.BotCommand{Command: "start", Description: "Start the bot"},
		tapi.BotCommand{Command: "subscribe", Description: "Subscribe to a course"},
		tapi.BotCommand{Command: "unsubscribe", Description: "Unsubscribe from a course"},
		tapi.BotCommand{Command: "list", Description: "List your subscriptions"},
		tapi.BotCommand{Command: "faq", Description: "Frequently Asked Questions"},
		// tapi.BotCommand{Command: "gatekeep", Description: "gatekeep your course and section of choice"},
		tapi.BotCommand{Command: "donate", Description: "Donate to the bot"},
		tapi.BotCommand{Command: "nextupdatetime", Description: "next sync time"},
	)
}

func (h *MessageHandler) HandleCommand(cmd *tapi.Message) []tapi.Chattable {
	slog.Debug("Handling command", "command", cmd.Command(), "text", cmd.Text, "username", cmd.From.UserName)
	mf := telegramfmt.NewMessageFormatter(cmd.From.ID)
	if !slices.Contains(knownCommands, cmd.Command()) {
		return mf.ImmediateMessage("❌ Unknown command " + cmd.Command())
	}

	h.StatisticsRepo.AddOne("command" + cmd.Command())
	h.StateRepo.Upsert(cmd.From.ID, "")
	switch cmd.Command() {
	case "start":
		return mf.ImmediateMessage(h.welcomeText)
	case "list":
		return h.ListSubscriptions(cmd)
	case "donate":
		return mf.ImmediateMessage(fmt.Sprintf("\n Toss a coin to your humble bot,\nO student of fate, \nWhen rivals draw near, and\nthe registration deadline won’t wait\\.\nA humble donation, a whisper, a nudge,\nTo tilt odds in your favor in timetable wars\n\nKaspi: `%s`\n\\[Click to the number to copy\\]", h.KaspiCard))
	case "faq":
		return mf.ImmediateMessage(h.faq)
	case "nextupdatetime":
		return mf.ImmediateMessage(fmt.Sprintf("Next update time is: %s", telegramfmt.Escape(h.CoursesRepo.NextTimeToParse.Format("15:04:05 02.01.2006"))))
	case "parsestat":
		return AuthAdmin(h.AdminID, h.parsestat)(cmd)
	}

	h.StateRepo.Upsert(cmd.From.ID, cmd.Command())
	switch cmd.Command() {
	case "subscribe":
		return mf.ImmediateMessage("Please provide a course abbr and section as in docs\\.\nFormat: `[Course Name] [Course Sections]`\\.\nExample: \\'PHYS 161 2L 1PLB 2R 2r 3plb 3L\\'\n\nYou can also provide \\.txt file from crashed\\.nu")
	case "unsubscribe":
		return mf.ImmediateMessage("Please provide a course abbr as in docs\\.\nFormat: `[Course Name]`\\.\nExample: \\'PHYS161\\'\\.")
	default:
		return h.HandleCommandUnknown(cmd)
	}
}

func (h *MessageHandler) HandleMessage(msg *tapi.Message) []tapi.Chattable {
	mf := telegramfmt.NewMessageFormatter(msg.From.ID)

	slog.Debug("Handling message", "text", msg.Text, "username", msg.From.UserName)
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
	if cmd.Document != nil {
		return h.HandleSubscribeFromCrashedNUFile(cmd)
	}

	mf := telegramfmt.NewMessageFormatter(cmd.From.ID)
	courseName, sectionNames, err := h.parseCommandArguments(cmd.Text)
	if err != nil {
		switch err {
		case ErrNotEnoughParams:
			return mf.ImmediateMessage("❌ You haven't provided  enough arguments\\. If you want to try again \\, first call /subscribe")
		case InvalidParams:
			return mf.ImmediateMessage("❌ You haven't provided valid parameters for the command\\. If you want to try again \\, first call /subscribe")
		}
		return mf.ImmediateMessage("❌ You haven't provided coursename\\. If you want to try again\\, first call /subscribe")
	}

	if _, exists := h.CoursesRepo.GetCourse(courseName); !exists {
		return mf.ImmediateNotFoundCourse(courseName, "for subscription")
	}

	if valid, sect := h.CoursesRepo.CheckForValidness(courseName, sectionNames); !valid {
		return mf.ImmediateNotFoundCourseSection(courseName, sect, "for subscription")
	}

	err = h.SubscriptionRepo.Subscribe(cmd.From.ID, courseName, sectionNames)
	if err != nil {
		slog.Error("Failed to subscribe",
			"error", err,
			"user_id", cmd.From.ID,
			"course", courseName)
		return mf.ImmediateMessage("⚠️ Failed to subscribe to the course\\. Please try again\\.")
	}

	return mf.ImmediateMessage(fmt.Sprintf("✅ Successfully subscribed to *%s \\(%s\\)*", courseName, strings.Join(sectionNames, ", ")))
}

func (h *MessageHandler) HandleSubscribeFromCrashedNUFile(cmd *tapi.Message) []tapi.Chattable {
	mf := telegramfmt.NewMessageFormatter(cmd.From.ID)

	buf, err := h.DownloadFile(cmd.Document.FileID)
	if err != nil {
		slog.Error("Failed to download file", "error", err, "file_id", cmd.Document.FileID)
		return mf.ImmediateMessage("⚠️ Failed to download the file\\. Please try again later\\.")
	}

	str := string(buf)
	lines := strings.Split(str, "\n")
	for ind, line := range lines {
		line = strings.ReplaceAll(line, ": ", "|")
		line = strings.ReplaceAll(line, ", ", "|")
		fields := strings.FieldsFunc(line, func(r rune) bool {
			return r == '|'
		})
		if len(fields) < 2 {
			mf.AddString(fmt.Sprintf("⚠️ Invalid line format\\: %s", lines[ind]))
			continue
		}

		courseName := fields[0]
		section := fields[1:]
		if valid, sect := h.CoursesRepo.CheckForValidness(courseName, section); !valid {
			mf.AddNotFoundCourseSection(courseName, sect)
			continue
		}

		err := h.SubscriptionRepo.Subscribe(cmd.From.ID, courseName, section)
		if err != nil {
			mf.AddString(fmt.Sprintf("⚠️ Failed to subscribe to %s\\. Please try again\\.", courseName))
			continue
		}
		mf.AddString(fmt.Sprintf("✅ Successfully subscribed to %s\n", lines[ind]))
	}

	return mf.Messages()
}

func (h *MessageHandler) parseCommandArguments(args string) (string, []string, error) {
	fields := strings.Fields(args)

	courseName := fields[0]
	ind := 1
	if len(fields) >= 2 && !isDigit(courseName[len(courseName)-1]) && isDigit(fields[1][0]) {
		courseName += fields[1]
		ind++
	}

	if ind == len(fields) {
		return "", nil, ErrNotEnoughParams
	}

	var section []string
	for i := ind; i < len(fields); i++ {
		if len(fields[i]) == 0 || !isDigit(fields[i][0]) {
			if len(section) == 0 {
				return "", nil, InvalidParams
			}
			section[len(section)-1] += fields[i]
		} else {
			section = append(section, fields[i])
		}
	}
	for i := range section {
		sec, ok := telegramfmt.StandartizeSectionName(section[i], h.CoursesRepo.SectionAbbrList)
		if !ok {
			return "", nil, InvalidParams
		}
		section[i] = sec
	}

	return telegramfmt.StandartizeCourseName(courseName), section, nil

}

func isDigit(b byte) bool {
	return '0' <= rune(b) && rune(b) <= '9'
}

func (h *MessageHandler) HandleUnsubscribe(cmd *tapi.Message) []tapi.Chattable {
	mf := telegramfmt.NewMessageFormatter(cmd.From.ID)
	courseName := telegramfmt.StandartizeCourseName(cmd.Text)
	if courseName == "" {
		return mf.ImmediateMessage("❌ You haven't provided coursename")
	}

	if _, exists := h.CoursesRepo.GetCourse(courseName); !exists {
		return mf.ImmediateNotFoundCourse(courseName, "for unsubscribing")
	}

	err := h.SubscriptionRepo.UnSubscribe(cmd.From.ID, courseName)
	if err != nil {
		slog.Error("Failed to subscribe",
			"error", err,
			"user_id", cmd.From.ID,
			"course", courseName)
		return mf.ImmediateMessage("⚠️ Failed to unsubscribe to the course\\. Please try again\\.")
	}

	return mf.ImmediateMessage(fmt.Sprintf("✅ Successfully unsubscribed from *%s*", courseName))
}

func (h *MessageHandler) Clear(cmd *tapi.Message) []tapi.Chattable {
	mf := telegramfmt.NewMessageFormatter(cmd.From.ID)

	err := h.SubscriptionRepo.ClearSubscriptions(cmd.From.ID)
	if err != nil {
		slog.Error("Failed to clear subscriptions",
			"error", err,
			"user_id", cmd.From.ID)
		return mf.ImmediateMessage("⚠️ Failed to clear subscriptions\\. Please try again\\.")
	}

	return mf.ImmediateMessage("✅ Successfully cleared")
}

func (h *MessageHandler) ListSubscriptions(cmd *tapi.Message) []tapi.Chattable {
	mf := telegramfmt.NewMessageFormatter(cmd.From.ID)
	subs, err := h.SubscriptionRepo.GetSubscriptions(cmd.From.ID)
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
			mf.AddNotFoundCourse(sub.Course)
			mf.UnsubscribeOrIgnoreCourse(sub.Course)
			continue
		}

		section, exists := h.CoursesRepo.GetSection(sub.Course, sub.Section)
		if !exists {
			mf.AddNotFoundCourseSection(sub.Course, sub.Section)
			mf.UnsubscribeOrIgnoreSection(sub.Course, sub.Section)
		} else {
			sb.WriteString(telegramfmt.FormatCourseSection(sub.Course, sub.Section, section.Size, section.Cap))
		}
	}
	mf.AddString(sb.String())
	return mf.Messages()
}

func (h *MessageHandler) HandleCourseCode(updateMsg *tapi.Message) []tapi.Chattable {
	mf := telegramfmt.NewMessageFormatter(updateMsg.From.ID)

	courseAbbr := telegramfmt.StandartizeCourseName(updateMsg.Text)
	course, exists := h.CoursesRepo.GetCourse(courseAbbr)
	h.StatisticsRepo.AddOne(courseAbbr)
	if !exists {
		return mf.ImmediateNotFoundCourse(courseAbbr, "")
	}

	return mf.ImmediateMessage(telegramfmt.FormatCourseInDetails(course, h.CoursesRepo.SemesterName, h.CoursesRepo.LastTimeParsed))
}

func (h *MessageHandler) HandleCommandUnknown(cmd *tapi.Message) []tapi.Chattable {
	mf := telegramfmt.NewMessageFormatter(cmd.From.ID)

	return mf.ImmediateMessage(fmt.Sprintf("⚠️ Unknown State \\(/%s\\)", cmd.Command()))
}

func (h *MessageHandler) HandleCommandStart(cmd *tapi.Message) []tapi.Chattable {
	mf := telegramfmt.NewMessageFormatter(cmd.From.ID)

	return mf.ImmediateMessage(h.welcomeText)
}

func (h *MessageHandler) HandleCallback(callback *tapi.CallbackQuery) []tapi.Chattable {
	mf := telegramfmt.NewMessageFormatter(callback.From.ID)

	cmds := strings.Split(callback.Data, ";")

	for _, cmd := range cmds {
		args := strings.Split(cmd, "_")

		switch args[0] {
		case "delete":
			deleteCFG := tapi.NewDeleteMessage(callback.From.ID, callback.Message.MessageID)
			mf.Add(deleteCFG)
		case "unsubscribe":
			if len(args) == 2 {
				err := h.SubscriptionRepo.UnSubscribe(
					callback.From.ID, args[1],
				)
				if err != nil {
					slog.Error("Failed to unsubscribe", "error", err, "course", args[1])
				}
			} else if len(args) == 3 {
				err := h.SubscriptionRepo.UnSubscribeSection(
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

	return mf.Messages()
}

func (h *MessageHandler) parsestat(cmd *tapi.Message) []tapi.Chattable {
	mf := telegramfmt.NewMessageFormatter(cmd.From.ID)

	err := h.StatisticsRepo.Upsert()
	if err != nil {
		return mf.ImmediateMessage("Statistics updated error\\.\n" + err.Error())
	} else {
		return mf.ImmediateMessage("Statistics updated successfully\\.")
	}
}

func (h *MessageHandler) DownloadFile(fileID string) ([]byte, error) {
	file, err := h.BotAPI.GetFile(tapi.FileConfig{FileID: fileID})
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	fileURL := file.Link(h.BotAPI.Token)

	resp, err := http.Get(fileURL)
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return buf.Bytes(), nil
}
