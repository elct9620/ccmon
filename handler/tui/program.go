package tui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/elct9620/ccmon/entity"
	"github.com/elct9620/ccmon/usecase"
)

// RunMonitor runs the TUI monitor mode with usecases
func RunMonitor(getFilteredQuery *usecase.GetFilteredApiRequestsQuery, calculateStatsQuery *usecase.CalculateStatsQuery, timezone *time.Location, block *entity.Block, tokenLimit int) error {
	// Create the view model (which now implements tea.Model directly)
	model := NewViewModel(getFilteredQuery, calculateStatsQuery, timezone, block, tokenLimit)

	// Create and run the Bubble Tea program
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running TUI: %w", err)
	}

	return nil
}
