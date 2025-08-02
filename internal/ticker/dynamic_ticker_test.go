package ticker

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseTimeConfig(t *testing.T) {
	tests := []struct {
		name string
		args []TickerIntervalConfig
		want []TickerInterval
	}{
		{
			name: "Test with valid time points",
			args: []TickerIntervalConfig{
				{
					Till:  parseTime("2025-09-06T09:00:00+05:00"),
					Label: "First Priority for 4,5,6 UG",
				},
			},

			want: []TickerInterval{
				{
					From:     parseTime("2025-09-06T08:00:00+05:00"),
					Till:     parseTime("2025-09-06T08:30:00+05:00"),
					Interval: 30 * time.Minute,
					Label:    "First Priority for 4,5,6 UG",
				},
				{
					From:     parseTime("2025-09-06T08:30:00+05:00"),
					Till:     parseTime("2025-09-06T08:45:00+05:00"),
					Interval: 15 * time.Minute,
					Label:    "First Priority for 4,5,6 UG",
				},
				{
					From:     parseTime("2025-09-06T08:45:00+05:00"),
					Till:     parseTime("2025-09-06T08:55:00+05:00"),
					Interval: 5 * time.Minute,
					Label:    "First Priority for 4,5,6 UG",
				},
				{
					From:     parseTime("2025-09-06T08:55:00+05:00"),
					Till:     parseTime("2025-09-06T09:05:00+05:00"),
					Interval: 1 * time.Minute,
					Label:    "First Priority for 4,5,6 UG",
				},
				{
					From:     parseTime("2025-09-06T09:05:00+05:00"),
					Till:     parseTime("2025-09-06T09:30:00+05:00"),
					Interval: 3 * time.Minute,
					Label:    "First Priority for 4,5,6 UG",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, ParseTimeConfig(tt.args), tt.want)
		})
	}
}
