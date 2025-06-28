package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// runMonitor runs the TUI monitor mode
func runMonitor() error {
	// Initialize database in read-only mode
	db, err := NewDatabaseReadOnly()
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer db.Close()

	// Create the Bubble Tea model
	model := NewModel(db)

	// Create and run the Bubble Tea program
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running TUI: %w", err)
	}

	return nil
}
