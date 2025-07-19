package cli

import (
	"context"
	"strings"
	"time"

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
	// Create context with timeout to prevent hanging
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	variableMap, err := r.usageVariablesQuery.Execute(ctx)
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
