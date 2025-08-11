package auth

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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/resolver"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Mock repository for testing
type mockAuthTestRepository struct {
	requests []entity.APIRequest
}

func (m *mockAuthTestRepository) Save(req entity.APIRequest) error {
	m.requests = append(m.requests, req)
	return nil
}

func (m *mockAuthTestRepository) FindByPeriodWithLimit(period entity.Period, limit int, offset int) ([]entity.APIRequest, error) {
	return m.requests, nil
}

func (m *mockAuthTestRepository) FindAll() ([]entity.APIRequest, error) {
	return m.requests, nil
}

func (m *mockAuthTestRepository) DeleteOlderThan(cutoffTime time.Time) (int, error) {
	return 0, nil
}

// Mock stats repository for testing
type mockAuthStatsRepository struct {
	repo *mockAuthTestRepository
}

func (m *mockAuthStatsRepository) GetStatsByPeriod(period entity.Period) (entity.Stats, error) {
	// Calculate stats from the requests in the repository
	requests := m.repo.requests
	baseTokens := entity.NewToken(0, 0, 0, 0)
	premiumTokens := entity.NewToken(0, 0, 0, 0)
	baseCost := entity.NewCost(0)
	premiumCost := entity.NewCost(0)
	baseRequests := 0
	premiumRequests := 0

	for _, req := range requests {
		if req.Model().IsBase() {
			baseRequests++
			baseTokens = baseTokens.Add(req.Tokens())
			baseCost = baseCost.Add(req.Cost())
		} else {
			premiumRequests++
			premiumTokens = premiumTokens.Add(req.Tokens())
			premiumCost = premiumCost.Add(req.Cost())
		}
	}

	return entity.NewStats(baseRequests, premiumRequests, baseTokens, premiumTokens, baseCost, premiumCost, period), nil
}

// Mock stats cache for testing
type mockStatsCache struct{}

func (m *mockStatsCache) Get(period entity.Period) *entity.Stats {
	return nil
}

func (m *mockStatsCache) Set(period entity.Period, stats *entity.Stats) {}

