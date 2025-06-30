package entity

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
			got, err := ParseBlockTime(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseBlockTime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseBlockTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlock_NewCurrentBlock(t *testing.T) {
	loc, _ := time.LoadLocation("UTC")

	tests := []struct {
		name      string
		startHour int
		now       time.Time
		wantStart time.Time
		wantEnd   time.Time
	}{
		{
			name:      "before start hour same day - shows upcoming block",
			startHour: 10,                                      // 10am
			now:       time.Date(2025, 1, 1, 9, 46, 0, 0, loc), // 9:46am
			wantStart: time.Date(2025, 1, 1, 10, 0, 0, 0, loc), // 10am
			wantEnd:   time.Date(2025, 1, 1, 15, 0, 0, 0, loc), // 3pm
		},
		{
			name:      "at start hour - shows current block",
			startHour: 10,
			now:       time.Date(2025, 1, 1, 10, 0, 0, 0, loc),
			wantStart: time.Date(2025, 1, 1, 10, 0, 0, 0, loc),
			wantEnd:   time.Date(2025, 1, 1, 15, 0, 0, 0, loc),
		},
		{
			name:      "within first block",
			startHour: 10,
			now:       time.Date(2025, 1, 1, 12, 30, 0, 0, loc), // 12:30pm
			wantStart: time.Date(2025, 1, 1, 10, 0, 0, 0, loc),
			wantEnd:   time.Date(2025, 1, 1, 15, 0, 0, 0, loc),
		},
		{
			name:      "in second block",
			startHour: 10,
			now:       time.Date(2025, 1, 1, 16, 0, 0, 0, loc), // 4pm
			wantStart: time.Date(2025, 1, 1, 15, 0, 0, 0, loc), // 3pm
			wantEnd:   time.Date(2025, 1, 1, 20, 0, 0, 0, loc), // 8pm
		},
		{
			name:      "late night crossing to next day",
			startHour: 23,                                        // 11pm
			now:       time.Date(2025, 1, 1, 1, 0, 0, 0, loc),    // 1am
			wantStart: time.Date(2024, 12, 31, 23, 0, 0, 0, loc), // 11pm yesterday
			wantEnd:   time.Date(2025, 1, 1, 4, 0, 0, 0, loc),    // 4am today
		},
		{
			name:      "early morning before 5am start",
			startHour: 5,
			now:       time.Date(2025, 1, 1, 3, 0, 0, 0, loc),  // 3am
			wantStart: time.Date(2025, 1, 1, 5, 0, 0, 0, loc),  // 5am (upcoming)
			wantEnd:   time.Date(2025, 1, 1, 10, 0, 0, 0, loc), // 10am
		},
		{
			name:      "very early morning - shows upcoming block",
			startHour: 10,
			now:       time.Date(2025, 1, 1, 2, 0, 0, 0, loc),  // 2am
			wantStart: time.Date(2025, 1, 1, 10, 0, 0, 0, loc), // 10am (upcoming)
			wantEnd:   time.Date(2025, 1, 1, 15, 0, 0, 0, loc), // 3pm
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			block := NewCurrentBlock(tt.startHour, loc, tt.now)

			if !block.StartAt().Equal(tt.wantStart) {
				t.Errorf("NewCurrentBlock() start = %v, want %v", block.StartAt(), tt.wantStart)
			}
			if !block.EndAt().Equal(tt.wantEnd) {
				t.Errorf("NewCurrentBlock() end = %v, want %v", block.EndAt(), tt.wantEnd)
			}
		})
	}
}

