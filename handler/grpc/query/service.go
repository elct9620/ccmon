package query

import (
	"context"
	"fmt"

	"github.com/elct9620/ccmon/db"
	"github.com/elct9620/ccmon/entity"
	pb "github.com/elct9620/ccmon/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Database interface to avoid circular dependency
type Database interface {
	GetAPIRequests(filter db.Filter) ([]db.APIRequest, error)
	GetAllRequests() ([]db.APIRequest, error)
	Close() error
}

// Service implements the QueryService gRPC interface
type Service struct {
	pb.UnimplementedQueryServiceServer
	db Database
}

// NewService creates a new query service instance
func NewService(database Database) *Service {
	return &Service{
		db: database,
	}
}

// GetStats returns aggregated statistics based on time filter
func (s *Service) GetStats(ctx context.Context, req *pb.GetStatsRequest) (*pb.GetStatsResponse, error) {
	// Convert proto time filter to db filter
	dbFilter := db.Filter{
		TimeFilter: convertTimeFilter(req.TimeFilter),
	}

	// Get requests from database
	requests, err := s.db.GetAPIRequests(dbFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to get requests: %w", err)
	}

	// Convert to entities and calculate stats
	entities := db.ToEntities(requests)
	stats := entity.CalculateStats(entities)

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
	// Convert proto time filter to db filter
	dbFilter := db.Filter{
		TimeFilter: convertTimeFilter(req.TimeFilter),
	}

	// Get requests from database
	requests, err := s.db.GetAPIRequests(dbFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to get requests: %w", err)
	}

	// Apply pagination if specified
	totalCount := len(requests)
	if req.Offset > 0 || req.Limit > 0 {
		start := int(req.Offset)
		if start > len(requests) {
			start = len(requests)
		}

		end := len(requests)
		if req.Limit > 0 {
			end = start + int(req.Limit)
			if end > len(requests) {
				end = len(requests)
			}
		}

		requests = requests[start:end]
	}

	// Convert to protobuf messages
	pbRequests := make([]*pb.APIRequest, len(requests))
	for i, req := range requests {
		pbRequests[i] = convertAPIRequestToProto(req)
	}

	return &pb.GetAPIRequestsResponse{
		Requests:   pbRequests,
		TotalCount: int32(totalCount),
	}, nil
}

// convertTimeFilter converts protobuf TimeFilter to db.TimeFilter
func convertTimeFilter(filter pb.TimeFilter) db.TimeFilter {
	switch filter {
	case pb.TimeFilter_TIME_FILTER_ALL:
		return db.FilterAll
	case pb.TimeFilter_TIME_FILTER_HOUR:
		return db.FilterHour
	case pb.TimeFilter_TIME_FILTER_DAY:
		return db.FilterDay
	case pb.TimeFilter_TIME_FILTER_WEEK:
		return db.FilterWeek
	case pb.TimeFilter_TIME_FILTER_MONTH:
		return db.FilterMonth
	default:
		return db.FilterAll
	}
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

// convertAPIRequestToProto converts db.APIRequest to protobuf APIRequest
func convertAPIRequestToProto(req db.APIRequest) *pb.APIRequest {
	return &pb.APIRequest{
		SessionId:           req.SessionID,
		Timestamp:           timestamppb.New(req.Timestamp),
		Model:               req.Model,
		InputTokens:         req.InputTokens,
		OutputTokens:        req.OutputTokens,
		CacheReadTokens:     req.CacheReadTokens,
		CacheCreationTokens: req.CacheCreationTokens,
		TotalTokens:         req.TotalTokens,
		CostUsd:             req.CostUSD,
		DurationMs:          req.DurationMS,
	}
}
