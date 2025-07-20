package usecase

import "github.com/elct9620/ccmon/entity"

// StatsCache defines the interface for caching statistics query results.
// Implementations should handle TTL-based expiration and thread-safe access.
type StatsCache interface {
	// Get retrieves cached statistics for the given period.
	// Returns nil if the cache entry doesn't exist or has expired.
	Get(period entity.Period) *entity.Stats

	// Set stores statistics in the cache for the given period.
	// The implementation determines the TTL for cache entries.
	Set(period entity.Period, stats *entity.Stats)
}
