package service

import "github.com/elct9620/ccmon/entity"

// NoOpStatsCache is a no-operation implementation that never caches.
// Used when caching is disabled via configuration.
type NoOpStatsCache struct{}

// Get always returns nil, indicating no cached data
func (c *NoOpStatsCache) Get(period *entity.Period) *entity.Stats {
	return nil
}

// Set does nothing, as caching is disabled
func (c *NoOpStatsCache) Set(period *entity.Period, stats *entity.Stats) {
	// No-op: caching is disabled
}
