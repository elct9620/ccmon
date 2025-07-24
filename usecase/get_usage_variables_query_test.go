package usecase_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/elct9620/ccmon/entity"
	"github.com/elct9620/ccmon/testutil"
	"github.com/elct9620/ccmon/usecase"
)

// MockPeriodFactory implements usecase.PeriodFactory for testing
type MockPeriodFactory struct {
	dailyPeriod   entity.Period
	monthlyPeriod entity.Period
}

func (m *MockPeriodFactory) CreateDaily() entity.Period {
	return m.dailyPeriod
}

func (m *MockPeriodFactory) CreateMonthly() entity.Period {
	return m.monthlyPeriod
}

// Helper function to calculate expected daily usage percentage based on current month
func calculateExpectedDailyUsage(dailyCost, planPrice float64) string {
	now := time.Now()
	daysInMonth := now.AddDate(0, 1, -now.Day()).Day()
	dailyBudget := planPrice / float64(daysInMonth)
	percentage := int((dailyCost / dailyBudget) * 100)
	return fmt.Sprintf("%d%%", percentage)
}

// Helper function to create test API requests
func createAPIRequests(baseCount, premiumCount int, baseCost, premiumCost float64) []entity.APIRequest {
	var requests []entity.APIRequest

	now := time.Now()

	// Create base model requests
	for i := 0; i < baseCount; i++ {
		req := entity.NewAPIRequest(
			"test-session",
			now,
			"claude-3-haiku-20240307",
			entity.NewToken(200, 160, 0, 0),
			entity.NewCost(baseCost/float64(baseCount)),
			1000, // 1 second duration
		)
		requests = append(requests, req)
	}

	// Create premium model requests
	for i := 0; i < premiumCount; i++ {
		req := entity.NewAPIRequest(
			"test-session",
			now,
			"claude-3-5-sonnet-20241022",
			entity.NewToken(666, 500, 0, 0),
			entity.NewCost(premiumCost/float64(premiumCount)),
			1000, // 1 second duration
		)
		requests = append(requests, req)
	}

	return requests
}

