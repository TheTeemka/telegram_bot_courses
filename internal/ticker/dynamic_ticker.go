package ticker

import (
	"log/slog"
	"time"
)

var registrationTickerIntervalConfig = []TickerIntervalConfig{
	{
		Till:  parseTime("2025-09-06T09:00:00+05:00"),
		Label: "First Priority for 4,5,6 UG",
	},
	{
		Till:  parseTime("2025-09-06T11:00:00+05:00"),
		Label: "First Priority for 3 UG",
	},
	{
		Till:  parseTime("2025-09-06T13:00:00+05:00"),
		Label: "First Priority for 2 UG",
	},
	{
		Till:  parseTime("2025-09-13T09:00:00+05:00"),
		Label: "First Priority for 1 UG",
	},

	{
		Till:  parseTime("2025-09-14T09:00:00+05:00"),
		Label: "Second Priority for 4,5,6 UG",
	},
	{
		Till:  parseTime("2025-09-14T11:00:00+05:00"),
		Label: "Second Priority for 3 UG",
	},
	{
		Till:  parseTime("2025-09-14T13:00:00+05:00"),
		Label: "Second Priority for 2 UG",
	},
	{
		Till:  parseTime("2025-09-14T15:00:00+05:00"),
		Label: "Second Priority for 1 UG",
	},

	{
		Till:  parseTime("2025-09-14T15:00:00+05:00"),
		Label: "Third Priority for All UG",
	},
}

var registrationTickerInterval = ParseTimeConfig(registrationTickerIntervalConfig)

type TickerIntervalConfig struct {
	Till  time.Time
	Label string
}

type TickerInterval struct {
	From     time.Time
	Till     time.Time
	Interval time.Duration
	Label    string
}

type DynamicTicker struct {
	C                   chan time.Time
	stop                chan struct{}
	TickerIntervals     []TickerInterval
	defaultTimeInterval time.Duration
}

func NewDynamicTicker(timeInterval time.Duration) *DynamicTicker {
	t := &DynamicTicker{
		C:                   make(chan time.Time, 1),
		stop:                make(chan struct{}),
		TickerIntervals:     registrationTickerInterval,
		defaultTimeInterval: timeInterval,
	}
	go t.run()
	return t
}

func (t *DynamicTicker) run() {
	for {
		d := t.getDuration()
		slog.Debug("Dynamic ticker started", "duration", d.String())
		timer := time.NewTimer(d)
		select {
		case now := <-timer.C:
			select {
			case t.C <- now:
			default: // drop tick if nobody is listening
			}
		case <-t.stop:
			timer.Stop()
			close(t.C)
			return
		}
	}
}

func (t *DynamicTicker) Stop() {
	close(t.stop)
}

func (t *DynamicTicker) getDuration() time.Duration {
	cur := t.defaultTimeInterval
	for _, tt := range t.TickerIntervals {
		if isWithin(time.Now(), tt) {
			cur = tt.Interval
		} else if isWithin(time.Now().Add(cur), tt) {
			cur = time.Until(tt.Till)
		}
	}

	return cur
}

func isWithin(t time.Time, cfg TickerInterval) bool {
	return (t.After(cfg.From) || t.Equal(cfg.From)) && (t.Before(cfg.Till) || t.Equal(cfg.Till))
}

func ParseTimeConfig(timeConfigs []TickerIntervalConfig) []TickerInterval {
	tt := make([]TickerInterval, 0, len(timeConfigs)*4)
	for _, t := range timeConfigs {
		tt = append(tt, TickerInterval{
			From:     t.Till.Add(-1 * time.Hour),
			Till:     t.Till.Add(-30 * time.Minute),
			Interval: 30 * time.Minute,
			Label:    t.Label,
		})

		tt = append(tt, TickerInterval{
			From:     t.Till.Add(-30 * time.Minute),
			Till:     t.Till.Add(-15 * time.Minute),
			Interval: 15 * time.Minute,
			Label:    t.Label,
		})

		tt = append(tt, TickerInterval{
			From:     t.Till.Add(-15 * time.Minute),
			Till:     t.Till.Add(-5 * time.Minute),
			Interval: 5 * time.Minute,
			Label:    t.Label,
		})

		tt = append(tt, TickerInterval{
			From:     t.Till.Add(-5 * time.Minute),
			Till:     t.Till.Add(0 * time.Minute),
			Interval: 1 * time.Minute,
			Label:    t.Label,
		})
	}

	return tt
}

func parseTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic("invalid time format: " + s)
	}
	return t
}

func main() {

}
