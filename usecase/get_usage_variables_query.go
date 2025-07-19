package usecase

import (
	"context"
	"fmt"
	"time"

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
	// Check if context is already cancelled
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context cancelled before execution: %w", err)
	}

	// Get configured plan for percentage calculations
	plan, err := q.planRepository.GetConfiguredPlan()
	if err != nil {
		// Don't fail the entire query if plan is not configured
		// Use an unset plan as fallback
		plan = entity.NewPlan("unset", entity.NewCost(0))
	}

	// Check if context was cancelled while getting plan
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context cancelled while getting plan: %w", err)
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

	// Check if context was cancelled between stats queries
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context cancelled between stats queries: %w", err)
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

	// Daily plan usage percentage - updated formula: dailyCost / (planPrice / daysInMonth)
	var dailyPercentage int
	if !plan.IsValid() || plan.Price().Amount() == 0 {
		dailyPercentage = 0
	} else {
		// Get days in current month
		now := time.Now()
		daysInMonth := now.AddDate(0, 1, -now.Day()).Day()

		// Calculate daily budget (plan price / days in month)
		dailyBudget := plan.Price().Amount() / float64(daysInMonth)

		// Calculate percentage: (daily cost / daily budget) * 100
		percentage := (dailyCost.Amount() / dailyBudget) * 100
		dailyPercentage = int(percentage)
	}
	variables[entity.DailyPlanUsageVariable.Key()] = fmt.Sprintf("%d%%", dailyPercentage)

	// Monthly plan usage percentage
	monthlyPercentage := plan.CalculateUsagePercentage(monthlyCost)
	variables[entity.MonthlyPlanUsageVariable.Key()] = fmt.Sprintf("%d%%", monthlyPercentage)

	return variables
}
