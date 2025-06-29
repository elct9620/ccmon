package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/elct9620/ccmon/entity"
	pb "github.com/elct9620/ccmon/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// GRPCAPIRequestRepository implements APIRequestRepository using gRPC client
type GRPCAPIRequestRepository struct {
	client pb.QueryServiceClient
	conn   *grpc.ClientConn
}

// NewGRPCAPIRequestRepository creates a new gRPC repository instance
func NewGRPCAPIRequestRepository(serverAddress string) (*GRPCAPIRequestRepository, error) {
	// Create connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, serverAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC server at %s: %w", serverAddress, err)
	}

	client := pb.NewQueryServiceClient(conn)

	return &GRPCAPIRequestRepository{
		client: client,
		conn:   conn,
	}, nil
}

// Save is not supported in monitor mode (read-only repository)
func (r *GRPCAPIRequestRepository) Save(req entity.APIRequest) error {
	return errors.New("save operation not supported in monitor mode (read-only repository)")
}

// FindByPeriod retrieves API requests filtered by time period via gRPC
func (r *GRPCAPIRequestRepository) FindByPeriod(period entity.Period) ([]entity.APIRequest, error) {
	// Convert entity.Period to protobuf TimeFilter
	timeFilter := convertPeriodToTimeFilter(period)

	// Create gRPC request
	req := &pb.GetAPIRequestsRequest{
		TimeFilter: timeFilter,
	}

	// Call gRPC service
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := r.client.GetAPIRequests(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get API requests via gRPC: %w", err)
	}

	// Convert protobuf responses to entities
	entities := make([]entity.APIRequest, len(resp.Requests))
	for i, pbReq := range resp.Requests {
		entities[i] = convertProtoToAPIRequest(pbReq)
	}

	return entities, nil
}

// FindAll retrieves all API requests via gRPC
func (r *GRPCAPIRequestRepository) FindAll() ([]entity.APIRequest, error) {
	// Use all-time period
	return r.FindByPeriod(entity.NewAllTimePeriod())
}

// Close closes the gRPC connection
func (r *GRPCAPIRequestRepository) Close() error {
	return r.conn.Close()
}

// convertPeriodToTimeFilter converts entity.Period to protobuf TimeFilter
func convertPeriodToTimeFilter(period entity.Period) pb.TimeFilter {
	if period.IsAllTime() {
		return pb.TimeFilter_TIME_FILTER_ALL
	}

	duration := period.Duration()
	switch {
	case duration <= time.Hour:
		return pb.TimeFilter_TIME_FILTER_HOUR
	case duration <= 24*time.Hour:
		return pb.TimeFilter_TIME_FILTER_DAY
	case duration <= 7*24*time.Hour:
		return pb.TimeFilter_TIME_FILTER_WEEK
	case duration <= 30*24*time.Hour:
		return pb.TimeFilter_TIME_FILTER_MONTH
	default:
		return pb.TimeFilter_TIME_FILTER_ALL
	}
}

// convertProtoToAPIRequest converts protobuf APIRequest to entity.APIRequest
func convertProtoToAPIRequest(pbReq *pb.APIRequest) entity.APIRequest {
	// Create token entity
	tokens := entity.NewToken(
		pbReq.InputTokens,
		pbReq.OutputTokens,
		pbReq.CacheReadTokens,
		pbReq.CacheCreationTokens,
	)

	// Create cost entity
	cost := entity.NewCost(pbReq.CostUsd)

	// Create API request entity
	return entity.NewAPIRequest(
		pbReq.SessionId,
		pbReq.Timestamp.AsTime(),
		pbReq.Model,
		tokens,
		cost,
		pbReq.DurationMs,
	)
}
