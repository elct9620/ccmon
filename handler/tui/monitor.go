package tui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/elct9620/ccmon/usecase"
)

// RunMonitor runs the TUI monitor mode with usecase
func RunMonitor(getFilteredQuery *usecase.GetFilteredApiRequestsQuery, timezone *time.Location) error {
	// Create the Bubble Tea model
	model := NewModel(getFilteredQuery, timezone)

	// Create and run the Bubble Tea program
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running TUI: %w", err)
	}

	return nil
}
