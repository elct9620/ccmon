package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	tracesv1 "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	metricsv1 "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	logsv1 "go.opentelemetry.io/proto/otlp/collector/logs/v1"
)

var rootCmd = &cobra.Command{
	Use:   "ccmon",
	Short: "Claude Code usage monitoring with OTEL collector",
	Long:  "A CLI tool to collect and display Claude Code usage data via OpenTelemetry",
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the OTEL gRPC receiver to receive Claude Code telemetry",
	Run: func(cmd *cobra.Command, args []string) {
		startReceiver()
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

type traceReceiver struct {
	tracesv1.UnimplementedTraceServiceServer
}

func (r *traceReceiver) Export(ctx context.Context, req *tracesv1.ExportTraceServiceRequest) (*tracesv1.ExportTraceServiceResponse, error) {
	fmt.Printf("Received trace data: %d resource spans\n", len(req.ResourceSpans))
	for _, rs := range req.ResourceSpans {
		fmt.Printf("  Resource: %v\n", rs.Resource)
		for _, ss := range rs.ScopeSpans {
			fmt.Printf("    Scope: %v, Spans: %d\n", ss.Scope, len(ss.Spans))
		}
	}
	return &tracesv1.ExportTraceServiceResponse{}, nil
}

type metricsReceiver struct {
	metricsv1.UnimplementedMetricsServiceServer
}

func (r *metricsReceiver) Export(ctx context.Context, req *metricsv1.ExportMetricsServiceRequest) (*metricsv1.ExportMetricsServiceResponse, error) {
	fmt.Printf("Received metrics data: %d resource metrics\n", len(req.ResourceMetrics))
	for _, rm := range req.ResourceMetrics {
		fmt.Printf("  Resource: %v\n", rm.Resource)
		for _, sm := range rm.ScopeMetrics {
			fmt.Printf("    Scope: %v, Metrics: %d\n", sm.Scope, len(sm.Metrics))
		}
	}
	return &metricsv1.ExportMetricsServiceResponse{}, nil
}

type logsReceiver struct {
	logsv1.UnimplementedLogsServiceServer
}

func (r *logsReceiver) Export(ctx context.Context, req *logsv1.ExportLogsServiceRequest) (*logsv1.ExportLogsServiceResponse, error) {
	fmt.Printf("Received logs data: %d resource logs\n", len(req.ResourceLogs))
	for _, rl := range req.ResourceLogs {
		fmt.Printf("  Resource: %v\n", rl.Resource)
		for _, sl := range rl.ScopeLogs {
			fmt.Printf("    Scope: %v, Logs: %d\n", sl.Scope, len(sl.LogRecords))
		}
	}
	return &logsv1.ExportLogsServiceResponse{}, nil
}

func startReceiver() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	lis, err := net.Listen("tcp", ":4317")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	
	tracesv1.RegisterTraceServiceServer(s, &traceReceiver{})
	metricsv1.RegisterMetricsServiceServer(s, &metricsReceiver{})
	logsv1.RegisterLogsServiceServer(s, &logsReceiver{})

	go func() {
		<-ctx.Done()
		fmt.Println("\nShutting down server...")
		s.GracefulStop()
	}()

	fmt.Println("Starting OTEL gRPC receiver on port 4317...")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}