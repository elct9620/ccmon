package entity

// Stats represents aggregated statistics for API requests
type Stats struct {
	BaseRequests    int
	PremiumRequests int
	BaseTokens      Token
	PremiumTokens   Token
	BaseCost        Cost
	PremiumCost     Cost
}

// TotalRequests returns the total number of requests
func (s Stats) TotalRequests() int {
	return s.BaseRequests + s.PremiumRequests
}

// TotalTokens returns the total tokens across all requests
func (s Stats) TotalTokens() Token {
	return s.BaseTokens.Add(s.PremiumTokens)
}

// TotalCost returns the total cost across all requests
func (s Stats) TotalCost() Cost {
	return s.BaseCost.Add(s.PremiumCost)
}

// CalculateStats calculates statistics for a set of API requests
func CalculateStats(requests []APIRequest) Stats {
	var stats Stats

	for _, req := range requests {
		if req.Model().IsBase() {
			stats.BaseRequests++
			stats.BaseTokens = stats.BaseTokens.Add(req.Tokens())
			stats.BaseCost = stats.BaseCost.Add(req.Cost())
		} else {
			stats.PremiumRequests++
			stats.PremiumTokens = stats.PremiumTokens.Add(req.Tokens())
			stats.PremiumCost = stats.PremiumCost.Add(req.Cost())
		}
	}

	return stats
}
