package testutil

import (
	"fmt"
	"time"

	"github.com/elct9620/ccmon/entity"
)

// MockError implements error interface for testing
type MockError struct {
	Message string
}

func (e *MockError) Error() string {
	return e.Message
}

// MockAPIRequestRepository implements usecase.APIRequestRepository for testing
type MockAPIRequestRepository struct {
	requests []entity.APIRequest
	err      error
}

// NewMockAPIRequestRepository creates a new mock API request repository
func NewMockAPIRequestRepository() *MockAPIRequestRepository {
	return &MockAPIRequestRepository{
		requests: []entity.APIRequest{},
	}
}

// NewMockAPIRequestRepositoryWithError creates a mock repository that returns errors
func NewMockAPIRequestRepositoryWithError(err error) *MockAPIRequestRepository {
	return &MockAPIRequestRepository{
		requests: []entity.APIRequest{},
		err:      err,
	}
}

// SetMockData sets the mock data to be returned by the repository
func (m *MockAPIRequestRepository) SetMockData(requests []entity.APIRequest) {
	m.requests = requests
}

// SetError sets an error to be returned by the repository methods
func (m *MockAPIRequestRepository) SetError(err error) {
	m.err = err
}

// Save implements usecase.APIRequestRepository
func (m *MockAPIRequestRepository) Save(req entity.APIRequest) error {
	if m.err != nil {
		return m.err
	}
	m.requests = append(m.requests, req)
	return nil
}

// FindByPeriodWithLimit implements usecase.APIRequestRepository
func (m *MockAPIRequestRepository) FindByPeriodWithLimit(period entity.Period, limit int, offset int) ([]entity.APIRequest, error) {
	if m.err != nil {
		return nil, m.err
	}

	var filtered []entity.APIRequest
	for _, req := range m.requests {
		if !period.IsAllTime() {
			if req.Timestamp().Before(period.StartAt()) || req.Timestamp().After(period.EndAt()) {
				continue
			}
		}
		filtered = append(filtered, req)
	}

	if offset > len(filtered) {
		return []entity.APIRequest{}, nil
	}
	filtered = filtered[offset:]

	if limit > 0 && limit < len(filtered) {
		filtered = filtered[:limit]
	}

	return filtered, nil
}

// FindAll implements usecase.APIRequestRepository
func (m *MockAPIRequestRepository) FindAll() ([]entity.APIRequest, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.requests, nil
}

// DeleteOlderThan implements usecase.APIRequestRepository
func (m *MockAPIRequestRepository) DeleteOlderThan(cutoffTime time.Time) (int, error) {
	if m.err != nil {
		return 0, m.err
	}

	var remaining []entity.APIRequest
	deletedCount := 0

	for _, req := range m.requests {
		if req.Timestamp().Before(cutoffTime) {
			deletedCount++
		} else {
			remaining = append(remaining, req)
		}
	}

	m.requests = remaining
	return deletedCount, nil
}

// MockStatsRepository wraps MockAPIRequestRepository to implement StatsRepository
type MockStatsRepository struct {
	apiRepo *MockAPIRequestRepository
}

// NewMockStatsRepository creates a new mock stats repository
func NewMockStatsRepository(apiRepo *MockAPIRequestRepository) *MockStatsRepository {
	return &MockStatsRepository{apiRepo: apiRepo}
}

// GetStatsByPeriod implements usecase.StatsRepository
func (m *MockStatsRepository) GetStatsByPeriod(period entity.Period) (entity.Stats, error) {
	requests, err := m.apiRepo.FindByPeriodWithLimit(period, 0, 0)
	if err != nil {
		return entity.Stats{}, err
	}
	return entity.NewStatsFromRequests(requests, period), nil
}

// InstrumentedRepository wraps a repository to count method calls for performance testing
type InstrumentedRepository struct {
	repo      *MockAPIRequestRepository
	callCount *int
}

// NewInstrumentedRepository creates a new instrumented repository
func NewInstrumentedRepository(repo *MockAPIRequestRepository, callCount *int) *InstrumentedRepository {
	return &InstrumentedRepository{
		repo:      repo,
		callCount: callCount,
	}
}

// Save implements usecase.APIRequestRepository
func (r *InstrumentedRepository) Save(req entity.APIRequest) error {
	return r.repo.Save(req)
}

// FindByPeriodWithLimit implements usecase.APIRequestRepository with call counting
func (r *InstrumentedRepository) FindByPeriodWithLimit(period entity.Period, limit int, offset int) ([]entity.APIRequest, error) {
	*r.callCount++
	return r.repo.FindByPeriodWithLimit(period, limit, offset)
}

