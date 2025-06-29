package receiver

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/elct9620/ccmon/entity"
	"github.com/elct9620/ccmon/usecase"
	logsv1 "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	metricsv1 "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	tracesv1 "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	commonv1 "go.opentelemetry.io/proto/otlp/common/v1"
	logsdata "go.opentelemetry.io/proto/otlp/logs/v1"
	metricsdata "go.opentelemetry.io/proto/otlp/metrics/v1"
	resourcev1 "go.opentelemetry.io/proto/otlp/resource/v1"
	tracesdata "go.opentelemetry.io/proto/otlp/trace/v1"
)

// Mock repository for testing
type mockAPIRequestRepository struct {
	requests []entity.APIRequest
	saveErr  error
}

func (m *mockAPIRequestRepository) Save(req entity.APIRequest) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.requests = append(m.requests, req)
	return nil
}

func (m *mockAPIRequestRepository) FindByPeriodWithLimit(period entity.Period, limit int, offset int) ([]entity.APIRequest, error) {
	return m.requests, nil
}

func (m *mockAPIRequestRepository) FindAll() ([]entity.APIRequest, error) {
	return m.requests, nil
}

// Helper function to create OTLP log request with Claude Code API request data
func createClaudeCodeLogRequest(sessionID, timestamp, model string, inputTokens, outputTokens, cacheRead, cacheCreation int64, costUSD float64, durationMS int64) *logsv1.ExportLogsServiceRequest {
	return &logsv1.ExportLogsServiceRequest{
		ResourceLogs: []*logsdata.ResourceLogs{
			{
				Resource: &resourcev1.Resource{
					Attributes: []*commonv1.KeyValue{},
				},
				ScopeLogs: []*logsdata.ScopeLogs{
					{
						LogRecords: []*logsdata.LogRecord{
							{
								Body: &commonv1.AnyValue{
									Value: &commonv1.AnyValue_StringValue{
										StringValue: "claude_code.api_request",
									},
								},
								Attributes: []*commonv1.KeyValue{
									{
										Key: "session.id",
										Value: &commonv1.AnyValue{
											Value: &commonv1.AnyValue_StringValue{
												StringValue: sessionID,
											},
										},
									},
									{
										Key: "event.timestamp",
										Value: &commonv1.AnyValue{
											Value: &commonv1.AnyValue_StringValue{
												StringValue: timestamp,
											},
										},
									},
									{
										Key: "model",
										Value: &commonv1.AnyValue{
											Value: &commonv1.AnyValue_StringValue{
												StringValue: model,
											},
										},
									},
									{
										Key: "input_tokens",
										Value: &commonv1.AnyValue{
											Value: &commonv1.AnyValue_StringValue{
												StringValue: fmt.Sprintf("%d", inputTokens),
											},
										},
									},
									{
										Key: "output_tokens",
										Value: &commonv1.AnyValue{
											Value: &commonv1.AnyValue_StringValue{
												StringValue: fmt.Sprintf("%d", outputTokens),
											},
										},
									},
									{
										Key: "cache_read_tokens",
										Value: &commonv1.AnyValue{
											Value: &commonv1.AnyValue_StringValue{
												StringValue: fmt.Sprintf("%d", cacheRead),
											},
										},
									},
									{
										Key: "cache_creation_tokens",
										Value: &commonv1.AnyValue{
											Value: &commonv1.AnyValue_StringValue{
												StringValue: fmt.Sprintf("%d", cacheCreation),
											},
										},
									},
									{
										Key: "cost_usd",
										Value: &commonv1.AnyValue{
											Value: &commonv1.AnyValue_StringValue{
												StringValue: fmt.Sprintf("%.6f", costUSD),
											},
										},
									},
									{
										Key: "duration_ms",
										Value: &commonv1.AnyValue{
											Value: &commonv1.AnyValue_StringValue{
												StringValue: fmt.Sprintf("%d", durationMS),
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func TestOTLPReceiver_LogsServiceExport(t *testing.T) {
	now := time.Now().UTC()
	validTimestamp := now.Format(time.RFC3339)

	tests := []struct {
		name               string
		request            *logsv1.ExportLogsServiceRequest
		expectedSavedCount int
		validateSaved      func(t *testing.T, saved entity.APIRequest)
	}{
		{
			name: "valid_claude_code_request",
			request: createClaudeCodeLogRequest(
				"test-session-123",
				validTimestamp,
				"claude-3-sonnet-20240229",
				1000, 500, 100, 50, // tokens
				2.50, // cost
				1500, // duration
			),
			expectedSavedCount: 1,
			validateSaved: func(t *testing.T, saved entity.APIRequest) {
				if saved.SessionID() != "test-session-123" {
					t.Errorf("Expected session ID 'test-session-123', got '%s'", saved.SessionID())
				}
				if string(saved.Model()) != "claude-3-sonnet-20240229" {
					t.Errorf("Expected model 'claude-3-sonnet-20240229', got '%s'", saved.Model())
				}
				tokens := saved.Tokens()
				if tokens.Input() != 1000 {
					t.Errorf("Expected 1000 input tokens, got %d", tokens.Input())
				}
				if tokens.Output() != 500 {
					t.Errorf("Expected 500 output tokens, got %d", tokens.Output())
				}
				if tokens.CacheRead() != 100 {
					t.Errorf("Expected 100 cache read tokens, got %d", tokens.CacheRead())
				}
				if tokens.CacheCreation() != 50 {
					t.Errorf("Expected 50 cache creation tokens, got %d", tokens.CacheCreation())
				}
				if saved.Cost().Amount() != 2.50 {
					t.Errorf("Expected cost $2.50, got $%.2f", saved.Cost().Amount())
				}
				if saved.DurationMS() != 1500 {
					t.Errorf("Expected duration 1500ms, got %dms", saved.DurationMS())
				}
			},
		},
		{
			name: "non_claude_code_request_ignored",
			request: &logsv1.ExportLogsServiceRequest{
				ResourceLogs: []*logsdata.ResourceLogs{
					{
						ScopeLogs: []*logsdata.ScopeLogs{
							{
								LogRecords: []*logsdata.LogRecord{
									{
										Body: &commonv1.AnyValue{
											Value: &commonv1.AnyValue_StringValue{
												StringValue: "some.other.log", // Not claude_code.api_request
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expectedSavedCount: 0,
			validateSaved:      nil, // No validation needed for ignored requests
		},
		{
			name: "haiku_model_request",
			request: createClaudeCodeLogRequest(
				"haiku-session",
				validTimestamp,
				"claude-3-haiku-20240307", // Base model
				500, 250, 50, 25,
				0.15,
				800,
			),
			expectedSavedCount: 1,
			validateSaved: func(t *testing.T, saved entity.APIRequest) {
				if string(saved.Model()) != "claude-3-haiku-20240307" {
					t.Errorf("Expected model 'claude-3-haiku-20240307', got '%s'", saved.Model())
				}
				if saved.Cost().Amount() != 0.15 {
					t.Errorf("Expected cost $0.15, got $%.2f", saved.Cost().Amount())
				}
			},
		},
		{
			name: "opus_model_request",
			request: createClaudeCodeLogRequest(
				"opus-session",
				validTimestamp,
				"claude-3-opus-20240229", // Premium model
				2000, 1000, 200, 100,
				5.00,
				3000,
			),
			expectedSavedCount: 1,
			validateSaved: func(t *testing.T, saved entity.APIRequest) {
				if string(saved.Model()) != "claude-3-opus-20240229" {
					t.Errorf("Expected model 'claude-3-opus-20240229', got '%s'", saved.Model())
				}
				if saved.Cost().Amount() != 5.00 {
					t.Errorf("Expected cost $5.00, got $%.2f", saved.Cost().Amount())
				}
			},
		},
		{
			name: "malformed_token_data_handled_gracefully",
			request: func() *logsv1.ExportLogsServiceRequest {
				req := createClaudeCodeLogRequest(
					"test-session",
					"invalid-timestamp",
					"claude-3-sonnet-20240229",
					1000, 500, 100, 50,
					2.50,
					1500,
				)
				// Modify input_tokens attribute to have invalid data
				req.ResourceLogs[0].ScopeLogs[0].LogRecords[0].Attributes[3].Value = &commonv1.AnyValue{
					Value: &commonv1.AnyValue_StringValue{
						StringValue: "invalid-number", // Should cause parsing error
					},
				}
				return req
			}(),
			expectedSavedCount: 1,
			validateSaved: func(t *testing.T, saved entity.APIRequest) {
				if saved.SessionID() != "test-session" {
					t.Errorf("Expected session ID 'test-session', got '%s'", saved.SessionID())
				}
				// The invalid token field should be 0 due to parsing error
				tokens := saved.Tokens()
				if tokens.Input() != 0 { // Should be 0 due to parsing error
					t.Errorf("Expected 0 input tokens (due to parsing error), got %d", tokens.Input())
				}
				// Other valid fields should still work
				if tokens.Output() != 500 {
					t.Errorf("Expected 500 output tokens, got %d", tokens.Output())
				}
			},
		},
		{
			name: "empty_request",
			request: &logsv1.ExportLogsServiceRequest{
				ResourceLogs: []*logsdata.ResourceLogs{},
			},
			expectedSavedCount: 0,
			validateSaved:      nil,
		},
		{
			name: "multiple_records_mixed",
			request: &logsv1.ExportLogsServiceRequest{
				ResourceLogs: []*logsdata.ResourceLogs{
					{
						ScopeLogs: []*logsdata.ScopeLogs{
							{
								LogRecords: []*logsdata.LogRecord{
									// First record: should be ignored
									{
										Body: &commonv1.AnyValue{
											Value: &commonv1.AnyValue_StringValue{
												StringValue: "other.log",
											},
										},
									},
									// Second record: should be processed
									createClaudeCodeLogRequest(
										"multi-session",
										validTimestamp,
										"claude-3-sonnet-20240229",
										100, 50, 10, 5,
										0.30,
										500,
									).ResourceLogs[0].ScopeLogs[0].LogRecords[0],
								},
							},
						},
					},
				},
			},
			expectedSavedCount: 1,
			validateSaved: func(t *testing.T, saved entity.APIRequest) {
				if saved.SessionID() != "multi-session" {
					t.Errorf("Expected session ID 'multi-session', got '%s'", saved.SessionID())
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock repository and usecase for each test
			mockRepo := &mockAPIRequestRepository{}
			appendCommand := usecase.NewAppendApiRequestCommand(mockRepo)

			// Create receiver
			receiver := NewReceiver(nil, nil, appendCommand)
			logsService := receiver.GetLogsServiceServer()

			// Export the log
			ctx := context.Background()
			resp, err := logsService.Export(ctx, tt.request)
			if err != nil {
				t.Fatalf("Export failed: %v", err)
			}

			// Verify response is not nil
			if resp == nil {
				t.Fatal("Expected non-nil response")
			}

			// Verify expected number of saved requests
			if len(mockRepo.requests) != tt.expectedSavedCount {
				t.Errorf("Expected %d requests in repository, got %d", tt.expectedSavedCount, len(mockRepo.requests))
			}

			// Validate saved request if validation function is provided
			if tt.validateSaved != nil && len(mockRepo.requests) > 0 {
				tt.validateSaved(t, mockRepo.requests[0])
			}
		})
	}
}

func TestOTLPReceiver_IgnoredServices(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T, receiver *Receiver)
	}{
		{
			name: "trace_service_export_ignored",
			test: func(t *testing.T, receiver *Receiver) {
				traceService := receiver.GetTraceServiceServer()
				req := &tracesv1.ExportTraceServiceRequest{}

				ctx := context.Background()
				resp, err := traceService.Export(ctx, req)
				if err != nil {
					t.Fatalf("Trace export failed: %v", err)
				}
				if resp == nil {
					t.Fatal("Expected non-nil response")
				}
			},
		},
		{
			name: "metrics_service_export_ignored",
			test: func(t *testing.T, receiver *Receiver) {
				metricsService := receiver.GetMetricsServiceServer()
				req := &metricsv1.ExportMetricsServiceRequest{}

				ctx := context.Background()
				resp, err := metricsService.Export(ctx, req)
				if err != nil {
					t.Fatalf("Metrics export failed: %v", err)
				}
				if resp == nil {
					t.Fatal("Expected non-nil response")
				}
			},
		},
		{
			name: "trace_service_with_data_ignored",
			test: func(t *testing.T, receiver *Receiver) {
				traceService := receiver.GetTraceServiceServer()
				req := &tracesv1.ExportTraceServiceRequest{
					ResourceSpans: []*tracesdata.ResourceSpans{},
				}

				ctx := context.Background()
				resp, err := traceService.Export(ctx, req)
				if err != nil {
					t.Fatalf("Trace export failed: %v", err)
				}
				if resp == nil {
					t.Fatal("Expected non-nil response")
				}
			},
		},
		{
			name: "metrics_service_with_data_ignored",
			test: func(t *testing.T, receiver *Receiver) {
				metricsService := receiver.GetMetricsServiceServer()
				req := &metricsv1.ExportMetricsServiceRequest{
					ResourceMetrics: []*metricsdata.ResourceMetrics{},
				}

				ctx := context.Background()
				resp, err := metricsService.Export(ctx, req)
				if err != nil {
					t.Fatalf("Metrics export failed: %v", err)
				}
				if resp == nil {
					t.Fatal("Expected non-nil response")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup receiver (no usecase needed for ignored services)
			receiver := NewReceiver(nil, nil, nil)

			// Run the specific test
			tt.test(t, receiver)
		})
	}
}
