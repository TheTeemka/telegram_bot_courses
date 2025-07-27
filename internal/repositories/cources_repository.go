package repositories

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/TheTeemka/telegram_bot_cources/internal/models"
	"github.com/TheTeemka/telegram_bot_cources/internal/ticker"
	"github.com/shakinm/xlsReader/xls"
	"github.com/shakinm/xlsReader/xls/structure"
)

type CourseRepository struct {
	CoursesURL    string
	IsExampleData bool

	Courses         map[string]*models.Course
	LastTimeParsed  time.Time
	SemesterName    string
	SectionAbbrList []string

	mutex  sync.RWMutex
	ticker *ticker.DynamicTicker
}

func NewCourseRepo(coursesAPIURL string, isExampleData bool) *CourseRepository {
	r := &CourseRepository{
		CoursesURL:    coursesAPIURL,
		IsExampleData: isExampleData,

		Courses: map[string]*models.Course{},
		ticker:  ticker.NewDynamicTicker(),
	}

	err := r.Parse()
	if err != nil {
		slog.Error("Failed to parse courses", "error", err)
		os.Exit(1)
	}

	return r
}

// func (r *CourseRepository) Watch() {
// 	for range r.ticker.C {
// 		if err := r.Parse(); err != nil {
// 			slog.Error("Failed to parse courses", "error", err)
// 			continue
// 		}
// 	}
// }

func (r *CourseRepository) Parse() error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.IsExampleData {
		return r.ParseExampleData()
	}

	semesterName, cources, sectionAbbrList, err := ParseCourses(r.CoursesURL)
	if err != nil {
		return err
	}

	r.Courses = cources
	r.LastTimeParsed = time.Now()
	r.SemesterName = semesterName
	r.SectionAbbrList = sectionAbbrList
	slog.Info("Courses parsed successfully")
	return nil

}

func (r *CourseRepository) ParseExampleData() error {
	if r.SemesterName != "" {
		return nil
	}

	var b []byte
	file, err := os.Open("example.xls")
	if err != nil {

		if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("opening example file: %w", err)
		}
		b, err = fetch(r.CoursesURL)
		file, err = os.Create("example.xls")
		if err != nil {
			return nil
		}
		file.Write(b)
	} else {
		byt := new(bytes.Buffer)
		io.Copy(byt, file)
		b = byt.Bytes()
	}

	semesterName, cources, sectionAbbrList, err := parseXLS(bytes.NewReader(b))
	r.SemesterName = semesterName
	r.Courses = cources
	r.SectionAbbrList = sectionAbbrList
	stat, err := file.Stat()
	if err != nil {
		return nil
	}
	r.LastTimeParsed = stat.ModTime()
	slog.Info("Example Courses parsed successfully")

	return nil
}

func (r *CourseRepository) GetCourse(name string) (*models.Course, bool) {
	if time.Since(r.LastTimeParsed) > 10*time.Minute {
		err := r.Parse()
		if err != nil {
			slog.Error("Failed to parse courses", "error", err)
		}
	}

	r.mutex.RLock()
	defer r.mutex.RUnlock()

	course, exists := r.Courses[name]
	return course, exists
}

func (r *CourseRepository) GetSection(courseName, SectionName string) (*models.Section, bool) {
	if time.Since(r.LastTimeParsed) > 10*time.Minute {
		err := r.Parse()
		if err != nil {
			slog.Error("Failed to parse courses", "error", err)

		}
	}

	r.mutex.RLock()
	defer r.mutex.RUnlock()

	course, exists := r.Courses[courseName]
	if !exists {
		return nil, false
	}
	for _, section := range course.Sections {
		if section.SectionName == SectionName {
			return section, true
		}
	}
	return nil, false
}

func ParseCourses(url string) (string, map[string]*models.Course, []string, error) {
	b, err := fetch(url)
	if err != nil {
		return "", nil, nil, err
	}

	return parseXLS(bytes.NewReader(b))
}

func fetch(url string) ([]byte, error) {
	buf := new(bytes.Buffer)

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := client.Get(url)
	if err != nil {
		panic(err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Bad response status: %s", resp.Status)
	}
	defer resp.Body.Close()

	io.Copy(buf, resp.Body)

	return buf.Bytes(), nil
}

func parseXLS(file io.ReadSeeker) (string, map[string]*models.Course, []string, error) {
	wb, err := xls.OpenReader(file)
	if err != nil {
		return "", nil, nil, err
	}

	sheet, err := wb.GetSheet(0)
	if err != nil {
		return "", nil, nil, err
	}
	rows := sheet.GetRows()

	semesterName, err := rows[0].GetCol(0)
	if err != nil {
		return "", nil, nil, err
	}

	duplicates := make(map[string]bool)
	courses := make(map[string]*models.Course)
	sectionName := make(map[string]bool)
	for _, row := range rows {
		abbrName, err := GetString(row.GetCol(2)) //Course Abbr
		if err != nil || len(abbrName) == 0 {
			continue
		}

		section, err := GetString(row.GetCol(3)) //S/T
		if err != nil {
			continue
		}

		courseKey := abbrName + "_" + section
		if _, ok := duplicates[courseKey]; ok {
			// slog.Warn("Duplicate course section found, skipping")
			continue
		}
		duplicates[courseKey] = true

		enrolled, err := GetString(row.GetCol(11)) //Enr
		if err != nil {
			continue
		}

		capacity, err := GetString(row.GetCol(12)) //Cap
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
		if _, ok := courses[abbrName]; !ok {
			fullName, err := GetString(row.GetCol(4)) //Course Title
			if err != nil {
				continue
			}
			courses[abbrName] = &models.Course{
				AbbrName: abbrName,
				FullName: fullName,
			}
		}

		sectionName[trimNumbersFromPrefix(section)] = true

		crs := courses[abbrName]
		crs.Sections = append(crs.Sections, &models.Section{
			SectionName: section,
			Size:        enNum,
			Cap:         enCap,
		})
		courses[abbrName] = crs
	}

	sectionAbbrList := make([]string, 0, len(sectionName))
	for abbr := range sectionName {
		sectionAbbrList = append(sectionAbbrList, abbr)
	}

	return semesterName.GetString(), courses, sectionAbbrList, nil
}

func GetString(s structure.CellData, err error) (string, error) {
	return s.GetString(), err
}

func trimNumbersFromPrefix(s string) string {
	return strings.TrimLeftFunc(s, func(r rune) bool {
		return (r >= '0' && r <= '9') || r == ' ' || r == '-'
	})
}
