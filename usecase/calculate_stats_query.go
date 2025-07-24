package usecase

import (
	"context"

	"github.com/elct9620/ccmon/entity"
)

// CalculateStatsQuery handles the calculation of statistics using StatsRepository
type CalculateStatsQuery struct {
	statsRepository StatsRepository
	cache           StatsCache
}

// NewCalculateStatsQuery creates a new CalculateStatsQuery with the given stats repository and cache
func NewCalculateStatsQuery(statsRepository StatsRepository, cache StatsCache) *CalculateStatsQuery {
	return &CalculateStatsQuery{
		statsRepository: statsRepository,
		cache:           cache,
	}
}

// CalculateStatsParams contains the parameters for calculating statistics
type CalculateStatsParams struct {
	Period entity.Period
}

// Execute executes the calculate statistics query
func (q *CalculateStatsQuery) Execute(ctx context.Context, params CalculateStatsParams) (entity.Stats, error) {
	if cachedStats := q.cache.Get(params.Period); cachedStats != nil {
		return *cachedStats, nil
	}

	stats, err := q.statsRepository.GetStatsByPeriod(params.Period)
	if err != nil {
		return entity.Stats{}, err
	}

	q.cache.Set(params.Period, &stats)

	return stats, nil
}
