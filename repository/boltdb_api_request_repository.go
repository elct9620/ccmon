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
	return r.db.Save(req)
}

// FindByPeriod retrieves API requests filtered by time period
func (r *BoltDBAPIRequestRepository) FindByPeriod(period entity.Period) ([]entity.APIRequest, error) {
	return r.db.FindByPeriod(period)
}

// FindAll retrieves all API requests (limited to prevent memory issues)
func (r *BoltDBAPIRequestRepository) FindAll() ([]entity.APIRequest, error) {
	return r.db.FindAll()
}

// Close closes the database connection
func (r *BoltDBAPIRequestRepository) Close() error {
	return r.db.Close()
}