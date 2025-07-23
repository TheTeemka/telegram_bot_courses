package ticker

import (
	"time"
)

var defaultTickerIntervalConfig = []TickerIntervalConfig{
	{from: parseTime("28 Jun 2025 00:00:00"), till: parseTime("28 Jun 2025 03:07:00"), interval: 10 * time.Second},
}

type TickerIntervalConfig struct {
	from     time.Time
	till     time.Time
	interval time.Duration
}

type DynamicTicker struct {
	C                    chan time.Time
	stop                 chan struct{}
	TickerIntervalConfig []TickerIntervalConfig
}

func NewDynamicTicker() *DynamicTicker {
	t := &DynamicTicker{
		C:                    make(chan time.Time, 1),
		stop:                 make(chan struct{}),
		TickerIntervalConfig: defaultTickerIntervalConfig,
	}
	go t.run()
	return t
}

func (t *DynamicTicker) run() {
	for {
		d := t.getDuration()
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
	cur := 10 * time.Minute
	for _, tt := range t.TickerIntervalConfig {
		if compare(time.Now(), tt) {
			cur = tt.interval
		} else if compare(time.Now().Add(cur), tt) {
			cur = time.Until(tt.till)
		}
	}

	return cur
}

func compare(t time.Time, cfg TickerIntervalConfig) bool {
	return t.After(cfg.from) && t.Before(cfg.till)
}

func parseTime(s string) time.Time {
	loc := time.Now().Location()
	t, err := time.ParseInLocation("2 Jan 2006 15:04:05", s, loc)
	if err != nil {
		panic("invalid time format: " + s)
	}
	return t
}
