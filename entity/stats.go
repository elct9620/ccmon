package entity

// Stats represents aggregated statistics for API requests
type Stats struct {
	baseRequests    int
	premiumRequests int
	baseTokens      Token
	premiumTokens   Token
	baseCost        Cost
	premiumCost     Cost
	period          Period
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

// Period returns the time period for these statistics
func (s Stats) Period() Period {
	return s.period
}

// PremiumTokenBurnRate returns the premium token consumption rate per minute
// Returns 0 for all-time periods or zero duration periods
func (s Stats) PremiumTokenBurnRate() float64 {
	// Skip calculation for all-time periods
	if s.period.IsAllTime() {
		return 0
	}

	// Calculate duration in minutes
	duration := s.period.EndAt().Sub(s.period.StartAt())
	if duration <= 0 {
		return 0
	}

	minutes := duration.Minutes()
	if minutes == 0 {
		return 0
	}

	// Use Limited() tokens as these count against Claude's rate limits
	limitedTokens := float64(s.premiumTokens.Limited())
	return limitedTokens / minutes
}

// NewStats creates a new Stats instance with the given values
func NewStats(baseRequests, premiumRequests int, baseTokens, premiumTokens Token, baseCost, premiumCost Cost, period Period) Stats {
	return Stats{
		baseRequests:    baseRequests,
		premiumRequests: premiumRequests,
		baseTokens:      baseTokens,
		premiumTokens:   premiumTokens,
		baseCost:        baseCost,
		premiumCost:     premiumCost,
		period:          period,
	}
}
