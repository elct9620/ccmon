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
		} else if req.Timestamp().After(period.StartAt()) && req.Timestamp().Before(period.EndAt()) {
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

	usage, err := query.ListByDay(context.Background(), 2) // Get last 2 days
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

	_, err := query.ListByDay(context.Background(), 30)
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
