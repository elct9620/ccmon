package entity

import (
	"testing"
)

func TestUsageVariableDefinitions(t *testing.T) {
	tests := []struct {
		name     string
		variable UsageVariable
		wantKey  string
		wantName string
	}{
		{
			name:     "daily cost variable",
			variable: DailyCostVariable,
			wantKey:  "@daily_cost",
			wantName: "Daily Cost",
		},
		{
			name:     "monthly cost variable",
			variable: MonthlyCostVariable,
			wantKey:  "@monthly_cost",
			wantName: "Monthly Cost",
		},
		{
			name:     "daily plan usage variable",
			variable: DailyPlanUsageVariable,
			wantKey:  "@daily_plan_usage",
			wantName: "Daily Plan Usage",
		},
		{
			name:     "monthly plan usage variable",
			variable: MonthlyPlanUsageVariable,
			wantKey:  "@monthly_plan_usage",
			wantName: "Monthly Plan Usage",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.variable.Key(); got != tt.wantKey {
				t.Errorf("Key() = %v, want %v", got, tt.wantKey)
			}
			if got := tt.variable.Name(); got != tt.wantName {
				t.Errorf("Name() = %v, want %v", got, tt.wantName)
			}
		})
	}
}

func TestGetAllUsageVariables(t *testing.T) {
	variables := GetAllUsageVariables()

	if len(variables) != 4 {
		t.Errorf("Expected 4 variables, got %d", len(variables))
	}

	expectedKeys := map[string]bool{
		"@daily_cost":         false,
		"@monthly_cost":       false,
		"@daily_plan_usage":   false,
		"@monthly_plan_usage": false,
	}

	for _, v := range variables {
		key := v.Key()
		if _, exists := expectedKeys[key]; !exists {
			t.Errorf("Unexpected variable key: %s", key)
		}
		expectedKeys[key] = true
	}

	for key, found := range expectedKeys {
		if !found {
			t.Errorf("Missing expected variable: %s", key)
		}
	}
}
