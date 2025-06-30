package entity

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// TimeBlockDuration represents the duration of each Claude token limit block
const TimeBlockDuration = 5 * time.Hour

// Block represents a specific 5-hour token limit block for Claude
// This is a value object representing a concrete time period with optional token limit
type Block struct {
	startAt    time.Time // Concrete timestamp when this block starts
	tokenLimit int       // Token limit for this block (0 = no limit)
}

// NewBlock creates a new Block from a concrete start timestamp without token limit
func NewBlock(startAt time.Time) Block {
	return Block{
		startAt:    startAt,
		tokenLimit: 0, // No token limit
	}
}

// NewBlockWithLimit creates a new Block from a concrete start timestamp with token limit
func NewBlockWithLimit(startAt time.Time, tokenLimit int) Block {
	return Block{
		startAt:    startAt,
		tokenLimit: tokenLimit,
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

// StartAt returns the start time of this block
func (b Block) StartAt() time.Time {
	return b.startAt
}

// EndAt returns the end time of this block
func (b Block) EndAt() time.Time {
	return b.startAt.Add(TimeBlockDuration)
}

// TokenLimit returns the token limit for this block (0 = no limit)
func (b Block) TokenLimit() int {
	return b.tokenLimit
}

// HasLimit returns true if this block has a token limit configured
func (b Block) HasLimit() bool {
	return b.tokenLimit > 0
}

// CalculateProgress calculates the progress percentage of premium token usage against the limit
// Returns 0.0 if no limit is configured, otherwise returns percentage (0.0 to 100.0+)
func (b Block) CalculateProgress(premiumTokens Token) float64 {
	if !b.HasLimit() {
		return 0.0
	}
	
	// Only premium tokens count toward limits (Haiku is free)
	used := premiumTokens.Limited()
	limit := int64(b.tokenLimit)
	
	if limit == 0 {
		return 0.0
	}
	
	percentage := float64(used) / float64(limit) * 100
	return percentage
}

// IsLimitExceeded returns true if the premium token usage exceeds the configured limit
func (b Block) IsLimitExceeded(premiumTokens Token) bool {
	if !b.HasLimit() {
		return false
	}
	
	used := premiumTokens.Limited()
	return used > int64(b.tokenLimit)
}

// Period returns the time period represented by this block
func (b Block) Period() Period {
	return NewPeriod(b.startAt, b.EndAt())
}

// NextBlock returns the appropriate block for the given time.
// If the current time is still within this block, returns self.
// If the current time is beyond this block, returns the next appropriate block.
func (b Block) NextBlock(now time.Time) Block {
	// If current time is still within this block, return self
	if now.Before(b.EndAt()) {
		return b
	}

	// Calculate which block the current time falls into
	delta := now.Sub(b.startAt)
	blockIndex := int(delta / TimeBlockDuration)

	// Create new block at the appropriate position, preserving token limit
	newStart := b.startAt.Add(time.Duration(blockIndex) * TimeBlockDuration)
	return NewBlockWithLimit(newStart, b.tokenLimit)
}

// NewCurrentBlock creates a Block representing the current 5-hour period
// based on user's specified start hour and timezone.
// Always returns a valid block - either the current block or the next upcoming block.
func NewCurrentBlock(userStartHour int, timezone *time.Location, now time.Time) Block {
	return NewCurrentBlockWithLimit(userStartHour, timezone, now, 0)
}

// NewCurrentBlockWithLimit creates a Block with token limit representing the current 5-hour period
// based on user's specified start hour and timezone.
// Always returns a valid block - either the current block or the next upcoming block.
func NewCurrentBlockWithLimit(userStartHour int, timezone *time.Location, now time.Time, tokenLimit int) Block {
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
			return NewBlockWithLimit(referenceTime.UTC(), tokenLimit)
		}
	}

	// Calculate which 5-hour block we're in based on the delta
	blockIndex := int(delta / TimeBlockDuration)
	blockStart := referenceTime.Add(time.Duration(blockIndex) * TimeBlockDuration)

	return NewBlockWithLimit(blockStart.UTC(), tokenLimit)
}

// FormatBlockTime formats the block period for display in the given timezone
func (b Block) FormatBlockTime(timezone *time.Location) string {
	startLocal := b.startAt.In(timezone)
	endLocal := b.EndAt().In(timezone)

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
