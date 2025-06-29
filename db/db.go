package db

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/elct9620/ccmon/entity"
	"go.etcd.io/bbolt"
)

const (
	requestsBucket = "requests"
	metadataBucket = "metadata"
)

// Database wraps bbolt DB with our specific methods
type Database struct {
	db *bbolt.DB
}

// NewDatabase creates a new database instance
func NewDatabase(dbPath string) (*Database, error) {
	return NewDatabaseWithOptions(dbPath, false)
}

// NewDatabaseReadOnly creates a new read-only database instance
func NewDatabaseReadOnly(dbPath string) (*Database, error) {
	// Check if database file exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("database file does not exist: %s (run server mode first to create it)", dbPath)
	}
	return NewDatabaseWithOptions(dbPath, true)
}

// NewDatabaseWithOptions creates a new database instance with specified options
func NewDatabaseWithOptions(dbPath string, readOnly bool) (*Database, error) {
	// Create directory if it doesn't exist (for write mode)
	if !readOnly {
		dir := filepath.Dir(dbPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create database directory: %w", err)
		}
	}

	options := &bbolt.Options{
		Timeout:  1 * time.Second,
		ReadOnly: readOnly,
	}

	db, err := bbolt.Open(dbPath, 0600, options)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Initialize buckets (only if not read-only)
	if !readOnly {
		err = db.Update(func(tx *bbolt.Tx) error {
			_, err := tx.CreateBucketIfNotExists([]byte(requestsBucket))
			if err != nil {
				return fmt.Errorf("failed to create requests bucket: %w", err)
			}
			_, err = tx.CreateBucketIfNotExists([]byte(metadataBucket))
			if err != nil {
				return fmt.Errorf("failed to create metadata bucket: %w", err)
			}
			return nil
		})
		if err != nil {
			db.Close()
			return nil, err
		}
	}

	return &Database{db: db}, nil
}

// Close closes the database
func (d *Database) Close() error {
	return d.db.Close()
}

// Save saves an API request to the database (implements usecase.APIRequestRepository interface)
func (d *Database) Save(req entity.APIRequest) error {
	dbReq := FromEntity(req)
	return d.SaveRequest(&dbReq)
}

// SaveAPIRequest saves an API request to the database (implements receiver.Database interface)
func (d *Database) SaveAPIRequest(req entity.APIRequest) error {
	return d.Save(req)
}

// SaveRequest saves an API request to the database
func (d *Database) SaveRequest(req *APIRequest) error {
	return d.db.Update(func(tx *bbolt.Tx) error {
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

// QueryTimeRange queries requests within a time range
func (d *Database) QueryTimeRange(start, end time.Time) ([]APIRequest, error) {
	var requests []APIRequest

	err := d.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(requestsBucket))
		c := bucket.Cursor()

		// Convert times to keys for range scan
		startKey := []byte(start.Format(time.RFC3339Nano))
		endKey := []byte(end.Format(time.RFC3339Nano) + "\xff") // \xff ensures we get all entries up to end time

		// Seek to start time and iterate until end time
		for k, v := c.Seek(startKey); k != nil && string(k) < string(endKey); k, v = c.Next() {
			var req APIRequest
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

// FindAll returns all requests (limited to last 10000 to prevent memory issues)
// Implements usecase.APIRequestRepository interface
func (d *Database) FindAll() ([]entity.APIRequest, error) {
	var requests []APIRequest

	err := d.db.View(func(tx *bbolt.Tx) error {
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

			var req APIRequest
			if err := json.Unmarshal(v, &req); err != nil {
				// Skip malformed entries
				continue
			}
			requests = append(requests, req)
			i++
		}

		return nil
	})

	// Convert to entities
	entities := ToEntities(requests)
	return entities, err
}

// CalculateStats calculates statistics for a set of requests
func CalculateStats(requests []APIRequest) (baseReqs, premiumReqs int, baseTokens, premiumTokens, baseLimited, premiumLimited, baseCache, premiumCache int64, baseCost, premiumCost float64) {
	// Convert to entities and calculate stats
	entities := ToEntities(requests)
	stats := entity.CalculateStats(entities)

	// Extract values for backward compatibility
	baseReqs = stats.BaseRequests()
	premiumReqs = stats.PremiumRequests()
	baseTokens = stats.BaseTokens().Total()
	premiumTokens = stats.PremiumTokens().Total()
	baseLimited = stats.BaseTokens().Limited()
	premiumLimited = stats.PremiumTokens().Limited()
	baseCache = stats.BaseTokens().Cache()
	premiumCache = stats.PremiumTokens().Cache()
	baseCost = stats.BaseCost().Amount()
	premiumCost = stats.PremiumCost().Amount()

	return
}

// FindByPeriod returns requests based on period (implements usecase.APIRequestRepository interface)
func (d *Database) FindByPeriod(period entity.Period) ([]entity.APIRequest, error) {
	var dbRequests []APIRequest
	var err error
	
	if period.IsAllTime() {
		// Get all requests and convert to entities
		entities, err := d.FindAll()
		return entities, err
	} else {
		// Query time range and convert to entities
		dbRequests, err = d.QueryTimeRange(period.StartAt(), period.EndAt())
		if err != nil {
			return nil, err
		}
		entities := ToEntities(dbRequests)
		return entities, nil
	}
}

// GetAllRequests returns all requests (backward compatibility method)
func (d *Database) GetAllRequests() ([]entity.APIRequest, error) {
	return d.FindAll()
}

// GetAPIRequests returns requests based on period (backward compatibility method)
func (d *Database) GetAPIRequests(period entity.Period) ([]entity.APIRequest, error) {
	return d.FindByPeriod(period)
}
