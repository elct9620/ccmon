package grpc

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/elct9620/ccmon/entity"
	"github.com/elct9620/ccmon/handler/grpc/query"
	"github.com/elct9620/ccmon/handler/grpc/receiver"
	pb "github.com/elct9620/ccmon/proto"
	"github.com/elct9620/ccmon/usecase"
	logsv1 "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	metricsv1 "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	tracesv1 "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/resolver"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Mock implementations for testing
type mockAPIRequestRepository struct {
	requests []entity.APIRequest
	saveErr  error
	findErr  error
}

func (m *mockAPIRequestRepository) Save(req entity.APIRequest) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.requests = append(m.requests, req)
	return nil
}

func (m *mockAPIRequestRepository) FindByPeriodWithLimit(period entity.Period, limit int, offset int) ([]entity.APIRequest, error) {
	return m.requests, m.findErr
}

func (m *mockAPIRequestRepository) FindAll() ([]entity.APIRequest, error) {
	return m.requests, m.findErr
}

// Test helper to create an in-memory gRPC server
func setupTestServer(t *testing.T) (*grpc.Server, *bufconn.Listener, pb.QueryServiceClient, *mockAPIRequestRepository) {
	// Create bufconn listener
	lis := bufconn.Listen(1024 * 1024)

	// Create mock repository
	mockRepo := &mockAPIRequestRepository{}

	// Create real usecases with mock repository
	appendCommand := usecase.NewAppendApiRequestCommand(mockRepo)
	getFilteredQuery := usecase.NewGetFilteredApiRequestsQuery(mockRepo)
	calculateStatsQuery := usecase.NewCalculateStatsQuery(mockRepo)

	// Create gRPC server and register services (same as RunServer but without lifecycle management)
	grpcServer := grpc.NewServer()

	// Create the OTLP receiver
	otlpReceiver := receiver.NewReceiver(nil, nil, appendCommand)

	// Create the query service
	queryService := query.NewService(getFilteredQuery, calculateStatsQuery)

	// Register OTLP services
	tracesv1.RegisterTraceServiceServer(grpcServer, otlpReceiver.GetTraceServiceServer())
	metricsv1.RegisterMetricsServiceServer(grpcServer, otlpReceiver.GetMetricsServiceServer())
	logsv1.RegisterLogsServiceServer(grpcServer, otlpReceiver.GetLogsServiceServer())

	// Register the query service
	pb.RegisterQueryServiceServer(grpcServer, queryService)

	// Start server in background
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			t.Logf("Server stopped: %v", err)
		}
	}()

	// Set resolver to passthrough for bufconn
	resolver.SetDefaultScheme("passthrough")
	
	// Create client connection
	conn, err := grpc.NewClient("bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("Failed to create client connection: %v", err)
	}

	client := pb.NewQueryServiceClient(conn)

	t.Cleanup(func() {
		conn.Close()
		grpcServer.Stop()
		lis.Close()
	})

	return grpcServer, lis, client, mockRepo
}

func TestGRPCServer_QueryService_GetStats(t *testing.T) {
	_, _, client, mockRepo := setupTestServer(t)

	// Populate mock repository with test data
	now := time.Now()
	
	// Add some base model requests
	for i := 0; i < 10; i++ {
		req := entity.NewAPIRequest(
			"session1", now.Add(-time.Duration(i)*time.Hour),
			"claude-3-haiku-20240307", // base model
			entity.NewToken(100, 50, 10, 5),
			entity.NewCost(0.15),
			1000,
		)
		mockRepo.requests = append(mockRepo.requests, req)
	}
	
	// Add some premium model requests
	for i := 0; i < 5; i++ {
		req := entity.NewAPIRequest(
			"session2", now.Add(-time.Duration(i)*time.Hour),
			"claude-3-sonnet-20240229", // premium model
			entity.NewToken(200, 100, 20, 10),
			entity.NewCost(0.70),
			1500,
		)
		mockRepo.requests = append(mockRepo.requests, req)
	}

	// Make gRPC call
	ctx := context.Background()
	req := &pb.GetStatsRequest{
		StartTime: timestamppb.New(time.Now().Add(-24 * time.Hour)),
		EndTime:   timestamppb.New(time.Now()),
	}

	resp, err := client.GetStats(ctx, req)
	if err != nil {
		t.Fatalf("GetStats failed: %v", err)
	}

	// Verify response
	if resp.Stats == nil {
		t.Fatal("Expected stats in response")
	}

	stats := resp.Stats
	if stats.TotalRequests != 15 {
		t.Errorf("Expected 15 total requests, got %d", stats.TotalRequests)
	}
	if stats.BaseRequests != 10 {
		t.Errorf("Expected 10 base requests, got %d", stats.BaseRequests)
	}
	if stats.PremiumRequests != 5 {
		t.Errorf("Expected 5 premium requests, got %d", stats.PremiumRequests)
	}

	// Verify token data 
	// Base: 10 requests * (100+50+10+5) = 10 * 165 = 1650 tokens
	// Premium: 5 requests * (200+100+20+10) = 5 * 330 = 1650 tokens
	// Total: 1650 + 1650 = 3300 tokens
	expectedTotalTokens := int64(10*165 + 5*330) // 3300
	if stats.TotalTokens.Total != expectedTotalTokens {
		t.Errorf("Expected %d total tokens, got %d", expectedTotalTokens, stats.TotalTokens.Total)
	}

	// Verify cost data (10 * 0.15 + 5 * 0.70 = 5.00)
	expectedTotalCost := 10*0.15 + 5*0.70 // 5.00
	if stats.TotalCost.Amount != expectedTotalCost {
		t.Errorf("Expected $%.2f total cost, got $%.2f", expectedTotalCost, stats.TotalCost.Amount)
	}
}

