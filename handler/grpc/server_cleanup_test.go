package grpc

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/elct9620/ccmon/entity"
	"github.com/elct9620/ccmon/repository"
	"github.com/elct9620/ccmon/repository/schema"
	"github.com/elct9620/ccmon/usecase"
	"go.etcd.io/bbolt"
)

// MockServerConfig implements ServerConfig interface for testing
type MockServerConfig struct {
	retention string
}

func (m MockServerConfig) IsRetentionEnabled() bool {
	return m.retention != "" && m.retention != "never"
}

func (m MockServerConfig) GetRetentionDuration() time.Duration {
	if !m.IsRetentionEnabled() {
		return 0
	}
	duration, err := time.ParseDuration(m.retention)
	if err != nil {
		return 0
	}
	return duration
}

func TestCleanupSchedulerIntegration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                  string
		serverConfig          MockServerConfig
		initialRecords        []schema.APIRequest
		expectedSchedulerRuns bool
		waitTime              time.Duration
		expectedDeleted       int
	}{
		{
			name:         "retention disabled - no cleanup",
			serverConfig: MockServerConfig{retention: "never"},
			initialRecords: []schema.APIRequest{
				createTestAPIRequest("session1", time.Now().Add(-48*time.Hour)),
				createTestAPIRequest("session2", time.Now().Add(-12*time.Hour)),
			},
			expectedSchedulerRuns: false,
			waitTime:              100 * time.Millisecond,
			expectedDeleted:       0,
		},
		{
			name:         "retention enabled - cleanup runs",
			serverConfig: MockServerConfig{retention: "24h"},
			initialRecords: []schema.APIRequest{
				createTestAPIRequest("session1", time.Now().Add(-48*time.Hour)), // Should be deleted
				createTestAPIRequest("session2", time.Now().Add(-12*time.Hour)), // Should remain
			},
			expectedSchedulerRuns: true,
			waitTime:              200 * time.Millisecond, // Wait for initial cleanup
			expectedDeleted:       1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create temporary database
			dbPath := createTempDBFile(t)
			defer os.Remove(dbPath)

			// Setup database with test data
			db := setupTestDatabase(t, dbPath, tt.initialRecords)
			defer db.Close()

			// Create repository and cleanup command
			repo := repository.NewBoltDBAPIRequestRepository(db)
			cleanupCommand := usecase.NewCleanupOldRecordsCommand(repo)

			// Create context for the scheduler
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// Start cleanup scheduler
			if tt.serverConfig.IsRetentionEnabled() {
				startCleanupScheduler(ctx, cleanupCommand, tt.serverConfig)
			}

			// Wait for cleanup to potentially run
			time.Sleep(tt.waitTime)

			// Cancel context to stop scheduler
			cancel()

			// Verify results
			remaining, err := repo.FindAll()
			if err != nil {
				t.Fatalf("Failed to fetch remaining records: %v", err)
			}

			initialCount := len(tt.initialRecords)
			remainingCount := len(remaining)
			actualDeleted := initialCount - remainingCount

			if actualDeleted != tt.expectedDeleted {
				t.Errorf("Cleanup deleted %d records, want %d", actualDeleted, tt.expectedDeleted)
			}

			// Additional verification for retention enabled case
			if tt.expectedSchedulerRuns && tt.expectedDeleted > 0 {
				// Verify that remaining records are newer than cutoff
				cutoffTime := time.Now().Add(-tt.serverConfig.GetRetentionDuration())
				for _, record := range remaining {
					if record.Timestamp().Before(cutoffTime) {
						t.Errorf("Record with timestamp %v should have been deleted (cutoff: %v)",
							record.Timestamp(), cutoffTime)
					}
				}
			}
		})
	}
}

