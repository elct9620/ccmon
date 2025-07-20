package service

import (
	"testing"
	"time"

	"github.com/elct9620/ccmon/entity"
)

func TestInMemoryStatsCache_LazyCleanup(t *testing.T) {
	// Test lazy cleanup mechanism prevents memory growth
	cache := NewInMemoryStatsCache(50 * time.Millisecond) // Very short TTL for testing

	// Create test period and stats
	period := entity.NewPeriod(time.Now().Add(-1*time.Hour), time.Now())
	stats := &entity.Stats{}

	// Add entries that will expire quickly
	cache.Set(period, stats)

	// Verify entry exists initially
	if result := cache.Get(period); result == nil {
		t.Error("Expected cached stats to be returned")
	}

	// Wait for expiration
	time.Sleep(60 * time.Millisecond)

	// Access cache to trigger lazy cleanup
	if result := cache.Get(period); result != nil {
		t.Error("Expected expired entry to return nil")
	}

	// Give cleanup goroutine time to complete
	time.Sleep(20 * time.Millisecond)

	// Verify cleanup removed expired entries
	cache.mutex.RLock()
	cacheSize := len(cache.cache)
	cache.mutex.RUnlock()

	if cacheSize != 0 {
		t.Errorf("Expected cache to be empty after cleanup, but found %d entries", cacheSize)
	}
}

func TestInMemoryStatsCache_ConcurrentAccess(t *testing.T) {
	// Test that concurrent access doesn't cause race conditions
	cache := NewInMemoryStatsCache(100 * time.Millisecond)

	period := entity.NewPeriod(time.Now().Add(-1*time.Hour), time.Now())
	stats := &entity.Stats{}

	// Simulate concurrent access
	done := make(chan bool, 2)

	// Concurrent Set operations
	go func() {
		for i := 0; i < 10; i++ {
			cache.Set(period, stats)
			time.Sleep(5 * time.Millisecond)
		}
		done <- true
	}()

	// Concurrent Get operations
	go func() {
		for i := 0; i < 10; i++ {
			cache.Get(period)
			time.Sleep(5 * time.Millisecond)
		}
		done <- true
	}()

	// Wait for both goroutines to complete
	<-done
	<-done

	// Test passes if no race condition panic occurred
	t.Log("Concurrent access test completed successfully")
}

func TestInMemoryStatsCache_CleanupGoroutineLimit(t *testing.T) {
	// Test that only one cleanup goroutine runs at a time
	cache := NewInMemoryStatsCache(10 * time.Millisecond)

	period := entity.NewPeriod(time.Now().Add(-1*time.Hour), time.Now())
	stats := &entity.Stats{}

	// Trigger multiple cleanup attempts rapidly
	for i := 0; i < 5; i++ {
		cache.Set(period, stats)
		cache.Get(period)
	}

	// Give time for any goroutines to complete
	time.Sleep(50 * time.Millisecond)

	// Verify atomic flag is reset (no orphaned goroutines)
	if cache.cleanupRunning != 0 {
		t.Error("Expected cleanup flag to be reset, indicating no orphaned goroutines")
	}
}
