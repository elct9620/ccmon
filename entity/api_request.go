package entity

import (
	"time"
)

// APIRequest represents a Claude Code API request entity
type APIRequest struct {
	sessionID string
	timestamp time.Time
	model     Model
	tokens    Token
	cost      Cost
	duration  time.Duration
}

// NewAPIRequest creates a new APIRequest entity
func NewAPIRequest(sessionID string, timestamp time.Time, model string, tokens Token, cost Cost, durationMS int64) APIRequest {
	return APIRequest{
		sessionID: sessionID,
		timestamp: timestamp,
		model:     Model(model),
		tokens:    tokens,
		cost:      cost,
		duration:  time.Duration(durationMS) * time.Millisecond,
	}
}

// SessionID returns the session ID
func (a APIRequest) SessionID() string {
	return a.sessionID
}

// Timestamp returns the request timestamp
func (a APIRequest) Timestamp() time.Time {
	return a.timestamp
}

// Model returns the AI model used
func (a APIRequest) Model() Model {
	return a.model
}

// Tokens returns the token usage
func (a APIRequest) Tokens() Token {
	return a.tokens
}

// Cost returns the cost of the request
func (a APIRequest) Cost() Cost {
	return a.cost
}

// DurationMS returns the request duration in milliseconds
func (a APIRequest) DurationMS() int64 {
	return int64(a.duration / time.Millisecond)
}
