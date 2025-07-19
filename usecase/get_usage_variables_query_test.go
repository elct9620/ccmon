package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/elct9620/ccmon/entity"
	"github.com/elct9620/ccmon/service"
	"github.com/elct9620/ccmon/usecase"
)

// MockPlanRepository implements usecase.PlanRepository for testing
type MockPlanRepository struct {
	plan entity.Plan
	err  error
}

func (m *MockPlanRepository) GetConfiguredPlan() (entity.Plan, error) {
	return m.plan, m.err
}

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

// MockAPIRequestRepository implements usecase.APIRequestRepository for testing stats query
type MockAPIRequestRepository struct {
	dailyRequests   []entity.APIRequest
	monthlyRequests []entity.APIRequest
	err             error
}

func (m *MockAPIRequestRepository) Save(req entity.APIRequest) error {
	return nil
}

func (m *MockAPIRequestRepository) FindByPeriodWithLimit(period entity.Period, limit int, offset int) ([]entity.APIRequest, error) {
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

func (m *MockAPIRequestRepository) FindAll() ([]entity.APIRequest, error) {
	return nil, nil
}

func (m *MockAPIRequestRepository) DeleteOlderThan(cutoffTime time.Time) (int, error) {
	return 0, nil
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
			dailyRequests:   createAPIRequests(5, 3, 5.0, 10.0),
			monthlyRequests: createAPIRequests(50, 30, 50.0, 90.0), // Changed to avoid float precision issues
			expectedVars: map[string]string{
				"@daily_cost":         "$15.0",
				"@monthly_cost":       "$140.0",
				"@daily_plan_usage":   "75%",
				"@monthly_plan_usage": "700%",
			},
		},
		{
			name:            "successful execution with unset plan",
			plan:            entity.NewPlan("unset", entity.NewCost(0)),
			dailyRequests:   createAPIRequests(5, 3, 5.0, 10.0),
			monthlyRequests: createAPIRequests(50, 30, 50.0, 90.0),
			expectedVars: map[string]string{
				"@daily_cost":         "$15.0",
				"@monthly_cost":       "$140.0",
				"@daily_plan_usage":   "0%",
				"@monthly_plan_usage": "0%",
			},
		},
		{
			name:            "plan repository error - fallback to unset",
			planErr:         errors.New("failed to get plan"),
			dailyRequests:   createAPIRequests(5, 3, 5.0, 10.0),
			monthlyRequests: createAPIRequests(50, 30, 50.0, 90.0),
			expectedVars: map[string]string{
				"@daily_cost":         "$15.0",
				"@monthly_cost":       "$140.0",
				"@daily_plan_usage":   "0%",
				"@monthly_plan_usage": "0%",
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
			mockPlanRepo := &MockPlanRepository{
				plan: tt.plan,
				err:  tt.planErr,
			}

			mockPeriodFactory := &MockPeriodFactory{
				dailyPeriod:   dailyPeriod,
				monthlyPeriod: monthlyPeriod,
			}

			// Create mock repository with appropriate requests
			mockRepo := &MockAPIRequestRepository{
				dailyRequests:   tt.dailyRequests,
				monthlyRequests: tt.monthlyRequests,
				err:             tt.statsErr,
			}

			// Real CalculateStatsQuery with mock repository
			statsQuery := usecase.NewCalculateStatsQuery(mockRepo)

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

func TestTimePeriodFactory(t *testing.T) {
	// Test with specific timezone
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatalf("Failed to load timezone: %v", err)
	}

	factory := service.NewTimePeriodFactory(loc)

	t.Run("CreateDaily", func(t *testing.T) {
		period := factory.CreateDaily()

		// Verify the period spans exactly one day
		duration := period.EndAt().Sub(period.StartAt())
		expectedDuration := 24*time.Hour - time.Nanosecond

		if duration != expectedDuration {
			t.Errorf("daily period duration: got %v, want %v", duration, expectedDuration)
		}

		// Verify times are in UTC (for database queries)
		if period.StartAt().Location() != time.UTC {
			t.Errorf("daily period start time not in UTC")
		}
		if period.EndAt().Location() != time.UTC {
			t.Errorf("daily period end time not in UTC")
		}
	})

	t.Run("CreateMonthly", func(t *testing.T) {
		period := factory.CreateMonthly()

		// Verify start is first day of month
		if period.StartAt().Day() != 1 {
			t.Errorf("monthly period should start on day 1, got %d", period.StartAt().Day())
		}

		// Verify end is last moment of month
		nextMonth := period.EndAt().Add(time.Nanosecond)
		if nextMonth.Day() != 1 {
			t.Errorf("monthly period should end at last moment of month")
		}

		// Verify times are in UTC (for database queries)
		if period.StartAt().Location() != time.UTC {
			t.Errorf("monthly period start time not in UTC")
		}
		if period.EndAt().Location() != time.UTC {
			t.Errorf("monthly period end time not in UTC")
		}
	})

	t.Run("CreateWithNilTimezone", func(t *testing.T) {
		// Should default to UTC
		factory := service.NewTimePeriodFactory(nil)
		period := factory.CreateDaily()

		if period.StartAt().Location() != time.UTC {
			t.Errorf("nil timezone should default to UTC")
		}
	})
}
