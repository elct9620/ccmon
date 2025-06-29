package usecase

import "github.com/elct9620/ccmon/entity"

// APIRequestRepository defines the repository interface for API request data access
type APIRequestRepository interface {
	// Save stores an API request entity
	Save(req entity.APIRequest) error

	// FindByPeriod retrieves API requests filtered by time period
	FindByPeriod(period entity.Period) ([]entity.APIRequest, error)

	// FindAll retrieves all API requests (limited to prevent memory issues)
	FindAll() ([]entity.APIRequest, error)
}
