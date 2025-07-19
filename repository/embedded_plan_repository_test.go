package repository

import (
	"embed"
	"testing"
)

//go:embed testdata/*
var testDataFS embed.FS

type testDataEmbedFS struct {
	embed.FS
}

func (fs testDataEmbedFS) ReadFile(name string) ([]byte, error) {
	if name == "data/plans.json" {
		return fs.FS.ReadFile("testdata/plans.json")
	}
	return fs.FS.ReadFile(name)
}

var mockDataFS = testDataEmbedFS{testDataFS}

type mockPlanConfig struct {
	plan string
}

func (m *mockPlanConfig) GetClaudePlan() string {
	return m.plan
}

func TestNewEmbeddedPlanRepository(t *testing.T) {
	config := &mockPlanConfig{plan: "pro"}

	repo, err := NewEmbeddedPlanRepository(config, mockDataFS)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if repo == nil {
		t.Fatal("Expected repository to be created")
	}

	if repo.config != config {
		t.Error("Expected config to be set")
	}

	expectedPlans := 4
	if len(repo.plans) != expectedPlans {
		t.Errorf("Expected %d plans, got %d", expectedPlans, len(repo.plans))
	}
}

func TestNewEmbeddedPlanRepositoryInvalidFile(t *testing.T) {
	config := &mockPlanConfig{plan: "pro"}

	var emptyFS embed.FS
	_, err := NewEmbeddedPlanRepository(config, emptyFS)
	if err == nil {
		t.Fatal("Expected error for missing file")
	}
}

func TestGetConfiguredPlan(t *testing.T) {
	tests := []struct {
		name         string
		configPlan   string
		expectedName string
		expectedCost float64
	}{
		{"pro plan configured", "pro", "pro", 20.0},
		{"max plan configured", "max", "max", 100.0},
		{"max20 plan configured", "max20", "max20", 200.0},
		{"unset plan configured", "unset", "unset", 0.0},
		{"empty plan defaults to unset", "", "unset", 0.0},
		{"invalid plan defaults to unset", "invalid", "unset", 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &mockPlanConfig{plan: tt.configPlan}

			repo, err := NewEmbeddedPlanRepository(config, mockDataFS)
			if err != nil {
				t.Fatalf("Failed to create repository: %v", err)
			}

			plan, err := repo.GetConfiguredPlan()
			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}

			if plan.Name() != tt.expectedName {
				t.Errorf("Expected plan name %s, got %s", tt.expectedName, plan.Name())
			}

			if plan.Price().Amount() != tt.expectedCost {
				t.Errorf("Expected plan cost %.1f, got %.1f", tt.expectedCost, plan.Price().Amount())
			}
		})
	}
}

func TestPlanRepositoryInterface(t *testing.T) {
	config := &mockPlanConfig{plan: "pro"}

	repo, err := NewEmbeddedPlanRepository(config, mockDataFS)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	var planRepo PlanRepository = repo

	plan, err := planRepo.GetConfiguredPlan()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if plan.Name() != "pro" {
		t.Errorf("Expected plan name pro, got %s", plan.Name())
	}
}
