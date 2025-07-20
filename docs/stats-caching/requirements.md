# Feature: Stats In-Memory Cache

Add simple in-memory caching mechanism for stats queries to improve response time and reduce CPU usage.

## 1. Basic Cache Mechanism

As a ccmon user
I want stats queries to use in-memory caching
So that repeated queries are faster and use less CPU

```gherkin
Scenario: Cache stats query results
  Given ccmon server is running with cache enabled
  When I query stats for a specific time period
  Then the result should be cached for 1 minute
  And subsequent identical queries should return cached data

Scenario: Cache expiration
  Given a stats result is cached
  When 1 minute has passed
  Then the cache should expire
  And next query should recalculate and cache new result
```

## 2. Configuration Options

As a ccmon system administrator
I want to configure stats caching behavior
So that I can control cache settings based on my needs

```gherkin
Scenario: Default cache settings
  Given ccmon server starts with default config
  Then stats cache should be enabled by default
  And cache TTL should be 1 minute by default

Scenario: Configure cache settings
  Given server config has "server.cache.stats.enabled = true"
  And server config has "server.cache.stats.ttl = 1m"
  When ccmon server starts
  Then stats cache should be active with 1 minute TTL

Scenario: Disable cache
  Given server config has "server.cache.stats.enabled = false"
  When ccmon server starts
  Then stats queries should not use cache
```

## Technical Notes

- **Cache Key**: Period Start + End timestamps
- **Cache Interval**: 1 minute (configurable)
- **Default Settings**: 
  - `server.cache.stats.enabled = true`
  - `server.cache.stats.ttl = 1m`
- **Implementation**: In-memory cache with TTL-based expiration