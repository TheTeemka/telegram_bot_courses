package courses

import (
	"bytes"
	"io"
	"strconv"

	"github.com/shakinm/xlsReader/xls"
	"github.com/shakinm/xlsReader/xls/structure"
)

type Section struct {
	SectionName string
	Size        int
	Cap         int
}

func GetCources(url string) (string, map[string][]Section, error) {
	b, err := fetch(url)
	if err != nil {
		return "", nil, err
	}

	return parse(bytes.NewReader(b))
}

func parse(file io.ReadSeeker) (string, map[string][]Section, error) {
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
	cources := make(map[string][]Section)

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
		cources[name] = append(cources[name], Section{
			SectionName: section,
			Size:        enNum,
			Cap:         enCap,
		})
		mp[name+section] = true
	}

	return semesterName.GetString(), cources, nil
}

func GetString(s structure.CellData, err error) (string, error) {
	return s.GetString(), err
}
