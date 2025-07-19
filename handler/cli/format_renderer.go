package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/elct9620/ccmon/usecase"
)

type FormatRenderer struct {
	statsQuery    *usecase.CalculateStatsQuery
	periodFactory usecase.PeriodFactory
}

func NewFormatRenderer(statsQuery *usecase.CalculateStatsQuery, periodFactory usecase.PeriodFactory) *FormatRenderer {
	return &FormatRenderer{
		statsQuery:    statsQuery,
		periodFactory: periodFactory,
	}
}

func (r *FormatRenderer) Render(formatString string) (string, error) {
	return r.substituteVariables(formatString)
}

func (r *FormatRenderer) substituteVariables(input string) (string, error) {
	result := input

	// Calculate daily and monthly periods using PeriodFactory
	dailyPeriod := r.periodFactory.CreateDaily()
	monthlyPeriod := r.periodFactory.CreateMonthly()

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
