package schema

import "time"

// APIRequest represents a single Claude Code API request
type APIRequest struct {
	SessionID           string
	Timestamp           time.Time
	Model               string
	InputTokens         int64
	OutputTokens        int64
	CacheReadTokens     int64
	CacheCreationTokens int64
	TotalTokens         int64
	CostUSD             float64
	DurationMS          int64
}

// TimeFilter represents available time filter options
type TimeFilter int

const (
	FilterAll TimeFilter = iota
	FilterHour
	FilterDay
	FilterWeek
	FilterMonth
)

// Filter represents the filter criteria for API requests
type Filter struct {
	TimeFilter TimeFilter
}
