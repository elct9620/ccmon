package usecase

import (
	"context"
	"fmt"

	"github.com/elct9620/ccmon/entity"
)

// PeriodFactory provides methods to create common time periods
type PeriodFactory interface {
	CreateDaily() entity.Period
	CreateMonthly() entity.Period
}

// GetUsageVariablesQuery retrieves usage variables for format string substitution
type GetUsageVariablesQuery struct {
	statsQuery     *CalculateStatsQuery
	planRepository PlanRepository
	periodFactory  PeriodFactory
}

// NewGetUsageVariablesQuery creates a new GetUsageVariablesQuery with the given dependencies
func NewGetUsageVariablesQuery(
	statsQuery *CalculateStatsQuery,
	planRepository PlanRepository,
	periodFactory PeriodFactory,
) *GetUsageVariablesQuery {
	return &GetUsageVariablesQuery{
		statsQuery:     statsQuery,
		planRepository: planRepository,
		periodFactory:  periodFactory,
	}
}

// Execute retrieves usage variables as a substitution map
func (q *GetUsageVariablesQuery) Execute(ctx context.Context) (map[string]string, error) {
	// Get configured plan for percentage calculations
	plan, err := q.planRepository.GetConfiguredPlan()
	if err != nil {
		// Don't fail the entire query if plan is not configured
		// Use an unset plan as fallback
		plan = entity.NewPlan("unset", entity.NewCost(0))
	}

	// Create periods for daily and monthly calculations
	dailyPeriod := q.periodFactory.CreateDaily()
	monthlyPeriod := q.periodFactory.CreateMonthly()

	// Get daily stats
	dailyStats, err := q.statsQuery.Execute(ctx, CalculateStatsParams{
		Period: dailyPeriod,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to calculate daily stats: %w", err)
	}

	// Get monthly stats
	monthlyStats, err := q.statsQuery.Execute(ctx, CalculateStatsParams{
		Period: monthlyPeriod,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to calculate monthly stats: %w", err)
	}

	// Generate the variable map
	return q.generateVariableMap(plan, dailyStats, monthlyStats), nil
}

// generateVariableMap creates the substitution map from stats and plan data
func (q *GetUsageVariablesQuery) generateVariableMap(
	plan entity.Plan,
	dailyStats entity.Stats,
	monthlyStats entity.Stats,
) map[string]string {
	variables := make(map[string]string)

	// Daily cost
	dailyCost := dailyStats.TotalCost()
	variables[entity.DailyCostVariable.Key()] = fmt.Sprintf("$%.1f", dailyCost.Amount())

	// Monthly cost
	monthlyCost := monthlyStats.TotalCost()
	variables[entity.MonthlyCostVariable.Key()] = fmt.Sprintf("$%.1f", monthlyCost.Amount())

	// Daily plan usage percentage
	dailyPercentage := plan.CalculateUsagePercentage(dailyCost)
	variables[entity.DailyPlanUsageVariable.Key()] = fmt.Sprintf("%d%%", dailyPercentage)

	// Monthly plan usage percentage
	monthlyPercentage := plan.CalculateUsagePercentage(monthlyCost)
	variables[entity.MonthlyPlanUsageVariable.Key()] = fmt.Sprintf("%d%%", monthlyPercentage)

	return variables
}
