package usecase

import (
	"context"
	"time"
)

// CleanupOldRecordsCommand handles the command to cleanup old API request records
type CleanupOldRecordsCommand struct {
	repository APIRequestRepository
}

// NewCleanupOldRecordsCommand creates a new CleanupOldRecordsCommand with the given repository
func NewCleanupOldRecordsCommand(repository APIRequestRepository) *CleanupOldRecordsCommand {
	return &CleanupOldRecordsCommand{
		repository: repository,
	}
}

// CleanupOldRecordsParams contains the parameters for cleaning up old records
type CleanupOldRecordsParams struct {
	CutoffTime time.Time
}

// CleanupOldRecordsResult contains the result of the cleanup operation
type CleanupOldRecordsResult struct {
	DeletedCount int
}

// Execute executes the cleanup old records command
func (c *CleanupOldRecordsCommand) Execute(ctx context.Context, params CleanupOldRecordsParams) (*CleanupOldRecordsResult, error) {
	// Delete records older than cutoff time via repository
	deletedCount, err := c.repository.DeleteOlderThan(params.CutoffTime)
	if err != nil {
		return nil, err
	}

	return &CleanupOldRecordsResult{
		DeletedCount: deletedCount,
	}, nil
}
