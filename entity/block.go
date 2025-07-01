package entity

import (
	"time"
)

// TimeBlockDuration represents the duration of each Claude token limit block
const TimeBlockDuration = 5 * time.Hour

// Block represents a specific 5-hour token limit block for Claude
// This is a value object representing a concrete time period with optional token limit
type Block struct {
	startAt    time.Time // Concrete timestamp when this block starts
	tokenLimit int       // Token limit for this block (0 = no limit)
}

// NewBlock creates a new Block from a concrete start timestamp without token limit
func NewBlock(startAt time.Time) Block {
	return Block{
		startAt:    startAt,
		tokenLimit: 0, // No token limit
	}
}

// NewBlockWithLimit creates a new Block from a concrete start timestamp with token limit
func NewBlockWithLimit(startAt time.Time, tokenLimit int) Block {
	return Block{
		startAt:    startAt,
		tokenLimit: tokenLimit,
	}
}

// StartAt returns the start time of this block
func (b Block) StartAt() time.Time {
	return b.startAt
}

// EndAt returns the end time of this block
func (b Block) EndAt() time.Time {
	return b.startAt.Add(TimeBlockDuration)
}

// TokenLimit returns the token limit for this block (0 = no limit)
func (b Block) TokenLimit() int {
	return b.tokenLimit
}

// HasLimit returns true if this block has a token limit configured
func (b Block) HasLimit() bool {
	return b.tokenLimit > 0
}

// CalculateProgress calculates the progress percentage of premium token usage against the limit
// Returns 0.0 if no limit is configured, otherwise returns percentage (0.0 to 100.0+)
func (b Block) CalculateProgress(premiumTokens Token) float64 {
	if !b.HasLimit() {
		return 0.0
	}

	// Only premium tokens count toward limits (Haiku is free)
	used := premiumTokens.Limited()
	limit := int64(b.tokenLimit)

	if limit == 0 {
		return 0.0
	}

	percentage := float64(used) / float64(limit) * 100
	return percentage
}

// IsLimitExceeded returns true if the premium token usage exceeds the configured limit
func (b Block) IsLimitExceeded(premiumTokens Token) bool {
	if !b.HasLimit() {
		return false
	}

	used := premiumTokens.Limited()
	return used > int64(b.tokenLimit)
}

// Period returns the time period represented by this block
func (b Block) Period() Period {
	return NewPeriod(b.startAt, b.EndAt())
}

// NextBlock returns the appropriate block for the given time.
// If the current time is still within this block, returns self.
// If the current time is beyond this block, returns the next appropriate block.
func (b Block) NextBlock(now time.Time) Block {
	// If current time is still within this block, return self
	if now.Before(b.EndAt()) {
		return b
	}

	// Calculate which block the current time falls into
	delta := now.Sub(b.startAt)
	blockIndex := int(delta / TimeBlockDuration)

	// Create new block at the appropriate position, preserving token limit
	newStart := b.startAt.Add(time.Duration(blockIndex) * TimeBlockDuration)
	return NewBlockWithLimit(newStart, b.tokenLimit)
}
