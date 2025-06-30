package entity

import "strings"

// Model represents the AI model used for the API request
type Model string

// NewModel creates a new Model, returning "unknown" for invalid inputs
func NewModel(modelName string) Model {
	// Return "unknown" for empty or whitespace-only model names
	trimmed := strings.TrimSpace(modelName)
	if trimmed == "" {
		return Model("unknown")
	}

	return Model(trimmed)
}

// IsBase returns true if this is a base model (Haiku)
func (m Model) IsBase() bool {
	return strings.Contains(strings.ToLower(string(m)), "haiku")
}

// String returns the string representation of the model
func (m Model) String() string {
	return string(m)
}