func TestBlock_NextBlock(t *testing.T) {
	loc, _ := time.LoadLocation("UTC")

	tests := []struct {
		name       string
		blockStart time.Time
		now        time.Time
		wantStart  time.Time
		wantEnd    time.Time
		shouldStay bool // true if block should remain the same
	}{
		{
			name:       "time still within current block",
			blockStart: time.Date(2025, 1, 1, 10, 0, 0, 0, loc),  // 10am block
			now:        time.Date(2025, 1, 1, 12, 30, 0, 0, loc), // 12:30pm (within block)
			wantStart:  time.Date(2025, 1, 1, 10, 0, 0, 0, loc),  // same block
			wantEnd:    time.Date(2025, 1, 1, 15, 0, 0, 0, loc),
			shouldStay: true,
		},
		{
			name:       "time exactly at block end",
			blockStart: time.Date(2025, 1, 1, 10, 0, 0, 0, loc), // 10am block
			now:        time.Date(2025, 1, 1, 15, 0, 0, 0, loc), // 3pm (exactly at end)
			wantStart:  time.Date(2025, 1, 1, 15, 0, 0, 0, loc), // next block
			wantEnd:    time.Date(2025, 1, 1, 20, 0, 0, 0, loc),
			shouldStay: false,
		},
		{
			name:       "time past current block - advance one block",
			blockStart: time.Date(2025, 1, 1, 10, 0, 0, 0, loc),  // 10am block
			now:        time.Date(2025, 1, 1, 16, 30, 0, 0, loc), // 4:30pm (in next block)
			wantStart:  time.Date(2025, 1, 1, 15, 0, 0, 0, loc),  // next block
			wantEnd:    time.Date(2025, 1, 1, 20, 0, 0, 0, loc),
			shouldStay: false,
		},
		{
			name:       "time far past - advance multiple blocks",
			blockStart: time.Date(2025, 1, 1, 10, 0, 0, 0, loc), // 10am block
			now:        time.Date(2025, 1, 2, 8, 0, 0, 0, loc),  // next day 8am (4 blocks later)
			wantStart:  time.Date(2025, 1, 2, 6, 0, 0, 0, loc),  // 6am next day (block 4)
			wantEnd:    time.Date(2025, 1, 2, 11, 0, 0, 0, loc), // 11am next day
			shouldStay: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalBlock := NewBlock(tt.blockStart.UTC())
			nextBlock := originalBlock.NextBlock(tt.now)

			if tt.shouldStay {
				// Block should remain the same
				if !nextBlock.StartAt().Equal(originalBlock.StartAt()) {
					t.Errorf("NextBlock() should stay same, got start = %v, want %v",
						nextBlock.StartAt(), originalBlock.StartAt())
				}
			}

			if !nextBlock.StartAt().Equal(tt.wantStart) {
				t.Errorf("NextBlock() start = %v, want %v", nextBlock.StartAt(), tt.wantStart)
			}
			if !nextBlock.EndAt().Equal(tt.wantEnd) {
				t.Errorf("NextBlock() end = %v, want %v", nextBlock.EndAt(), tt.wantEnd)
			}
		})
	}
}

func TestBlock_FormatBlockTime(t *testing.T) {
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
			block := NewBlock(tt.blockStart.UTC())
			got := block.FormatBlockTime(loc)
			if got != tt.want {
				t.Errorf("FormatBlockTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlock_ValueObjectBehavior(t *testing.T) {
	loc, _ := time.LoadLocation("UTC")

	t.Run("Block represents specific time period", func(t *testing.T) {
		start := time.Date(2025, 1, 1, 10, 0, 0, 0, loc)
		block := NewBlock(start.UTC())

		// Test getter methods
		if !block.StartAt().Equal(start.UTC()) {
			t.Errorf("StartAt() = %v, want %v", block.StartAt(), start.UTC())
		}

		expectedEnd := start.Add(TimeBlockDuration)
		if !block.EndAt().Equal(expectedEnd.UTC()) {
			t.Errorf("EndAt() = %v, want %v", block.EndAt(), expectedEnd.UTC())
		}

		// Test Period method
		period := block.Period()
		if !period.StartAt().Equal(start.UTC()) {
			t.Errorf("Period().StartAt() = %v, want %v", period.StartAt(), start.UTC())
		}
		if !period.EndAt().Equal(expectedEnd.UTC()) {
			t.Errorf("Period().EndAt() = %v, want %v", period.EndAt(), expectedEnd.UTC())
		}
	})

	t.Run("Block formatting works correctly", func(t *testing.T) {
		// Test the key scenario: 9:46am with 10am start should show "10am - 3pm"
		start := time.Date(2025, 1, 1, 10, 0, 0, 0, loc) // 10am start
		block := NewBlock(start.UTC())

		formatted := block.FormatBlockTime(loc)
		expected := "10am - 3pm"

		if formatted != expected {
			t.Errorf("FormatBlockTime() = %v, want %v", formatted, expected)
		}
	})
}
