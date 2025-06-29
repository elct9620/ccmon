package receiver

import (
	"context"
	"fmt"
	"log"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/elct9620/ccmon/db"
	"github.com/elct9620/ccmon/entity"
	logsv1 "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	metricsv1 "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	tracesv1 "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	commonv1 "go.opentelemetry.io/proto/otlp/common/v1"
	logsdata "go.opentelemetry.io/proto/otlp/logs/v1"
)

// Receiver handles OTLP message processing
type Receiver struct {
	requestChan chan db.APIRequest
	program     *tea.Program
	db          Database
}

// Database interface to avoid circular dependency
type Database interface {
	SaveAPIRequest(req db.APIRequest) error
	GetAllRequests() ([]db.APIRequest, error)
	Close() error
}

// NewReceiver creates a new OTLP receiver
func NewReceiver(requestChan chan db.APIRequest, program *tea.Program, database Database) *Receiver {
	return &Receiver{
		requestChan: requestChan,
		program:     program,
		db:          database,
	}
}

// GetTraceServiceServer returns the trace service implementation
func (r *Receiver) GetTraceServiceServer() tracesv1.TraceServiceServer {
	return &traceReceiver{}
}

// GetMetricsServiceServer returns the metrics service implementation
func (r *Receiver) GetMetricsServiceServer() metricsv1.MetricsServiceServer {
	return &metricsReceiver{}
}

// GetLogsServiceServer returns the logs service implementation
func (r *Receiver) GetLogsServiceServer() logsv1.LogsServiceServer {
	return &logsReceiver{receiver: r}
}

// traceReceiver handles trace exports (ignored)
type traceReceiver struct {
	tracesv1.UnimplementedTraceServiceServer
}

func (r *traceReceiver) Export(ctx context.Context, req *tracesv1.ExportTraceServiceRequest) (*tracesv1.ExportTraceServiceResponse, error) {
	// Silently ignore traces - we're focusing on logs for token usage
	return &tracesv1.ExportTraceServiceResponse{}, nil
}

// metricsReceiver handles metrics exports (ignored)
type metricsReceiver struct {
	metricsv1.UnimplementedMetricsServiceServer
}

func (r *metricsReceiver) Export(ctx context.Context, req *metricsv1.ExportMetricsServiceRequest) (*metricsv1.ExportMetricsServiceResponse, error) {
	// Silently ignore metrics - we're focusing on logs for token usage
	return &metricsv1.ExportMetricsServiceResponse{}, nil
}

// logsReceiver handles log exports
type logsReceiver struct {
	logsv1.UnimplementedLogsServiceServer
	receiver *Receiver
}

func (r *logsReceiver) Export(ctx context.Context, req *logsv1.ExportLogsServiceRequest) (*logsv1.ExportLogsServiceResponse, error) {
	for _, rl := range req.ResourceLogs {
		for _, sl := range rl.ScopeLogs {
			for _, logRecord := range sl.LogRecords {
				// Check if this is an API request log
				if body, ok := logRecord.Body.Value.(*commonv1.AnyValue_StringValue); ok && body.StringValue == "claude_code.api_request" {
					apiReq := r.parseAPIRequest(logRecord)
					if apiReq != nil {
						// Save to database
						if r.receiver.db != nil {
							dbReq := db.FromEntity(*apiReq)
							if err := r.receiver.db.SaveAPIRequest(dbReq); err != nil {
								log.Printf("Failed to save request to database: %v", err)
							} else {
								// Log the request in server mode
								if r.receiver.requestChan == nil && r.receiver.program == nil {
									log.Printf("Request received - Model: %s | Limited tokens: %d | Cache tokens: %d | Cost: $%.6f",
										apiReq.Model(), apiReq.Tokens().Limited(),
										apiReq.Tokens().Cache(),
										apiReq.Cost().Amount())
								}
							}
						}

						// Send to channel (non-blocking) - only used in old architecture
						if r.receiver.requestChan != nil {
							dbReq := db.FromEntity(*apiReq)
							select {
							case r.receiver.requestChan <- dbReq:
							default:
								// Channel is full, drop the request
							}
						}
					}
				}
			}
		}
	}

	return &logsv1.ExportLogsServiceResponse{}, nil
}

// parseAPIRequest extracts API request data from a log record
func (r *logsReceiver) parseAPIRequest(logRecord *logsdata.LogRecord) *entity.APIRequest {
	var sessionID, timestampStr, model string
	var inputTokens, outputTokens, cacheReadTokens, cacheCreationTokens int64
	var costUSD float64
	var durationMS int64

	for _, attr := range logRecord.Attributes {
		switch attr.Key {
		case "session.id":
			if v, ok := attr.Value.Value.(*commonv1.AnyValue_StringValue); ok {
				sessionID = v.StringValue
			}
		case "event.timestamp":
			if v, ok := attr.Value.Value.(*commonv1.AnyValue_StringValue); ok {
				timestampStr = v.StringValue
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

	// Parse timestamp
	timestamp, err := time.Parse(time.RFC3339, timestampStr)
	if err != nil {
		timestamp = time.Now()
	}

	tokens := entity.NewToken(inputTokens, outputTokens, cacheReadTokens, cacheCreationTokens)
	cost := entity.NewCost(costUSD)
	req := entity.NewAPIRequest(sessionID, timestamp, model, tokens, cost, durationMS)
	return &req
}
