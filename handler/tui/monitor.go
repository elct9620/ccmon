package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/elct9620/ccmon/db"
)

// Database interface to avoid circular dependency
type Database interface {
	GetAPIRequests(filter db.Filter) ([]db.APIRequest, error)
	Close() error
}

// RunMonitor runs the TUI monitor mode
func RunMonitor(newDBReadOnly func() (Database, error)) error {
	// Initialize database in read-only mode
	db, err := newDBReadOnly()
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
