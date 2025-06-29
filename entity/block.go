package entity

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Block represents a 5-hour token limit block for Claude
type Block struct {
	startHour int            // 0-23, hour when block starts
	timezone  *time.Location // timezone for block calculations
}

// NewBlock creates a new Block from hour and timezone
func NewBlock(startHour int, timezone *time.Location) Block {
	return Block{
		startHour: startHour,
		timezone:  timezone,
	}
}

// ParseBlockTime parses simple time format like "5am", "11pm" into hour (0-23)
func ParseBlockTime(timeStr string) (int, error) {
	timeStr = strings.ToLower(strings.TrimSpace(timeStr))

	var hour int
	var ampm string

	// Parse formats like "5am", "11pm"
	if strings.HasSuffix(timeStr, "am") || strings.HasSuffix(timeStr, "pm") {
		ampm = timeStr[len(timeStr)-2:]
		hourStr := timeStr[:len(timeStr)-2]

		var err error
		hour, err = strconv.Atoi(hourStr)
		if err != nil {
			return 0, fmt.Errorf("invalid hour format: %s", timeStr)
		}

		// Validate hour range
		if hour < 1 || hour > 12 {
			return 0, fmt.Errorf("hour must be between 1-12: %d", hour)
		}

		// Convert to 24-hour format
		if ampm == "am" {
			if hour == 12 {
				hour = 0 // 12am = 0
			}
		} else { // pm
			if hour != 12 {
				hour += 12 // 1pm = 13, 11pm = 23
			}
			// 12pm = 12 (no change)
		}
	} else {
		return 0, fmt.Errorf("time must end with 'am' or 'pm': %s", timeStr)
	}

	return hour, nil
}

// CurrentBlock calculates the current 5-hour block period
func (b Block) CurrentBlock(now time.Time) Period {
	nowInTz := now.In(b.timezone)

	// Find which 5-hour block we're in
	currentHour := nowInTz.Hour()

	// Calculate how many 5-hour blocks have passed since start hour
	hoursSinceStart := (currentHour - b.startHour + 24) % 24
	blockIndex := hoursSinceStart / 5

	// Calculate start hour of current block
	blockStartHour := (b.startHour + blockIndex*5) % 24

	// Create start and end times for current block
	today := time.Date(nowInTz.Year(), nowInTz.Month(), nowInTz.Day(), blockStartHour, 0, 0, 0, b.timezone)

	// If current time is before the calculated start, we're in yesterday's block
	if nowInTz.Before(today) {
		today = today.AddDate(0, 0, -1)
	}

	blockStart := today.UTC()
	blockEnd := today.Add(5 * time.Hour).UTC()

	return NewPeriod(blockStart, blockEnd)
}

// FormatBlockTime formats the block period for display
func (b Block) FormatBlockTime(now time.Time) string {
	currentBlock := b.CurrentBlock(now)
	startLocal := currentBlock.StartAt().In(b.timezone)
	endLocal := currentBlock.EndAt().In(b.timezone)

	startStr := formatHour(startLocal.Hour())
	endStr := formatHour(endLocal.Hour())

	return fmt.Sprintf("%s - %s", startStr, endStr)
}

// formatHour formats hour (0-23) into 12-hour format with am/pm
func formatHour(hour int) string {
	if hour == 0 {
		return "12am"
	} else if hour < 12 {
		return fmt.Sprintf("%dam", hour)
	} else if hour == 12 {
		return "12pm"
	} else {
		return fmt.Sprintf("%dpm", hour-12)
	}
}
