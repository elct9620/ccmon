package db

import (
	"github.com/elct9620/ccmon/entity"
)

// ToEntity converts a database APIRequest to an entity APIRequest
func (r APIRequest) ToEntity() entity.APIRequest {
	tokens := entity.NewToken(
		r.InputTokens,
		r.OutputTokens,
		r.CacheReadTokens,
		r.CacheCreationTokens,
	)
	cost := entity.NewCost(r.CostUSD)

	return entity.NewAPIRequest(
		r.SessionID,
		r.Timestamp,
		r.Model,
		tokens,
		cost,
		r.DurationMS,
	)
}

// FromEntity converts an entity APIRequest to a database APIRequest
func FromEntity(e entity.APIRequest) APIRequest {
	return APIRequest{
		SessionID:           e.SessionID(),
		Timestamp:           e.Timestamp(),
		Model:               e.Model().String(),
		InputTokens:         e.Tokens().Input(),
		OutputTokens:        e.Tokens().Output(),
		CacheReadTokens:     e.Tokens().CacheRead(),
		CacheCreationTokens: e.Tokens().CacheCreation(),
		TotalTokens:         e.Tokens().Total(),
		CostUSD:             e.Cost().Amount(),
		DurationMS:          e.DurationMS(),
	}
}

// ToEntities converts a slice of database APIRequests to entity APIRequests
func ToEntities(requests []APIRequest) []entity.APIRequest {
	entities := make([]entity.APIRequest, len(requests))
	for i, req := range requests {
		entities[i] = req.ToEntity()
	}
	return entities
}

// FromEntities converts a slice of entity APIRequests to database APIRequests
func FromEntities(entities []entity.APIRequest) []APIRequest {
	requests := make([]APIRequest, len(entities))
	for i, e := range entities {
		requests[i] = FromEntity(e)
	}
	return requests
}
