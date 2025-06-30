package query

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/elct9620/ccmon/entity"
	pb "github.com/elct9620/ccmon/proto"
	"github.com/elct9620/ccmon/usecase"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Helper function for creating API requests in tests
func mustCreateAPIRequest(sessionID string, timestamp time.Time, model string, tokens entity.Token, cost entity.Cost, durationMS int64) entity.APIRequest {
	return entity.NewAPIRequest(sessionID, timestamp, model, tokens, cost, durationMS)
}

// Mock repository for testing
type mockAPIRequestRepository struct {
	requests []entity.APIRequest
	findErr  error
}

func (m *mockAPIRequestRepository) Save(req entity.APIRequest) error {
	m.requests = append(m.requests, req)
	return nil
}

func (m *mockAPIRequestRepository) FindByPeriodWithLimit(period entity.Period, limit int, offset int) ([]entity.APIRequest, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}

	// Filter requests by period
	var filtered []entity.APIRequest
	for _, req := range m.requests {
		if !period.IsAllTime() {
			if req.Timestamp().Before(period.StartAt()) || req.Timestamp().After(period.EndAt()) {
				continue
			}
		}
		filtered = append(filtered, req)
	}

	// Apply offset
	if offset > len(filtered) {
		return []entity.APIRequest{}, nil
	}
	filtered = filtered[offset:]

	// Apply limit
	if limit > 0 && limit < len(filtered) {
		filtered = filtered[:limit]
	}

	return filtered, nil
}

func (m *mockAPIRequestRepository) FindAll() ([]entity.APIRequest, error) {
	return m.requests, m.findErr
}