// Test helper to create gRPC server with auth
func setupAuthTestServer(t *testing.T, authToken string) (*grpc.Server, *bufconn.Listener, pb.QueryServiceClient, *mockAuthTestRepository) {
	lis := bufconn.Listen(1024 * 1024)
	mockRepo := &mockAuthTestRepository{}

	// Create usecases
	appendCommand := usecase.NewAppendApiRequestCommand(mockRepo)
	getFilteredQuery := usecase.NewGetFilteredApiRequestsQuery(mockRepo)
	mockStatsRepo := &mockAuthStatsRepository{repo: mockRepo}
	mockCache := &mockStatsCache{}
	calculateStatsQuery := usecase.NewCalculateStatsQuery(mockStatsRepo, mockCache)

	// Configure server with auth interceptors
	opts := []grpc.ServerOption{}
	if authToken != "" {
		opts = append(opts,
			grpc.UnaryInterceptor(UnaryServerInterceptor(authToken)),
			grpc.StreamInterceptor(StreamServerInterceptor(authToken)),
		)
	}

	grpcServer := grpc.NewServer(opts...)

	// Register services
	otlpReceiver := receiver.NewReceiver(nil, nil, appendCommand)
	queryService := query.NewService(getFilteredQuery, calculateStatsQuery)

	tracesv1.RegisterTraceServiceServer(grpcServer, otlpReceiver.GetTraceServiceServer())
	metricsv1.RegisterMetricsServiceServer(grpcServer, otlpReceiver.GetMetricsServiceServer())
	logsv1.RegisterLogsServiceServer(grpcServer, otlpReceiver.GetLogsServiceServer())
	pb.RegisterQueryServiceServer(grpcServer, queryService)

	// Start server
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			t.Logf("Server stopped: %v", err)
		}
	}()

	// Set resolver to passthrough for bufconn
	resolver.SetDefaultScheme("passthrough")

	// Create client connection with auth
	clientOpts := []grpc.DialOption{
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	if authToken != "" {
		clientOpts = append(clientOpts, grpc.WithUnaryInterceptor(ClientInterceptor(authToken)))
	}

	conn, err := grpc.NewClient("bufnet", clientOpts...)
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

func TestAuth_NoAuthentication(t *testing.T) {
	t.Parallel()

	// Setup server without auth
	_, _, client, mockRepo := setupAuthTestServer(t, "")

	// Add test data
	req := entity.NewAPIRequest(
		"session1",
		time.Now(),
		"claude-3-sonnet-20240229",
		entity.NewToken(100, 50, 10, 5),
		entity.NewCost(0.50),
		1500,
	)
	mockRepo.requests = []entity.APIRequest{req}

	// Make request (should work without auth)
	ctx := context.Background()
	request := &pb.GetStatsRequest{
		StartTime: timestamppb.New(time.Now().Add(-24 * time.Hour)),
		EndTime:   timestamppb.New(time.Now()),
	}

	resp, err := client.GetStats(ctx, request)
	if err != nil {
		t.Fatalf("Expected request to succeed without auth, got error: %v", err)
	}

	if resp.Stats.TotalRequests != 1 {
		t.Errorf("Expected 1 request, got %d", resp.Stats.TotalRequests)
	}
}

func TestAuth_ValidAuthentication(t *testing.T) {
	t.Parallel()

	authToken := "Bearer test-secret-token"

	// Setup server with auth
	_, _, client, mockRepo := setupAuthTestServer(t, authToken)

	// Add test data
	req := entity.NewAPIRequest(
		"session1",
		time.Now(),
		"claude-3-sonnet-20240229",
		entity.NewToken(100, 50, 10, 5),
		entity.NewCost(0.50),
		1500,
	)
	mockRepo.requests = []entity.APIRequest{req}

	// Make request with matching auth token (should work)
	ctx := context.Background()
	request := &pb.GetStatsRequest{
		StartTime: timestamppb.New(time.Now().Add(-24 * time.Hour)),
		EndTime:   timestamppb.New(time.Now()),
	}

	resp, err := client.GetStats(ctx, request)
	if err != nil {
		t.Fatalf("Expected request to succeed with valid auth, got error: %v", err)
	}

	if resp.Stats.TotalRequests != 1 {
		t.Errorf("Expected 1 request, got %d", resp.Stats.TotalRequests)
	}
}

func TestAuth_InvalidAuthentication(t *testing.T) {
	t.Parallel()

	serverToken := "Bearer server-secret-token"
	clientToken := "Bearer wrong-token"

	// Setup server with one token
	lis := bufconn.Listen(1024 * 1024)
	mockRepo := &mockAuthTestRepository{}

	appendCommand := usecase.NewAppendApiRequestCommand(mockRepo)
	getFilteredQuery := usecase.NewGetFilteredApiRequestsQuery(mockRepo)
	mockStatsRepo := &mockAuthStatsRepository{repo: mockRepo}
	mockCache := &mockStatsCache{}
	calculateStatsQuery := usecase.NewCalculateStatsQuery(mockStatsRepo, mockCache)

	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(UnaryServerInterceptor(serverToken)),
		grpc.StreamInterceptor(StreamServerInterceptor(serverToken)),
	}

	grpcServer := grpc.NewServer(opts...)

	otlpReceiver := receiver.NewReceiver(nil, nil, appendCommand)
	queryService := query.NewService(getFilteredQuery, calculateStatsQuery)

	tracesv1.RegisterTraceServiceServer(grpcServer, otlpReceiver.GetTraceServiceServer())
	metricsv1.RegisterMetricsServiceServer(grpcServer, otlpReceiver.GetMetricsServiceServer())
	logsv1.RegisterLogsServiceServer(grpcServer, otlpReceiver.GetLogsServiceServer())
	pb.RegisterQueryServiceServer(grpcServer, queryService)

	go func() {
		grpcServer.Serve(lis)
	}()

	// Set resolver to passthrough for bufconn
	resolver.SetDefaultScheme("passthrough")

	// Create client with different token
	clientOpts := []grpc.DialOption{
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(ClientInterceptor(clientToken)),
	}

	conn, err := grpc.NewClient("bufnet", clientOpts...)
	if err != nil {
		t.Fatalf("Failed to create client connection: %v", err)
	}
	defer conn.Close()
	defer grpcServer.Stop()
	defer lis.Close()

	client := pb.NewQueryServiceClient(conn)

	// Make request with wrong token (should fail)
	ctx := context.Background()
	request := &pb.GetStatsRequest{
		StartTime: timestamppb.New(time.Now().Add(-24 * time.Hour)),
		EndTime:   timestamppb.New(time.Now()),
	}

	_, err = client.GetStats(ctx, request)
	if err == nil {
		t.Fatal("Expected request to fail with invalid auth, but it succeeded")
	}

	// Check that it's an authentication error
	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("Expected gRPC status error")
	}

	if st.Code() != codes.Unauthenticated {
		t.Errorf("Expected Unauthenticated error, got %v", st.Code())
	}
}

