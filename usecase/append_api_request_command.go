package usecase

import (
	"context"
	"time"

	"github.com/elct9620/ccmon/entity"
)

// AppendApiRequestCommand handles the command to append a new API request
type AppendApiRequestCommand struct {
	repository APIRequestRepository
}

// NewAppendApiRequestCommand creates a new AppendApiRequestCommand with the given repository
func NewAppendApiRequestCommand(repository APIRequestRepository) *AppendApiRequestCommand {
	return &AppendApiRequestCommand{
		repository: repository,
	}
}

// AppendApiRequestParams contains the parameters for appending an API request
type AppendApiRequestParams struct {
	SessionID  string
	Timestamp  time.Time
	Model      string
	Tokens     entity.Token
	Cost       entity.Cost
	DurationMS int64
}

// Execute executes the append API request command
func (c *AppendApiRequestCommand) Execute(ctx context.Context, params AppendApiRequestParams) error {
	// Create the API request entity
	apiRequest := entity.NewAPIRequest(
		params.SessionID,
		params.Timestamp,
		params.Model,
		params.Tokens,
		params.Cost,
		params.DurationMS,
	)

	// Save the API request via repository
	return c.repository.Save(apiRequest)
}
