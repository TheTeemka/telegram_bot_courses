package telegram

import (
	"fmt"
	"log/slog"
	"slices"
	"strings"
)

func (bot *TelegramBot) beatify(name string, sections []Section) string {
	builder := strings.Builder{}

	builder.WriteString(fmt.Sprintf("%s: %s\n", Standartize(name), Standartize(bot.CoursesRepo.SemesterName)))

	sections = sortSections(sections)

	var s string
	for _, section := range sections {
		if s != trimNumbersFromPrefix(section.SectionName) {
			s = trimNumbersFromPrefix(section.SectionName)
			builder.WriteRune('\n')
		}
		if section.Size >= section.Cap {
			builder.WriteString(fmt.Sprintf("  ~%-7s \\(%d/%d\\)\n~", section.SectionName, section.Size, section.Cap))
		} else {
			builder.WriteString(fmt.Sprintf("  %-7s \\(%d/%d\\)\n", section.SectionName, section.Size, section.Cap))
		}
	}
	builder.WriteString(bot.CoursesRepo.LastTimeParsed.Format("\n_\\Last Updated on:  15:04:05 02\\.01\\.2006 _"))

	return builder.String()
}

func sortSections(sections []Section) []Section {
	slices.SortFunc(sections, func(a, b Section) int {
		atrim, btrim := trimNumbersFromPrefix(a.SectionName), trimNumbersFromPrefix(b.SectionName)
		if atrim == btrim {
			an, bn := getPrefixNumbers(a.SectionName), getPrefixNumbers(b.SectionName)
			if an < bn {
				return -1
			} else if an > bn {
				return 1
			} else {
				slog.Error("Sections have the same name and prefix numbers", "sectionA", a.SectionName, "sectionB", b.SectionName)
			}
			return 0
		}

		return strings.Compare(atrim, btrim)
	})
	return sections
}

func trimNumbersFromPrefix(s string) string {
	return strings.TrimLeftFunc(s, func(r rune) bool {
		return (r >= '0' && r <= '9') || r == ' ' || r == '-'
	})
}

func Standartize(s string) string {
	s = strings.ToUpper(s)
	s = strings.Join(strings.Fields(s), "")

	var result strings.Builder
	var numStart bool
	for _, r := range s {
		if r >= '0' && r <= '9' && !numStart {
			result.WriteRune(' ')
			numStart = true
		}
		result.WriteRune(r)
	}
	return strings.Join(strings.Fields(result.String()), " ")
}

func getPrefixNumbers(s string) int {
	prefix := 0
	for _, r := range s {
		if r >= '0' && r <= '9' {
			prefix = prefix*10 + int(r-'0')
		} else {
			break
		}
	}
	return prefix
}