func TestAuth_MissingAuthenticationHeader(t *testing.T) {
	t.Parallel()

	authToken := "Bearer test-secret-token"

	// Setup server with auth
	lis := bufconn.Listen(1024 * 1024)
	mockRepo := &mockAuthTestRepository{}

	appendCommand := usecase.NewAppendApiRequestCommand(mockRepo)
	getFilteredQuery := usecase.NewGetFilteredApiRequestsQuery(mockRepo)
	mockStatsRepo := &mockAuthStatsRepository{repo: mockRepo}
	mockCache := &mockStatsCache{}
	calculateStatsQuery := usecase.NewCalculateStatsQuery(mockStatsRepo, mockCache)

	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(UnaryServerInterceptor(authToken)),
		grpc.StreamInterceptor(StreamServerInterceptor(authToken)),
	}

	grpcServer := grpc.NewServer(opts...)

	otlpReceiver := receiver.NewReceiver(nil, nil, appendCommand)
	queryService := query.NewService(getFilteredQuery, calculateStatsQuery)

	tracesv1.RegisterTraceServiceServer(grpcServer, otlpReceiver.GetTraceServiceServer())
	metricsv1.RegisterMetricsServiceServer(grpcServer, otlpReceiver.GetMetricsServiceServer())
	logsv1.RegisterLogsServiceServer(grpcServer, otlpReceiver.GetLogsServiceServer())
	pb.RegisterQueryServiceServer(grpcServer, queryService)

	go func() {
		grpcServer.Serve(lis)
	}()

	// Set resolver to passthrough for bufconn
	resolver.SetDefaultScheme("passthrough")

	// Create client without auth interceptor
	clientOpts := []grpc.DialOption{
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	conn, err := grpc.NewClient("bufnet", clientOpts...)
	if err != nil {
		t.Fatalf("Failed to create client connection: %v", err)
	}
	defer conn.Close()
	defer grpcServer.Stop()
	defer lis.Close()

	client := pb.NewQueryServiceClient(conn)

	// Make request without auth header (should fail)
	ctx := context.Background()
	request := &pb.GetStatsRequest{
		StartTime: timestamppb.New(time.Now().Add(-24 * time.Hour)),
		EndTime:   timestamppb.New(time.Now()),
	}

	_, err = client.GetStats(ctx, request)
	if err == nil {
		t.Fatal("Expected request to fail without auth header, but it succeeded")
	}

	// Check that it's an authentication error
	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("Expected gRPC status error")
	}

	if st.Code() != codes.Unauthenticated {
		t.Errorf("Expected Unauthenticated error, got %v", st.Code())
	}
}
