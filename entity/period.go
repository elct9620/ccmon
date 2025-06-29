package entity

import "time"

// Period represents a time range value object
type Period struct {
	startAt time.Time
	endAt   time.Time
}

// NewPeriod creates a new Period value object
func NewPeriod(startAt, endAt time.Time) Period {
	return Period{
		startAt: startAt,
		endAt:   endAt,
	}
}

// NewPeriodFromDuration creates a Period from current time minus duration
func NewPeriodFromDuration(duration time.Duration) Period {
	now := time.Now().UTC()
	return Period{
		startAt: now.Add(-duration),
		endAt:   now,
	}
}

// NewPeriodFromDurationWithTimezone creates a Period from current time minus duration in specified timezone
func NewPeriodFromDurationWithTimezone(duration time.Duration, timezone *time.Location) Period {
	now := time.Now().In(timezone)
	startAt := now.Add(-duration).UTC()
	endAt := now.UTC()
	return Period{
		startAt: startAt,
		endAt:   endAt,
	}
}

// NewAllTimePeriod creates a Period representing all time (zero start time)
func NewAllTimePeriod() Period {
	return Period{
		startAt: time.Time{}, // Zero time represents "all time"
		endAt:   time.Now().UTC(),
	}
}

// StartAt returns the start time of the period
func (p Period) StartAt() time.Time {
	return p.startAt
}

// EndAt returns the end time of the period
func (p Period) EndAt() time.Time {
	return p.endAt
}

// IsAllTime returns true if this period represents all time
func (p Period) IsAllTime() bool {
	return p.startAt.IsZero()
}

// Contains checks if the given time is within this period
func (p Period) Contains(t time.Time) bool {
	if p.IsAllTime() {
		return t.Before(p.endAt) || t.Equal(p.endAt)
	}
	return (t.After(p.startAt) || t.Equal(p.startAt)) && (t.Before(p.endAt) || t.Equal(p.endAt))
}

// Duration returns the duration of this period
func (p Period) Duration() time.Duration {
	if p.IsAllTime() {
		return 0 // Undefined for all-time periods
	}
	return p.endAt.Sub(p.startAt)
}
