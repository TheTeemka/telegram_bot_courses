package repo

import (
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/TheTeemka/telegram_bot_cources/internal/courses"
)

type CourceRepo struct {
	CourcesAPIURL string
	SemesterName  string

	Cources        map[string][]courses.Section
	LastTimeParsed time.Time
	mutex          sync.RWMutex
	ticker         *time.Ticker
}

func NewCourceRepo(courcesAPIURL string, duration time.Duration) *CourceRepo {
	r := &CourceRepo{
		CourcesAPIURL: courcesAPIURL,
		Cources:       map[string][]courses.Section{},
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

func (r *CourceRepo) Watch() {
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

func (r *CourceRepo) Parse() error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	semesterName, crs, err := courses.GetCources(r.CourcesAPIURL)
	if err != nil {
		return err
	}

	r.Cources = crs
	r.LastTimeParsed = time.Now()
	r.SemesterName = semesterName
	slog.Info("Courses parsed successfully")
	return nil
}

func (r *CourceRepo) GetCourse(name string) ([]courses.Section, bool) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	sections, exists := r.Cources[name]
	return sections, exists
}
