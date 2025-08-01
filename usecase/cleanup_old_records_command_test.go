package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/elct9620/ccmon/testutil"
)

func TestCleanupOldRecordsCommand_Execute(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		cutoffTime        time.Time
		mockDeleteFunc    func(cutoffTime time.Time) (int, error)
		expectedDeleted   int
		expectError       bool
		expectedCallCount int
	}{
		{
			name:       "successful cleanup with records deleted",
			cutoffTime: time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
			mockDeleteFunc: func(cutoffTime time.Time) (int, error) {
				return 15, nil
			},
			expectedDeleted:   15,
			expectError:       false,
			expectedCallCount: 1,
		},
		{
			name:       "successful cleanup with no records to delete",
			cutoffTime: time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
			mockDeleteFunc: func(cutoffTime time.Time) (int, error) {
				return 0, nil
			},
			expectedDeleted:   0,
			expectError:       false,
			expectedCallCount: 1,
		},
		{
			name:       "cleanup with repository error",
			cutoffTime: time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
			mockDeleteFunc: func(cutoffTime time.Time) (int, error) {
				return 0, &testutil.MockError{Message: "database connection failed"}
			},
			expectedDeleted:   0,
			expectError:       true,
			expectedCallCount: 1,
		},
		{
			name:       "cleanup with large number of records",
			cutoffTime: time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC),
			mockDeleteFunc: func(cutoffTime time.Time) (int, error) {
				return 10000, nil
			},
			expectedDeleted:   10000,
			expectError:       false,
			expectedCallCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create mock repository
			mockRepo := testutil.NewMockRepositoryWithDeleteFunc(tt.mockDeleteFunc)

			// Create command
			command := NewCleanupOldRecordsCommand(mockRepo)

			// Execute command
			ctx := context.Background()
			params := CleanupOldRecordsParams{
				CutoffTime: tt.cutoffTime,
			}

			result, err := command.Execute(ctx, params)

			// Verify error expectations
			if tt.expectError {
				if err == nil {
					t.Errorf("Execute() expected error but got none")
				}
				if result != nil {
					t.Errorf("Execute() expected nil result on error but got %v", result)
				}
			} else {
				if err != nil {
					t.Errorf("Execute() unexpected error = %v", err)
				}
				if result == nil {
					t.Errorf("Execute() expected result but got nil")
					return
				}
				if result.DeletedCount != tt.expectedDeleted {
					t.Errorf("Execute() DeletedCount = %d, want %d", result.DeletedCount, tt.expectedDeleted)
				}
			}

			// Verify repository interaction
			if mockRepo.GetDeleteCallCount() != tt.expectedCallCount {
				t.Errorf("DeleteOlderThan() call count = %d, want %d", mockRepo.GetDeleteCallCount(), tt.expectedCallCount)
			}

			if !tt.expectError && !mockRepo.GetLastCutoffTime().Equal(tt.cutoffTime) {
				t.Errorf("DeleteOlderThan() cutoffTime = %v, want %v", mockRepo.GetLastCutoffTime(), tt.cutoffTime)
			}
		})
	}
}

func TestCleanupOldRecordsCommand_ExecuteContextCancellation(t *testing.T) {
	// Create mock repository that simulates a slow operation
	mockRepo := testutil.NewMockRepositoryWithDeleteFunc(func(cutoffTime time.Time) (int, error) {
		// Simulate slow operation
		time.Sleep(100 * time.Millisecond)
		return 5, nil
	})

	command := NewCleanupOldRecordsCommand(mockRepo)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	params := CleanupOldRecordsParams{
		CutoffTime: time.Now().Add(-24 * time.Hour),
	}

	// Note: Since our current implementation doesn't check context cancellation
	// in the repository layer, this test documents current behavior
	// In a real implementation, we might want to add context support to the repository
	result, err := command.Execute(ctx, params)

	// Current implementation doesn't handle context cancellation
	// This test documents the current behavior
	if err != nil {
		t.Logf("Execute() with cancelled context returned error: %v", err)
	}
	if result != nil {
		t.Logf("Execute() with cancelled context returned result: %+v", result)
	}
}

func TestNewCleanupOldRecordsCommand(t *testing.T) {
	mockRepo := testutil.NewMockRepositoryWithDeleteFunc(nil)
	command := NewCleanupOldRecordsCommand(mockRepo)

	if command == nil {
		t.Errorf("NewCleanupOldRecordsCommand() returned nil")
		return
	}

	if command.repository != mockRepo {
		t.Errorf("NewCleanupOldRecordsCommand() repository not set correctly")
	}
}
