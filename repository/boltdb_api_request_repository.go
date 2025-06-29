package repository

import (
	"github.com/elct9620/ccmon/db"
	"github.com/elct9620/ccmon/entity"
)

// BoltDBAPIRequestRepository implements APIRequestRepository using BoltDB
type BoltDBAPIRequestRepository struct {
	db *db.Database
}

// NewBoltDBAPIRequestRepository creates a new BoltDB repository instance
func NewBoltDBAPIRequestRepository(dbPath string) (*BoltDBAPIRequestRepository, error) {
	database, err := db.NewDatabase(dbPath)
	if err != nil {
		return nil, err
	}
	return &BoltDBAPIRequestRepository{
		db: database,
	}, nil
}

// Save stores an API request entity
func (r *BoltDBAPIRequestRepository) Save(req entity.APIRequest) error {
	dbReq := r.convertFromEntity(req)
	return r.db.SaveRequest(&dbReq)
}

// FindByPeriod retrieves API requests filtered by time period
func (r *BoltDBAPIRequestRepository) FindByPeriod(period entity.Period) ([]entity.APIRequest, error) {
	var dbRequests []db.APIRequest
	var err error

	if period.IsAllTime() {
		// Get all requests
		dbRequests, err = r.db.GetAllRequests()
		if err != nil {
			return nil, err
		}
	} else {
		// Query time range
		dbRequests, err = r.db.QueryTimeRange(period.StartAt(), period.EndAt())
		if err != nil {
			return nil, err
		}
	}

	// Convert to entities
	return r.convertToEntities(dbRequests), nil
}

// FindAll retrieves all API requests (limited to prevent memory issues)
func (r *BoltDBAPIRequestRepository) FindAll() ([]entity.APIRequest, error) {
	dbRequests, err := r.db.GetAllRequests()
	if err != nil {
		return nil, err
	}
	return r.convertToEntities(dbRequests), nil
}

// Close closes the database connection
func (r *BoltDBAPIRequestRepository) Close() error {
	return r.db.Close()
}

// convertToEntity converts a database APIRequest to an entity APIRequest
func (r *BoltDBAPIRequestRepository) convertToEntity(dbReq db.APIRequest) entity.APIRequest {
	tokens := entity.NewToken(
		dbReq.InputTokens,
		dbReq.OutputTokens,
		dbReq.CacheReadTokens,
		dbReq.CacheCreationTokens,
	)
	cost := entity.NewCost(dbReq.CostUSD)

	return entity.NewAPIRequest(
		dbReq.SessionID,
		dbReq.Timestamp,
		dbReq.Model,
		tokens,
		cost,
		dbReq.DurationMS,
	)
}

// convertFromEntity converts an entity APIRequest to a database APIRequest
func (r *BoltDBAPIRequestRepository) convertFromEntity(e entity.APIRequest) db.APIRequest {
	return db.APIRequest{
		SessionID:           e.SessionID(),
		Timestamp:           e.Timestamp(),
		Model:               e.Model().String(),
		InputTokens:         e.Tokens().Input(),
		OutputTokens:        e.Tokens().Output(),
		CacheReadTokens:     e.Tokens().CacheRead(),
		CacheCreationTokens: e.Tokens().CacheCreation(),
		TotalTokens:         e.Tokens().Total(),
		CostUSD:             e.Cost().Amount(),
		DurationMS:          e.DurationMS(),
	}
}

// convertToEntities converts a slice of database APIRequests to entity APIRequests
func (r *BoltDBAPIRequestRepository) convertToEntities(requests []db.APIRequest) []entity.APIRequest {
	entities := make([]entity.APIRequest, len(requests))
	for i, req := range requests {
		entities[i] = r.convertToEntity(req)
	}
	return entities
}

// convertFromEntities converts a slice of entity APIRequests to database APIRequests
func (r *BoltDBAPIRequestRepository) convertFromEntities(entities []entity.APIRequest) []db.APIRequest {
	requests := make([]db.APIRequest, len(entities))
	for i, e := range entities {
		requests[i] = r.convertFromEntity(e)
	}
	return requests
}
