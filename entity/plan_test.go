package entity

import (
	"testing"
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
