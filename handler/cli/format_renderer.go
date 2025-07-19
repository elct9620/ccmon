package cli

import (
	"strings"
)

type FormatRenderer struct{}

func NewFormatRenderer() *FormatRenderer {
	return &FormatRenderer{}
}

func (r *FormatRenderer) Render(formatString string) (string, error) {
	return r.substituteVariables(formatString)
}

func (r *FormatRenderer) substituteVariables(input string) (string, error) {
	result := input

	// Variable substitution map with hardcoded values for initial implementation
	variables := map[string]string{
		"@daily_cost":         "$10.0",
		"@monthly_cost":       "$150.0",
		"@daily_plan_usage":   "50%",
		"@monthly_plan_usage": "50%",
	}

	// Replace all variables in the format string
	for variable, value := range variables {
		result = strings.ReplaceAll(result, variable, value)
	}

	return result, nil
}
