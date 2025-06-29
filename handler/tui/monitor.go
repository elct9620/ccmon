package tui

import (
	"context"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/elct9620/ccmon/entity"
	pb "github.com/elct9620/ccmon/proto"
	"github.com/elct9620/ccmon/usecase"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

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
func (qc *QueryClient) GetAPIRequests(period entity.Period) ([]entity.APIRequest, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Convert entity.Period to proto TimeFilter
	var protoFilter pb.TimeFilter
	if period.IsAllTime() {
		protoFilter = pb.TimeFilter_TIME_FILTER_ALL
	} else {
		// Calculate duration to determine filter type
		duration := period.Duration()
		if duration <= time.Hour {
			protoFilter = pb.TimeFilter_TIME_FILTER_HOUR
		} else if duration <= 24*time.Hour {
			protoFilter = pb.TimeFilter_TIME_FILTER_DAY
		} else if duration <= 7*24*time.Hour {
			protoFilter = pb.TimeFilter_TIME_FILTER_WEEK
		} else if duration <= 30*24*time.Hour {
			protoFilter = pb.TimeFilter_TIME_FILTER_MONTH
		} else {
			protoFilter = pb.TimeFilter_TIME_FILTER_ALL
		}
	}

	req := &pb.GetAPIRequestsRequest{
		TimeFilter: protoFilter,
	}

	resp, err := qc.client.GetAPIRequests(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get API requests: %w", err)
	}

	// Convert proto responses to entity.APIRequest
	requests := make([]entity.APIRequest, len(resp.Requests))
	for i, protoReq := range resp.Requests {
		tokens := entity.NewToken(
			protoReq.InputTokens,
			protoReq.OutputTokens,
			protoReq.CacheReadTokens,
			protoReq.CacheCreationTokens,
		)
		cost := entity.NewCost(protoReq.CostUsd)
		requests[i] = entity.NewAPIRequest(
			protoReq.SessionId,
			protoReq.Timestamp.AsTime(),
			protoReq.Model,
			tokens,
			cost,
			protoReq.DurationMs,
		)
	}

	return requests, nil
}

// Close closes the gRPC connection
func (qc *QueryClient) Close() error {
	return qc.conn.Close()
}

// RunMonitor runs the TUI monitor mode with usecase
func RunMonitor(getFilteredQuery *usecase.GetFilteredApiRequestsQuery) error {
	// Create the Bubble Tea model
	model := NewModel(getFilteredQuery)

	// Create and run the Bubble Tea program
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running TUI: %w", err)
	}

	return nil
}
