package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/elct9620/ccmon/entity"
)

// mockAPIRequestRepository for testing - implements both APIRequestRepository and StatsRepository
type mockAPIRequestRepository struct {
	findByPeriodWithLimitFunc func(period entity.Period, limit int, offset int) ([]entity.APIRequest, error)
}

func (m *mockAPIRequestRepository) Save(request entity.APIRequest) error {
	return nil
}

func (m *mockAPIRequestRepository) FindByPeriodWithLimit(period entity.Period, limit int, offset int) ([]entity.APIRequest, error) {
	if m.findByPeriodWithLimitFunc != nil {
		return m.findByPeriodWithLimitFunc(period, limit, offset)
	}
	return []entity.APIRequest{}, nil
}

func (m *mockAPIRequestRepository) FindAll() ([]entity.APIRequest, error) {
	return []entity.APIRequest{}, nil
}

func (m *mockAPIRequestRepository) DeleteOlderThan(timestamp time.Time) (int, error) {
	return 0, nil
}

// mockStatsRepository for testing - wraps mockAPIRequestRepository
type mockStatsRepository struct {
	apiRepo *mockAPIRequestRepository
}

func newMockStatsRepository(apiRepo *mockAPIRequestRepository) *mockStatsRepository {
	return &mockStatsRepository{apiRepo: apiRepo}
}

func (m *mockStatsRepository) GetStatsByPeriod(period entity.Period) (entity.Stats, error) {
	requests, err := m.apiRepo.FindByPeriodWithLimit(period, 0, 0)
	if err != nil {
		return entity.Stats{}, err
	}
	return entity.NewStatsFromRequests(requests, period), nil
}

// mockStatsCache for testing
type mockStatsCache struct {
	getFunc   func(period entity.Period) *entity.Stats
	setFunc   func(period entity.Period, stats *entity.Stats)
	getCalled int
	setCalled int
}

func (m *mockStatsCache) Get(period entity.Period) *entity.Stats {
	m.getCalled++
	if m.getFunc != nil {
		return m.getFunc(period)
	}
	return nil
}

func (m *mockStatsCache) Set(period entity.Period, stats *entity.Stats) {
	m.setCalled++
	if m.setFunc != nil {
		m.setFunc(period, stats)
	}
}

