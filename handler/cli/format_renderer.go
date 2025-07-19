package cli

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/elct9620/ccmon/entity"
	"github.com/elct9620/ccmon/usecase"
)

type FormatRenderer struct {
	statsQuery *usecase.CalculateStatsQuery
	timezone   *time.Location
}

func NewFormatRenderer(statsQuery *usecase.CalculateStatsQuery, timezone *time.Location) *FormatRenderer {
	return &FormatRenderer{
		statsQuery: statsQuery,
		timezone:   timezone,
	}
}

func (r *FormatRenderer) Render(formatString string) (string, error) {
	return r.substituteVariables(formatString)
}

func (r *FormatRenderer) substituteVariables(input string) (string, error) {
	result := input
	now := time.Now()

	// Calculate daily and monthly periods using same logic as TUI
	dailyPeriod := r.createDailyPeriod(now)
	monthlyPeriod := r.createMonthlyPeriod(now)

	// Get real data from CalculateStatsQuery
	variables := make(map[string]string)

	// Get daily cost
	if strings.Contains(input, "@daily_cost") {
		dailyStats, err := r.statsQuery.Execute(context.Background(), usecase.CalculateStatsParams{
			Period: dailyPeriod,
		})
		if err != nil {
			return "", fmt.Errorf("failed to get daily stats: %w", err)
		}
		totalCost := dailyStats.BaseCost().Amount() + dailyStats.PremiumCost().Amount()
		variables["@daily_cost"] = fmt.Sprintf("$%.1f", totalCost)
	}

	// Get monthly cost
	if strings.Contains(input, "@monthly_cost") {
		monthlyStats, err := r.statsQuery.Execute(context.Background(), usecase.CalculateStatsParams{
			Period: monthlyPeriod,
		})
		if err != nil {
			return "", fmt.Errorf("failed to get monthly stats: %w", err)
		}
		totalCost := monthlyStats.BaseCost().Amount() + monthlyStats.PremiumCost().Amount()
		variables["@monthly_cost"] = fmt.Sprintf("$%.1f", totalCost)
	}

	// Plan usage percentages still hardcoded (will be implemented in later tasks)
	if strings.Contains(input, "@daily_plan_usage") {
		variables["@daily_plan_usage"] = "50%"
	}
	if strings.Contains(input, "@monthly_plan_usage") {
		variables["@monthly_plan_usage"] = "50%"
	}

	// Replace all variables in the format string
	for variable, value := range variables {
		result = strings.ReplaceAll(result, variable, value)
	}

	return result, nil
}

// createDailyPeriod creates a period for today using timezone-aware boundaries
func (r *FormatRenderer) createDailyPeriod(now time.Time) entity.Period {
	// Use same logic as GetUsageQuery: today from 00:00:00 to 23:59:59 in user's timezone
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, r.timezone)
	dayEnd := dayStart.Add(24*time.Hour - time.Nanosecond)

	// Convert to UTC for database queries but maintain timezone-aware boundaries
	return entity.NewPeriod(dayStart.UTC(), dayEnd.UTC())
}

// createMonthlyPeriod creates a period for current month using timezone-aware boundaries
func (r *FormatRenderer) createMonthlyPeriod(now time.Time) entity.Period {
	// First day of current month at 00:00:00 in user's timezone
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, r.timezone)
	// First day of next month minus 1 nanosecond to get end of current month
	monthEnd := monthStart.AddDate(0, 1, 0).Add(-time.Nanosecond)

	// Convert to UTC for database queries but maintain timezone-aware boundaries
	return entity.NewPeriod(monthStart.UTC(), monthEnd.UTC())
}
