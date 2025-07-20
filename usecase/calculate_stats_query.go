package usecase

import (
	"context"

	"github.com/elct9620/ccmon/entity"
)

// CalculateStatsQuery handles the calculation of statistics for API requests
type CalculateStatsQuery struct {
	repository APIRequestRepository
	cache      StatsCache
}

// NewCalculateStatsQuery creates a new CalculateStatsQuery with the given repository and cache
func NewCalculateStatsQuery(repository APIRequestRepository, cache StatsCache) *CalculateStatsQuery {
	return &CalculateStatsQuery{
		repository: repository,
		cache:      cache,
	}
}

// CalculateStatsParams contains the parameters for calculating statistics
type CalculateStatsParams struct {
	Period entity.Period
}

// Execute executes the calculate statistics query
func (q *CalculateStatsQuery) Execute(ctx context.Context, params CalculateStatsParams) (entity.Stats, error) {
	if cachedStats := q.cache.Get(&params.Period); cachedStats != nil {
		return *cachedStats, nil
	}

	requests, err := q.repository.FindByPeriodWithLimit(params.Period, 0, 0)
	if err != nil {
		return entity.Stats{}, err
	}

	var baseRequests, premiumRequests int
	var baseTokens, premiumTokens entity.Token
	var baseCost, premiumCost entity.Cost

	for _, req := range requests {
		if req.Model().IsBase() {
			baseRequests++
			baseTokens = baseTokens.Add(req.Tokens())
			baseCost = baseCost.Add(req.Cost())
		} else {
			premiumRequests++
			premiumTokens = premiumTokens.Add(req.Tokens())
			premiumCost = premiumCost.Add(req.Cost())
		}
	}

	stats := entity.NewStats(
		baseRequests,
		premiumRequests,
		baseTokens,
		premiumTokens,
		baseCost,
		premiumCost,
		params.Period,
	)

	q.cache.Set(&params.Period, &stats)

	return stats, nil
}
