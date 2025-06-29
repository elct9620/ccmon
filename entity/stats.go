package entity

import "time"

// Stats represents aggregated statistics for API requests
type Stats struct {
	baseRequests    int
	premiumRequests int
	baseTokens      Token
	premiumTokens   Token
	baseCost        Cost
	premiumCost     Cost
	// Block-related fields
	blockTokenLimit int
	blockStartTime  time.Time
	blockEndTime    time.Time
	isBlockActive   bool
}

// BaseRequests returns the number of base model requests
func (s Stats) BaseRequests() int {
	return s.baseRequests
}

// PremiumRequests returns the number of premium model requests
func (s Stats) PremiumRequests() int {
	return s.premiumRequests
}

// BaseTokens returns the token usage for base models
func (s Stats) BaseTokens() Token {
	return s.baseTokens
}

// PremiumTokens returns the token usage for premium models
func (s Stats) PremiumTokens() Token {
	return s.premiumTokens
}

// BaseCost returns the cost for base model usage
func (s Stats) BaseCost() Cost {
	return s.baseCost
}

// PremiumCost returns the cost for premium model usage
func (s Stats) PremiumCost() Cost {
	return s.premiumCost
}

// TotalRequests returns the total number of requests
func (s Stats) TotalRequests() int {
	return s.baseRequests + s.premiumRequests
}

// TotalTokens returns the total tokens across all requests
func (s Stats) TotalTokens() Token {
	return s.baseTokens.Add(s.premiumTokens)
}

// TotalCost returns the total cost across all requests
func (s Stats) TotalCost() Cost {
	return s.baseCost.Add(s.premiumCost)
}

// BlockProgress returns the current block progress (used tokens, limit, percentage)
func (s Stats) BlockProgress() (used int64, limit int, percentage float64) {
	if !s.isBlockActive || s.blockTokenLimit == 0 {
		return 0, 0, 0
	}

	used = s.premiumTokens.Limited() // Only premium tokens count toward limits
	limit = s.blockTokenLimit
	percentage = float64(used) / float64(limit) * 100

	if percentage > 100 {
		percentage = 100
	}

	return used, limit, percentage
}

// BlockTimeRemaining returns time remaining until next block
func (s Stats) BlockTimeRemaining() time.Duration {
	if !s.isBlockActive {
		return 0
	}

	now := time.Now().UTC()
	if now.After(s.blockEndTime) {
		return 0
	}

	return s.blockEndTime.Sub(now)
}

// IsInActiveBlock returns true if block tracking is active
func (s Stats) IsInActiveBlock() bool {
	return s.isBlockActive
}

// BlockStartTime returns the current block start time
func (s Stats) BlockStartTime() time.Time {
	return s.blockStartTime
}

// BlockEndTime returns the current block end time
func (s Stats) BlockEndTime() time.Time {
	return s.blockEndTime
}

// BlockTokenLimit returns the token limit for the current block
func (s Stats) BlockTokenLimit() int {
	return s.blockTokenLimit
}

// NewStats creates a new Stats instance with the given values
func NewStats(baseRequests, premiumRequests int, baseTokens, premiumTokens Token, baseCost, premiumCost Cost, blockTokenLimit int, blockStartTime, blockEndTime time.Time) Stats {
	return Stats{
		baseRequests:    baseRequests,
		premiumRequests: premiumRequests,
		baseTokens:      baseTokens,
		premiumTokens:   premiumTokens,
		baseCost:        baseCost,
		premiumCost:     premiumCost,
		blockTokenLimit: blockTokenLimit,
		blockStartTime:  blockStartTime,
		blockEndTime:    blockEndTime,
		isBlockActive:   blockTokenLimit > 0 && !blockStartTime.IsZero(),
	}
}
