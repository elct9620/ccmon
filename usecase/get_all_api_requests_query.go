package usecase

import (
	"context"

	"github.com/elct9620/ccmon/entity"
)

// GetAllApiRequestsQuery handles the query to get all API requests
type GetAllApiRequestsQuery struct {
	repository APIRequestRepository
}

// NewGetAllApiRequestsQuery creates a new GetAllApiRequestsQuery with the given repository
func NewGetAllApiRequestsQuery(repository APIRequestRepository) *GetAllApiRequestsQuery {
	return &GetAllApiRequestsQuery{
		repository: repository,
	}
}

// Execute executes the get all API requests query
func (q *GetAllApiRequestsQuery) Execute(ctx context.Context) ([]entity.APIRequest, error) {
	return q.repository.FindAll()
}