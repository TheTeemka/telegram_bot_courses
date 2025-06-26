package handlers

import (
	"fmt"
	"strings"

	"github.com/TheTeemka/telegram_bot_cources/internal/models"
)

func (h *MessageHandler) beatify(course models.Course) string {
	builder := strings.Builder{}

	builder.WriteString(fmt.Sprintf("%s\n", h.CoursesRepo.SemesterName))
	builder.WriteString(fmt.Sprintf("%s: %s\n", course.AbbrName, course.FullName))

	course.Sections = models.SortSections(course.Sections)

	var s string
	for _, section := range course.Sections {
		if s != trimNumbersFromPrefix(section.SectionName) {
			s = trimNumbersFromPrefix(section.SectionName)
			builder.WriteRune('\n')
		}
		if section.Size >= section.Cap {
			builder.WriteString(fmt.Sprintf("•   ~%-7s \\(%d/%d\\)~\n", section.SectionName, section.Size, section.Cap))
		} else {
			builder.WriteString(fmt.Sprintf("•   %-7s \\(%d/%d\\)\n", section.SectionName, section.Size, section.Cap))
		}
	}
	builder.WriteString(h.CoursesRepo.LastTimeParsed.Format("\n_\\Last Updated on:  15:04:05 02\\.01\\.2006 _"))

	return builder.String()
}

func trimNumbersFromPrefix(s string) string {
	return strings.TrimLeftFunc(s, func(r rune) bool {
		return (r >= '0' && r <= '9') || r == ' ' || r == '-'
	})
}

func StandartizeCourseName(s string) string {
	s = strings.ToUpper(s)

	var result strings.Builder
	var numStart bool
	for _, r := range s {
		if r == ' ' {
			continue
		}
		if r >= '0' && r <= '9' && !numStart {
			result.WriteRune(' ')
			numStart = true
		}
		result.WriteRune(r)
	}
	return strings.Join(strings.Fields(result.String()), " ")
}

func StandartizeSectionName(s string, sectionAbbrList []string) (string, bool) {
	trimmedS := trimNumbersFromPrefix(s)
	for _, sectionAbbr := range sectionAbbrList {
		if strings.EqualFold(trimmedS, sectionAbbr) {
			return retrieveNumbersFromPrefix(s) + sectionAbbr, true
		}
	}
	return "", false
}

func retrieveNumbersFromPrefix(s string) string {
	for i, r := range s {
		if !(r >= '0' && r <= '9') {
			return s[:i]
		}
	}
	return s
}
