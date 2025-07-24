package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/elct9620/ccmon/entity"
	pb "github.com/elct9620/ccmon/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// GRPCStatsRepository implements usecase.StatsRepository using gRPC GetStats call
// This is used on the client side to get pre-calculated stats from the server
type GRPCStatsRepository struct {
	client pb.QueryServiceClient
	conn   *grpc.ClientConn
}

// NewGRPCStatsRepository creates a new gRPC stats repository instance
func NewGRPCStatsRepository(serverAddress string) (*GRPCStatsRepository, error) {
	// Create connection
	conn, err := grpc.NewClient(serverAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC server at %s: %w", serverAddress, err)
	}

	client := pb.NewQueryServiceClient(conn)

	return &GRPCStatsRepository{
		client: client,
		conn:   conn,
	}, nil
}

// GetStatsByPeriod retrieves stats for a given period via gRPC GetStats
func (r *GRPCStatsRepository) GetStatsByPeriod(period entity.Period) (entity.Stats, error) {
	// Convert entity.Period to protobuf timestamps
	var startTime, endTime *timestamppb.Timestamp

	if !period.IsAllTime() {
		startTime = timestamppb.New(period.StartAt())
	}
	endTime = timestamppb.New(period.EndAt())

	// Create gRPC request
	req := &pb.GetStatsRequest{
		StartTime: startTime,
		EndTime:   endTime,
	}

	// Call gRPC service
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := r.client.GetStats(ctx, req)
	if err != nil {
		return entity.Stats{}, fmt.Errorf("failed to get stats via gRPC: %w", err)
	}

	// Convert protobuf response to entity
	return convertProtoToStats(resp.Stats, period), nil
}

// Close closes the gRPC connection
func (r *GRPCStatsRepository) Close() error {
	return r.conn.Close()
}

// convertProtoToStats converts protobuf Stats to entity.Stats
func convertProtoToStats(pbStats *pb.Stats, period entity.Period) entity.Stats {
	// Convert protobuf tokens to entities
	baseTokens := entity.NewToken(
		pbStats.BaseTokens.Input,
		pbStats.BaseTokens.Output,
		pbStats.BaseTokens.CacheRead,
		pbStats.BaseTokens.CacheCreation,
	)

	premiumTokens := entity.NewToken(
		pbStats.PremiumTokens.Input,
		pbStats.PremiumTokens.Output,
		pbStats.PremiumTokens.CacheRead,
		pbStats.PremiumTokens.CacheCreation,
	)

	// Convert protobuf costs to entities
	baseCost := entity.NewCost(pbStats.BaseCost.Amount)
	premiumCost := entity.NewCost(pbStats.PremiumCost.Amount)

	// Create stats entity
	return entity.NewStats(
		int(pbStats.BaseRequests),
		int(pbStats.PremiumRequests),
		baseTokens,
		premiumTokens,
		baseCost,
		premiumCost,
		period,
	)
}
