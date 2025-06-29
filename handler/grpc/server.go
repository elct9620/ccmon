package grpc

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/elct9620/ccmon/handler/grpc/query"
	"github.com/elct9620/ccmon/handler/grpc/receiver"
	pb "github.com/elct9620/ccmon/proto"
	"github.com/elct9620/ccmon/usecase"
	logsv1 "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	metricsv1 "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	tracesv1 "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	"google.golang.org/grpc"
)

// RunServer runs the headless OTLP server mode
func RunServer(address string, appendCommand *usecase.AppendApiRequestCommand, getFilteredQuery *usecase.GetFilteredApiRequestsQuery, calculateStatsQuery *usecase.CalculateStatsQuery) error {
	log.Println("Starting ccmon in server mode...")

	// Create the OTLP receiver
	otlpReceiver := receiver.NewReceiver(nil, nil, appendCommand) // No channel or TUI program needed

	// Create the query service
	queryService := query.NewService(getFilteredQuery, calculateStatsQuery)

	// Set up gRPC server
	lis, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	grpcServer := grpc.NewServer()

	// Register the OTLP services
	tracesv1.RegisterTraceServiceServer(grpcServer, otlpReceiver.GetTraceServiceServer())
	metricsv1.RegisterMetricsServiceServer(grpcServer, otlpReceiver.GetMetricsServiceServer())
	logsv1.RegisterLogsServiceServer(grpcServer, otlpReceiver.GetLogsServiceServer())

	// Register the query service
	pb.RegisterQueryServiceServer(grpcServer, queryService)

	// Create a context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("Shutting down server...")
		cancel()
	}()

	// Handle graceful shutdown
	go func() {
		<-ctx.Done()
		grpcServer.GracefulStop()
	}()

	// Start the gRPC server
	log.Printf("gRPC server (OTLP + Query) listening on %s\n", address)
	if err := grpcServer.Serve(lis); err != nil {
		return fmt.Errorf("failed to start gRPC server: %w", err)
	}
	log.Println("Server stopped")
	return nil
}
