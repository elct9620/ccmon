package repository

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/elct9620/ccmon/entity"
	"github.com/elct9620/ccmon/repository/schema"
	"go.etcd.io/bbolt"
)

func TestBoltDBAPIRequestRepository_DeleteOlderThan(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		setupRecords      []schema.APIRequest
		cutoffTime        time.Time
		expectedDeleted   int
		expectedRemaining []string // session IDs that should remain
	}{
		{
			name: "delete older records",
			setupRecords: []schema.APIRequest{
				createTestRecord("session1", time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)),
				createTestRecord("session2", time.Date(2025, 1, 2, 10, 0, 0, 0, time.UTC)),
				createTestRecord("session3", time.Date(2025, 1, 3, 10, 0, 0, 0, time.UTC)),
				createTestRecord("session4", time.Date(2025, 1, 4, 10, 0, 0, 0, time.UTC)),
			},
			cutoffTime:        time.Date(2025, 1, 2, 12, 0, 0, 0, time.UTC),
			expectedDeleted:   2,
			expectedRemaining: []string{"session3", "session4"},
		},
		{
			name: "delete all records",
			setupRecords: []schema.APIRequest{
				createTestRecord("session1", time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)),
				createTestRecord("session2", time.Date(2025, 1, 2, 10, 0, 0, 0, time.UTC)),
			},
			cutoffTime:        time.Date(2025, 1, 3, 0, 0, 0, 0, time.UTC),
			expectedDeleted:   2,
			expectedRemaining: []string{},
		},
		{
			name: "delete no records (all newer)",
			setupRecords: []schema.APIRequest{
				createTestRecord("session1", time.Date(2025, 1, 3, 10, 0, 0, 0, time.UTC)),
				createTestRecord("session2", time.Date(2025, 1, 4, 10, 0, 0, 0, time.UTC)),
			},
			cutoffTime:        time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC),
			expectedDeleted:   0,
			expectedRemaining: []string{"session1", "session2"},
		},
		{
			name:              "empty database",
			setupRecords:      []schema.APIRequest{},
			cutoffTime:        time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC),
			expectedDeleted:   0,
			expectedRemaining: []string{},
		},
		{
			name: "exact cutoff time boundary",
			setupRecords: []schema.APIRequest{
				createTestRecord("session1", time.Date(2025, 1, 2, 10, 0, 0, 0, time.UTC)),
				createTestRecord("session2", time.Date(2025, 1, 2, 10, 0, 0, 1, time.UTC)), // 1 nanosecond later
			},
			cutoffTime:        time.Date(2025, 1, 2, 10, 0, 0, 0, time.UTC),
			expectedDeleted:   0, // Records at exact cutoff time should not be deleted
			expectedRemaining: []string{"session1", "session2"},
		},
		{
			name: "records with same timestamp different sessions",
			setupRecords: []schema.APIRequest{
				createTestRecord("sessionA", time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)),
				createTestRecord("sessionB", time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)),
				createTestRecord("sessionC", time.Date(2025, 1, 3, 10, 0, 0, 0, time.UTC)),
			},
			cutoffTime:        time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC),
			expectedDeleted:   2,
			expectedRemaining: []string{"sessionC"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create temporary database
			dbPath := createTempDB(t)
			defer func() {
				if err := os.Remove(dbPath); err != nil {
					t.Logf("Failed to remove temp database: %v", err)
				}
			}()

			// Open database
			db, err := bbolt.Open(dbPath, 0600, nil)
			if err != nil {
				t.Fatalf("Failed to open database: %v", err)
			}
			defer func() {
				if err := db.Close(); err != nil {
					t.Logf("Failed to close database: %v", err)
				}
			}()

			// Initialize bucket
			err = db.Update(func(tx *bbolt.Tx) error {
				_, err := tx.CreateBucket([]byte(requestsBucket))
				return err
			})
			if err != nil {
				t.Fatalf("Failed to create bucket: %v", err)
			}

			// Create repository
			repo := NewBoltDBAPIRequestRepository(db)

			// Setup test records
			for _, record := range tt.setupRecords {
				entity := createTestEntity(record.SessionID, record.Timestamp)
				err := repo.Save(entity)
				if err != nil {
					t.Fatalf("Failed to save test record: %v", err)
				}
			}

			// Execute DeleteOlderThan
			deletedCount, err := repo.DeleteOlderThan(tt.cutoffTime)
			if err != nil {
				t.Fatalf("DeleteOlderThan() error = %v", err)
			}

			// Verify deleted count
			if deletedCount != tt.expectedDeleted {
				t.Errorf("DeleteOlderThan() deleted count = %d, want %d", deletedCount, tt.expectedDeleted)
			}

			// Verify remaining records
			remaining, err := repo.FindAll()
			if err != nil {
				t.Fatalf("Failed to fetch remaining records: %v", err)
			}

			if len(remaining) != len(tt.expectedRemaining) {
				t.Errorf("Remaining records count = %d, want %d", len(remaining), len(tt.expectedRemaining))
			}

			// Verify specific remaining sessions
			remainingSessions := make(map[string]bool)
			for _, record := range remaining {
				remainingSessions[record.SessionID()] = true
			}

			for _, expectedSession := range tt.expectedRemaining {
				if !remainingSessions[expectedSession] {
					t.Errorf("Expected session %s to remain but it was deleted", expectedSession)
				}
			}

			// Verify no unexpected sessions remain
			for sessionID := range remainingSessions {
				found := false
				for _, expected := range tt.expectedRemaining {
					if sessionID == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Unexpected session %s remained after deletion", sessionID)
				}
			}
		})
	}
}

