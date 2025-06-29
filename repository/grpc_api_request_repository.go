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
	"google.golang.org/protobuf/types/known/timestamppb"
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

// FindByPeriodWithLimit retrieves API requests filtered by time period with limit and offset via gRPC
// Use limit = 0 for no limit (fetch all records)
// Use offset = 0 when no offset is needed
func (r *GRPCAPIRequestRepository) FindByPeriodWithLimit(period entity.Period, limit int, offset int) ([]entity.APIRequest, error) {
	// Convert entity.Period to protobuf timestamps
	var startTime, endTime *timestamppb.Timestamp

	if !period.IsAllTime() {
		startTime = timestamppb.New(period.StartAt())
	}
	endTime = timestamppb.New(period.EndAt())

	// Create gRPC request with timestamps, limit and offset
	req := &pb.GetAPIRequestsRequest{
		StartTime: startTime,
		EndTime:   endTime,
		Limit:     int32(limit),
		Offset:    int32(offset),
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
	// Use all-time period with no limit
	return r.FindByPeriodWithLimit(entity.NewAllTimePeriod(), 0, 0)
}

// Close closes the gRPC connection
func (r *GRPCAPIRequestRepository) Close() error {
	return r.conn.Close()
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
