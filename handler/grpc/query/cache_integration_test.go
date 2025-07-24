package query

import (
	"context"
	"testing"
	"time"

	"github.com/elct9620/ccmon/entity"
	pb "github.com/elct9620/ccmon/proto"
	"github.com/elct9620/ccmon/service"
	"github.com/elct9620/ccmon/usecase"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// instrumentedRepository wraps a repository to count method calls
type instrumentedRepository struct {
	repo      usecase.APIRequestRepository
	callCount *int
}

func (r *instrumentedRepository) Save(req entity.APIRequest) error {
	return r.repo.Save(req)
}

func (r *instrumentedRepository) FindByPeriodWithLimit(period entity.Period, limit int, offset int) ([]entity.APIRequest, error) {
	*r.callCount++
	return r.repo.FindByPeriodWithLimit(period, limit, offset)
}

func (r *instrumentedRepository) FindAll() ([]entity.APIRequest, error) {
	return r.repo.FindAll()
}

func (r *instrumentedRepository) DeleteOlderThan(cutoffTime time.Time) (int, error) {
	return r.repo.DeleteOlderThan(cutoffTime)
}

// mockStatsRepository wraps mockAPIRequestRepository to implement StatsRepository
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

// instrumentedStatsRepository wraps instrumentedRepository to implement StatsRepository
type instrumentedStatsRepository struct {
	apiRepo *instrumentedRepository
}

func newInstrumentedStatsRepository(apiRepo *instrumentedRepository) *instrumentedStatsRepository {
	return &instrumentedStatsRepository{apiRepo: apiRepo}
}

func (m *instrumentedStatsRepository) GetStatsByPeriod(period entity.Period) (entity.Stats, error) {
	requests, err := m.apiRepo.FindByPeriodWithLimit(period, 0, 0)
	if err != nil {
		return entity.Stats{}, err
	}
	return entity.NewStatsFromRequests(requests, period), nil
}

// TestQueryService_GetStats_CacheIntegration tests end-to-end cache behavior through gRPC service
func TestQueryService_GetStats_CacheIntegration(t *testing.T) {
	baseTime := time.Date(2024, 6, 29, 12, 0, 0, 0, time.UTC)

	t.Run("first_query_hits_repository_and_populates_cache", func(t *testing.T) {
		// Setup mock repository with test data
		testRequest := mustCreateAPIRequest(
			"session1", baseTime,
			"claude-3-sonnet-20240229",
			entity.NewToken(100, 50, 10, 5),
			entity.NewCost(0.50),
			1000,
		)
		mockRepo := &mockAPIRequestRepository{
			requests: []entity.APIRequest{testRequest},
		}

		// Create cache with 1-minute TTL
		cache := service.NewInMemoryStatsCache(1 * time.Minute)
		mockStatsRepo := newMockStatsRepository(mockRepo)
		calculateStatsQuery := usecase.NewCalculateStatsQuery(mockStatsRepo, cache)
		queryService := NewService(nil, calculateStatsQuery)

		// Create request for specific time period
		req := &pb.GetStatsRequest{
			StartTime: timestamppb.New(baseTime),
			EndTime:   timestamppb.New(baseTime.Add(time.Hour)),
		}

		// First query should hit repository
		ctx := context.Background()
		resp1, err := queryService.GetStats(ctx, req)
		if err != nil {
			t.Fatalf("First query failed: %v", err)
		}

		// Verify correct stats returned
		if resp1.Stats.TotalRequests != 1 {
			t.Errorf("Expected 1 request, got %d", resp1.Stats.TotalRequests)
		}
		if resp1.Stats.PremiumRequests != 1 {
			t.Errorf("Expected 1 premium request, got %d", resp1.Stats.PremiumRequests)
		}
		if resp1.Stats.TotalCost.Amount != 0.50 {
			t.Errorf("Expected $0.50 cost, got $%.2f", resp1.Stats.TotalCost.Amount)
		}

		// Second identical query should return cached data
		resp2, err := queryService.GetStats(ctx, req)
		if err != nil {
			t.Fatalf("Second query failed: %v", err)
		}

		// Verify same results (cached)
		if resp2.Stats.TotalRequests != resp1.Stats.TotalRequests {
			t.Errorf("Cached response differs: expected %d requests, got %d",
				resp1.Stats.TotalRequests, resp2.Stats.TotalRequests)
		}
		if resp2.Stats.TotalCost.Amount != resp1.Stats.TotalCost.Amount {
			t.Errorf("Cached response differs: expected $%.2f cost, got $%.2f",
				resp1.Stats.TotalCost.Amount, resp2.Stats.TotalCost.Amount)
		}
	})

	t.Run("subsequent_identical_queries_return_cached_data", func(t *testing.T) {
		// Create instrumented repository to count calls
		callCount := 0
		mockRepo := &mockAPIRequestRepository{
			requests: []entity.APIRequest{
				mustCreateAPIRequest(
					"cached", baseTime,
					"claude-3-haiku-20240307",
					entity.NewToken(200, 100, 20, 10),
					entity.NewCost(0.25),
					800,
				),
			},
		}

		// Create instrumented repository wrapper
		instrumentedRepo := &instrumentedRepository{
			repo:      mockRepo,
			callCount: &callCount,
		}

		cache := service.NewInMemoryStatsCache(1 * time.Minute)
		instrumentedStatsRepo := newInstrumentedStatsRepository(instrumentedRepo)
		calculateStatsQuery := usecase.NewCalculateStatsQuery(instrumentedStatsRepo, cache)
		queryService := NewService(nil, calculateStatsQuery)

		req := &pb.GetStatsRequest{
			StartTime: timestamppb.New(baseTime),
			EndTime:   timestamppb.New(baseTime.Add(time.Hour)),
		}

		ctx := context.Background()

		// Make multiple identical queries
		for i := 0; i < 3; i++ {
			resp, err := queryService.GetStats(ctx, req)
			if err != nil {
				t.Fatalf("Query %d failed: %v", i+1, err)
			}

			// Verify consistent results
			if resp.Stats.BaseRequests != 1 {
				t.Errorf("Query %d: expected 1 base request, got %d", i+1, resp.Stats.BaseRequests)
			}
			if resp.Stats.TotalCost.Amount != 0.25 {
				t.Errorf("Query %d: expected $0.25 cost, got $%.2f", i+1, resp.Stats.TotalCost.Amount)
			}
		}

		// Repository should only be called once (first query)
		if callCount != 1 {
			t.Errorf("Expected repository to be called once, but was called %d times", callCount)
		}
	})

	t.Run("cache_expiration_after_1_minute_returns_fresh_data", func(t *testing.T) {
		// Setup repository with initial data
		initialRequest := mustCreateAPIRequest(
			"initial", baseTime,
			"claude-3-sonnet-20240229",
			entity.NewToken(100, 50, 10, 5),
			entity.NewCost(0.50),
			1000,
		)

		updatedRequest := mustCreateAPIRequest(
			"updated", baseTime.Add(30*time.Minute),
			"claude-3-opus-20240229",
			entity.NewToken(200, 100, 20, 10),
			entity.NewCost(1.50),
			2000,
		)

		mockRepo := &mockAPIRequestRepository{
			requests: []entity.APIRequest{initialRequest},
		}

		// Use very short TTL for testing
		cache := service.NewInMemoryStatsCache(50 * time.Millisecond)
		mockStatsRepo := newMockStatsRepository(mockRepo)
		calculateStatsQuery := usecase.NewCalculateStatsQuery(mockStatsRepo, cache)
		queryService := NewService(nil, calculateStatsQuery)

		req := &pb.GetStatsRequest{
			StartTime: timestamppb.New(baseTime),
			EndTime:   timestamppb.New(baseTime.Add(time.Hour)),
		}

		ctx := context.Background()

		// First query - populate cache
		resp1, err := queryService.GetStats(ctx, req)
		if err != nil {
			t.Fatalf("First query failed: %v", err)
		}
		if resp1.Stats.TotalRequests != 1 {
			t.Errorf("First query: expected 1 request, got %d", resp1.Stats.TotalRequests)
		}
		if resp1.Stats.TotalCost.Amount != 0.50 {
			t.Errorf("First query: expected $0.50 cost, got $%.2f", resp1.Stats.TotalCost.Amount)
		}

		// Wait for cache expiration
		time.Sleep(60 * time.Millisecond)

		// Update repository data
		mockRepo.requests = []entity.APIRequest{initialRequest, updatedRequest}

		// Query after expiration should return fresh data
		resp2, err := queryService.GetStats(ctx, req)
		if err != nil {
			t.Fatalf("Post-expiration query failed: %v", err)
		}

		// Verify updated stats
		if resp2.Stats.TotalRequests != 2 {
			t.Errorf("Post-expiration: expected 2 requests, got %d", resp2.Stats.TotalRequests)
		}
		if resp2.Stats.TotalCost.Amount != 2.00 { // 0.50 + 1.50
			t.Errorf("Post-expiration: expected $2.00 cost, got $%.2f", resp2.Stats.TotalCost.Amount)
		}
		if resp2.Stats.PremiumRequests != 2 {
			t.Errorf("Post-expiration: expected 2 premium requests, got %d", resp2.Stats.PremiumRequests)
		}
	})

	t.Run("cache_can_be_disabled_via_configuration", func(t *testing.T) {
		callCount := 0
		mockRepo := &mockAPIRequestRepository{
			requests: []entity.APIRequest{
				mustCreateAPIRequest(
					"nocache", baseTime,
					"claude-3-haiku-20240307",
					entity.NewToken(150, 75, 15, 8),
					entity.NewCost(0.15),
					900,
				),
			},
		}

		// Create instrumented repository wrapper
		instrumentedRepo := &instrumentedRepository{
			repo:      mockRepo,
			callCount: &callCount,
		}

		// Use NoOpStatsCache to simulate disabled cache
		noOpCache := &service.NoOpStatsCache{}
		instrumentedStatsRepo := newInstrumentedStatsRepository(instrumentedRepo)
		calculateStatsQuery := usecase.NewCalculateStatsQuery(instrumentedStatsRepo, noOpCache)
		queryService := NewService(nil, calculateStatsQuery)

		req := &pb.GetStatsRequest{
			StartTime: timestamppb.New(baseTime),
			EndTime:   timestamppb.New(baseTime.Add(time.Hour)),
		}

		ctx := context.Background()

		// Make multiple identical queries
		for i := 0; i < 3; i++ {
			resp, err := queryService.GetStats(ctx, req)
			if err != nil {
				t.Fatalf("Query %d failed: %v", i+1, err)
			}

			// Verify consistent results
			if resp.Stats.BaseRequests != 1 {
				t.Errorf("Query %d: expected 1 base request, got %d", i+1, resp.Stats.BaseRequests)
			}
			if resp.Stats.TotalCost.Amount != 0.15 {
				t.Errorf("Query %d: expected $0.15 cost, got $%.2f", i+1, resp.Stats.TotalCost.Amount)
			}
		}

		// Repository should be called for every query (no caching)
		if callCount != 3 {
			t.Errorf("Expected repository to be called 3 times, but was called %d times", callCount)
		}
	})

	t.Run("different_time_periods_have_separate_cache_entries", func(t *testing.T) {
		// Setup repository with data for different periods
		period1Request := mustCreateAPIRequest(
			"period1", baseTime,
			"claude-3-sonnet-20240229",
			entity.NewToken(100, 50, 10, 5),
			entity.NewCost(0.50),
			1000,
		)

		period2Request := mustCreateAPIRequest(
			"period2", baseTime.Add(2*time.Hour),
			"claude-3-opus-20240229",
			entity.NewToken(200, 100, 20, 10),
			entity.NewCost(1.50),
			2000,
		)

		mockRepo := &mockAPIRequestRepository{
			requests: []entity.APIRequest{period1Request, period2Request},
		}

		cache := service.NewInMemoryStatsCache(1 * time.Minute)
		mockStatsRepo := newMockStatsRepository(mockRepo)
		calculateStatsQuery := usecase.NewCalculateStatsQuery(mockStatsRepo, cache)
		queryService := NewService(nil, calculateStatsQuery)

		ctx := context.Background()

		// Query for period 1
		req1 := &pb.GetStatsRequest{
			StartTime: timestamppb.New(baseTime),
			EndTime:   timestamppb.New(baseTime.Add(time.Hour)),
		}
		resp1, err := queryService.GetStats(ctx, req1)
		if err != nil {
			t.Fatalf("Period 1 query failed: %v", err)
		}
		if resp1.Stats.TotalRequests != 1 {
			t.Errorf("Period 1: expected 1 request, got %d", resp1.Stats.TotalRequests)
		}

		// Query for period 2
		req2 := &pb.GetStatsRequest{
			StartTime: timestamppb.New(baseTime.Add(2 * time.Hour)),
			EndTime:   timestamppb.New(baseTime.Add(3 * time.Hour)),
		}
		resp2, err := queryService.GetStats(ctx, req2)
		if err != nil {
			t.Fatalf("Period 2 query failed: %v", err)
		}
		if resp2.Stats.TotalRequests != 1 {
			t.Errorf("Period 2: expected 1 request, got %d", resp2.Stats.TotalRequests)
		}

		// Verify different costs confirm different cache entries
		if resp1.Stats.TotalCost.Amount == resp2.Stats.TotalCost.Amount {
			t.Error("Expected different costs for different periods, but got same values")
		}
		if resp1.Stats.TotalCost.Amount != 0.50 {
			t.Errorf("Period 1: expected $0.50 cost, got $%.2f", resp1.Stats.TotalCost.Amount)
		}
		if resp2.Stats.TotalCost.Amount != 1.50 {
			t.Errorf("Period 2: expected $1.50 cost, got $%.2f", resp2.Stats.TotalCost.Amount)
		}
	})
}
