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

	dbpkg "github.com/elct9620/ccmon/db"
	"github.com/elct9620/ccmon/handler/grpc/receiver"
	logsv1 "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	metricsv1 "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	tracesv1 "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	"google.golang.org/grpc"
)

// Database interface to avoid circular dependency
type Database interface {
	SaveAPIRequest(req dbpkg.APIRequest) error
	GetAllRequests() ([]dbpkg.APIRequest, error)
	Close() error
}

// RunServer runs the headless OTLP server mode
func RunServer(address string, newDB func() (Database, error)) error {
	log.Println("Starting ccmon in server mode...")

	// Initialize database
	db, err := newDB()
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer db.Close()

	// Create the OTLP receiver
	otlpReceiver := receiver.NewReceiver(nil, nil, db) // No channel or TUI program needed

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

	// Start request counter
	go logRequestStats(ctx, db)

	// Handle graceful shutdown
	go func() {
		<-ctx.Done()
		grpcServer.GracefulStop()
	}()

	// Start the gRPC server
	log.Printf("OTLP receiver listening on %s\n", address)
	if err := grpcServer.Serve(lis); err != nil {
		return fmt.Errorf("failed to start gRPC server: %w", err)
	}
	log.Println("Server stopped")
	return nil
}

// logRequestStats periodically logs request statistics
func logRequestStats(ctx context.Context, db Database) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Get all requests to calculate stats
			requests, err := db.GetAllRequests()
			if err != nil {
				log.Printf("Error reading stats: %v", err)
				continue
			}

			// Calculate stats
			baseReqs, premiumReqs, baseTokens, premiumTokens, _, _, _, _, baseCost, premiumCost := dbpkg.CalculateStats(requests)
			totalReqs := baseReqs + premiumReqs
			totalTokens := baseTokens + premiumTokens
			totalCost := baseCost + premiumCost

			log.Printf("Stats - Requests: %d (Base: %d, Premium: %d) | Tokens: %d | Cost: $%.6f",
				totalReqs, baseReqs, premiumReqs, totalTokens, totalCost)
		}
	}
}