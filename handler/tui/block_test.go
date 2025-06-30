package tui

import (
	"testing"
	"time"
)

func TestParseBlockTime(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int
		wantErr bool
	}{
		{
			name:    "parse 5am",
			input:   "5am",
			want:    5,
			wantErr: false,
		},
		{
			name:    "parse 10am",
			input:   "10am",
			want:    10,
			wantErr: false,
		},
		{
			name:    "parse 12am (midnight)",
			input:   "12am",
			want:    0,
			wantErr: false,
		},
		{
			name:    "parse 12pm (noon)",
			input:   "12pm",
			want:    12,
			wantErr: false,
		},
		{
			name:    "parse 1pm",
			input:   "1pm",
			want:    13,
			wantErr: false,
		},
		{
			name:    "parse 11pm",
			input:   "11pm",
			want:    23,
			wantErr: false,
		},
		{
			name:    "parse with spaces",
			input:   " 5am ",
			want:    5,
			wantErr: false,
		},
		{
			name:    "parse uppercase",
			input:   "5AM",
			want:    5,
			wantErr: false,
		},
		{
			name:    "invalid - no am/pm",
			input:   "5",
			want:    0,
			wantErr: true,
		},
		{
			name:    "invalid - hour out of range",
			input:   "13am",
			want:    0,
			wantErr: true,
		},
		{
			name:    "invalid - hour zero",
			input:   "0am",
			want:    0,
			wantErr: true,
		},
		{
			name:    "invalid - not a number",
			input:   "fiveam",
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := parseBlockTime(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseBlockTime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parseBlockTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCalculateCurrentBlock(t *testing.T) {
	loc, _ := time.LoadLocation("UTC")

	tests := []struct {
		name       string
		startHour  int
		now        time.Time
		tokenLimit int
		wantStart  time.Time
		wantEnd    time.Time
		wantLimit  int
	}{
		{
			name:       "before start hour same day - shows upcoming block",
			startHour:  10,                                      // 10am
			now:        time.Date(2025, 1, 1, 9, 46, 0, 0, loc), // 9:46am
			tokenLimit: 7000,
			wantStart:  time.Date(2025, 1, 1, 10, 0, 0, 0, loc), // 10am
			wantEnd:    time.Date(2025, 1, 1, 15, 0, 0, 0, loc), // 3pm
			wantLimit:  7000,
		},
		{
			name:       "at start hour - shows current block",
			startHour:  10,
			now:        time.Date(2025, 1, 1, 10, 0, 0, 0, loc),
			tokenLimit: 7000,
			wantStart:  time.Date(2025, 1, 1, 10, 0, 0, 0, loc),
			wantEnd:    time.Date(2025, 1, 1, 15, 0, 0, 0, loc),
			wantLimit:  7000,
		},
		{
			name:       "within first block",
			startHour:  10,
			now:        time.Date(2025, 1, 1, 12, 30, 0, 0, loc), // 12:30pm
			tokenLimit: 35000,
			wantStart:  time.Date(2025, 1, 1, 10, 0, 0, 0, loc),
			wantEnd:    time.Date(2025, 1, 1, 15, 0, 0, 0, loc),
			wantLimit:  35000,
		},
		{
			name:       "in second block",
			startHour:  10,
			now:        time.Date(2025, 1, 1, 16, 0, 0, 0, loc), // 4pm
			tokenLimit: 0,                                       // no limit
			wantStart:  time.Date(2025, 1, 1, 15, 0, 0, 0, loc), // 3pm
			wantEnd:    time.Date(2025, 1, 1, 20, 0, 0, 0, loc), // 8pm
			wantLimit:  0,
		},
		{
			name:       "late night crossing to next day",
			startHour:  23,                                     // 11pm
			now:        time.Date(2025, 1, 1, 1, 0, 0, 0, loc), // 1am
			tokenLimit: 7000,
			wantStart:  time.Date(2024, 12, 31, 23, 0, 0, 0, loc), // 11pm yesterday
			wantEnd:    time.Date(2025, 1, 1, 4, 0, 0, 0, loc),    // 4am today
			wantLimit:  7000,
		},
		{
			name:       "early morning before 5am start",
			startHour:  5,
			now:        time.Date(2025, 1, 1, 3, 0, 0, 0, loc), // 3am
			tokenLimit: 7000,
			wantStart:  time.Date(2025, 1, 1, 5, 0, 0, 0, loc),  // 5am (upcoming)
			wantEnd:    time.Date(2025, 1, 1, 10, 0, 0, 0, loc), // 10am
			wantLimit:  7000,
		},
		{
			name:       "very early morning - shows upcoming block",
			startHour:  10,
			now:        time.Date(2025, 1, 1, 2, 0, 0, 0, loc), // 2am
			tokenLimit: 140000,
			wantStart:  time.Date(2025, 1, 1, 10, 0, 0, 0, loc), // 10am (upcoming)
			wantEnd:    time.Date(2025, 1, 1, 15, 0, 0, 0, loc), // 3pm
			wantLimit:  140000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			block := calculateCurrentBlock(tt.startHour, loc, tt.now, tt.tokenLimit)

			if !block.StartAt().Equal(tt.wantStart) {
				t.Errorf("calculateCurrentBlock() start = %v, want %v", block.StartAt(), tt.wantStart)
			}
			if !block.EndAt().Equal(tt.wantEnd) {
				t.Errorf("calculateCurrentBlock() end = %v, want %v", block.EndAt(), tt.wantEnd)
			}
			if block.TokenLimit() != tt.wantLimit {
				t.Errorf("calculateCurrentBlock() token limit = %v, want %v", block.TokenLimit(), tt.wantLimit)
			}
		})
	}
}

func TestCalculateCurrentBlock_TimezoneHandling(t *testing.T) {
	// Test different timezones to ensure proper handling
	utc, _ := time.LoadLocation("UTC")
	est, _ := time.LoadLocation("America/New_York")
	pst, _ := time.LoadLocation("America/Los_Angeles")

	tests := []struct {
		name       string
		startHour  int
		timezone   *time.Location
		nowUTC     time.Time
		wantStartH int // expected start hour in the timezone
	}{
		{
			name:       "UTC timezone",
			startHour:  10,
			timezone:   utc,
			nowUTC:     time.Date(2025, 1, 1, 12, 0, 0, 0, utc), // 12pm UTC
			wantStartH: 10,                                      // 10am UTC
		},
		{
			name:       "EST timezone morning",
			startHour:  9,
			timezone:   est,
			nowUTC:     time.Date(2025, 1, 1, 16, 0, 0, 0, utc), // 4pm UTC = 11am EST
			wantStartH: 9,                                       // 9am EST
		},
		{
			name:       "PST timezone afternoon",
			startHour:  14,
			timezone:   pst,
			nowUTC:     time.Date(2025, 1, 1, 23, 0, 0, 0, utc), // 11pm UTC = 3pm PST
			wantStartH: 14,                                      // 2pm PST
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			block := calculateCurrentBlock(tt.startHour, tt.timezone, tt.nowUTC, 7000)

			// Convert block start time to the test timezone for verification
			blockStartInTz := block.StartAt().In(tt.timezone)
			if blockStartInTz.Hour() != tt.wantStartH {
				t.Errorf("calculateCurrentBlock() start hour in timezone = %v, want %v",
					blockStartInTz.Hour(), tt.wantStartH)
			}

			// Verify block is always returned in UTC internally
			if block.StartAt().Location() != utc {
				t.Errorf("calculateCurrentBlock() should return UTC time, got %v",
					block.StartAt().Location())
			}
		})
	}
}
