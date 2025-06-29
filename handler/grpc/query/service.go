package query

import (
	"context"
	"fmt"
	"time"

	"github.com/elct9620/ccmon/entity"
	pb "github.com/elct9620/ccmon/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Database interface to avoid circular dependency
type Database interface {
	GetAPIRequests(period entity.Period) ([]entity.APIRequest, error)
	GetAllRequests() ([]entity.APIRequest, error)
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
	// Convert proto time filter to entity.Period
	period := convertTimeFilterToPeriod(req.TimeFilter)

	// Get requests from database
	requests, err := s.db.GetAPIRequests(period)
	if err != nil {
		return nil, fmt.Errorf("failed to get requests: %w", err)
	}

	// Calculate stats
	stats := entity.CalculateStats(requests)

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
	// Convert proto time filter to entity.Period
	period := convertTimeFilterToPeriod(req.TimeFilter)

	// Get requests from database
	requests, err := s.db.GetAPIRequests(period)
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
	for i, apiReq := range requests {
		pbRequests[i] = convertAPIRequestToProto(apiReq)
	}

	return &pb.GetAPIRequestsResponse{
		Requests:   pbRequests,
		TotalCount: int32(totalCount),
	}, nil
}

// convertTimeFilterToPeriod converts protobuf TimeFilter to entity.Period
func convertTimeFilterToPeriod(filter pb.TimeFilter) entity.Period {
	switch filter {
	case pb.TimeFilter_TIME_FILTER_ALL:
		return entity.NewAllTimePeriod()
	case pb.TimeFilter_TIME_FILTER_HOUR:
		return entity.NewPeriodFromDuration(time.Hour)
	case pb.TimeFilter_TIME_FILTER_DAY:
		return entity.NewPeriodFromDuration(24 * time.Hour)
	case pb.TimeFilter_TIME_FILTER_WEEK:
		return entity.NewPeriodFromDuration(7 * 24 * time.Hour)
	case pb.TimeFilter_TIME_FILTER_MONTH:
		return entity.NewPeriodFromDuration(30 * 24 * time.Hour)
	default:
		return entity.NewAllTimePeriod()
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