func TestRunCleanupFunction(t *testing.T) {
	t.Parallel()

	// Create temporary database
	dbPath := createTempDBFile(t)
	defer os.Remove(dbPath)

	// Setup test records
	testRecords := []schema.APIRequest{
		createTestAPIRequest("old1", time.Now().Add(-48*time.Hour)),
		createTestAPIRequest("old2", time.Now().Add(-36*time.Hour)),
		createTestAPIRequest("new1", time.Now().Add(-12*time.Hour)),
	}

	db := setupTestDatabase(t, dbPath, testRecords)
	defer db.Close()

	// Create repository and cleanup command
	repo := repository.NewBoltDBAPIRequestRepository(db)
	cleanupCommand := usecase.NewCleanupOldRecordsCommand(repo)

	// Run cleanup with 24 hour retention
	ctx := context.Background()
	retentionDuration := 24 * time.Hour

	// Verify initial count
	initial, err := repo.FindAll()
	if err != nil {
		t.Fatalf("Failed to fetch initial records: %v", err)
	}
	if len(initial) != 3 {
		t.Fatalf("Expected 3 initial records, got %d", len(initial))
	}

	// Run cleanup
	runCleanup(ctx, cleanupCommand, retentionDuration)

	// Verify cleanup results
	remaining, err := repo.FindAll()
	if err != nil {
		t.Fatalf("Failed to fetch remaining records: %v", err)
	}

	expectedRemaining := 1 // Only "new1" should remain
	if len(remaining) != expectedRemaining {
		t.Errorf("Expected %d remaining records, got %d", expectedRemaining, len(remaining))
	}

	// Verify the correct record remains
	if len(remaining) > 0 && remaining[0].SessionID() != "new1" {
		t.Errorf("Expected 'new1' to remain, got %s", remaining[0].SessionID())
	}
}

func TestCleanupSchedulerCancellation(t *testing.T) {
	t.Parallel()

	// Create temporary database
	dbPath := createTempDBFile(t)
	defer os.Remove(dbPath)

	db := setupTestDatabase(t, dbPath, []schema.APIRequest{})
	defer db.Close()

	// Create repository and cleanup command
	repo := repository.NewBoltDBAPIRequestRepository(db)
	cleanupCommand := usecase.NewCleanupOldRecordsCommand(repo)

	// Create context and cancel immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Start cleanup scheduler with cancelled context
	serverConfig := MockServerConfig{retention: "24h"}

	// This should return quickly due to cancelled context
	start := time.Now()
	startCleanupScheduler(ctx, cleanupCommand, serverConfig)

	// Give it a moment to process the cancellation
	time.Sleep(50 * time.Millisecond)

	elapsed := time.Since(start)

	// Scheduler should handle cancellation gracefully and not hang
	if elapsed > 1*time.Second {
		t.Errorf("Cleanup scheduler took too long to handle cancellation: %v", elapsed)
	}
}

// Helper functions

func createTempDBFile(t *testing.T) string {
	tempDir := t.TempDir()
	return filepath.Join(tempDir, "test_cleanup.db")
}

func setupTestDatabase(t *testing.T, dbPath string, records []schema.APIRequest) *bbolt.DB {
	// Open database
	db, err := bbolt.Open(dbPath, 0600, nil)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// Initialize bucket
	err = db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucket([]byte("requests"))
		return err
	})
	if err != nil {
		t.Fatalf("Failed to create bucket: %v", err)
	}

	// Insert test records
	repo := repository.NewBoltDBAPIRequestRepository(db)
	for _, record := range records {
		entity := convertSchemaToEntity(record)
		err := repo.Save(entity)
		if err != nil {
			t.Fatalf("Failed to save test record: %v", err)
		}
	}

	return db
}

func createTestAPIRequest(sessionID string, timestamp time.Time) schema.APIRequest {
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

func convertSchemaToEntity(s schema.APIRequest) entity.APIRequest {
	tokens := entity.NewToken(s.InputTokens, s.OutputTokens, s.CacheReadTokens, s.CacheCreationTokens)
	cost := entity.NewCost(s.CostUSD)
	return entity.NewAPIRequest(s.SessionID, s.Timestamp, s.Model, tokens, cost, s.DurationMS)
}
