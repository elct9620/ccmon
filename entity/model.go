package entity

import "strings"

// Model represents the AI model used for the API request
type Model string

// IsBase returns true if this is a base model (Haiku)
func (m Model) IsBase() bool {
	return strings.Contains(strings.ToLower(string(m)), "haiku")
}

// IsPremium returns true if this is a premium model (Sonnet, Opus, etc.)
func (m Model) IsPremium() bool {
	return !m.IsBase()
}

// String returns the string representation of the model
func (m Model) String() string {
	return string(m)
}
