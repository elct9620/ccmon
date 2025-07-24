package repository

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/elct9620/ccmon/entity"
	pb "github.com/elct9620/ccmon/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

// MockQueryServiceServer for testing GRPCStatsRepository
type MockQueryServiceServer struct {
	pb.UnimplementedQueryServiceServer
	stats *pb.Stats
	err   error
}

func (m *MockQueryServiceServer) GetStats(ctx context.Context, req *pb.GetStatsRequest) (*pb.GetStatsResponse, error) {
	if m.err != nil {
		return nil, m.err
	}

	return &pb.GetStatsResponse{
		Stats: m.stats,
	}, nil
}

func (m *MockQueryServiceServer) GetAPIRequests(ctx context.Context, req *pb.GetAPIRequestsRequest) (*pb.GetAPIRequestsResponse, error) {
	return &pb.GetAPIRequestsResponse{}, nil
}

// setupMockGRPCServer creates a mock gRPC server for testing
func setupMockGRPCServer(mockStats *pb.Stats, mockErr error) (*grpc.Server, *bufconn.Listener) {
	listener := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()

	mockService := &MockQueryServiceServer{
		stats: mockStats,
		err:   mockErr,
	}
	pb.RegisterQueryServiceServer(server, mockService)

	go func() {
		_ = server.Serve(listener) // Expected to fail when test completes
	}()

	return server, listener
}

// createGRPCStatsRepository creates a GRPCStatsRepository connected to the mock server
func createGRPCStatsRepository(listener *bufconn.Listener) (*GRPCStatsRepository, error) {
	conn, err := grpc.NewClient("passthrough://bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return listener.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	client := pb.NewQueryServiceClient(conn)
	return &GRPCStatsRepository{
		client: client,
		conn:   conn,
	}, nil
}

func TestGRPCStatsRepository_GetStatsByPeriod(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		mockStats     *pb.Stats
		mockErr       error
		period        entity.Period
		expectedStats entity.Stats
		expectError   bool
	}{
		{
			name: "successful stats retrieval",
			mockStats: &pb.Stats{
				BaseRequests:    2,
				PremiumRequests: 3,
				TotalRequests:   5,
				BaseTokens: &pb.Token{
					Input:         100,
					Output:        80,
					CacheRead:     0,
					CacheCreation: 0,
					Total:         180,
					Limited:       180,
					Cache:         0,
				},
				PremiumTokens: &pb.Token{
					Input:         300,
					Output:        250,
					CacheRead:     0,
					CacheCreation: 0,
					Total:         550,
					Limited:       550,
					Cache:         0,
				},
				TotalTokens: &pb.Token{
					Input:         400,
					Output:        330,
					CacheRead:     0,
					CacheCreation: 0,
					Total:         730,
					Limited:       730,
					Cache:         0,
				},
				BaseCost:    &pb.Cost{Amount: 5.0},
				PremiumCost: &pb.Cost{Amount: 15.0},
				TotalCost:   &pb.Cost{Amount: 20.0},
			},
			period: entity.NewPeriod(
				time.Date(2025, 7, 24, 0, 0, 0, 0, time.UTC),
				time.Date(2025, 7, 24, 23, 59, 59, 999999999, time.UTC),
			),
			expectedStats: entity.NewStats(
				2,                               // base requests
				3,                               // premium requests
				entity.NewToken(100, 80, 0, 0),  // base tokens
				entity.NewToken(300, 250, 0, 0), // premium tokens
				entity.NewCost(5.0),             // base cost
				entity.NewCost(15.0),            // premium cost
				entity.NewPeriod(
					time.Date(2025, 7, 24, 0, 0, 0, 0, time.UTC),
					time.Date(2025, 7, 24, 23, 59, 59, 999999999, time.UTC),
				),
			),
			expectError: false,
		},
		{
			name: "zero stats",
			mockStats: &pb.Stats{
				BaseRequests:    0,
				PremiumRequests: 0,
				TotalRequests:   0,
				BaseTokens: &pb.Token{
					Input:         0,
					Output:        0,
					CacheRead:     0,
					CacheCreation: 0,
					Total:         0,
					Limited:       0,
					Cache:         0,
				},
				PremiumTokens: &pb.Token{
					Input:         0,
					Output:        0,
					CacheRead:     0,
					CacheCreation: 0,
					Total:         0,
					Limited:       0,
					Cache:         0,
				},
				TotalTokens: &pb.Token{
					Input:         0,
					Output:        0,
					CacheRead:     0,
					CacheCreation: 0,
					Total:         0,
					Limited:       0,
					Cache:         0,
				},
				BaseCost:    &pb.Cost{Amount: 0.0},
				PremiumCost: &pb.Cost{Amount: 0.0},
				TotalCost:   &pb.Cost{Amount: 0.0},
			},
			period: entity.NewAllTimePeriod(time.Now()),
			expectedStats: entity.NewStats(
				0, 0, // no requests
				entity.NewToken(0, 0, 0, 0), entity.NewToken(0, 0, 0, 0), // no tokens
				entity.NewCost(0), entity.NewCost(0), // no cost
				entity.NewAllTimePeriod(time.Now()),
			),
			expectError: false,
		},
		{
			name:        "gRPC error",
			mockStats:   nil,
			mockErr:     fmt.Errorf("connection refused"),
			period:      entity.NewAllTimePeriod(time.Now()),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup mock gRPC server
			server, listener := setupMockGRPCServer(tt.mockStats, tt.mockErr)
			defer server.Stop()

			// Create GRPCStatsRepository
			statsRepo, err := createGRPCStatsRepository(listener)
			if err != nil {
				t.Fatalf("Failed to create GRPCStatsRepository: %v", err)
			}
			defer func() {
				if err := statsRepo.Close(); err != nil {
					t.Logf("Failed to close statsRepo: %v", err)
				}
			}()

			// Execute
			result, err := statsRepo.GetStatsByPeriod(tt.period)

			// Verify error expectation
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Verify stats (only check key fields to avoid overly complex comparison)
			if result.BaseRequests() != tt.expectedStats.BaseRequests() {
				t.Errorf("Base requests: expected %d, got %d", tt.expectedStats.BaseRequests(), result.BaseRequests())
			}
			if result.PremiumRequests() != tt.expectedStats.PremiumRequests() {
				t.Errorf("Premium requests: expected %d, got %d", tt.expectedStats.PremiumRequests(), result.PremiumRequests())
			}
			if result.BaseCost().Amount() != tt.expectedStats.BaseCost().Amount() {
				t.Errorf("Base cost: expected %.1f, got %.1f", tt.expectedStats.BaseCost().Amount(), result.BaseCost().Amount())
			}
			if result.PremiumCost().Amount() != tt.expectedStats.PremiumCost().Amount() {
				t.Errorf("Premium cost: expected %.1f, got %.1f", tt.expectedStats.PremiumCost().Amount(), result.PremiumCost().Amount())
			}
		})
	}
}

func TestGRPCStatsRepository_Close(t *testing.T) {
	// Setup mock gRPC server
	server, listener := setupMockGRPCServer(&pb.Stats{}, nil)
	defer server.Stop()

	// Create GRPCStatsRepository
	statsRepo, err := createGRPCStatsRepository(listener)
	if err != nil {
		t.Fatalf("Failed to create GRPCStatsRepository: %v", err)
	}

	// Test Close
	err = statsRepo.Close()
	if err != nil {
		t.Errorf("Unexpected error closing repository: %v", err)
	}
}