// FindAll implements usecase.APIRequestRepository
func (r *InstrumentedRepository) FindAll() ([]entity.APIRequest, error) {
	return r.repo.FindAll()
}

// DeleteOlderThan implements usecase.APIRequestRepository
func (r *InstrumentedRepository) DeleteOlderThan(cutoffTime time.Time) (int, error) {
	return r.repo.DeleteOlderThan(cutoffTime)
}

// InstrumentedStatsRepository wraps InstrumentedRepository to implement StatsRepository
type InstrumentedStatsRepository struct {
	apiRepo *InstrumentedRepository
}

// NewInstrumentedStatsRepository creates a new instrumented stats repository
func NewInstrumentedStatsRepository(apiRepo *InstrumentedRepository) *InstrumentedStatsRepository {
	return &InstrumentedStatsRepository{apiRepo: apiRepo}
}

// GetStatsByPeriod implements usecase.StatsRepository with call counting
func (m *InstrumentedStatsRepository) GetStatsByPeriod(period entity.Period) (entity.Stats, error) {
	requests, err := m.apiRepo.FindByPeriodWithLimit(period, 0, 0)
	if err != nil {
		return entity.Stats{}, err
	}
	return entity.NewStatsFromRequests(requests, period), nil
}

// Factory Methods for Convenience

// NewMockRepositoryPair creates a connected pair of API request and stats repositories
func NewMockRepositoryPair() (*MockAPIRequestRepository, *MockStatsRepository) {
	apiRepo := NewMockAPIRequestRepository()
	statsRepo := NewMockStatsRepository(apiRepo)
	return apiRepo, statsRepo
}

// NewInstrumentedRepositoryPair creates instrumented repositories for performance testing
func NewInstrumentedRepositoryPair() (*MockAPIRequestRepository, *InstrumentedStatsRepository, *int) {
	callCount := 0
	apiRepo := NewMockAPIRequestRepository()
	instrumentedRepo := NewInstrumentedRepository(apiRepo, &callCount)
	statsRepo := NewInstrumentedStatsRepository(instrumentedRepo)
	return apiRepo, statsRepo, &callCount
}

// NewMockRepositoryWithData creates repositories pre-populated with test data
func NewMockRepositoryWithData(requests []entity.APIRequest) (*MockAPIRequestRepository, *MockStatsRepository) {
	apiRepo := NewMockAPIRequestRepository()
	apiRepo.SetMockData(requests)
	statsRepo := NewMockStatsRepository(apiRepo)
	return apiRepo, statsRepo
}

// Test Data Creation Helpers

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

// NewMockRepositoryWithTestData creates repositories with a standard test dataset
func NewMockRepositoryWithTestData() (*MockAPIRequestRepository, *MockStatsRepository) {
	return NewMockRepositoryWithData(CreateTestRequestsSet())
}

// CreateTestStats creates test stats with sample data
func CreateTestStats() entity.Stats {
	baseTokens := entity.NewToken(1500, 700, 0, 0)     // Haiku tokens
	premiumTokens := entity.NewToken(3500, 1800, 0, 0) // Sonnet/Opus tokens
	baseCost := entity.NewCost(0.0015)
	premiumCost := entity.NewCost(0.025)

	return entity.NewStats(2, 2, baseTokens, premiumTokens, baseCost, premiumCost, entity.NewAllTimePeriod(time.Now()))
}

// MockStatsCache implements usecase.StatsCache for testing
type MockStatsCache struct {
	getFunc   func(period entity.Period) *entity.Stats
	setFunc   func(period entity.Period, stats *entity.Stats)
	getCalled int
	setCalled int
}

// NewMockStatsCache creates a new mock stats cache
func NewMockStatsCache() *MockStatsCache {
	return &MockStatsCache{}
}

// NewMockStatsCacheWithData creates a mock cache that returns specific stats for Get calls
func NewMockStatsCacheWithData(getFunc func(period entity.Period) *entity.Stats) *MockStatsCache {
	return &MockStatsCache{getFunc: getFunc}
}

// SetGetFunc sets the function to be called for Get operations
func (m *MockStatsCache) SetGetFunc(f func(period entity.Period) *entity.Stats) {
	m.getFunc = f
}

// SetSetFunc sets the function to be called for Set operations
func (m *MockStatsCache) SetSetFunc(f func(period entity.Period, stats *entity.Stats)) {
	m.setFunc = f
}

// GetCallCount returns the number of times Get was called
func (m *MockStatsCache) GetCallCount() int {
	return m.getCalled
}

// SetCallCount returns the number of times Set was called
func (m *MockStatsCache) SetCallCount() int {
	return m.setCalled
}

