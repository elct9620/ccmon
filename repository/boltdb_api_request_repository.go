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

// FindByPeriodWithLimit retrieves API requests filtered by time period with limit and offset
// Use limit = 0 for no limit (fetch all records)
// Use offset = 0 when no offset is needed
func (r *BoltDBAPIRequestRepository) FindByPeriodWithLimit(period entity.Period, limit int, offset int) ([]entity.APIRequest, error) {
	var dbRequests []schema.APIRequest
	var err error

	if period.IsAllTime() {
		// Get all requests with limit/offset
		dbRequests, err = r.getAllRequestsWithLimit(limit, offset)
		if err != nil {
			return nil, err
		}
	} else {
		// Query time range with limit/offset
		dbRequests, err = r.queryTimeRangeWithLimit(period.StartAt(), period.EndAt(), limit, offset)
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

// DeleteOlderThan deletes API requests older than the specified cutoff time
// Returns the number of deleted records and any error
func (r *BoltDBAPIRequestRepository) DeleteOlderThan(cutoffTime time.Time) (int, error) {
	deletedCount := 0

	err := r.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(requestsBucket))
		c := bucket.Cursor()

		// Collect keys to delete
		var keysToDelete [][]byte
		for k, v := c.First(); k != nil; k, v = c.Next() {
			// Parse the timestamp from the stored record to compare properly
			var req schema.APIRequest
			if err := json.Unmarshal(v, &req); err != nil {
				// Skip malformed entries
				continue
			}

			// Only delete records that are strictly older than cutoff time
			if req.Timestamp.Before(cutoffTime) {
				// Make a copy of the key since it's only valid for the life of the transaction
				keyToDelete := make([]byte, len(k))
				copy(keyToDelete, k)
				keysToDelete = append(keysToDelete, keyToDelete)
			}
		}

		// Delete collected keys
		for _, key := range keysToDelete {
			if err := bucket.Delete(key); err != nil {
				return fmt.Errorf("failed to delete key %s: %w", string(key), err)
			}
			deletedCount++
		}

		return nil
	})

	return deletedCount, err
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

// queryTimeRangeWithLimit queries requests within a time range with limit and offset
// limit = 0 means no limit, offset = 0 means no offset
func (r *BoltDBAPIRequestRepository) queryTimeRangeWithLimit(start, end time.Time, limit int, offset int) ([]schema.APIRequest, error) {
	var requests []schema.APIRequest

	err := r.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(requestsBucket))
		c := bucket.Cursor()

		// Convert times to keys for range scan
		startKey := []byte(start.Format(time.RFC3339Nano))
		endKey := []byte(end.Format(time.RFC3339Nano) + "\xff") // \xff ensures we get all entries up to end time

		// If no limit specified, collect all records in range
		if limit == 0 {
			for k, v := c.Seek(startKey); k != nil && string(k) < string(endKey); k, v = c.Next() {
				var req schema.APIRequest
				if err := json.Unmarshal(v, &req); err != nil {
					// Skip malformed entries
					continue
				}
				requests = append(requests, req)
			}
			return nil
		}

		// With limit: first collect all matching records, then apply limit/offset
		var allRequests []schema.APIRequest
		for k, v := c.Seek(startKey); k != nil && string(k) < string(endKey); k, v = c.Next() {
			var req schema.APIRequest
			if err := json.Unmarshal(v, &req); err != nil {
				// Skip malformed entries
				continue
			}
			allRequests = append(allRequests, req)
		}

		// Apply offset and limit (get latest entries if no offset)
		totalFound := len(allRequests)
		if totalFound == 0 {
			return nil
		}

		startIdx := 0
		if offset > 0 {
			// With offset, calculate from the end
			startIdx = totalFound - offset - limit
			if startIdx < 0 {
				startIdx = 0
			}
		} else {
			// No offset, get latest entries
			if totalFound > limit {
				startIdx = totalFound - limit
			}
		}

		endIdx := startIdx + limit
		if endIdx > totalFound {
			endIdx = totalFound
		}

		if startIdx < totalFound {
			requests = allRequests[startIdx:endIdx]
		}

		return nil
	})

	return requests, err
}

// getAllRequests returns all requests (limited to last 10000 to prevent memory issues)
func (r *BoltDBAPIRequestRepository) getAllRequests() ([]schema.APIRequest, error) {
	return r.getAllRequestsWithLimit(10000, 0)
}

// getAllRequestsWithLimit returns requests with limit and offset
// limit = 0 means no limit, offset = 0 means no offset
func (r *BoltDBAPIRequestRepository) getAllRequestsWithLimit(limit int, offset int) ([]schema.APIRequest, error) {
	var requests []schema.APIRequest

	err := r.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(requestsBucket))
		c := bucket.Cursor()

		// If no limit specified, use default 10000 to prevent memory issues
		if limit == 0 {
			limit = 10000
		}

		// Count total entries first if we need to apply offset
		totalCount := 0
		if offset > 0 {
			for k, _ := c.First(); k != nil; k, _ = c.Next() {
				totalCount++
			}
		}

		// Calculate where to start
		skipCount := 0
		if offset > 0 && totalCount > offset {
			skipCount = totalCount - offset - limit
			if skipCount < 0 {
				skipCount = 0
			}
		} else if offset == 0 {
			// No offset, get latest entries by skipping older ones
			for k, _ := c.First(); k != nil; k, _ = c.Next() {
				totalCount++
			}
			if totalCount > limit {
				skipCount = totalCount - limit
			}
		}

		// Iterate and collect requested records
		i := 0
		collected := 0
		for k, v := c.First(); k != nil && collected < limit; k, v = c.Next() {
			if i < skipCount {
				i++
				continue
			}

			var req schema.APIRequest
			if err := json.Unmarshal(v, &req); err != nil {
				// Skip malformed entries
				i++
				continue
			}
			requests = append(requests, req)
			collected++
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
