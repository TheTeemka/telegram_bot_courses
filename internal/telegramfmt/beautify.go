package telegramfmt

import (
	"fmt"
	"strings"
	"time"

	"github.com/TheTeemka/telegram_bot_cources/internal/models"
)

func FormatCourseInDetails(course *models.Course, semesterName string, lastTimeParse time.Time) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("%s\n", semesterName))
	sb.WriteString(fmt.Sprintf("%s: %s\n", course.AbbrName, course.FullName))

	course.Sections = models.SortSections(course.Sections)

	var s string
	for _, section := range course.Sections {
		if s != trimNumbersFromPrefix(section.SectionName) {
			if s != "" {
				sb.WriteRune('\n')
			}
			s = trimNumbersFromPrefix(section.SectionName)
		}

		sb.WriteString(formatSection(section.SectionName, section.Size, section.Cap))
	}
	sb.WriteString(lastTimeParse.Format("\n_\\Last Update on:  15:04:05 02\\.01\\.2006 _"))

	return sb.String()
}

func formatSection(sectionName string, sectionSize, sectionCap int) string {
	if sectionSize >= sectionCap {
		return (fmt.Sprintf("• `%-7s \\(%d/%d\\)   \\[FULL\\]`\n", sectionName, sectionSize, sectionCap))
	} else {
		return (fmt.Sprintf("• `%-7s \\(%d/%d\\)\n`", sectionName, sectionSize, sectionCap))
	}
}

func FormatCourseSection(courseName, sectionName string, sectionSize, sectionCap int) string {
	if sectionSize >= sectionCap {
		return (fmt.Sprintf("•` ~%-10s %-6s\\(%d/%d\\)~\n`", courseName, sectionName, sectionSize, sectionCap))
	} else {
		return (fmt.Sprintf("•` %-10s %-6s\\(%d/%d\\)\n`", courseName, sectionName, sectionSize, sectionCap))
	}
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
