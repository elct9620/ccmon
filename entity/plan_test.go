package entity

import (
	"testing"
	"time"
)

func TestNewPlan(t *testing.T) {
	name := "pro"
	price := NewCost(20.0)

	plan := NewPlan(name, price)

	if plan.Name() != name {
		t.Errorf("Expected name %s, got %s", name, plan.Name())
	}

	if plan.Price().Amount() != price.Amount() {
		t.Errorf("Expected price %f, got %f", price.Amount(), plan.Price().Amount())
	}
}

func TestPlanIsValid(t *testing.T) {
	tests := []struct {
		name     string
		planName string
		expected bool
	}{
		{"unset plan is valid", "unset", true},
		{"pro plan is valid", "pro", true},
		{"max plan is valid", "max", true},
		{"max20 plan is valid", "max20", true},
		{"invalid plan is not valid", "invalid", false},
		{"empty plan is not valid", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plan := NewPlan(tt.planName, NewCost(10.0))
			if plan.IsValid() != tt.expected {
				t.Errorf("Expected IsValid() to return %v for plan %s", tt.expected, tt.planName)
			}
		})
	}
}

func TestPlanCalculateUsagePercentage(t *testing.T) {
	tests := []struct {
		name       string
		planName   string
		planPrice  float64
		actualCost float64
		expected   int
	}{
		{"10% usage", "pro", 20.0, 2.0, 10},
		{"50% usage", "pro", 20.0, 10.0, 50},
		{"100% usage", "pro", 20.0, 20.0, 100},
		{"150% usage", "max", 100.0, 150.0, 150},
		{"zero cost", "pro", 20.0, 0.0, 0},
		{"invalid plan returns 0", "invalid", 20.0, 10.0, 0},
		{"unset plan returns 0", "unset", 0.0, 10.0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plan := NewPlan(tt.planName, NewCost(tt.planPrice))
			actualCost := NewCost(tt.actualCost)
			result := plan.CalculateUsagePercentage(actualCost)

			if result != tt.expected {
				t.Errorf("Expected %d%%, got %d%% for plan %s with cost $%.2f/$%.2f",
					tt.expected, result, tt.planName, tt.actualCost, tt.planPrice)
			}
		})
	}
}

func TestPlanCalculateUsagePercentageInPeriod(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		planName    string
		planPrice   float64
		actualCost  float64
		periodYear  int
		periodMonth time.Month
		periodDay   int
		expected    int
		testReason  string
	}{
		{
			name:        "February 2024 (leap year) - 29 days",
			planName:    "pro",
			planPrice:   20.0,
			actualCost:  1.0,
			periodYear:  2024,
			periodMonth: time.February,
			periodDay:   15,
			expected:    145, // 1.0 / (20.0/29) * 100 = 145%
			testReason:  "Leap year February should have 29 days",
		},
		{
			name:        "February 2023 (non-leap year) - 28 days",
			planName:    "pro",
			planPrice:   20.0,
			actualCost:  1.0,
			periodYear:  2023,
			periodMonth: time.February,
			periodDay:   15,
			expected:    140, // 1.0 / (20.0/28) * 100 = 140%
			testReason:  "Non-leap year February should have 28 days",
		},
		{
			name:        "January 2024 - 31 days",
			planName:    "max",
			planPrice:   100.0,
			actualCost:  5.0,
			periodYear:  2024,
			periodMonth: time.January,
			periodDay:   10,
			expected:    155, // 5.0 / (100.0/31) * 100 = 155%
			testReason:  "January should have 31 days",
		},
		{
			name:        "April 2024 - 30 days",
			planName:    "max",
			planPrice:   100.0,
			actualCost:  5.0,
			periodYear:  2024,
			periodMonth: time.April,
			periodDay:   15,
			expected:    150, // 5.0 / (100.0/30) * 100 = 150%
			testReason:  "April should have 30 days",
		},
		{
			name:        "zero cost",
			planName:    "pro",
			planPrice:   20.0,
			actualCost:  0.0,
			periodYear:  2024,
			periodMonth: time.March,
			periodDay:   15,
			expected:    0,
			testReason:  "Zero cost should result in 0%",
		},
		{
			name:        "invalid plan returns 0",
			planName:    "invalid",
			planPrice:   20.0,
			actualCost:  1.0,
			periodYear:  2024,
			periodMonth: time.March,
			periodDay:   15,
			expected:    0,
			testReason:  "Invalid plan should return 0%",
		},
		{
			name:        "unset plan returns 0",
			planName:    "unset",
			planPrice:   0.0,
			actualCost:  1.0,
			periodYear:  2024,
			periodMonth: time.March,
			periodDay:   15,
			expected:    0,
			testReason:  "Unset plan with zero price should return 0%",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			plan := NewPlan(tt.planName, NewCost(tt.planPrice))
			cost := NewCost(tt.actualCost)

			// Create period for the specific day
			periodStart := time.Date(tt.periodYear, tt.periodMonth, tt.periodDay, 0, 0, 0, 0, time.UTC)
			periodEnd := periodStart.Add(24 * time.Hour)
			period := NewPeriod(periodStart, periodEnd)

			result := plan.CalculateUsagePercentageInPeriod(cost, period)

			if result != tt.expected {
				// Calculate expected values for debugging
				daysInMonth := time.Date(tt.periodYear, tt.periodMonth+1, 0, 0, 0, 0, 0, time.UTC).Day()
				dailyBudget := tt.planPrice / float64(daysInMonth)
				expectedFloat := (tt.actualCost / dailyBudget) * 100

				t.Errorf("%s: Expected %d%%, got %d%%", tt.testReason, tt.expected, result)
				t.Logf("Debug info:")
				t.Logf("  Month: %s %d", tt.periodMonth, tt.periodYear)
				t.Logf("  Days in month: %d", daysInMonth)
				t.Logf("  Plan price: $%.1f", tt.planPrice)
				t.Logf("  Daily budget: $%.6f", dailyBudget)
				t.Logf("  Actual cost: $%.1f", tt.actualCost)
				t.Logf("  Expected percentage (float): %.2f%%", expectedFloat)
			}
		})
	}
}
