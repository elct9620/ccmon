package tui

import (
	"context"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/elct9620/ccmon/db"
	pb "github.com/elct9620/ccmon/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Database interface to avoid circular dependency
type Database interface {
	GetAPIRequests(filter db.Filter) ([]db.APIRequest, error)
	Close() error
}

// QueryClient wraps the gRPC client for querying data
type QueryClient struct {
	client pb.QueryServiceClient
	conn   *grpc.ClientConn
}

// NewQueryClient creates a new query client
func NewQueryClient(address string) (*QueryClient, error) {
	// Create connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, address, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC server: %w", err)
	}

	client := pb.NewQueryServiceClient(conn)
	return &QueryClient{
		client: client,
		conn:   conn,
	}, nil
}

// GetAPIRequests implements the Database interface using gRPC
func (qc *QueryClient) GetAPIRequests(filter db.Filter) ([]db.APIRequest, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Convert db.TimeFilter to proto TimeFilter
	var protoFilter pb.TimeFilter
	switch filter.TimeFilter {
	case db.FilterAll:
		protoFilter = pb.TimeFilter_TIME_FILTER_ALL
	case db.FilterHour:
		protoFilter = pb.TimeFilter_TIME_FILTER_HOUR
	case db.FilterDay:
		protoFilter = pb.TimeFilter_TIME_FILTER_DAY
	case db.FilterWeek:
		protoFilter = pb.TimeFilter_TIME_FILTER_WEEK
	case db.FilterMonth:
		protoFilter = pb.TimeFilter_TIME_FILTER_MONTH
	default:
		protoFilter = pb.TimeFilter_TIME_FILTER_ALL
	}

	req := &pb.GetAPIRequestsRequest{
		TimeFilter: protoFilter,
	}

	resp, err := qc.client.GetAPIRequests(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get API requests: %w", err)
	}

	// Convert proto responses to db.APIRequest
	requests := make([]db.APIRequest, len(resp.Requests))
	for i, protoReq := range resp.Requests {
		requests[i] = db.APIRequest{
			SessionID:           protoReq.SessionId,
			Timestamp:           protoReq.Timestamp.AsTime(),
			Model:               protoReq.Model,
			InputTokens:         protoReq.InputTokens,
			OutputTokens:        protoReq.OutputTokens,
			CacheReadTokens:     protoReq.CacheReadTokens,
			CacheCreationTokens: protoReq.CacheCreationTokens,
			TotalTokens:         protoReq.TotalTokens,
			CostUSD:             protoReq.CostUsd,
			DurationMS:          protoReq.DurationMs,
		}
	}

	return requests, nil
}

// Close closes the gRPC connection
func (qc *QueryClient) Close() error {
	return qc.conn.Close()
}

// RunMonitor runs the TUI monitor mode with gRPC client
func RunMonitor(serverAddress string) error {
	// Create gRPC client
	client, err := NewQueryClient(serverAddress)
	if err != nil {
		return fmt.Errorf("failed to create query client: %w", err)
	}
	defer client.Close()

	// Create the Bubble Tea model
	model := NewModel(client)

	// Create and run the Bubble Tea program
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running TUI: %w", err)
	}

	return nil
}
