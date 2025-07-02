package grpc

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/elct9620/ccmon/handler/grpc/query"
	"github.com/elct9620/ccmon/handler/grpc/receiver"
	pb "github.com/elct9620/ccmon/proto"
	"github.com/elct9620/ccmon/usecase"
	logsv1 "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	metricsv1 "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	tracesv1 "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	"google.golang.org/grpc"
)

// ServerConfig interface to avoid import cycle
type ServerConfig interface {
	IsRetentionEnabled() bool
	GetRetentionDuration() time.Duration
}

// RunServer runs the headless OTLP server mode
func RunServer(address string, appendCommand *usecase.AppendApiRequestCommand, getFilteredQuery *usecase.GetFilteredApiRequestsQuery, calculateStatsQuery *usecase.CalculateStatsQuery, cleanupCommand *usecase.CleanupOldRecordsCommand, serverConfig ServerConfig) error {
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

	// Start cleanup scheduler if retention is enabled
	if serverConfig.IsRetentionEnabled() {
		startCleanupScheduler(ctx, cleanupCommand, serverConfig)
	}

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

// startCleanupScheduler starts a background cleanup scheduler
func startCleanupScheduler(ctx context.Context, cleanupCommand *usecase.CleanupOldRecordsCommand, serverConfig ServerConfig) {
	retentionDuration := serverConfig.GetRetentionDuration()
	cleanupInterval := 6 * time.Hour // Run cleanup every 6 hours

	log.Printf("Starting cleanup scheduler: retention=%v, interval=%v", retentionDuration, cleanupInterval)

	go func() {
		ticker := time.NewTicker(cleanupInterval)
		defer ticker.Stop()

		// Run initial cleanup
		runCleanup(ctx, cleanupCommand, retentionDuration)

		for {
			select {
			case <-ctx.Done():
				log.Println("Cleanup scheduler stopped")
				return
			case <-ticker.C:
				runCleanup(ctx, cleanupCommand, retentionDuration)
			}
		}
	}()
}

// runCleanup performs a single cleanup operation
func runCleanup(ctx context.Context, cleanupCommand *usecase.CleanupOldRecordsCommand, retentionDuration time.Duration) {
	cutoffTime := time.Now().Add(-retentionDuration)

	log.Printf("Running cleanup: deleting records older than %v", cutoffTime)

	params := usecase.CleanupOldRecordsParams{
		CutoffTime: cutoffTime,
	}

	result, err := cleanupCommand.Execute(ctx, params)
	if err != nil {
		log.Printf("Cleanup failed: %v", err)
		return
	}

	if result.DeletedCount > 0 {
		log.Printf("Cleanup completed: deleted %d records", result.DeletedCount)
	} else {
		log.Printf("Cleanup completed: no records to delete")
	}
}
