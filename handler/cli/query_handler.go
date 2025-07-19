package cli

import (
	"fmt"
)

type QueryHandler struct{}

func NewQueryHandler() *QueryHandler {
	return &QueryHandler{}
}

func (h *QueryHandler) HandleFormatQuery(formatString string) error {
	result, err := h.processFormat(formatString)
	h.outputResult(result, err)
	return err
}

func (h *QueryHandler) processFormat(formatString string) (string, error) {
	// For task 2: Return hardcoded values for testing
	// This will be replaced with real format rendering in later tasks
	switch formatString {
	case "@daily_cost":
		return "$10.0", nil
	case "@monthly_cost":
		return "$150.0", nil
	case "@daily_plan_usage":
		return "50%", nil
	case "@monthly_plan_usage":
		return "50%", nil
	default:
		// For any other format string, return it as-is for now
		// This handles cases like "💰 @daily_cost" which will be processed by FormatRenderer later
		return formatString, nil
	}
}

func (h *QueryHandler) outputResult(result string, err error) {
	if err != nil {
		// Output error message for connection simulation
		fmt.Print("❌ ERROR")
	} else {
		fmt.Print(result)
	}
}
