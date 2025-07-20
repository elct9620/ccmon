# Stats In-Memory Cache - Task Breakdown

- [x] 1. Add cache configuration to existing config system with default values
  - Add ServerCache and CacheStats structs to config.go
  - Update config.toml.example with cache settings (enabled=true, ttl="1m")
  - Verify config loads correctly with defaults when cache settings are missing
  - _Requirements: R2

- [ ] 2. Create StatsCache interface in usecase layer
  - Create usecase/stats_cache.go with Get/Set interface methods
  - Add NoOpStatsCache implementation for disabled cache scenario
  - Document interface contract with parameter types
  - _Requirements: R1, R2

- [ ] 3. Implement InMemoryStatsCache service with TTL support
  - Create service/in_memory_stats_cache.go with cache map and RWMutex
  - Implement Get method that returns nil for expired/missing entries
  - Implement Set method that stores stats with expiration timestamp
  - Add generateKey method using period start/end timestamps
  - _Requirements: R1

- [ ] 4. Integrate StatsCache into CalculateStatsQuery with cache-first strategy
  - Add cache dependency to CalculateStatsQuery constructor
  - Modify Execute to check cache before repository query
  - Store calculated results in cache after repository fetch
  - Update main.go to wire cache service based on config
  - _Requirements: R1

- [ ] 5. Add lazy cleanup mechanism to prevent memory growth
  - Implement tryCleanupExpired with atomic flag for single goroutine
  - Add cleanupExpired method to remove expired entries
  - Trigger cleanup on Get/Set operations
  - Verify no orphaned goroutines or memory leaks
  - _Requirements: R1

- [ ] 6. Test end-to-end cache behavior through gRPC service
  - Verify first stats query hits repository and populates cache
  - Confirm subsequent identical queries return cached data
  - Test cache expiration after 1 minute returns fresh data
  - Validate cache can be disabled via configuration
  - _Requirements: R1, R2