// Get implements usecase.StatsCache
func (m *MockStatsCache) Get(period entity.Period) *entity.Stats {
	m.getCalled++
	if m.getFunc != nil {
		return m.getFunc(period)
	}
	return nil
}

// Set implements usecase.StatsCache
func (m *MockStatsCache) Set(period entity.Period, stats *entity.Stats) {
	m.setCalled++
	if m.setFunc != nil {
		m.setFunc(period, stats)
	}
}

// NoOpStatsCache creates a cache that does nothing (for testing when caching is disabled)
func NewNoOpStatsCache() *MockStatsCache {
	return &MockStatsCache{
		getFunc: func(period entity.Period) *entity.Stats { return nil },
		setFunc: func(period entity.Period, stats *entity.Stats) {},
	}
}

// MockPlanRepository implements usecase.PlanRepository for testing
type MockPlanRepository struct {
	plan entity.Plan
	err  error
}

// NewMockPlanRepository creates a new mock plan repository
func NewMockPlanRepository(plan entity.Plan) *MockPlanRepository {
	return &MockPlanRepository{plan: plan}
}

// NewMockPlanRepositoryWithError creates a mock plan repository that returns errors
func NewMockPlanRepositoryWithError(err error) *MockPlanRepository {
	return &MockPlanRepository{err: err}
}

// SetPlan sets the plan to be returned by the repository
func (m *MockPlanRepository) SetPlan(plan entity.Plan) {
	m.plan = plan
}

// SetError sets an error to be returned by the repository
func (m *MockPlanRepository) SetError(err error) {
	m.err = err
}

// GetConfiguredPlan implements usecase.PlanRepository
func (m *MockPlanRepository) GetConfiguredPlan() (entity.Plan, error) {
	return m.plan, m.err
}

// MockRepositoryWithDeleteFunc allows customization of DeleteOlderThan behavior for cleanup testing
type MockRepositoryWithDeleteFunc struct {
	*MockAPIRequestRepository
	deleteOlderThanFunc func(cutoffTime time.Time) (int, error)
	deleteCallCount     int
	lastCutoffTime      time.Time
}

// NewMockRepositoryWithDeleteFunc creates a repository with customizable DeleteOlderThan behavior
func NewMockRepositoryWithDeleteFunc(deleteFunc func(cutoffTime time.Time) (int, error)) *MockRepositoryWithDeleteFunc {
	return &MockRepositoryWithDeleteFunc{
		MockAPIRequestRepository: NewMockAPIRequestRepository(),
		deleteOlderThanFunc:      deleteFunc,
	}
}

// DeleteOlderThan overrides the base implementation with custom behavior
func (m *MockRepositoryWithDeleteFunc) DeleteOlderThan(cutoffTime time.Time) (int, error) {
	m.deleteCallCount++
	m.lastCutoffTime = cutoffTime
	if m.deleteOlderThanFunc != nil {
		return m.deleteOlderThanFunc(cutoffTime)
	}
	return m.MockAPIRequestRepository.DeleteOlderThan(cutoffTime)
}

// GetDeleteCallCount returns the number of DeleteOlderThan calls
func (m *MockRepositoryWithDeleteFunc) GetDeleteCallCount() int {
	return m.deleteCallCount
}

// GetLastCutoffTime returns the last cutoff time passed to DeleteOlderThan
func (m *MockRepositoryWithDeleteFunc) GetLastCutoffTime() time.Time {
	return m.lastCutoffTime
}

// MockRepositoryWithCustomFunc allows custom behavior for FindByPeriodWithLimit
type MockRepositoryWithCustomFunc struct {
	*MockAPIRequestRepository
	findByPeriodWithLimitFunc func(period entity.Period, limit int, offset int) ([]entity.APIRequest, error)
}

// NewMockRepositoryWithCustomFunc creates a repository with customizable FindByPeriodWithLimit behavior
func NewMockRepositoryWithCustomFunc(findFunc func(period entity.Period, limit int, offset int) ([]entity.APIRequest, error)) *MockRepositoryWithCustomFunc {
	return &MockRepositoryWithCustomFunc{
		MockAPIRequestRepository:  NewMockAPIRequestRepository(),
		findByPeriodWithLimitFunc: findFunc,
	}
}

// FindByPeriodWithLimit overrides the base implementation with custom behavior
func (m *MockRepositoryWithCustomFunc) FindByPeriodWithLimit(period entity.Period, limit int, offset int) ([]entity.APIRequest, error) {
	if m.findByPeriodWithLimitFunc != nil {
		return m.findByPeriodWithLimitFunc(period, limit, offset)
	}
	return m.MockAPIRequestRepository.FindByPeriodWithLimit(period, limit, offset)
}

