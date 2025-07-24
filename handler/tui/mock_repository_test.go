package tui_test

import (
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/elct9620/ccmon/entity"
	"github.com/elct9620/ccmon/service"
	"github.com/elct9620/ccmon/testutil"
	"github.com/elct9620/ccmon/usecase"
	"github.com/muesli/termenv"
)

// NewMockAPIRequestRepository creates a new mock repository with test data
func NewMockAPIRequestRepository() *testutil.MockAPIRequestRepository {
	return testutil.NewMockAPIRequestRepository()
}

// NewMockStatsRepository creates a new mock stats repository
func NewMockStatsRepository(apiRepo *testutil.MockAPIRequestRepository) *testutil.MockStatsRepository {
	return testutil.NewMockStatsRepository(apiRepo)
}

// Test data helpers - use testutil functions

// CreateTestAPIRequest creates a test API request with default values
func CreateTestAPIRequest(sessionID string, timestamp time.Time, model string, inputTokens, outputTokens int64, cost float64) entity.APIRequest {
	return testutil.CreateTestAPIRequest(sessionID, timestamp, model, inputTokens, outputTokens, cost)
}

// CreateTestRequestsSet creates a set of test requests with different models and times
func CreateTestRequestsSet() []entity.APIRequest {
	return testutil.CreateTestRequestsSet()
}

// CreateEmptyStats creates empty stats for testing
func CreateEmptyStats() entity.Stats {
	return entity.NewStats(0, 0, entity.Token{}, entity.Token{}, entity.Cost{}, entity.Cost{}, entity.NewAllTimePeriod(time.Now()))
}

// CreateTestStats creates test stats with sample data
func CreateTestStats() entity.Stats {
	return testutil.CreateTestStats()
}

// CreateTestBlock creates a test block for testing block tracking
func CreateTestBlock() *entity.Block {
	// Create a block starting at 5am today in UTC with 7000 token limit
	now := time.Now().UTC()
	start := time.Date(now.Year(), now.Month(), now.Day(), 5, 0, 0, 0, time.UTC)
	block := entity.NewBlockWithLimit(start, 7000)
	return &block
}

// CreateTestUsage creates test usage data for testing
func CreateTestUsage() entity.Usage {
	// Create daily stats for last few days
	now := time.Now().UTC()
	stats := []entity.Stats{}

	for i := 0; i < 5; i++ {
		date := now.AddDate(0, 0, -i)
		period := entity.NewPeriod(
			time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC),
			time.Date(date.Year(), date.Month(), date.Day(), 23, 59, 59, 0, time.UTC),
		)

		baseTokens := entity.NewToken(int64(1000*(i+1)), int64(500*(i+1)), 0, 0)
		premiumTokens := entity.NewToken(int64(2000*(i+1)), int64(1000*(i+1)), int64(100*(i+1)), int64(50*(i+1)))
		baseCost := entity.NewCost(0.001 * float64(i+1))
		premiumCost := entity.NewCost(0.010 * float64(i+1))

		stat := entity.NewStats(i+1, i+2, baseTokens, premiumTokens, baseCost, premiumCost, period)
		stats = append(stats, stat)
	}

	return entity.NewUsage(stats)
}

// CreateTestUsageQuery creates a test usage query for testing
func CreateTestUsageQuery() *usecase.GetUsageQuery {
	apiRepo := NewMockAPIRequestRepository()
	periodFactory := service.NewTimePeriodFactory(time.UTC)
	return usecase.NewGetUsageQuery(apiRepo, periodFactory)
}

// setupTestEnvironment configures the environment for testing in CI/GitHub Actions
func setupTestEnvironment() {
	// Set color profile to ASCII for GitHub Actions compatibility
	lipgloss.SetColorProfile(termenv.Ascii)
}
