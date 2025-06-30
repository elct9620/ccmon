package entity

// Usage represents usage statistics grouped by periods
type Usage struct {
	stats []Stats
}

// NewUsage creates a new Usage instance with the given stats
func NewUsage(stats []Stats) Usage {
	return Usage{
		stats: stats,
	}
}

// GetStats returns the statistics grouped by periods
func (u Usage) GetStats() []Stats {
	return u.stats
}
