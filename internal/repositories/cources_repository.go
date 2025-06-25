package repositories

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/TheTeemka/telegram_bot_cources/internal/models"
	"github.com/shakinm/xlsReader/xls"
	"github.com/shakinm/xlsReader/xls/structure"
)

type CourseRepository struct {
	CoursesAPIURL string
	SemesterName  string

	Courses        map[string]models.Course
	LastTimeParsed time.Time
	mutex          sync.RWMutex
	ticker         *time.Ticker
}

func NewCourseRepo(coursesAPIURL string, duration time.Duration) *CourseRepository {
	r := &CourseRepository{
		CoursesAPIURL: coursesAPIURL,
		Courses:       map[string]models.Course{},
		ticker:        time.NewTicker(duration),
	}

	err := r.Parse()
	if err != nil {
		slog.Error("Failed to parse courses", "error", err)
		os.Exit(1)
	}

	r.ticker.Reset(duration)
	return r
}

func (r *CourseRepository) Watch() {
	for range r.ticker.C {
		if err := r.Parse(); err != nil {
			slog.Error("Failed to parse courses", "error", err)
			continue
		}
	}
}

func (r *CourseRepository) Parse() error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	semesterName, cources, err := ParseCourses(r.CoursesAPIURL)
	if err != nil {
		return err
	}

	r.Courses = cources
	r.LastTimeParsed = time.Now()
	r.SemesterName = semesterName
	slog.Info("Courses parsed successfully")
	return nil
}

func (r *CourseRepository) GetCourse(name string) (models.Course, bool) {
	if time.Since(r.LastTimeParsed) > 10*time.Minute {
		err := r.Parse()
		if err != nil {
			slog.Error("Failed to parse courses", "error", err)
		}
	}

	r.mutex.RLock()
	defer r.mutex.RUnlock()

	sections, exists := r.Courses[name]
	return sections, exists
}

func (r *CourseRepository) GetSection(courseName, SectionName string) (models.Section, bool) {
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
		return models.Section{}, false
	}
	for _, section := range course.Sections {
		if section.SectionName == SectionName {
			return section, true
		}
	}
	return models.Section{}, false
}

func ParseCourses(url string) (string, map[string]models.Course, error) {
	b, err := fetch(url)
	if err != nil {
		return "", nil, err
	}

	return parseXLS(bytes.NewReader(b))
}

func fetch(url string) ([]byte, error) {
	buf := new(bytes.Buffer)

	if url != "" {
		resp, err := http.Get(url)
		if err != nil {
			panic(err)
		}
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("Bad response status: %s", resp.Status)
		}
		defer resp.Body.Close()

		io.Copy(buf, resp.Body)
	} else {
		f, err := os.Open("example.xls")
		if err != nil {
			return nil, err
		}
		defer f.Close()

		io.Copy(buf, f)
	}

	return buf.Bytes(), nil
}

func parseXLS(file io.ReadSeeker) (string, map[string]models.Course, error) {
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

	duplicates := make(map[string]bool)
	courses := make(map[string]models.Course)

	for _, row := range rows {
		abbrName, err := GetString(row.GetCol(2)) //Course Abbr
		if err != nil {
			continue
		}
		if len(abbrName) == 0 {
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
			courses[abbrName] = models.Course{
				AbbrName: abbrName,
				FullName: fullName,
			}
		}

		crs := courses[abbrName]
		crs.Sections = append(crs.Sections, models.Section{
			SectionName: section,
			Size:        enNum,
			Cap:         enCap,
		})
		courses[abbrName] = crs
	}

	return semesterName.GetString(), courses, nil
}

func GetString(s structure.CellData, err error) (string, error) {
	return s.GetString(), err
}
