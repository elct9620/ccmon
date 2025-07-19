package service

import (
	"time"

	"github.com/elct9620/ccmon/entity"
)

// TimePeriodFactory implements PeriodFactory using timezone-aware calculations
type TimePeriodFactory struct {
	timezone *time.Location
}

// NewTimePeriodFactory creates a new TimePeriodFactory with the given timezone
func NewTimePeriodFactory(timezone *time.Location) *TimePeriodFactory {
	if timezone == nil {
		timezone = time.UTC
	}
	return &TimePeriodFactory{
		timezone: timezone,
	}
}

// CreateDaily creates a period for today using timezone-aware boundaries
func (f *TimePeriodFactory) CreateDaily() entity.Period {
	now := time.Now().In(f.timezone)
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, f.timezone)
	dayEnd := dayStart.Add(24*time.Hour - time.Nanosecond)

	// Convert to UTC for database queries but maintain timezone-aware boundaries
	return entity.NewPeriod(dayStart.UTC(), dayEnd.UTC())
}

// CreateMonthly creates a period for current month using timezone-aware boundaries
func (f *TimePeriodFactory) CreateMonthly() entity.Period {
	now := time.Now().In(f.timezone)
	// First day of current month at 00:00:00 in user's timezone
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, f.timezone)
	// First day of next month minus 1 nanosecond to get end of current month
	monthEnd := monthStart.AddDate(0, 1, 0).Add(-time.Nanosecond)

	// Convert to UTC for database queries but maintain timezone-aware boundaries
	return entity.NewPeriod(monthStart.UTC(), monthEnd.UTC())
}
