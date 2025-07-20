package service

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/elct9620/ccmon/entity"
)

// InMemoryStatsCache implements TTL-based in-memory caching for statistics.
// It provides thread-safe access and lazy cleanup of expired entries.
type InMemoryStatsCache struct {
	cache          map[string]*CachedStats
	mutex          sync.RWMutex
	ttl            time.Duration
	cleanupRunning int32 // atomic flag for cleanup goroutine
}

// CachedStats represents a cached statistics entry with expiration time.
type CachedStats struct {
	Stats     *entity.Stats
	ExpiresAt time.Time
}

// NewInMemoryStatsCache creates a new in-memory cache instance.
func NewInMemoryStatsCache(ttl time.Duration) *InMemoryStatsCache {
	return &InMemoryStatsCache{
		cache: make(map[string]*CachedStats),
		ttl:   ttl,
	}
}

// Get retrieves cached statistics for the given period.
// Returns nil if entry doesn't exist or has expired.
func (c *InMemoryStatsCache) Get(period *entity.Period) *entity.Stats {
	c.tryCleanupExpired()

	key := c.generateKey(period)

	c.mutex.RLock()
	cached, exists := c.cache[key]
	c.mutex.RUnlock()

	if !exists {
		return nil
	}

	// Check if expired
	if time.Now().After(cached.ExpiresAt) {
		return nil
	}

	return cached.Stats
}

// Set stores statistics in the cache for the given period.
func (c *InMemoryStatsCache) Set(period *entity.Period, stats *entity.Stats) {
	c.tryCleanupExpired()

	key := c.generateKey(period)
	expiresAt := time.Now().Add(c.ttl)

	c.mutex.Lock()
	c.cache[key] = &CachedStats{
		Stats:     stats,
		ExpiresAt: expiresAt,
	}
	c.mutex.Unlock()
}

// generateKey creates a unique cache key from the period timestamps.
func (c *InMemoryStatsCache) generateKey(period *entity.Period) string {
	return fmt.Sprintf("%d_%d", period.StartAt().Unix(), period.EndAt().Unix())
}

// tryCleanupExpired attempts to start a cleanup goroutine if none is running.
func (c *InMemoryStatsCache) tryCleanupExpired() {
	// Try to set cleanupRunning from 0 to 1
	if atomic.CompareAndSwapInt32(&c.cleanupRunning, 0, 1) {
		go func() {
			defer atomic.StoreInt32(&c.cleanupRunning, 0)
			c.cleanupExpired()
		}()
	}
}

// cleanupExpired removes all expired entries from the cache.
func (c *InMemoryStatsCache) cleanupExpired() {
	now := time.Now()
	expiredKeys := make([]string, 0)

	// First pass: identify expired keys
	c.mutex.RLock()
	for key, cached := range c.cache {
		if now.After(cached.ExpiresAt) {
			expiredKeys = append(expiredKeys, key)
		}
	}
	c.mutex.RUnlock()

	// Second pass: remove expired entries
	if len(expiredKeys) > 0 {
		c.mutex.Lock()
		for _, key := range expiredKeys {
			// Double-check expiration in case another goroutine updated it
			if cached, exists := c.cache[key]; exists && now.After(cached.ExpiresAt) {
				delete(c.cache, key)
			}
		}
		c.mutex.Unlock()
	}
}
