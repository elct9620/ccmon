package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/elct9620/ccmon/entity"
)

// MockAPIRequestRepositoryForUsage implements APIRequestRepository for usage testing
type MockAPIRequestRepositoryForUsage struct {
	requests []entity.APIRequest
	err      error
}

func (m *MockAPIRequestRepositoryForUsage) FindByPeriodWithLimit(period entity.Period, limit int, offset int) ([]entity.APIRequest, error) {
	if m.err != nil {
		return nil, m.err
	}

	// Filter requests by period
	var filtered []entity.APIRequest
	for _, req := range m.requests {
		if period.IsAllTime() {
			filtered = append(filtered, req)
		} else if !req.Timestamp().Before(period.StartAt()) && !req.Timestamp().After(period.EndAt()) {
			// Request timestamp is within the period (inclusive of start, inclusive of end)
			filtered = append(filtered, req)
		}
	}

	return filtered, nil
}

func (m *MockAPIRequestRepositoryForUsage) FindAll() ([]entity.APIRequest, error) {
	return m.requests, nil
}

func (m *MockAPIRequestRepositoryForUsage) Save(req entity.APIRequest) error {
	m.requests = append(m.requests, req)
	return nil
}

func (m *MockAPIRequestRepositoryForUsage) DeleteOlderThan(cutoffTime time.Time) (int, error) {
	return 0, nil
}

func TestGetUsageQuery_ListByDay(t *testing.T) {
	// Create test API requests for a specific day
	now := time.Now().UTC()
	yesterday := now.AddDate(0, 0, -1)

	// Create requests for yesterday
	req1 := entity.NewAPIRequest("session1", yesterday.Add(2*time.Hour), "claude-3-5-haiku-20241022", entity.NewToken(100, 50, 0, 0), entity.NewCost(0.001), 1500)
	req2 := entity.NewAPIRequest("session2", yesterday.Add(4*time.Hour), "claude-3-5-sonnet-20241022", entity.NewToken(200, 100, 0, 0), entity.NewCost(0.002), 2000)

	repo := &MockAPIRequestRepositoryForUsage{
		requests: []entity.APIRequest{req1, req2},
	}
	query := NewGetUsageQuery(repo)

	usage, err := query.ListByDay(context.Background(), 2, time.UTC) // Get last 2 days
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(usage.GetStats()) != 2 {
		t.Errorf("Expected 2 stats, got %d", len(usage.GetStats()))
	}

	// The first stat should be for today (empty)
	todayStats := usage.GetStats()[0]
	if todayStats.TotalRequests() != 0 {
		t.Errorf("Expected 0 requests for today, got %d", todayStats.TotalRequests())
	}

	// The second stat should be for yesterday (with our test data)
	yesterdayStats := usage.GetStats()[1]
	if yesterdayStats.TotalRequests() != 2 {
		t.Errorf("Expected 2 requests for yesterday, got %d", yesterdayStats.TotalRequests())
	}
}

func TestGetUsageQuery_ListByDay_Error(t *testing.T) {
	repo := &MockAPIRequestRepositoryForUsage{err: &MockError{message: "database error"}}
	query := NewGetUsageQuery(repo)

	_, err := query.ListByDay(context.Background(), 30, time.UTC)
	if err == nil {
		t.Error("Expected error, got nil")
	}

	if err.Error() != "database error" {
		t.Errorf("Expected 'database error', got '%s'", err.Error())
	}
}

// MockError implements error interface for testing
type MockError struct {
	message string
}

func (e *MockError) Error() string {
	return e.message
}

func TestGetUsageQuery_ListByDay_Timezone(t *testing.T) {
	// Test that requests are grouped by days in the specified timezone
	// Create a request near midnight UTC that would be on different days in different timezones

	// Use a time yesterday at 11:30 PM UTC
	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)
	utcTime := time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 23, 30, 0, 0, time.UTC)

	// Load timezones
	tokyoTZ, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		t.Fatalf("Failed to load Tokyo timezone: %v", err)
	}

	newYorkTZ, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatalf("Failed to load New York timezone: %v", err)
	}

	// Create requests at 11:30 PM UTC yesterday
	// In Tokyo (UTC+9), this is 8:30 AM today
	// In New York (UTC-5), this is 6:30 PM yesterday
	req := entity.NewAPIRequest("session1", utcTime, "claude-3-5-sonnet-20241022", entity.NewToken(100, 50, 0, 0), entity.NewCost(0.001), 1500)

	repo := &MockAPIRequestRepositoryForUsage{
		requests: []entity.APIRequest{req},
	}
	query := NewGetUsageQuery(repo)

	tests := []struct {
		name             string
		timezone         *time.Location
		expectedDayStart time.Time
		description      string
	}{
		{
			name:             "UTC timezone - request on yesterday",
			timezone:         time.UTC,
			expectedDayStart: time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, time.UTC),
			description:      "11:30 PM UTC should be on the same day in UTC",
		},
		{
			name:             "New York timezone - request on yesterday",
			timezone:         newYorkTZ,
			expectedDayStart: time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, newYorkTZ).UTC(),
			description:      "11:30 PM UTC is 6:30 PM in New York, still yesterday",
		},
		{
			name:             "Tokyo timezone - request on today",
			timezone:         tokyoTZ,
			expectedDayStart: time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, tokyoTZ).UTC(),
			description:      "11:30 PM UTC is 8:30 AM next day in Tokyo",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			usage, err := query.ListByDay(context.Background(), 2, tc.timezone)
			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}

			// Find which day has the request
			var dayWithRequest time.Time
			found := false
			for _, stat := range usage.GetStats() {
				if stat.TotalRequests() > 0 {
					dayWithRequest = stat.Period().StartAt()
					found = true
					break
				}
			}

			if !found {
				t.Fatal("Expected to find a day with the request")
			}

			if !dayWithRequest.Equal(tc.expectedDayStart) {
				t.Errorf("%s: expected request period to start at %v (UTC), got %v", tc.description, tc.expectedDayStart, dayWithRequest)
			}
		})
	}
}

func TestGetUsageQuery_ListByDay_NilTimezone(t *testing.T) {
	// Test that nil timezone defaults to UTC
	req := entity.NewAPIRequest("session1", time.Now(), "claude-3-5-haiku-20241022", entity.NewToken(100, 50, 0, 0), entity.NewCost(0.001), 1500)

	repo := &MockAPIRequestRepositoryForUsage{
		requests: []entity.APIRequest{req},
	}
	query := NewGetUsageQuery(repo)

	// Pass nil timezone
	usage, err := query.ListByDay(context.Background(), 1, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(usage.GetStats()) != 1 {
		t.Errorf("Expected 1 stat, got %d", len(usage.GetStats()))
	}

	// Verify request was counted
	stat := usage.GetStats()[0]
	if stat.TotalRequests() != 1 {
		t.Errorf("Expected 1 request, got %d", stat.TotalRequests())
	}
}
