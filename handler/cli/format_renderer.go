package cli

import (
	"context"
	"strings"

	"github.com/elct9620/ccmon/usecase"
)

type FormatRenderer struct {
	usageVariablesQuery *usecase.GetUsageVariablesQuery
}

func NewFormatRenderer(usageVariablesQuery *usecase.GetUsageVariablesQuery) *FormatRenderer {
	return &FormatRenderer{
		usageVariablesQuery: usageVariablesQuery,
	}
}

func (r *FormatRenderer) Render(formatString string) (string, error) {
	variableMap, err := r.usageVariablesQuery.Execute(context.Background())
	if err != nil {
		return "", err
	}

	return r.substituteVariables(formatString, variableMap), nil
}

func (r *FormatRenderer) substituteVariables(input string, variableMap map[string]string) string {
	result := input

	// Replace all variables in the format string
	for variable, value := range variableMap {
		result = strings.ReplaceAll(result, variable, value)
	}

	return result
}
