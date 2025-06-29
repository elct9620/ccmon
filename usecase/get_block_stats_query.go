package usecase

import (
	"context"

	"github.com/elct9620/ccmon/entity"
)

// GetBlockStatsQuery handles the query to get block statistics
type GetBlockStatsQuery struct {
	repository BlockStatsRepository
}

// NewGetBlockStatsQuery creates a new GetBlockStatsQuery with the given repository
func NewGetBlockStatsQuery(repository BlockStatsRepository) *GetBlockStatsQuery {
	return &GetBlockStatsQuery{
		repository: repository,
	}
}

// GetBlockStatsParams contains the parameters for getting block statistics
type GetBlockStatsParams struct {
	Block entity.Block
}

// Execute executes the get block statistics query
func (q *GetBlockStatsQuery) Execute(ctx context.Context, params GetBlockStatsParams) (entity.Stats, error) {
	// Get block stats via repository
	return q.repository.GetBlockStats(params.Block)
}