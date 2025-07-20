package usecase

import (
	"context"
	"time"

	"github.com/elct9620/ccmon/entity"
)

// GetUsageQuery handles retrieving usage statistics grouped by periods
type GetUsageQuery struct {
	repository    APIRequestRepository
	periodFactory PeriodFactory
}

// NewGetUsageQuery creates a new GetUsageQuery with the given dependencies
func NewGetUsageQuery(repository APIRequestRepository, periodFactory PeriodFactory) *GetUsageQuery {
	return &GetUsageQuery{
		repository:    repository,
		periodFactory: periodFactory,
	}
}

// ListByDay retrieves usage statistics grouped by daily periods
func (q *GetUsageQuery) ListByDay(ctx context.Context, days int, timezone *time.Location) (entity.Usage, error) {
	var dailyStats []entity.Stats

	for i := 0; i < days; i++ {
		// Create historical daily period (today minus i days)
		period := q.createHistoricalDailyPeriod(i)

		// Get requests for this day using the API request repository
		requests, err := q.repository.FindByPeriodWithLimit(period, 0, 0) // No limit for stats calculation
		if err != nil {
			return entity.Usage{}, err
		}

		// Calculate stats for this day
		stats := q.calculateStatsFromRequests(requests, period)
		dailyStats = append(dailyStats, stats)
	}

	return entity.NewUsage(dailyStats), nil
}

// createHistoricalDailyPeriod creates a daily period for i days ago using PeriodFactory
func (q *GetUsageQuery) createHistoricalDailyPeriod(daysAgo int) entity.Period {
	// Get today's period from the factory
	todayPeriod := q.periodFactory.CreateDaily()

	// Calculate the offset for historical days
	dayOffset := time.Duration(daysAgo) * 24 * time.Hour

	// Create historical period by shifting both start and end times
	startAt := todayPeriod.StartAt().Add(-dayOffset)
	endAt := todayPeriod.EndAt().Add(-dayOffset)

	return entity.NewPeriod(startAt, endAt)
}

// calculateStatsFromRequests calculates statistics from a list of requests
func (q *GetUsageQuery) calculateStatsFromRequests(requests []entity.APIRequest, period entity.Period) entity.Stats {
	return entity.NewStatsFromRequests(requests, period)
}