func TestCalculateStatsQuery_Execute(t *testing.T) {
	now := time.Now()
	period := entity.NewPeriod(now.Add(-1*time.Hour), now)

	// Create test data
	cachedStats := entity.NewStats(
		10, 5,
		entity.NewToken(1000, 500, 50, 1550),
		entity.NewToken(2000, 1000, 100, 3100),
		entity.NewCost(0.05),
		entity.NewCost(0.15),
		period,
	)

	baseRequest := entity.NewAPIRequest(
		"req1",
		now.Add(-30*time.Minute),
		"claude-3-5-haiku-20241022",
		entity.NewToken(100, 50, 5, 0), // Total: 155
		entity.NewCost(0.01),
		1000,
	)

	premiumRequest := entity.NewAPIRequest(
		"req2",
		now.Add(-20*time.Minute),
		"claude-3-5-sonnet-20241022",
		entity.NewToken(200, 100, 10, 0), // Total: 310
		entity.NewCost(0.03),
		2000,
	)

	testError := errors.New("repository error")

	tests := []struct {
		name             string
		cacheGet         *entity.Stats
		repositoryData   []entity.APIRequest
		repositoryError  error
		expectError      bool
		expectRepoCalled bool
		expectCacheSet   bool
		validateResult   func(*testing.T, entity.Stats)
	}{
		{
			name:             "cache hit returns cached stats",
			cacheGet:         &cachedStats,
			expectRepoCalled: false,
			expectCacheSet:   false,
			validateResult: func(t *testing.T, result entity.Stats) {
				if result != cachedStats {
					t.Error("Expected cached stats to be returned")
				}
			},
		},
		{
			name:             "cache miss calculates and caches stats",
			cacheGet:         nil,
			repositoryData:   []entity.APIRequest{baseRequest, premiumRequest},
			expectRepoCalled: true,
			expectCacheSet:   true,
			validateResult: func(t *testing.T, result entity.Stats) {
				if result.BaseRequests() != 1 {
					t.Errorf("Expected 1 base request, got %d", result.BaseRequests())
				}
				if result.PremiumRequests() != 1 {
					t.Errorf("Expected 1 premium request, got %d", result.PremiumRequests())
				}
				if result.BaseTokens().Total() != 155 {
					t.Errorf("Expected base tokens total 155, got %d", result.BaseTokens().Total())
				}
				if result.PremiumTokens().Total() != 310 {
					t.Errorf("Expected premium tokens total 310, got %d", result.PremiumTokens().Total())
				}
				if result.BaseCost().Amount() != 0.01 {
					t.Errorf("Expected base cost 0.01, got %f", result.BaseCost().Amount())
				}
				if result.PremiumCost().Amount() != 0.03 {
					t.Errorf("Expected premium cost 0.03, got %f", result.PremiumCost().Amount())
				}
			},
		},
		{
			name:             "empty requests returns zero stats and caches",
			cacheGet:         nil,
			repositoryData:   []entity.APIRequest{},
			expectRepoCalled: true,
			expectCacheSet:   true,
			validateResult: func(t *testing.T, result entity.Stats) {
				if result.TotalRequests() != 0 {
					t.Errorf("Expected 0 total requests, got %d", result.TotalRequests())
				}
				if result.TotalTokens().Total() != 0 {
					t.Errorf("Expected 0 total tokens, got %d", result.TotalTokens().Total())
				}
				if result.TotalCost().Amount() != 0 {
					t.Errorf("Expected 0 total cost, got %f", result.TotalCost().Amount())
				}
			},
		},
		{
			name:             "repository error does not cache",
			cacheGet:         nil,
			repositoryError:  testError,
			expectError:      true,
			expectRepoCalled: true,
			expectCacheSet:   false,
			validateResult: func(t *testing.T, result entity.Stats) {
				if result != (entity.Stats{}) {
					t.Error("Expected empty stats on error")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			repoCalled := false
			mockRepo := &mockAPIRequestRepository{
				findByPeriodWithLimitFunc: func(p entity.Period, limit int, offset int) ([]entity.APIRequest, error) {
					repoCalled = true
					if tt.repositoryError != nil {
						return nil, tt.repositoryError
					}
					return tt.repositoryData, nil
				},
			}

			mockCache := &mockStatsCache{
				getFunc: func(p entity.Period) *entity.Stats {
					return tt.cacheGet
				},
			}

			// Execute
			mockStatsRepo := newMockStatsRepository(mockRepo)
			query := NewCalculateStatsQuery(mockStatsRepo, mockCache)
			ctx := context.Background()
			params := CalculateStatsParams{Period: period}
			result, err := query.Execute(ctx, params)

			// Verify error
			if tt.expectError && err == nil {
				t.Fatal("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Verify repository call
			if tt.expectRepoCalled != repoCalled {
				t.Errorf("Repository called = %v, want %v", repoCalled, tt.expectRepoCalled)
			}

			// Verify cache interactions
			if mockCache.getCalled != 1 {
				t.Errorf("Cache.Get called %d times, want 1", mockCache.getCalled)
			}
			if tt.expectCacheSet && mockCache.setCalled != 1 {
				t.Errorf("Cache.Set called %d times, want 1", mockCache.setCalled)
			}
			if !tt.expectCacheSet && mockCache.setCalled != 0 {
				t.Errorf("Cache.Set called %d times, want 0", mockCache.setCalled)
			}

			// Validate result
			if tt.validateResult != nil {
				tt.validateResult(t, result)
			}
		})
	}
}
