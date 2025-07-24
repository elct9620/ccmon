package repository

import (
	"github.com/elct9620/ccmon/entity"
	"github.com/elct9620/ccmon/usecase"
)

// BoltDBStatsRepository implements usecase.StatsRepository by calculating stats from BoltDB APIRequestRepository
// This is used on the server side where we have direct access to the BoltDB request data
type BoltDBStatsRepository struct {
	apiRequestRepository usecase.APIRequestRepository
}

// NewBoltDBStatsRepository creates a new BoltDBStatsRepository
func NewBoltDBStatsRepository(apiRequestRepository usecase.APIRequestRepository) *BoltDBStatsRepository {
	return &BoltDBStatsRepository{
		apiRequestRepository: apiRequestRepository,
	}
}

// GetStatsByPeriod retrieves statistics by calculating them from API requests
func (r *BoltDBStatsRepository) GetStatsByPeriod(period entity.Period) (entity.Stats, error) {
	// Get all requests for the period (no limit)
	requests, err := r.apiRequestRepository.FindByPeriodWithLimit(period, 0, 0)
	if err != nil {
		return entity.Stats{}, err
	}

	// Calculate stats from requests
	return entity.NewStatsFromRequests(requests, period), nil
}