func TestGetUsageVariablesQuery_Execute(t *testing.T) {
	now := time.Now()
	dailyPeriod := entity.NewPeriod(
		time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC),
		time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, time.UTC),
	)
	monthlyPeriod := entity.NewPeriod(
		time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC),
		time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, time.UTC).Add(-time.Nanosecond),
	)

	tests := []struct {
		name            string
		plan            entity.Plan
		planErr         error
		dailyRequests   []entity.APIRequest
		monthlyRequests []entity.APIRequest
		statsErr        error
		expectedVars    map[string]string
		expectedErr     bool
	}{
		{
			name:            "successful execution with pro plan",
			plan:            entity.NewPlan("pro", entity.NewCost(20.0)),
			dailyRequests:   createAPIRequests(5, 3, 0.5, 0.5),     // $1.0 total daily cost
			monthlyRequests: createAPIRequests(50, 30, 50.0, 90.0), // $140.0 total monthly cost
			expectedVars: map[string]string{
				"@daily_cost":         "$1.0",
				"@monthly_cost":       "$140.0",
				"@daily_plan_usage":   calculateExpectedDailyUsage(1.0, 20.0), // Calculate based on current month
				"@monthly_plan_usage": "700%",                                 // (140/20)*100 = 700%
			},
		},
		{
			name:            "successful execution with unset plan",
			plan:            entity.NewPlan("unset", entity.NewCost(0)),
			dailyRequests:   createAPIRequests(5, 3, 0.5, 0.5),     // $1.0 total daily cost
			monthlyRequests: createAPIRequests(50, 30, 50.0, 90.0), // $140.0 total monthly cost
			expectedVars: map[string]string{
				"@daily_cost":         "$1.0",
				"@monthly_cost":       "$140.0",
				"@daily_plan_usage":   "0%", // unset plan always returns 0%
				"@monthly_plan_usage": "0%", // unset plan always returns 0%
			},
		},
		{
			name:            "plan repository error - fallback to unset",
			planErr:         errors.New("failed to get plan"),
			dailyRequests:   createAPIRequests(5, 3, 0.5, 0.5),     // $1.0 total daily cost
			monthlyRequests: createAPIRequests(50, 30, 50.0, 90.0), // $140.0 total monthly cost
			expectedVars: map[string]string{
				"@daily_cost":         "$1.0",
				"@monthly_cost":       "$140.0",
				"@daily_plan_usage":   "0%", // fallback to unset plan always returns 0%
				"@monthly_plan_usage": "0%", // fallback to unset plan always returns 0%
			},
		},
		{
			name:        "stats query error",
			plan:        entity.NewPlan("pro", entity.NewCost(20.0)),
			statsErr:    errors.New("failed to calculate stats"),
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockPlanRepo := testutil.NewMockPlanRepository(tt.plan)
			if tt.planErr != nil {
				mockPlanRepo.SetError(tt.planErr)
			}

			mockPeriodFactory := &MockPeriodFactory{
				dailyPeriod:   dailyPeriod,
				monthlyPeriod: monthlyPeriod,
			}

			// Create mock repository with appropriate requests
			mockRepo := testutil.NewMockPeriodBasedRepository(tt.dailyRequests, tt.monthlyRequests)
			if tt.statsErr != nil {
				mockRepo.SetError(tt.statsErr)
			}

			// No-op cache for testing (caching disabled)
			noOpCache := testutil.NewNoOpStatsCache()
			// Real CalculateStatsQuery with mock repository
			statsQuery := usecase.NewCalculateStatsQuery(mockRepo, noOpCache)

			// Create query
			query := usecase.NewGetUsageVariablesQuery(
				statsQuery,
				mockPlanRepo,
				mockPeriodFactory,
			)

			// Execute
			vars, err := query.Execute(context.Background())

			// Verify
			if tt.expectedErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Check all expected variables
			for key, expectedValue := range tt.expectedVars {
				if actual, ok := vars[key]; !ok {
					t.Errorf("missing variable %s", key)
				} else if actual != expectedValue {
					t.Errorf("variable %s: got %s, want %s", key, actual, expectedValue)
				}
			}

			// Ensure no extra variables
			if len(vars) != len(tt.expectedVars) {
				t.Errorf("got %d variables, want %d", len(vars), len(tt.expectedVars))
			}
		})
	}
}

// TestDailyPlanUsageFormulaExamples tests specific examples mentioned in requirements
func TestDailyPlanUsageFormulaExamples(t *testing.T) {
	// Test the exact examples from the requirements to ensure formula is correct

	// Calculate expected percentages based on different month lengths
	calculateExpectedPercentage := func(dailyCost, planPrice float64, daysInMonth int) string {
		dailyBudget := planPrice / float64(daysInMonth)
		percentage := int((dailyCost / dailyBudget) * 100)
		return fmt.Sprintf("%d%%", percentage)
	}

	tests := []struct {
		name                   string
		dailyCost              float64
		planPrice              float64
		monthDays              int
		expectedPercentageNote string
	}{
		{
			name:                   "Pro plan $1.0 daily cost with 31 days example",
			dailyCost:              1.0,
			planPrice:              20.0,
			monthDays:              31,
			expectedPercentageNote: "155%", // $1.0 / ($20 / 31) = $1.0 / $0.645 = 155%
		},
		{
			name:                   "Pro plan $2.0 daily cost with 28 days example",
			dailyCost:              2.0,
			planPrice:              20.0,
			monthDays:              28,
			expectedPercentageNote: "280%", // $2.0 / ($20 / 28) = $2.0 / $0.714 = 280%
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expected := calculateExpectedPercentage(tt.dailyCost, tt.planPrice, tt.monthDays)

			// Verify the calculation matches the expected percentage from requirements
			if expected != tt.expectedPercentageNote {
				t.Logf("Formula verification for %d-day month:", tt.monthDays)
				t.Logf("  Daily Cost: $%.1f", tt.dailyCost)
				t.Logf("  Plan Price: $%.1f", tt.planPrice)
				t.Logf("  Daily Budget: $%.3f", tt.planPrice/float64(tt.monthDays))
				t.Logf("  Calculated: %s", expected)
				t.Logf("  Expected: %s", tt.expectedPercentageNote)
				t.Errorf("formula calculation mismatch: got %s, want %s", expected, tt.expectedPercentageNote)
			}
		})
	}
}

