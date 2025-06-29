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
