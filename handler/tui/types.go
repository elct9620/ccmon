package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// ComponentModel defines the interface that all TUI component models must implement
// This follows the Bubble Tea MVU pattern where each component manages its own state
type ComponentModel interface {
	// Update handles messages and updates the model state
	Update(msg tea.Msg) (ComponentModel, tea.Cmd)
	// View renders the component to a string
	View() string
	// Init initializes the component (optional, can return nil)
	Init() tea.Cmd
}

// TabModel defines the interface for tab components that can be switched between
type TabModel interface {
	ComponentModel
	// SetSize updates the component size when terminal is resized
	SetSize(width, height int)
}

// Message types for component communication
type RefreshMsg struct{}
type ResizeMsg struct {
	Width  int
	Height int
}
