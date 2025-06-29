package tui_test

import (
	"context"
	"time"

	"github.com/elct9620/ccmon/entity"
	"github.com/elct9620/ccmon/usecase"
)

// MockAPIRequestRepository implements the APIRequestRepository interface for testing
type MockAPIRequestRepository struct {
	requests []entity.APIRequest
	stats    entity.Stats
}

// NewMockAPIRequestRepository creates a new mock repository with test data
func NewMockAPIRequestRepository() *MockAPIRequestRepository {
	return &MockAPIRequestRepository{
		requests: []entity.APIRequest{},
		stats:    entity.Stats{},
	}
}

// SetMockData sets the mock data to be returned by the repository
func (m *MockAPIRequestRepository) SetMockData(requests []entity.APIRequest, stats entity.Stats) {
	m.requests = requests
	m.stats = stats
}

// FindByPeriodWithLimit implements the repository interface
func (m *MockAPIRequestRepository) FindByPeriodWithLimit(period entity.Period, limit int, offset int) ([]entity.APIRequest, error) {
	// Filter requests by period
	var filtered []entity.APIRequest
	for _, req := range m.requests {
		if period.IsAllTime() {
			filtered = append(filtered, req)
		} else if req.Timestamp().After(period.StartAt()) && req.Timestamp().Before(period.EndAt()) {
			filtered = append(filtered, req)
		}
	}

	// Apply limit and offset
	if offset >= len(filtered) {
		return []entity.APIRequest{}, nil
	}

	end := len(filtered)
	if limit > 0 && offset+limit < len(filtered) {
		end = offset + limit
	}

	return filtered[offset:end], nil
}

// FindAll implements the repository interface
func (m *MockAPIRequestRepository) FindAll() ([]entity.APIRequest, error) {
	return m.requests, nil
}

// Save implements the repository interface
func (m *MockAPIRequestRepository) Save(req entity.APIRequest) error {
	m.requests = append(m.requests, req)
	return nil
}

// Mock query interfaces to avoid direct dependency on usecase types

// GetFilteredQueryInterface represents the interface for getting filtered API requests
type GetFilteredQueryInterface interface {
	Execute(ctx context.Context, params usecase.GetFilteredApiRequestsParams) ([]entity.APIRequest, error)
}

// CalculateStatsQueryInterface represents the interface for calculating stats
type CalculateStatsQueryInterface interface {
	Execute(ctx context.Context, params usecase.CalculateStatsParams) (entity.Stats, error)
}

// MockGetFilteredQuery implements GetFilteredQueryInterface for testing
type MockGetFilteredQuery struct {
	repository *MockAPIRequestRepository
}

// NewMockGetFilteredQuery creates a new mock query
func NewMockGetFilteredQuery(repository *MockAPIRequestRepository) *MockGetFilteredQuery {
	return &MockGetFilteredQuery{
		repository: repository,
	}
}

// Execute implements the query interface
func (m *MockGetFilteredQuery) Execute(ctx context.Context, params usecase.GetFilteredApiRequestsParams) ([]entity.APIRequest, error) {
	return m.repository.FindByPeriodWithLimit(params.Period, params.Limit, params.Offset)
}

// MockCalculateStatsQuery implements CalculateStatsQueryInterface for testing
type MockCalculateStatsQuery struct {
	repository *MockAPIRequestRepository
}

// NewMockCalculateStatsQuery creates a new mock calculate stats query
func NewMockCalculateStatsQuery(repository *MockAPIRequestRepository) *MockCalculateStatsQuery {
	return &MockCalculateStatsQuery{
		repository: repository,
	}
}

// Execute implements the query interface
func (m *MockCalculateStatsQuery) Execute(ctx context.Context, params usecase.CalculateStatsParams) (entity.Stats, error) {
	// Get requests filtered by period
	requests, err := m.repository.FindByPeriodWithLimit(params.Period, 0, 0)
	if err != nil {
		return entity.Stats{}, err
	}

	// Calculate statistics from requests
	var baseRequests, premiumRequests int
	var baseTokens, premiumTokens entity.Token
	var baseCost, premiumCost entity.Cost

	for _, req := range requests {
		if req.Model().IsBase() {
			baseRequests++
			baseTokens = baseTokens.Add(req.Tokens())
			baseCost = baseCost.Add(req.Cost())
		} else {
			premiumRequests++
			premiumTokens = premiumTokens.Add(req.Tokens())
			premiumCost = premiumCost.Add(req.Cost())
		}
	}

	// Create and return stats with block information
	return entity.NewStats(
		baseRequests,
		premiumRequests,
		baseTokens,
		premiumTokens,
		baseCost,
		premiumCost,
		params.BlockTokenLimit,
		params.BlockStartTime,
		params.BlockEndTime,
	), nil
}

// Test data helpers

// CreateTestAPIRequest creates a test API request with default values
func CreateTestAPIRequest(sessionID string, timestamp time.Time, model string, inputTokens, outputTokens int64, cost float64) entity.APIRequest {
	tokens := entity.NewToken(inputTokens, outputTokens, 0, 0) // No cache tokens for simplicity
	costEntity := entity.NewCost(cost)
	return entity.NewAPIRequest(sessionID, timestamp, model, tokens, costEntity, 1500) // 1.5s duration
}

// CreateTestRequestsSet creates a set of test requests with different models and times
func CreateTestRequestsSet() []entity.APIRequest {
	now := time.Now()
	return []entity.APIRequest{
		CreateTestAPIRequest("session1", now.Add(-1*time.Hour), "claude-3-5-haiku-20241022", 1000, 500, 0.001),
		CreateTestAPIRequest("session2", now.Add(-30*time.Minute), "claude-3-5-sonnet-20241022", 2000, 1000, 0.010),
		CreateTestAPIRequest("session3", now.Add(-15*time.Minute), "claude-3-opus-20240229", 1500, 800, 0.015),
		CreateTestAPIRequest("session4", now.Add(-5*time.Minute), "claude-3-5-haiku-20241022", 500, 200, 0.0005),
	}
}

// CreateEmptyStats creates empty stats for testing
func CreateEmptyStats() entity.Stats {
	return entity.NewStats(0, 0, entity.Token{}, entity.Token{}, entity.Cost{}, entity.Cost{}, 0, time.Time{}, time.Time{})
}

// CreateTestStats creates test stats with sample data
func CreateTestStats() entity.Stats {
	baseTokens := entity.NewToken(1500, 700, 0, 0)     // Haiku tokens
	premiumTokens := entity.NewToken(3500, 1800, 0, 0) // Sonnet/Opus tokens
	baseCost := entity.NewCost(0.0015)
	premiumCost := entity.NewCost(0.025)

	return entity.NewStats(2, 2, baseTokens, premiumTokens, baseCost, premiumCost, 0, time.Time{}, time.Time{})
}

// CreateTestBlock creates a test block for testing block tracking
func CreateTestBlock() *entity.Block {
	// Create a block starting at 5am in UTC
	startHour := 5
	block := entity.NewBlock(startHour, time.UTC)
	return &block
}
