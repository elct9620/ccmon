package usecase

import "github.com/elct9620/ccmon/entity"

// APIRequestRepository defines the repository interface for API request data access
type APIRequestRepository interface {
	// Save stores an API request entity
	Save(req entity.APIRequest) error

	// FindByPeriodWithLimit retrieves API requests filtered by time period with limit and offset
	// Use limit = 0 for no limit (fetch all records)
	// Use offset = 0 when no offset is needed
	FindByPeriodWithLimit(period entity.Period, limit int, offset int) ([]entity.APIRequest, error)

	// FindAll retrieves all API requests (limited to prevent memory issues)
	FindAll() ([]entity.APIRequest, error)
}

// BlockStatsRepository defines the repository interface for block tracking statistics
type BlockStatsRepository interface {
	// GetBlockStats returns statistics for a specific block period
	// This is used for progress bar calculation and should always return
	// stats for the exact block period regardless of UI filters
	GetBlockStats(block entity.Block) (entity.Stats, error)
}
