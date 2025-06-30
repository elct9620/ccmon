package usecase

import (
	"context"
	"time"

	"github.com/elct9620/ccmon/entity"
)

// GetUsageQuery handles retrieving usage statistics grouped by periods
type GetUsageQuery struct {
	repository APIRequestRepository
}

// NewGetUsageQuery creates a new GetUsageQuery with the given repository
func NewGetUsageQuery(repository APIRequestRepository) *GetUsageQuery {
	return &GetUsageQuery{
		repository: repository,
	}
}

// ListByDay retrieves usage statistics grouped by daily periods
func (q *GetUsageQuery) ListByDay(ctx context.Context, days int, timezone *time.Location) (entity.Usage, error) {
	// Use provided timezone, fallback to UTC if nil
	if timezone == nil {
		timezone = time.UTC
	}

	now := time.Now().In(timezone)
	var dailyStats []entity.Stats

	for i := 0; i < days; i++ {
		// Calculate day period in the user's timezone (from 00:00:00 to 23:59:59)
		dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, timezone).AddDate(0, 0, -i)
		dayEnd := dayStart.Add(24*time.Hour - time.Nanosecond)

		// Convert to UTC for database queries but maintain timezone-aware boundaries
		period := entity.NewPeriod(dayStart.UTC(), dayEnd.UTC())

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

// calculateStatsFromRequests calculates statistics from a list of requests
func (q *GetUsageQuery) calculateStatsFromRequests(requests []entity.APIRequest, period entity.Period) entity.Stats {
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

	return entity.NewStats(
		baseRequests,
		premiumRequests,
		baseTokens,
		premiumTokens,
		baseCost,
		premiumCost,
		period,
	)
}
