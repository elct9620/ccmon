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

// TimePeriodFactory implements PeriodFactory using timezone-aware calculations
type TimePeriodFactory struct {
	timezone *time.Location
}

// NewTimePeriodFactory creates a new TimePeriodFactory with the given timezone
func NewTimePeriodFactory(timezone *time.Location) *TimePeriodFactory {
	if timezone == nil {
		timezone = time.UTC
	}
	return &TimePeriodFactory{
		timezone: timezone,
	}
}

// CreateDaily creates a period for today using timezone-aware boundaries
func (f *TimePeriodFactory) CreateDaily() entity.Period {
	now := time.Now().In(f.timezone)
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, f.timezone)
	dayEnd := dayStart.Add(24*time.Hour - time.Nanosecond)

	// Convert to UTC for database queries but maintain timezone-aware boundaries
	return entity.NewPeriod(dayStart.UTC(), dayEnd.UTC())
}

// CreateMonthly creates a period for current month using timezone-aware boundaries
func (f *TimePeriodFactory) CreateMonthly() entity.Period {
	now := time.Now().In(f.timezone)
	// First day of current month at 00:00:00 in user's timezone
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, f.timezone)
	// First day of next month minus 1 nanosecond to get end of current month
	monthEnd := monthStart.AddDate(0, 1, 0).Add(-time.Nanosecond)

	// Convert to UTC for database queries but maintain timezone-aware boundaries
	return entity.NewPeriod(monthStart.UTC(), monthEnd.UTC())
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
