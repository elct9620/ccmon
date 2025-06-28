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
	logsv1 "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	metricsv1 "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	tracesv1 "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	commonv1 "go.opentelemetry.io/proto/otlp/common/v1"
	"google.golang.org/grpc"
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
	// Silently ignore traces - we're focusing on logs for token usage
	return &tracesv1.ExportTraceServiceResponse{}, nil
}

type metricsReceiver struct {
	metricsv1.UnimplementedMetricsServiceServer
}

func (r *metricsReceiver) Export(ctx context.Context, req *metricsv1.ExportMetricsServiceRequest) (*metricsv1.ExportMetricsServiceResponse, error) {
	// Silently ignore metrics - we're focusing on logs for token usage
	return &metricsv1.ExportMetricsServiceResponse{}, nil
}

type logsReceiver struct {
	logsv1.UnimplementedLogsServiceServer
}

func (r *logsReceiver) Export(ctx context.Context, req *logsv1.ExportLogsServiceRequest) (*logsv1.ExportLogsServiceResponse, error) {
	for _, rl := range req.ResourceLogs {
		for _, sl := range rl.ScopeLogs {
			for _, log := range sl.LogRecords {
				// Check if this is an API request log
				if body, ok := log.Body.Value.(*commonv1.AnyValue_StringValue); ok && body.StringValue == "claude_code.api_request" {
					fmt.Printf("\n=== Claude Code API Request ===\n")

					// Extract attributes we care about
					var sessionID, timestamp, model string
					var inputTokens, outputTokens, cacheReadTokens, cacheCreationTokens int64
					var costUSD float64
					var durationMS int64

					// Debug mode - uncomment to see all attributes
					// fmt.Println("DEBUG: All attributes:")
					// for _, attr := range log.Attributes {
					// 	fmt.Printf("  %s = %v (type: %T)\n", attr.Key, attr.Value.Value, attr.Value.Value)
					// }

					for _, attr := range log.Attributes {
						switch attr.Key {
						case "session.id":
							if v, ok := attr.Value.Value.(*commonv1.AnyValue_StringValue); ok {
								sessionID = v.StringValue
							}
						case "event.timestamp":
							if v, ok := attr.Value.Value.(*commonv1.AnyValue_StringValue); ok {
								timestamp = v.StringValue
							}
						case "model":
							if v, ok := attr.Value.Value.(*commonv1.AnyValue_StringValue); ok {
								model = v.StringValue
							}
						case "input_tokens":
							if v, ok := attr.Value.Value.(*commonv1.AnyValue_StringValue); ok {
								fmt.Sscanf(v.StringValue, "%d", &inputTokens)
							}
						case "output_tokens":
							if v, ok := attr.Value.Value.(*commonv1.AnyValue_StringValue); ok {
								fmt.Sscanf(v.StringValue, "%d", &outputTokens)
							}
						case "cache_read_tokens":
							if v, ok := attr.Value.Value.(*commonv1.AnyValue_StringValue); ok {
								fmt.Sscanf(v.StringValue, "%d", &cacheReadTokens)
							}
						case "cache_creation_tokens":
							if v, ok := attr.Value.Value.(*commonv1.AnyValue_StringValue); ok {
								fmt.Sscanf(v.StringValue, "%d", &cacheCreationTokens)
							}
						case "cost_usd":
							if v, ok := attr.Value.Value.(*commonv1.AnyValue_StringValue); ok {
								fmt.Sscanf(v.StringValue, "%f", &costUSD)
							}
						case "duration_ms":
							if v, ok := attr.Value.Value.(*commonv1.AnyValue_StringValue); ok {
								fmt.Sscanf(v.StringValue, "%d", &durationMS)
							}
						}
					}

					// Display the extracted information
					fmt.Printf("Timestamp:    %s\n", timestamp)
					if sessionID != "" {
						fmt.Printf("Session ID:   %s\n", sessionID)
					}
					fmt.Printf("Model:        %s\n", model)
					fmt.Printf("Duration:     %dms\n", durationMS)
					fmt.Printf("\nToken Usage:\n")
					fmt.Printf("  Input:          %d\n", inputTokens)
					fmt.Printf("  Output:         %d\n", outputTokens)
					fmt.Printf("  Cache Read:     %d\n", cacheReadTokens)
					fmt.Printf("  Cache Creation: %d\n", cacheCreationTokens)
					fmt.Printf("  Total:          %d\n", inputTokens+outputTokens+cacheReadTokens+cacheCreationTokens)
					fmt.Printf("\nCost: $%.8f\n", costUSD)
					fmt.Printf("==============================\n")
				}
			}
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

	fmt.Println("Starting Claude Code monitor on port 4317...")
	fmt.Println("Tracking API requests and token usage...")
	fmt.Println("\nTo enable Claude Code telemetry, set these environment variables:")
	fmt.Println("  export CLAUDE_CODE_ENABLE_TELEMETRY=1")
	fmt.Println("  export OTEL_METRICS_EXPORTER=otlp")
	fmt.Println("  export OTEL_LOGS_EXPORTER=otlp")
	fmt.Println("  export OTEL_EXPORTER_OTLP_PROTOCOL=grpc")
	fmt.Println("  export OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317")

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