func TestGRPCServer_QueryService_GetAPIRequests(t *testing.T) {
	_, _, client, mockRepo := setupTestServer(t)

	// Set up mock data in repository
	now := time.Now()
	expectedRequests := []entity.APIRequest{
		entity.NewAPIRequest(
			"session1", now,
			"claude-3-sonnet-20240229",
			entity.NewToken(100, 50, 10, 5),
			entity.NewCost(0.50),
			1500,
		),
		entity.NewAPIRequest(
			"session2", now.Add(-time.Hour),
			"claude-3-haiku-20240307",
			entity.NewToken(200, 100, 20, 10),
			entity.NewCost(0.25),
			800,
		),
	}
	mockRepo.requests = expectedRequests

	// Make gRPC call
	ctx := context.Background()
	req := &pb.GetAPIRequestsRequest{
		StartTime: timestamppb.New(now.Add(-24 * time.Hour)),
		EndTime:   timestamppb.New(now),
		Limit:     10,
		Offset:    0,
	}

	resp, err := client.GetAPIRequests(ctx, req)
	if err != nil {
		t.Fatalf("GetAPIRequests failed: %v", err)
	}

	// Verify response
	if len(resp.Requests) != 2 {
		t.Fatalf("Expected 2 requests, got %d", len(resp.Requests))
	}

	// Verify first request
	req1 := resp.Requests[0]
	if req1.SessionId != "session1" {
		t.Errorf("Expected session1, got %s", req1.SessionId)
	}
	if req1.Model != "claude-3-sonnet-20240229" {
		t.Errorf("Expected claude-3-sonnet-20240229, got %s", req1.Model)
	}
	if req1.TotalTokens != 165 { // 100+50+10+5
		t.Errorf("Expected 165 total tokens, got %d", req1.TotalTokens)
	}
	if req1.CostUsd != 0.50 {
		t.Errorf("Expected $0.50 cost, got $%.2f", req1.CostUsd)
	}

	// Verify second request
	req2 := resp.Requests[1]
	if req2.SessionId != "session2" {
		t.Errorf("Expected session2, got %s", req2.SessionId)
	}
	if req2.Model != "claude-3-haiku-20240307" {
		t.Errorf("Expected claude-3-haiku-20240307, got %s", req2.Model)
	}
}

func TestGRPCServer_QueryService_AllTimeRequests(t *testing.T) {
	_, _, client, mockRepo := setupTestServer(t)

	// Set up mock data with empty requests
	mockRepo.requests = []entity.APIRequest{}

	// Make gRPC call with nil timestamps (all time)
	ctx := context.Background()
	req := &pb.GetAPIRequestsRequest{
		StartTime: nil, // All time from beginning
		EndTime:   nil, // All time to current
		Limit:     100,
		Offset:    0,
	}

	resp, err := client.GetAPIRequests(ctx, req)
	if err != nil {
		t.Fatalf("GetAPIRequests failed: %v", err)
	}

	// Verify empty response
	if len(resp.Requests) != 0 {
		t.Errorf("Expected 0 requests, got %d", len(resp.Requests))
	}
	if resp.TotalCount != 0 {
		t.Errorf("Expected 0 total count, got %d", resp.TotalCount)
	}
}

func TestGRPCServer_ServiceRegistration(t *testing.T) {
	grpcServer, _, _, _ := setupTestServer(t)

	// Verify server has services registered
	services := grpcServer.GetServiceInfo()

	// Check if QueryService is registered
	if _, exists := services["ccmon.v1.QueryService"]; !exists {
		t.Error("QueryService not registered")
	}

	// Check if OTLP services are registered
	if _, exists := services["opentelemetry.proto.collector.logs.v1.LogsService"]; !exists {
		t.Error("LogsService not registered")
	}
	if _, exists := services["opentelemetry.proto.collector.trace.v1.TraceService"]; !exists {
		t.Error("TraceService not registered")
	}
	if _, exists := services["opentelemetry.proto.collector.metrics.v1.MetricsService"]; !exists {
		t.Error("MetricsService not registered")
	}
}