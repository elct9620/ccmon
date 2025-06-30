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

// NewPeriodFromDuration creates a Period from given time minus duration
func NewPeriodFromDuration(now time.Time, duration time.Duration) Period {
	return NewPeriod(now.Add(-duration), now)
}

// NewAllTimePeriod creates a Period representing all time (zero start time)
func NewAllTimePeriod(now time.Time) Period {
	return NewPeriod(time.Time{}, now)
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
