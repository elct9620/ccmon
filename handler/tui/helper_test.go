package tui

import (
	"testing"
	"time"

	"github.com/elct9620/ccmon/entity"
)

func TestFormatBlockTime(t *testing.T) {
	loc, _ := time.LoadLocation("UTC")

	tests := []struct {
		name       string
		blockStart time.Time
		want       string
	}{
		{
			name:       "morning block",
			blockStart: time.Date(2025, 1, 1, 10, 0, 0, 0, loc), // 10am start
			want:       "10am - 3pm",
		},
		{
			name:       "afternoon block",
			blockStart: time.Date(2025, 1, 1, 15, 0, 0, 0, loc), // 3pm start
			want:       "3pm - 8pm",
		},
		{
			name:       "late night block",
			blockStart: time.Date(2024, 12, 31, 23, 0, 0, 0, loc), // 11pm start
			want:       "11pm - 4am",
		},
		{
			name:       "midnight start",
			blockStart: time.Date(2025, 1, 1, 0, 0, 0, 0, loc), // 12am start
			want:       "12am - 5am",
		},
		{
			name:       "noon crossing",
			blockStart: time.Date(2025, 1, 1, 10, 0, 0, 0, loc), // 10am start
			want:       "10am - 3pm",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			block := entity.NewBlock(tt.blockStart.UTC())
			got := FormatBlockTime(block, loc)
			if got != tt.want {
				t.Errorf("FormatBlockTime() = %v, want %v", got, tt.want)
			}
		})
	}
}
