package models

import (
	"slices"
	"strings"
)

type Course struct {
	FullName string
	AbbrName string
	Sections []*Section
}

type Section struct {
	SectionName string
	Days        []string
	Size        int
	Cap         int
}

func SortSections(sections []*Section) []*Section {
	slices.SortFunc(sections, func(a, b *Section) int {
		atrim, btrim := trimNumbersFromPrefix(a.SectionName), trimNumbersFromPrefix(b.SectionName)
		if atrim == btrim {
			an, bn := getPrefixNumbers(a.SectionName), getPrefixNumbers(b.SectionName)
			if an < bn {
				return -1
			} else if an > bn {
				return 1
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
