package query

import (
	"context"
	"fmt"
	"time"

	"github.com/elct9620/ccmon/entity"
	pb "github.com/elct9620/ccmon/proto"
	"github.com/elct9620/ccmon/usecase"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Service implements the QueryService gRPC interface
type Service struct {
	pb.UnimplementedQueryServiceServer
	getFilteredQuery    *usecase.GetFilteredApiRequestsQuery
	calculateStatsQuery *usecase.CalculateStatsQuery
}

// NewService creates a new query service instance
func NewService(getFilteredQuery *usecase.GetFilteredApiRequestsQuery, calculateStatsQuery *usecase.CalculateStatsQuery) *Service {
	return &Service{
		getFilteredQuery:    getFilteredQuery,
		calculateStatsQuery: calculateStatsQuery,
	}
}

// GetStats returns aggregated statistics based on time range
func (s *Service) GetStats(ctx context.Context, req *pb.GetStatsRequest) (*pb.GetStatsResponse, error) {
	// Convert proto timestamps to entity.Period
	period := convertTimestampsToPeriod(req.StartTime, req.EndTime)

	// Get stats via usecase
	params := usecase.CalculateStatsParams{Period: period}
	stats, err := s.calculateStatsQuery.Execute(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}

	// Convert to protobuf response
	pbStats := &pb.Stats{
		BaseRequests:    int32(stats.BaseRequests()),
		PremiumRequests: int32(stats.PremiumRequests()),
		TotalRequests:   int32(stats.TotalRequests()),
		BaseTokens:      convertTokenToProto(stats.BaseTokens()),
		PremiumTokens:   convertTokenToProto(stats.PremiumTokens()),
		TotalTokens:     convertTokenToProto(stats.TotalTokens()),
		BaseCost:        convertCostToProto(stats.BaseCost()),
		PremiumCost:     convertCostToProto(stats.PremiumCost()),
		TotalCost:       convertCostToProto(stats.TotalCost()),
	}

	return &pb.GetStatsResponse{
		Stats: pbStats,
	}, nil
}

// GetAPIRequests returns API request records based on filters
func (s *Service) GetAPIRequests(ctx context.Context, req *pb.GetAPIRequestsRequest) (*pb.GetAPIRequestsResponse, error) {
	// Convert proto timestamps to entity.Period
	period := convertTimestampsToPeriod(req.StartTime, req.EndTime)

	// Get requests via usecase with limit and offset
	params := usecase.GetFilteredApiRequestsParams{
		Period: period,
		Limit:  int(req.Limit),
		Offset: int(req.Offset),
	}
	requests, err := s.getFilteredQuery.Execute(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get requests: %w", err)
	}

	// Note: TotalCount is now the count of returned records since pagination
	// is handled at repository level. For true total count, we'd need a separate query.
	totalCount := len(requests)

	// Convert to protobuf messages
	pbRequests := make([]*pb.APIRequest, len(requests))
	for i, apiReq := range requests {
		pbRequests[i] = convertAPIRequestToProto(apiReq)
	}

	return &pb.GetAPIRequestsResponse{
		Requests:   pbRequests,
		TotalCount: int32(totalCount),
	}, nil
}

// convertTimestampsToPeriod converts protobuf timestamps to entity.Period
func convertTimestampsToPeriod(startTime, endTime *timestamppb.Timestamp) entity.Period {
	// Handle nil timestamps - use all time period
	if startTime == nil && endTime == nil {
		return entity.NewAllTimePeriod(time.Now().UTC())
	}

	var start, end time.Time

	// Convert startTime, default to zero time for "all time from beginning"
	if startTime != nil {
		start = startTime.AsTime()
	} else {
		start = time.Time{} // Zero time represents "all time"
	}

	// Convert endTime, default to current time
	if endTime != nil {
		end = endTime.AsTime()
	} else {
		end = time.Now().UTC()
	}

	return entity.NewPeriod(start, end)
}

// convertTokenToProto converts entity.Token to protobuf Token
func convertTokenToProto(token entity.Token) *pb.Token {
	return &pb.Token{
		Total:         token.Total(),
		Input:         token.Input(),
		Output:        token.Output(),
		CacheRead:     token.CacheRead(),
		CacheCreation: token.CacheCreation(),
		Limited:       token.Limited(),
		Cache:         token.Cache(),
	}
}

// convertCostToProto converts entity.Cost to protobuf Cost
func convertCostToProto(cost entity.Cost) *pb.Cost {
	return &pb.Cost{
		Amount: cost.Amount(),
	}
}

// convertAPIRequestToProto converts entity.APIRequest to protobuf APIRequest
func convertAPIRequestToProto(req entity.APIRequest) *pb.APIRequest {
	return &pb.APIRequest{
		SessionId:           req.SessionID(),
		Timestamp:           timestamppb.New(req.Timestamp()),
		Model:               string(req.Model()),
		InputTokens:         req.Tokens().Input(),
		OutputTokens:        req.Tokens().Output(),
		CacheReadTokens:     req.Tokens().CacheRead(),
		CacheCreationTokens: req.Tokens().CacheCreation(),
		TotalTokens:         req.Tokens().Total(),
		CostUsd:             req.Cost().Amount(),
		DurationMs:          req.DurationMS(),
	}
}
