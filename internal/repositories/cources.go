package repositories

import (
	"bytes"
	"io"
	"strconv"

	"github.com/TheTeemka/telegram_bot_cources/internal/models"
	"github.com/shakinm/xlsReader/xls"
	"github.com/shakinm/xlsReader/xls/structure"
)

func GetCourses(url string) (string, map[string][]models.Section, error) {
	b, err := fetch(url)
	if err != nil {
		return "", nil, err
	}

	return parseXLS(bytes.NewReader(b))
}

func parseXLS(file io.ReadSeeker) (string, map[string][]models.Section, error) {
	wb, err := xls.OpenReader(file)
	if err != nil {
		return "", nil, err
	}

	sheet, err := wb.GetSheet(0)
	if err != nil {
		return "", nil, err
	}
	rows := sheet.GetRows()

	semesterName, err := rows[0].GetCol(0)
	if err != nil {
		return "", nil, err
	}

	mp := make(map[string]bool)
	courses := make(map[string][]models.Section)

	for _, row := range rows {
		name, err := GetString(row.GetCol(2))
		if err != nil {
			continue
		}
		if len(name) == 0 {
			continue
		}

		section, err := GetString(row.GetCol(3))
		if err != nil {
			continue
		}

		if _, ok := mp[name+section]; ok {
			continue
		}

		enrolled, err := GetString(row.GetCol(11))
		if err != nil {
			continue
		}

		capacity, err := GetString(row.GetCol(12))
		if err != nil {
			continue
		}

		enNum, err := strconv.Atoi(enrolled)
		if err != nil {
			continue
		}

		enCap, err := strconv.Atoi(capacity)
		if err != nil {
			continue
		}
		courses[name] = append(courses[name], models.Section{
			SectionName: section,
			Size:        enNum,
			Cap:         enCap,
		})
		mp[name+section] = true
	}

	return semesterName.GetString(), courses, nil
}

func GetString(s structure.CellData, err error) (string, error) {
	return s.GetString(), err
}
