package usecase

import (
	"context"

	"github.com/elct9620/ccmon/entity"
)

// GetPlanQuery handles the retrieval of plan configuration
type GetPlanQuery struct {
	planRepository PlanRepository
}

// NewGetPlanQuery creates a new GetPlanQuery with the given repository
func NewGetPlanQuery(planRepository PlanRepository) *GetPlanQuery {
	return &GetPlanQuery{
		planRepository: planRepository,
	}
}

// Execute executes the get plan query
func (q *GetPlanQuery) Execute(ctx context.Context) (entity.Plan, error) {
	return q.planRepository.GetConfiguredPlan()
}
