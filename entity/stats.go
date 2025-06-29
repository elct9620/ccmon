package entity

// Stats represents aggregated statistics for API requests
type Stats struct {
	baseRequests    int
	premiumRequests int
	baseTokens      Token
	premiumTokens   Token
	baseCost        Cost
	premiumCost     Cost
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

// CalculateStats calculates statistics for a set of API requests
func CalculateStats(requests []APIRequest) Stats {
	var stats Stats

	for _, req := range requests {
		if req.Model().IsBase() {
			stats.baseRequests++
			stats.baseTokens = stats.baseTokens.Add(req.Tokens())
			stats.baseCost = stats.baseCost.Add(req.Cost())
		} else {
			stats.premiumRequests++
			stats.premiumTokens = stats.premiumTokens.Add(req.Tokens())
			stats.premiumCost = stats.premiumCost.Add(req.Cost())
		}
	}

	return stats
}
