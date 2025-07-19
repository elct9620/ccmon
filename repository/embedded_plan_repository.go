package repository

import (
	"encoding/json"
	"fmt"

	"github.com/elct9620/ccmon/entity"
)

type FileSystem interface {
	ReadFile(name string) ([]byte, error)
}

type PlanRepository interface {
	GetConfiguredPlan() (entity.Plan, error)
}

type EmbeddedPlanRepository struct {
	config PlanConfig
	dataFS FileSystem
	plans  map[string]PlanData
}

type PlanData struct {
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

type PlansDocument struct {
	Plans map[string]PlanData `json:"plans"`
}

type PlanConfig interface {
	GetClaudePlan() string
}

func NewEmbeddedPlanRepository(config PlanConfig, dataFS FileSystem) (*EmbeddedPlanRepository, error) {
	plansData, err := dataFS.ReadFile("data/plans.json")
	if err != nil {
		return nil, fmt.Errorf("failed to read plans.json: %w", err)
	}

	var doc PlansDocument
	if err := json.Unmarshal(plansData, &doc); err != nil {
		return nil, fmt.Errorf("failed to unmarshal plans data: %w", err)
	}

	return &EmbeddedPlanRepository{
		config: config,
		dataFS: dataFS,
		plans:  doc.Plans,
	}, nil
}

func (r *EmbeddedPlanRepository) GetConfiguredPlan() (entity.Plan, error) {
	planName := r.config.GetClaudePlan()
	if planName == "" {
		planName = "unset"
	}

	planData, exists := r.plans[planName]
	if !exists {
		planData = r.plans["unset"]
	}

	cost := entity.NewCost(planData.Price)
	return entity.NewPlan(planData.Name, cost), nil
}
