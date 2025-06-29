package repository

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/elct9620/ccmon/entity"
	"github.com/elct9620/ccmon/repository/schema"
	"go.etcd.io/bbolt"
)

const (
	requestsBucket = "requests"
	metadataBucket = "metadata"
)

// BoltDBAPIRequestRepository implements APIRequestRepository using BoltDB
type BoltDBAPIRequestRepository struct {
	db *bbolt.DB
}

// NewBoltDBAPIRequestRepository creates a new BoltDB repository instance
func NewBoltDBAPIRequestRepository(db *bbolt.DB) *BoltDBAPIRequestRepository {
	return &BoltDBAPIRequestRepository{
		db: db,
	}
}

// Save stores an API request entity
func (r *BoltDBAPIRequestRepository) Save(req entity.APIRequest) error {
	dbReq := r.convertFromEntity(req)
	return r.saveRequest(&dbReq)
}

// FindByPeriod retrieves API requests filtered by time period
func (r *BoltDBAPIRequestRepository) FindByPeriod(period entity.Period) ([]entity.APIRequest, error) {
	var dbRequests []schema.APIRequest
	var err error

	if period.IsAllTime() {
		// Get all requests
		dbRequests, err = r.getAllRequests()
		if err != nil {
			return nil, err
		}
	} else {
		// Query time range
		dbRequests, err = r.queryTimeRange(period.StartAt(), period.EndAt())
		if err != nil {
			return nil, err
		}
	}

	// Convert to entities
	return r.convertToEntities(dbRequests), nil
}

// FindAll retrieves all API requests (limited to prevent memory issues)
func (r *BoltDBAPIRequestRepository) FindAll() ([]entity.APIRequest, error) {
	dbRequests, err := r.getAllRequests()
	if err != nil {
		return nil, err
	}
	return r.convertToEntities(dbRequests), nil
}

// Close closes the database connection
func (r *BoltDBAPIRequestRepository) Close() error {
	return r.db.Close()
}

// saveRequest saves an API request to the database
func (r *BoltDBAPIRequestRepository) saveRequest(req *schema.APIRequest) error {
	return r.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(requestsBucket))

		// Create timestamp-based key for chronological ordering
		// Format: RFC3339Nano ensures lexicographic ordering matches chronological ordering
		key := fmt.Sprintf("%s_%s", req.Timestamp.Format(time.RFC3339Nano), req.SessionID)

		// Serialize request to JSON
		data, err := json.Marshal(req)
		if err != nil {
			return fmt.Errorf("failed to serialize request: %w", err)
		}

		return bucket.Put([]byte(key), data)
	})
}

// queryTimeRange queries requests within a time range
func (r *BoltDBAPIRequestRepository) queryTimeRange(start, end time.Time) ([]schema.APIRequest, error) {
	var requests []schema.APIRequest

	err := r.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(requestsBucket))
		c := bucket.Cursor()

		// Convert times to keys for range scan
		startKey := []byte(start.Format(time.RFC3339Nano))
		endKey := []byte(end.Format(time.RFC3339Nano) + "\xff") // \xff ensures we get all entries up to end time

		// Seek to start time and iterate until end time
		for k, v := c.Seek(startKey); k != nil && string(k) < string(endKey); k, v = c.Next() {
			var req schema.APIRequest
			if err := json.Unmarshal(v, &req); err != nil {
				// Skip malformed entries
				continue
			}
			requests = append(requests, req)
		}

		return nil
	})

	return requests, err
}

// getAllRequests returns all requests (limited to last 10000 to prevent memory issues)
func (r *BoltDBAPIRequestRepository) getAllRequests() ([]schema.APIRequest, error) {
	var requests []schema.APIRequest

	err := r.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(requestsBucket))
		c := bucket.Cursor()

		// Count total entries first
		count := 0
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			count++
		}

		// If more than 10000, skip to last 10000
		skip := 0
		if count > 10000 {
			skip = count - 10000
		}

		// Iterate through all keys
		i := 0
		for k, v := c.First(); k != nil; k, v = c.Next() {
			if i < skip {
				i++
				continue
			}

			var req schema.APIRequest
			if err := json.Unmarshal(v, &req); err != nil {
				// Skip malformed entries
				continue
			}
			requests = append(requests, req)
			i++
		}

		return nil
	})

	return requests, err
}

// convertToEntity converts a database APIRequest to an entity APIRequest
func (r *BoltDBAPIRequestRepository) convertToEntity(dbReq schema.APIRequest) entity.APIRequest {
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
func (r *BoltDBAPIRequestRepository) convertFromEntity(e entity.APIRequest) schema.APIRequest {
	return schema.APIRequest{
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
func (r *BoltDBAPIRequestRepository) convertToEntities(requests []schema.APIRequest) []entity.APIRequest {
	entities := make([]entity.APIRequest, len(requests))
	for i, req := range requests {
		entities[i] = r.convertToEntity(req)
	}
	return entities
}

// convertFromEntities converts a slice of entity APIRequests to database APIRequests
func (r *BoltDBAPIRequestRepository) convertFromEntities(entities []entity.APIRequest) []schema.APIRequest {
	requests := make([]schema.APIRequest, len(entities))
	for i, e := range entities {
		requests[i] = r.convertFromEntity(e)
	}
	return requests
}
