package repository

import (
	"fmt"
	"testing"
	"time"

	"github.com/elct9620/ccmon/entity"
	"github.com/elct9620/ccmon/testutil"
)

func TestBoltDBStatsRepository_GetStatsByPeriod(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		requests      []entity.APIRequest
		period        entity.Period
		repositoryErr error
		expectedStats entity.Stats
		expectError   bool
	}{
		{
			name: "calculate stats from mixed requests",
			requests: []entity.APIRequest{
				entity.NewAPIRequest(
					"session1",
					time.Date(2025, 7, 24, 10, 0, 0, 0, time.UTC),
					"claude-3-haiku-20240307", // base model
					entity.NewToken(100, 80, 0, 0),
					entity.NewCost(5.0),
					1000,
				),
				entity.NewAPIRequest(
					"session2",
					time.Date(2025, 7, 24, 11, 0, 0, 0, time.UTC),
					"claude-3-5-sonnet-20241022", // premium model
					entity.NewToken(200, 150, 0, 0),
					entity.NewCost(10.0),
					2000,
				),
			},
			period: entity.NewPeriod(
				time.Date(2025, 7, 24, 0, 0, 0, 0, time.UTC),
				time.Date(2025, 7, 24, 23, 59, 59, 999999999, time.UTC),
			),
			expectedStats: entity.NewStats(
				1,                               // base requests
				1,                               // premium requests
				entity.NewToken(100, 80, 0, 0),  // base tokens
				entity.NewToken(200, 150, 0, 0), // premium tokens
				entity.NewCost(5.0),             // base cost
				entity.NewCost(10.0),            // premium cost
				entity.NewPeriod(
					time.Date(2025, 7, 24, 0, 0, 0, 0, time.UTC),
					time.Date(2025, 7, 24, 23, 59, 59, 999999999, time.UTC),
				),
			),
			expectError: false,
		},
		{
			name: "empty period returns zero stats",
			requests: []entity.APIRequest{
				entity.NewAPIRequest(
					"session1",
					time.Date(2025, 7, 23, 10, 0, 0, 0, time.UTC), // outside period
					"claude-3-haiku-20240307",
					entity.NewToken(100, 80, 0, 0),
					entity.NewCost(5.0),
					1000,
				),
			},
			period: entity.NewPeriod(
				time.Date(2025, 7, 24, 0, 0, 0, 0, time.UTC),
				time.Date(2025, 7, 24, 23, 59, 59, 999999999, time.UTC),
			),
			expectedStats: entity.NewStats(
				0, 0, // no requests
				entity.NewToken(0, 0, 0, 0), entity.NewToken(0, 0, 0, 0), // no tokens
				entity.NewCost(0), entity.NewCost(0), // no cost
				entity.NewPeriod(
					time.Date(2025, 7, 24, 0, 0, 0, 0, time.UTC),
					time.Date(2025, 7, 24, 23, 59, 59, 999999999, time.UTC),
				),
			),
			expectError: false,
		},
		{
			name:          "repository error",
			requests:      []entity.APIRequest{},
			period:        entity.NewAllTimePeriod(time.Now()),
			repositoryErr: fmt.Errorf("database connection failed"),
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup mock repository using testutil factory
			mockRepo := testutil.NewMockAPIRequestRepository()
			mockRepo.SetMockData(tt.requests)
			if tt.repositoryErr != nil {
				mockRepo.SetError(tt.repositoryErr)
			}

			// Create BoltDBStatsRepository
			statsRepo := NewBoltDBStatsRepository(mockRepo)

			// Execute
			result, err := statsRepo.GetStatsByPeriod(tt.period)

			// Verify error expectation
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Verify stats
			if result.BaseRequests() != tt.expectedStats.BaseRequests() {
				t.Errorf("Base requests: expected %d, got %d", tt.expectedStats.BaseRequests(), result.BaseRequests())
			}
			if result.PremiumRequests() != tt.expectedStats.PremiumRequests() {
				t.Errorf("Premium requests: expected %d, got %d", tt.expectedStats.PremiumRequests(), result.PremiumRequests())
			}
			if result.TotalCost().Amount() != tt.expectedStats.TotalCost().Amount() {
				t.Errorf("Total cost: expected %.1f, got %.1f", tt.expectedStats.TotalCost().Amount(), result.TotalCost().Amount())
			}
		})
	}
}
