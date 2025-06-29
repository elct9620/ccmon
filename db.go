package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.etcd.io/bbolt"
)

const (
	RequestsBucket = "requests"
	MetadataBucket = "metadata"
)

// NewDatabase creates a new database instance
func NewDatabase(dbPath string) (*bbolt.DB, error) {
	return NewDatabaseWithOptions(dbPath, false)
}

// NewDatabaseReadOnly creates a new read-only database instance
func NewDatabaseReadOnly(dbPath string) (*bbolt.DB, error) {
	// Check if database file exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("database file does not exist: %s (run server mode first to create it)", dbPath)
	}
	return NewDatabaseWithOptions(dbPath, true)
}

// NewDatabaseWithOptions creates a new database instance with specified options
func NewDatabaseWithOptions(dbPath string, readOnly bool) (*bbolt.DB, error) {
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
		err = InitializeBuckets(db)
		if err != nil {
			db.Close()
			return nil, err
		}
	}

	return db, nil
}

// InitializeBuckets creates the necessary buckets for the database
func InitializeBuckets(db *bbolt.DB) error {
	return db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(RequestsBucket))
		if err != nil {
			return fmt.Errorf("failed to create requests bucket: %w", err)
		}
		_, err = tx.CreateBucketIfNotExists([]byte(MetadataBucket))
		if err != nil {
			return fmt.Errorf("failed to create metadata bucket: %w", err)
		}
		return nil
	})
}