// GetStatsByPeriod implements StatsRepository interface by calculating stats from requests
func (m *MockRepositoryWithCustomFunc) GetStatsByPeriod(period entity.Period) (entity.Stats, error) {
	requests, err := m.FindByPeriodWithLimit(period, 0, 0)
	if err != nil {
		return entity.Stats{}, err
	}
	return entity.NewStatsFromRequests(requests, period), nil
}

// MockPeriodBasedRepository allows different data for different periods (for usage variables testing)
type MockPeriodBasedRepository struct {
	*MockAPIRequestRepository
	dailyRequests   []entity.APIRequest
	monthlyRequests []entity.APIRequest
}

// NewMockPeriodBasedRepository creates a repository with different data for daily vs monthly periods
func NewMockPeriodBasedRepository(dailyRequests, monthlyRequests []entity.APIRequest) *MockPeriodBasedRepository {
	return &MockPeriodBasedRepository{
		MockAPIRequestRepository: NewMockAPIRequestRepository(),
		dailyRequests:            dailyRequests,
		monthlyRequests:          monthlyRequests,
	}
}

// FindByPeriodWithLimit overrides the base implementation to return different data based on period
func (m *MockPeriodBasedRepository) FindByPeriodWithLimit(period entity.Period, limit int, offset int) ([]entity.APIRequest, error) {
	if m.err != nil {
		return nil, m.err
	}

	// Return different requests based on period
	if period.StartAt().Day() == 1 {
		// Monthly period (starts on day 1)
		return m.monthlyRequests, nil
	}
	// Daily period
	return m.dailyRequests, nil
}

// GetStatsByPeriod implements StatsRepository interface by calculating stats from requests
func (m *MockPeriodBasedRepository) GetStatsByPeriod(period entity.Period) (entity.Stats, error) {
	requests, err := m.FindByPeriodWithLimit(period, 0, 0)
	if err != nil {
		return entity.Stats{}, err
	}
	return entity.NewStatsFromRequests(requests, period), nil
}

// Helper function to create API requests for testing - matches the pattern from CLI tests
func CreateTestAPIRequestsSet(dailyBaseRequests, dailyPremiumRequests, monthlyBaseRequests, monthlyPremiumRequests int,
	dailyBaseCost, dailyPremiumCost, monthlyBaseCost, monthlyPremiumCost float64) []entity.APIRequest {
	var requests []entity.APIRequest

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0, 0, time.UTC)
	monthStart := time.Date(now.Year(), now.Month(), 1, 12, 0, 0, 0, time.UTC)

	// Create daily base requests
	for i := 0; i < dailyBaseRequests; i++ {
		req := entity.NewAPIRequest(
			fmt.Sprintf("daily-base-%d", i),
			today,
			"claude-3-haiku-20240307",
			entity.NewToken(200, 160, 0, 0),
			entity.NewCost(dailyBaseCost/float64(dailyBaseRequests)),
			1000,
		)
		requests = append(requests, req)
	}

	// Create daily premium requests
	for i := 0; i < dailyPremiumRequests; i++ {
		req := entity.NewAPIRequest(
			fmt.Sprintf("daily-premium-%d", i),
			today,
			"claude-3-5-sonnet-20241022",
			entity.NewToken(666, 500, 0, 0),
			entity.NewCost(dailyPremiumCost/float64(dailyPremiumRequests)),
			1000,
		)
		requests = append(requests, req)
	}

	// Create monthly base requests (excluding daily ones)
	for i := 0; i < monthlyBaseRequests; i++ {
		req := entity.NewAPIRequest(
			fmt.Sprintf("monthly-base-%d", i),
			monthStart.Add(time.Duration(i)*time.Hour),
			"claude-3-haiku-20240307",
			entity.NewToken(200, 160, 0, 0),
			entity.NewCost(monthlyBaseCost/float64(monthlyBaseRequests)),
			1000,
		)
		requests = append(requests, req)
	}

	// Create monthly premium requests (excluding daily ones)
	for i := 0; i < monthlyPremiumRequests; i++ {
		req := entity.NewAPIRequest(
			fmt.Sprintf("monthly-premium-%d", i),
			monthStart.Add(time.Duration(i)*time.Hour),
			"claude-3-5-sonnet-20241022",
			entity.NewToken(666, 500, 0, 0),
			entity.NewCost(monthlyPremiumCost/float64(monthlyPremiumRequests)),
			1000,
		)
		requests = append(requests, req)
	}

	return requests
}