func TestBoltDBAPIRequestRepository_DeleteOlderThanWithLargeDataset(t *testing.T) {
	t.Parallel()

	// Create temporary database
	dbPath := createTempDB(t)
	defer func() {
		if err := os.Remove(dbPath); err != nil {
			t.Logf("Failed to remove temp database: %v", err)
		}
	}()

	// Open database
	db, err := bbolt.Open(dbPath, 0600, nil)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("Failed to close database: %v", err)
		}
	}()

	// Initialize bucket
	err = db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucket([]byte(requestsBucket))
		return err
	})
	if err != nil {
		t.Fatalf("Failed to create bucket: %v", err)
	}

	repo := NewBoltDBAPIRequestRepository(db)

	// Create 1000 records over different time periods
	baseTime := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	recordCount := 1000

	for i := 0; i < recordCount; i++ {
		timestamp := baseTime.Add(time.Duration(i) * time.Hour)
		entity := createTestEntity(fmt.Sprintf("session%d", i), timestamp)
		err := repo.Save(entity)
		if err != nil {
			t.Fatalf("Failed to save test record %d: %v", i, err)
		}
	}

	// Delete records older than 500 hours from base time
	cutoffTime := baseTime.Add(500 * time.Hour)
	deletedCount, err := repo.DeleteOlderThan(cutoffTime)
	if err != nil {
		t.Fatalf("DeleteOlderThan() error = %v", err)
	}

	// Should delete 500 records (0-499 hours)
	expectedDeleted := 500
	if deletedCount != expectedDeleted {
		t.Errorf("DeleteOlderThan() deleted count = %d, want %d", deletedCount, expectedDeleted)
	}

	// Verify remaining count
	remaining, err := repo.FindAll()
	if err != nil {
		t.Fatalf("Failed to fetch remaining records: %v", err)
	}

	expectedRemaining := recordCount - expectedDeleted
	if len(remaining) != expectedRemaining {
		t.Errorf("Remaining records count = %d, want %d", len(remaining), expectedRemaining)
	}
}

func TestBoltDBAPIRequestRepository_DeleteOlderThanDatabaseError(t *testing.T) {
	t.Parallel()

	// Create temporary database
	dbPath := createTempDB(t)
	defer func() {
		if err := os.Remove(dbPath); err != nil {
			t.Logf("Failed to remove temp database: %v", err)
		}
	}()

	// Open database
	db, err := bbolt.Open(dbPath, 0600, nil)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// Initialize bucket
	err = db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucket([]byte(requestsBucket))
		return err
	})
	if err != nil {
		t.Fatalf("Failed to create bucket: %v", err)
	}

	repo := NewBoltDBAPIRequestRepository(db)

	// Close database to simulate error
	if err := db.Close(); err != nil {
		t.Logf("Failed to close database (expected for test): %v", err)
	}

	// Attempt to delete should return error
	cutoffTime := time.Now().Add(-24 * time.Hour)
	deletedCount, err := repo.DeleteOlderThan(cutoffTime)

	if err == nil {
		t.Errorf("DeleteOlderThan() expected error when database is closed but got none")
	}

	if deletedCount != 0 {
		t.Errorf("DeleteOlderThan() deleted count = %d, want 0 when error occurs", deletedCount)
	}
}

// Helper functions

func createTempDB(t *testing.T) string {
	tempDir := t.TempDir()
	return filepath.Join(tempDir, "test.db")
}

func createTestRecord(sessionID string, timestamp time.Time) schema.APIRequest {
	return schema.APIRequest{
		SessionID:           sessionID,
		Timestamp:           timestamp,
		Model:               "claude-3-sonnet",
		InputTokens:         100,
		OutputTokens:        50,
		CacheReadTokens:     0,
		CacheCreationTokens: 0,
		TotalTokens:         150,
		CostUSD:             0.001,
		DurationMS:          1000,
	}
}

func createTestEntity(sessionID string, timestamp time.Time) entity.APIRequest {
	tokens := entity.NewToken(100, 50, 0, 0)
	cost := entity.NewCost(0.001)
	return entity.NewAPIRequest(sessionID, timestamp, "claude-3-sonnet", tokens, cost, 1000)
}
