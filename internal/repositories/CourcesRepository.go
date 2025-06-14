package repositories

import (
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/TheTeemka/telegram_bot_cources/internal/models"
)

type CourseRepository struct {
	CoursesAPIURL string
	SemesterName  string

	Courses        map[string][]models.Section
	LastTimeParsed time.Time
	mutex          sync.RWMutex
	ticker         *time.Ticker
}

func NewCourseRepo(coursesAPIURL string, duration time.Duration) *CourseRepository {
	r := &CourseRepository{
		CoursesAPIURL: coursesAPIURL,
		Courses:       map[string][]models.Section{},
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
	for {
		select {
		case <-r.ticker.C:
			if err := r.Parse(); err != nil {
				slog.Error("Failed to parse courses", "error", err)
				continue
			}
		}
	}
}

func (r *CourseRepository) Parse() error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	semesterName, crs, err := GetCourses(r.CoursesAPIURL)
	if err != nil {
		return err
	}

	r.Courses = crs
	r.LastTimeParsed = time.Now()
	r.SemesterName = semesterName
	slog.Info("Courses parsed successfully")
	return nil
}

func (r *CourseRepository) GetCourse(name string) ([]models.Section, bool) {
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
