package usecase

import (
	"context"

	"github.com/elct9620/ccmon/entity"
)

// GetFilteredApiRequestsQuery handles the query to get filtered API requests
type GetFilteredApiRequestsQuery struct {
	repository APIRequestRepository
}

// NewGetFilteredApiRequestsQuery creates a new GetFilteredApiRequestsQuery with the given repository
func NewGetFilteredApiRequestsQuery(repository APIRequestRepository) *GetFilteredApiRequestsQuery {
	return &GetFilteredApiRequestsQuery{
		repository: repository,
	}
}

// GetFilteredApiRequestsParams contains the parameters for getting filtered API requests
type GetFilteredApiRequestsParams struct {
	Period entity.Period
	Limit  int // Use 0 for no limit
	Offset int // Use 0 for no offset
}

// Execute executes the get filtered API requests query
func (q *GetFilteredApiRequestsQuery) Execute(ctx context.Context, params GetFilteredApiRequestsParams) ([]entity.APIRequest, error) {
	return q.repository.FindByPeriodWithLimit(params.Period, params.Limit, params.Offset)
}
