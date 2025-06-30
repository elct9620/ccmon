package tui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/elct9620/ccmon/entity"
	"github.com/elct9620/ccmon/usecase"
)

// MonitorConfig represents monitor configuration for TUI
type MonitorConfig struct {
	Server          string
	Timezone        string
	RefreshInterval string
	TokenLimit      int
	BlockTime       string
}

// RunMonitor runs the TUI monitor mode with usecases and config
func RunMonitor(getFilteredQuery *usecase.GetFilteredApiRequestsQuery, calculateStatsQuery *usecase.CalculateStatsQuery, monitorConfig MonitorConfig) error {
	// Load timezone for monitor mode
	timezone, err := time.LoadLocation(monitorConfig.Timezone)
	if err != nil {
		return fmt.Errorf("failed to load timezone %s: %w", monitorConfig.Timezone, err)
	}

	// Parse refresh interval
	refreshInterval, err := time.ParseDuration(monitorConfig.RefreshInterval)
	if err != nil {
		return fmt.Errorf("invalid refresh interval format %s: %w", monitorConfig.RefreshInterval, err)
	}

	// Validate refresh interval bounds
	if refreshInterval < time.Second {
		return fmt.Errorf("refresh interval too short (%v), minimum is 1 second", refreshInterval)
	}
	if refreshInterval > 5*time.Minute {
		return fmt.Errorf("refresh interval too long (%v), maximum is 5 minutes", refreshInterval)
	}

	// Parse block configuration if provided
	var block *entity.Block
	if monitorConfig.BlockTime != "" {
		startHour, err := parseBlockTime(monitorConfig.BlockTime)
		if err != nil {
			return fmt.Errorf("invalid block time format %s: %w", monitorConfig.BlockTime, err)
		}

		if monitorConfig.TokenLimit == 0 {
			fmt.Printf("Warning: No token limit configured. Set claude.plan or claude.max_tokens in config.\n")
		}

		// Create current block with token limit based on user's start hour
		blockEntity := calculateCurrentBlock(startHour, timezone, time.Now(), monitorConfig.TokenLimit)
		block = &blockEntity
	}

	// Create the view model (which now implements tea.Model directly)
	model := NewViewModel(getFilteredQuery, calculateStatsQuery, timezone, block, refreshInterval)

	// Create and run the Bubble Tea program
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running TUI: %w", err)
	}

	return nil
}
