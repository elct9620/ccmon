package cli

import (
	"fmt"
	"time"

	"github.com/elct9620/ccmon/usecase"
)

type QueryHandler struct {
	renderer *FormatRenderer
}

func NewQueryHandler(statsQuery *usecase.CalculateStatsQuery, timezone *time.Location) *QueryHandler {
	return &QueryHandler{
		renderer: NewFormatRenderer(statsQuery, timezone),
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
		// Output error message for connection simulation
		fmt.Print("❌ ERROR")
	} else {
		fmt.Print(result)
	}
}