func TestQueryService_GetStats(t *testing.T) {
	baseTime := time.Date(2024, 6, 29, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name          string
		requests      []entity.APIRequest
		startTime     *time.Time
		endTime       *time.Time
		expectedStats func(t *testing.T, stats *pb.Stats)
		expectError   bool
	}{
		{
			name: "mixed_requests_all_time",
			requests: []entity.APIRequest{
				mustCreateAPIRequest(
					"session1", baseTime,
					"claude-3-haiku-20240307", // base model
					entity.NewToken(100, 50, 10, 5),
					entity.NewCost(0.15),
					1000,
				),
				mustCreateAPIRequest(
					"session2", baseTime.Add(time.Hour),
					"claude-3-sonnet-20240229", // premium model
					entity.NewToken(200, 100, 20, 10),
					entity.NewCost(0.70),
					1500,
				),
				mustCreateAPIRequest(
					"session3", baseTime.Add(2*time.Hour),
					"claude-3-opus-20240229", // premium model
					entity.NewToken(300, 150, 30, 15),
					entity.NewCost(1.50),
					2000,
				),
			},
			startTime: nil, // All time
			endTime:   nil,
			expectedStats: func(t *testing.T, stats *pb.Stats) {
				if stats.BaseRequests != 1 {
					t.Errorf("Expected 1 base request, got %d", stats.BaseRequests)
				}
				if stats.PremiumRequests != 2 {
					t.Errorf("Expected 2 premium requests, got %d", stats.PremiumRequests)
				}
				if stats.TotalRequests != 3 {
					t.Errorf("Expected 3 total requests, got %d", stats.TotalRequests)
				}

				// Base tokens: 100+50+10+5 = 165
				if stats.BaseTokens.Total != 165 {
					t.Errorf("Expected 165 base tokens, got %d", stats.BaseTokens.Total)
				}

				// Premium tokens: (200+100+20+10) + (300+150+30+15) = 330 + 495 = 825
				if stats.PremiumTokens.Total != 825 {
					t.Errorf("Expected 825 premium tokens, got %d", stats.PremiumTokens.Total)
				}

				// Total tokens: 165 + 825 = 990
				if stats.TotalTokens.Total != 990 {
					t.Errorf("Expected 990 total tokens, got %d", stats.TotalTokens.Total)
				}

				// Total cost: 0.15 + 0.70 + 1.50 = 2.35
				if stats.TotalCost.Amount != 2.35 {
					t.Errorf("Expected $2.35 total cost, got $%.2f", stats.TotalCost.Amount)
				}
			},
			expectError: false,
		},
		{
			name: "time_filtered_requests",
			requests: []entity.APIRequest{
				mustCreateAPIRequest(
					"old", baseTime.Add(-2*time.Hour), // Outside range
					"claude-3-sonnet-20240229",
					entity.NewToken(100, 50, 10, 5),
					entity.NewCost(0.50),
					1000,
				),
				mustCreateAPIRequest(
					"current", baseTime,
					"claude-3-sonnet-20240229",
					entity.NewToken(200, 100, 20, 10),
					entity.NewCost(1.00),
					1500,
				),
				mustCreateAPIRequest(
					"future", baseTime.Add(2*time.Hour), // Outside range
					"claude-3-sonnet-20240229",
					entity.NewToken(300, 150, 30, 15),
					entity.NewCost(1.50),
					2000,
				),
			},
			startTime: &baseTime,
			endTime:   func() *time.Time { t := baseTime.Add(time.Hour); return &t }(),
			expectedStats: func(t *testing.T, stats *pb.Stats) {
				// Only the "current" request should be included
				if stats.TotalRequests != 1 {
					t.Errorf("Expected 1 filtered request, got %d", stats.TotalRequests)
				}
				if stats.PremiumRequests != 1 {
					t.Errorf("Expected 1 premium request, got %d", stats.PremiumRequests)
				}
				if stats.BaseRequests != 0 {
					t.Errorf("Expected 0 base requests, got %d", stats.BaseRequests)
				}
				if stats.TotalCost.Amount != 1.00 {
					t.Errorf("Expected $1.00 cost, got $%.2f", stats.TotalCost.Amount)
				}
			},
			expectError: false,
		},
		{
			name:      "empty_repository",
			requests:  []entity.APIRequest{},
			startTime: nil,
			endTime:   nil,
			expectedStats: func(t *testing.T, stats *pb.Stats) {
				if stats.TotalRequests != 0 {
					t.Errorf("Expected 0 requests, got %d", stats.TotalRequests)
				}
				if stats.TotalTokens.Total != 0 {
					t.Errorf("Expected 0 tokens, got %d", stats.TotalTokens.Total)
				}
				if stats.TotalCost.Amount != 0 {
					t.Errorf("Expected $0 cost, got $%.2f", stats.TotalCost.Amount)
				}
			},
			expectError: false,
		},
		{
			name: "only_base_models",
			requests: []entity.APIRequest{
				mustCreateAPIRequest(
					"haiku1", baseTime,
					"claude-3-haiku-20240307",
					entity.NewToken(100, 50, 10, 5),
					entity.NewCost(0.10),
					800,
				),
				mustCreateAPIRequest(
					"haiku2", baseTime.Add(time.Hour),
					"claude-3-haiku-20240307",
					entity.NewToken(150, 75, 15, 8),
					entity.NewCost(0.12),
					900,
				),
			},
			startTime: nil,
			endTime:   nil,
			expectedStats: func(t *testing.T, stats *pb.Stats) {
				if stats.BaseRequests != 2 {
					t.Errorf("Expected 2 base requests, got %d", stats.BaseRequests)
				}
				if stats.PremiumRequests != 0 {
					t.Errorf("Expected 0 premium requests, got %d", stats.PremiumRequests)
				}
				if stats.TotalRequests != 2 {
					t.Errorf("Expected 2 total requests, got %d", stats.TotalRequests)
				}
				// Base cost should equal total cost
				if stats.BaseCost.Amount != stats.TotalCost.Amount {
					t.Errorf("Base cost (%.2f) should equal total cost (%.2f)",
						stats.BaseCost.Amount, stats.TotalCost.Amount)
				}
			},
			expectError: false,
		},
		{
			name: "only_premium_models",
			requests: []entity.APIRequest{
				mustCreateAPIRequest(
					"sonnet", baseTime,
					"claude-3-sonnet-20240229",
					entity.NewToken(200, 100, 20, 10),
					entity.NewCost(1.00),
					1500,
				),
				mustCreateAPIRequest(
					"opus", baseTime.Add(time.Hour),
					"claude-3-opus-20240229",
					entity.NewToken(400, 200, 40, 20),
					entity.NewCost(3.00),
					2500,
				),
			},
			startTime: nil,
			endTime:   nil,
			expectedStats: func(t *testing.T, stats *pb.Stats) {
				if stats.BaseRequests != 0 {
					t.Errorf("Expected 0 base requests, got %d", stats.BaseRequests)
				}
				if stats.PremiumRequests != 2 {
					t.Errorf("Expected 2 premium requests, got %d", stats.PremiumRequests)
				}
				if stats.TotalRequests != 2 {
					t.Errorf("Expected 2 total requests, got %d", stats.TotalRequests)
				}
				// Premium cost should equal total cost
				if stats.PremiumCost.Amount != stats.TotalCost.Amount {
					t.Errorf("Premium cost (%.2f) should equal total cost (%.2f)",
						stats.PremiumCost.Amount, stats.TotalCost.Amount)
				}
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock repository
			mockRepo := &mockAPIRequestRepository{
				requests: tt.requests,
			}

			// Create usecases
			calculateStatsQuery := usecase.NewCalculateStatsQuery(mockRepo)

			// Create service
			service := NewService(nil, calculateStatsQuery) // getFilteredQuery not needed for this test

			// Create request
			req := &pb.GetStatsRequest{}
			if tt.startTime != nil {
				req.StartTime = timestamppb.New(*tt.startTime)
			}
			if tt.endTime != nil {
				req.EndTime = timestamppb.New(*tt.endTime)
			}

			// Call service
			ctx := context.Background()
			resp, err := service.GetStats(ctx, req)

			// Check error expectation
			if tt.expectError && err == nil {
				t.Fatal("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Validate response if no error expected
			if !tt.expectError {
				if resp == nil {
					t.Fatal("Expected non-nil response")
				}
				if resp.Stats == nil {
					t.Fatal("Expected non-nil stats")
				}
				tt.expectedStats(t, resp.Stats)
			}
		})
	}
}

func TestQueryService_GetAPIRequests(t *testing.T) {
	baseTime := time.Date(2024, 6, 29, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name             string
		requests         []entity.APIRequest
		requestParams    *pb.GetAPIRequestsRequest
		expectedCount    int
		validateFirstReq func(t *testing.T, req *pb.APIRequest)
		expectError      bool
	}{
		{
			name: "all_requests_no_pagination",
			requests: []entity.APIRequest{
				mustCreateAPIRequest(
					"session1", baseTime,
					"claude-3-sonnet-20240229",
					entity.NewToken(100, 50, 10, 5),
					entity.NewCost(0.50),
					1000,
				),
				mustCreateAPIRequest(
					"session2", baseTime.Add(time.Hour),
					"claude-3-haiku-20240307",
					entity.NewToken(200, 100, 20, 10),
					entity.NewCost(0.25),
					800,
				),
			},
			requestParams: &pb.GetAPIRequestsRequest{
				StartTime: nil,
				EndTime:   nil,
				Limit:     0, // No limit
				Offset:    0,
			},
			expectedCount: 2,
			validateFirstReq: func(t *testing.T, req *pb.APIRequest) {
				if req.SessionId != "session1" {
					t.Errorf("Expected session1, got %s", req.SessionId)
				}
				if req.Model != "claude-3-sonnet-20240229" {
					t.Errorf("Expected claude-3-sonnet-20240229, got %s", req.Model)
				}
				if req.TotalTokens != 165 { // 100+50+10+5
					t.Errorf("Expected 165 total tokens, got %d", req.TotalTokens)
				}
			},
			expectError: false,
		},
		{
			name: "pagination_with_limit",
			requests: func() []entity.APIRequest {
				var reqs []entity.APIRequest
				for i := 0; i < 5; i++ {
					reqs = append(reqs, mustCreateAPIRequest(
						fmt.Sprintf("session%d", i),
						baseTime.Add(time.Duration(i)*time.Hour),
						"claude-3-sonnet-20240229",
						entity.NewToken(100, 50, 10, 5),
						entity.NewCost(0.50),
						1000,
					))
				}
				return reqs
			}(),
			requestParams: &pb.GetAPIRequestsRequest{
				StartTime: nil,
				EndTime:   nil,
				Limit:     3, // Limit to 3 results
				Offset:    0,
			},
			expectedCount: 3,
			validateFirstReq: func(t *testing.T, req *pb.APIRequest) {
				if req.SessionId != "session0" {
					t.Errorf("Expected session0, got %s", req.SessionId)
				}
			},
			expectError: false,
		},
		{
			name: "pagination_with_offset",
			requests: func() []entity.APIRequest {
				var reqs []entity.APIRequest
				for i := 0; i < 5; i++ {
					reqs = append(reqs, mustCreateAPIRequest(
						fmt.Sprintf("session%d", i),
						baseTime.Add(time.Duration(i)*time.Hour),
						"claude-3-sonnet-20240229",
						entity.NewToken(100, 50, 10, 5),
						entity.NewCost(0.50),
						1000,
					))
				}
				return reqs
			}(),
			requestParams: &pb.GetAPIRequestsRequest{
				StartTime: nil,
				EndTime:   nil,
				Limit:     2,
				Offset:    2, // Skip first 2
			},
			expectedCount: 2,
			validateFirstReq: func(t *testing.T, req *pb.APIRequest) {
				if req.SessionId != "session2" {
					t.Errorf("Expected session2 (after offset), got %s", req.SessionId)
				}
			},
			expectError: false,
		},
		{
			name: "time_filtering",
			requests: []entity.APIRequest{
				mustCreateAPIRequest(
					"old", baseTime.Add(-2*time.Hour),
					"claude-3-sonnet-20240229",
					entity.NewToken(100, 50, 10, 5),
					entity.NewCost(0.50),
					1000,
				),
				mustCreateAPIRequest(
					"current", baseTime,
					"claude-3-sonnet-20240229",
					entity.NewToken(200, 100, 20, 10),
					entity.NewCost(1.00),
					1500,
				),
				mustCreateAPIRequest(
					"future", baseTime.Add(2*time.Hour),
					"claude-3-sonnet-20240229",
					entity.NewToken(300, 150, 30, 15),
					entity.NewCost(1.50),
					2000,
				),
			},
			requestParams: &pb.GetAPIRequestsRequest{
				StartTime: timestamppb.New(baseTime),
				EndTime:   timestamppb.New(baseTime.Add(time.Hour)),
				Limit:     0,
				Offset:    0,
			},
			expectedCount: 1,
			validateFirstReq: func(t *testing.T, req *pb.APIRequest) {
				if req.SessionId != "current" {
					t.Errorf("Expected current (time filtered), got %s", req.SessionId)
				}
			},
			expectError: false,
		},
		{
			name:     "empty_repository",
			requests: []entity.APIRequest{},
			requestParams: &pb.GetAPIRequestsRequest{
				StartTime: nil,
				EndTime:   nil,
				Limit:     0,
				Offset:    0,
			},
			expectedCount:    0,
			validateFirstReq: nil,
			expectError:      false,
		},
		{
			name: "offset_beyond_results",
			requests: []entity.APIRequest{
				mustCreateAPIRequest(
					"session1", baseTime,
					"claude-3-sonnet-20240229",
					entity.NewToken(100, 50, 10, 5),
					entity.NewCost(0.50),
					1000,
				),
			},
			requestParams: &pb.GetAPIRequestsRequest{
				StartTime: nil,
				EndTime:   nil,
				Limit:     10,
				Offset:    5, // Beyond available results
			},
			expectedCount:    0,
			validateFirstReq: nil,
			expectError:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock repository
			mockRepo := &mockAPIRequestRepository{
				requests: tt.requests,
			}

			// Create usecases
			getFilteredQuery := usecase.NewGetFilteredApiRequestsQuery(mockRepo)

			// Create service
			service := NewService(getFilteredQuery, nil) // calculateStatsQuery not needed for this test

			// Call service
			ctx := context.Background()
			resp, err := service.GetAPIRequests(ctx, tt.requestParams)

			// Check error expectation
			if tt.expectError && err == nil {
				t.Fatal("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Validate response if no error expected
			if !tt.expectError {
				if resp == nil {
					t.Fatal("Expected non-nil response")
				}

				if len(resp.Requests) != tt.expectedCount {
					t.Errorf("Expected %d requests, got %d", tt.expectedCount, len(resp.Requests))
				}

				if int(resp.TotalCount) != len(resp.Requests) {
					t.Errorf("TotalCount (%d) should match returned count (%d)", resp.TotalCount, len(resp.Requests))
				}

				// Validate first request if validation function provided and requests exist
				if tt.validateFirstReq != nil && len(resp.Requests) > 0 {
					tt.validateFirstReq(t, resp.Requests[0])
				}
			}
		})
	}
}

func TestQueryService_ConvertTimestampsToPeriod(t *testing.T) {
	baseTime := time.Date(2024, 6, 29, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		startTime *timestamppb.Timestamp
		endTime   *timestamppb.Timestamp
		validate  func(t *testing.T, period entity.Period)
	}{
		{
			name:      "both_nil_all_time",
			startTime: nil,
			endTime:   nil,
			validate: func(t *testing.T, period entity.Period) {
				if !period.IsAllTime() {
					t.Error("Expected all time period")
				}
			},
		},
		{
			name:      "start_nil_end_set",
			startTime: nil,
			endTime:   timestamppb.New(baseTime),
			validate: func(t *testing.T, period entity.Period) {
				if period.StartAt() != (time.Time{}) {
					t.Error("Expected zero start time")
				}
				if !period.EndAt().Equal(baseTime) {
					t.Errorf("Expected end time %v, got %v", baseTime, period.EndAt())
				}
			},
		},
		{
			name:      "start_set_end_nil",
			startTime: timestamppb.New(baseTime),
			endTime:   nil,
			validate: func(t *testing.T, period entity.Period) {
				if !period.StartAt().Equal(baseTime) {
					t.Errorf("Expected start time %v, got %v", baseTime, period.StartAt())
				}
				// End time should be approximately now
				if period.EndAt().Before(time.Now().Add(-time.Minute)) {
					t.Error("End time should be close to current time")
				}
			},
		},
		{
			name:      "both_set",
			startTime: timestamppb.New(baseTime),
			endTime:   timestamppb.New(baseTime.Add(time.Hour)),
			validate: func(t *testing.T, period entity.Period) {
				if !period.StartAt().Equal(baseTime) {
					t.Errorf("Expected start time %v, got %v", baseTime, period.StartAt())
				}
				if !period.EndAt().Equal(baseTime.Add(time.Hour)) {
					t.Errorf("Expected end time %v, got %v", baseTime.Add(time.Hour), period.EndAt())
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			period := convertTimestampsToPeriod(tt.startTime, tt.endTime)
			tt.validate(t, period)
		})
	}
}
