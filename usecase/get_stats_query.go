package usecase

import (
	"context"

	"github.com/elct9620/ccmon/entity"
)

// GetStatsQuery handles the query to get statistics
type GetStatsQuery struct {
	repository APIRequestRepository
}

// NewGetStatsQuery creates a new GetStatsQuery with the given repository
func NewGetStatsQuery(repository APIRequestRepository) *GetStatsQuery {
	return &GetStatsQuery{
		repository: repository,
	}
}

// GetStatsParams contains the parameters for getting statistics
type GetStatsParams struct {
	Period entity.Period
}

// Execute executes the get statistics query
func (q *GetStatsQuery) Execute(ctx context.Context, params GetStatsParams) (entity.Stats, error) {
	// Get requests filtered by period
	requests, err := q.repository.FindByPeriod(params.Period)
	if err != nil {
		return entity.Stats{}, err
	}

	// Calculate and return statistics
	return entity.CalculateStats(requests), nil
}