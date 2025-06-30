package entity

import (
	"testing"
	"time"
)

func TestNewUsage(t *testing.T) {
	// Create test stats with periods
	now := time.Now()
	period1 := NewPeriod(now.AddDate(0, 0, -2), now.AddDate(0, 0, -1))
	period2 := NewPeriod(now.AddDate(0, 0, -1), now)

	stats1 := NewStats(1, 2, NewToken(100, 50, 0, 0), NewToken(200, 100, 0, 0), NewCost(0.001), NewCost(0.002), period1)
	stats2 := NewStats(2, 1, NewToken(150, 75, 0, 0), NewToken(100, 50, 0, 0), NewCost(0.0015), NewCost(0.001), period2)

	stats := []Stats{stats1, stats2}
	usage := NewUsage(stats)

	if len(usage.GetStats()) != 2 {
		t.Errorf("Expected 2 stats, got %d", len(usage.GetStats()))
	}

	if usage.GetStats()[0].TotalRequests() != 3 {
		t.Errorf("Expected 3 total requests for first stat, got %d", usage.GetStats()[0].TotalRequests())
	}

	if usage.GetStats()[1].TotalRequests() != 3 {
		t.Errorf("Expected 3 total requests for second stat, got %d", usage.GetStats()[1].TotalRequests())
	}
}

func TestUsageGetStats(t *testing.T) {
	// Create empty usage
	usage := NewUsage([]Stats{})

	if len(usage.GetStats()) != 0 {
		t.Errorf("Expected 0 stats, got %d", len(usage.GetStats()))
	}
}
