package tui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/elct9620/ccmon/entity"
)

// parseBlockTime parses simple time format like "5am", "11pm" into hour (0-23)
func parseBlockTime(timeStr string) (int, error) {
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

// calculateCurrentBlock calculates the current 5-hour block based on user's start hour and timezone
// Always returns a valid block - either the current block or the next upcoming block.
func calculateCurrentBlock(userStartHour int, timezone *time.Location, now time.Time, tokenLimit int) entity.Block {
	nowInTz := now.In(timezone)

	// Create reference timestamp at start hour today
	referenceTime := time.Date(nowInTz.Year(), nowInTz.Month(), nowInTz.Day(),
		userStartHour, 0, 0, 0, timezone)

	// Calculate time difference from reference
	delta := nowInTz.Sub(referenceTime)

	// If we're before the start time today, check if we're within a reasonable range
	if delta < 0 {
		// If we're before today's start hour, check if yesterday's sequence makes more sense
		// This handles cases like current time 1am with 11pm start hour
		if delta < -12*time.Hour {
			// Use yesterday's reference instead
			referenceTime = referenceTime.AddDate(0, 0, -1)
			delta = nowInTz.Sub(referenceTime)
		}

		// If still negative, we're before the start time - show the upcoming block
		if delta < 0 {
			return entity.NewBlockWithLimit(referenceTime.UTC(), tokenLimit)
		}
	}

	// Calculate which 5-hour block we're in based on the delta
	blockIndex := int(delta / entity.TimeBlockDuration)
	blockStart := referenceTime.Add(time.Duration(blockIndex) * entity.TimeBlockDuration)

	return entity.NewBlockWithLimit(blockStart.UTC(), tokenLimit)
}
