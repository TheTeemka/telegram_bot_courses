package ticker

import (
	"testing"

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
					Interval: 1800000000000,
					Label:    "First Priority for 4,5,6 UG",
				},
				{
					From:     parseTime("2025-09-06T08:30:00+05:00"),
					Till:     parseTime("2025-09-06T08:45:00+05:00"),
					Interval: 900000000000,
					Label:    "First Priority for 4,5,6 UG",
				},
				{
					From:     parseTime("2025-09-06T08:45:00+05:00"),
					Till:     parseTime("2025-09-06T08:55:00+05:00"),
					Interval: 300000000000,
					Label:    "First Priority for 4,5,6 UG",
				},
				{
					From:     parseTime("2025-09-06T08:55:00+05:00"),
					Till:     parseTime("2025-09-06T09:00:00+05:00"),
					Interval: 60000000000,
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
