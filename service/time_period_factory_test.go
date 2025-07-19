package service

import (
	"testing"
	"time"
)

func TestTimePeriodFactory(t *testing.T) {
	// Test with specific timezone
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatalf("Failed to load timezone: %v", err)
	}

	factory := NewTimePeriodFactory(loc)

	t.Run("CreateDaily", func(t *testing.T) {
		period := factory.CreateDaily()

		// Verify the period spans exactly one day
		duration := period.EndAt().Sub(period.StartAt())
		expectedDuration := 24*time.Hour - time.Nanosecond

		if duration != expectedDuration {
			t.Errorf("daily period duration: got %v, want %v", duration, expectedDuration)
		}

		// Verify times are in UTC (for database queries)
		if period.StartAt().Location() != time.UTC {
			t.Errorf("daily period start time not in UTC")
		}
		if period.EndAt().Location() != time.UTC {
			t.Errorf("daily period end time not in UTC")
		}
	})

	t.Run("CreateMonthly", func(t *testing.T) {
		period := factory.CreateMonthly()

		// Verify start is first day of month
		if period.StartAt().Day() != 1 {
			t.Errorf("monthly period should start on day 1, got %d", period.StartAt().Day())
		}

		// Verify end is last moment of month
		nextMonth := period.EndAt().Add(time.Nanosecond)
		if nextMonth.Day() != 1 {
			t.Errorf("monthly period should end at last moment of month")
		}

		// Verify times are in UTC (for database queries)
		if period.StartAt().Location() != time.UTC {
			t.Errorf("monthly period start time not in UTC")
		}
		if period.EndAt().Location() != time.UTC {
			t.Errorf("monthly period end time not in UTC")
		}
	})

	t.Run("CreateWithNilTimezone", func(t *testing.T) {
		// Should default to UTC
		factory := NewTimePeriodFactory(nil)
		period := factory.CreateDaily()

		if period.StartAt().Location() != time.UTC {
			t.Errorf("nil timezone should default to UTC")
		}
	})
}