// TestGetUsageVariablesQuery_CurrentMonthCalculation tests with current month day count
func TestGetUsageVariablesQuery_CurrentMonthCalculation(t *testing.T) {
	// Test using current month's actual day count to verify the implementation
	now := time.Now()
	daysInCurrentMonth := now.AddDate(0, 1, -now.Day()).Day()

	tests := []struct {
		name          string
		plan          entity.Plan
		dailyCost     float64
		expectedDaily string
	}{
		{
			name:          "Pro plan with $1.0 daily cost (current month)",
			plan:          entity.NewPlan("pro", entity.NewCost(20.0)),
			dailyCost:     1.0,
			expectedDaily: fmt.Sprintf("%d%%", int((1.0/(20.0/float64(daysInCurrentMonth)))*100)),
		},
		{
			name:          "Max plan with $5.0 daily cost (current month)",
			plan:          entity.NewPlan("max", entity.NewCost(100.0)),
			dailyCost:     5.0,
			expectedDaily: fmt.Sprintf("%d%%", int((5.0/(100.0/float64(daysInCurrentMonth)))*100)),
		},
		{
			name:          "Zero daily cost should return 0%",
			plan:          entity.NewPlan("pro", entity.NewCost(20.0)),
			dailyCost:     0.0,
			expectedDaily: "0%",
		},
		{
			name:          "Unset plan should return 0%",
			plan:          entity.NewPlan("unset", entity.NewCost(0)),
			dailyCost:     5.0,
			expectedDaily: "0%",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			now := time.Now()
			dailyPeriod := entity.NewPeriod(
				time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC),
				time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, time.UTC),
			)
			monthlyPeriod := entity.NewPeriod(
				time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC),
				time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, time.UTC).Add(-time.Nanosecond),
			)

			// Create API requests with the specified daily cost
			dailyRequests := createAPIRequests(5, 3, tt.dailyCost/2, tt.dailyCost/2) // Split cost evenly
			monthlyRequests := createAPIRequests(50, 30, 50.0, 90.0)                 // Monthly cost for other tests

			// Setup mocks
			mockPlanRepo := testutil.NewMockPlanRepository(tt.plan)

			mockPeriodFactory := &MockPeriodFactory{
				dailyPeriod:   dailyPeriod,
				monthlyPeriod: monthlyPeriod,
			}

			mockRepo := testutil.NewMockPeriodBasedRepository(dailyRequests, monthlyRequests)

			noOpCache := testutil.NewNoOpStatsCache()
			statsQuery := usecase.NewCalculateStatsQuery(mockRepo, noOpCache)

			query := usecase.NewGetUsageVariablesQuery(
				statsQuery,
				mockPlanRepo,
				mockPeriodFactory,
			)

			// Execute query
			vars, err := query.Execute(context.Background())
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Verify daily plan usage with new formula
			if actual, ok := vars["@daily_plan_usage"]; !ok {
				t.Errorf("missing @daily_plan_usage variable")
			} else if actual != tt.expectedDaily {
				t.Errorf("daily plan usage: got %s, want %s (for %d-day month)", actual, tt.expectedDaily, daysInCurrentMonth)
			}

			// Verify monthly calculation remains unchanged
			if tt.plan.Name() != "unset" && tt.plan.Price().Amount() > 0 {
				if actual, ok := vars["@monthly_plan_usage"]; ok {
					monthlyCost := 140.0 // From createAPIRequests: 50.0 + 90.0
					expectedMonthly := int((monthlyCost / tt.plan.Price().Amount()) * 100)
					expectedMonthlyStr := fmt.Sprintf("%d%%", expectedMonthly)
					if actual != expectedMonthlyStr {
						t.Errorf("monthly plan usage calculation changed unexpectedly: got %s, want %s", actual, expectedMonthlyStr)
					}
				}
			}
		})
	}
}
