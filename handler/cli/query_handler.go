package cli

import (
	"fmt"
)

type QueryHandler struct {
	renderer *FormatRenderer
}

func NewQueryHandler(renderer *FormatRenderer) *QueryHandler {
	return &QueryHandler{
		renderer: renderer,
	}
}

func (h *QueryHandler) HandleFormatQuery(formatString string) error {
	result, err := h.processFormat(formatString)
	h.outputResult(result, err)
	return err
}

func (h *QueryHandler) processFormat(formatString string) (string, error) {
	// Use FormatRenderer to handle variable substitution
	return h.renderer.Render(formatString)
}

func (h *QueryHandler) outputResult(result string, err error) {
	if err != nil {
		// Output consistent error message for all failure scenarios
		// This provides graceful degradation as specified in requirements
		fmt.Print("❌ ERROR")
	} else {
		fmt.Print(result)
	}
}
