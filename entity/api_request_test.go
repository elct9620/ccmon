package entity

import (
	"strings"
	"testing"
	"time"
)

func TestAPIRequest_ID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		sessionID    string
		timestamp    time.Time
		wantContains []string
	}{
		{
			name:      "generates ID with timestamp and session",
			sessionID: "session123",
			timestamp: time.Date(2024, 1, 1, 10, 30, 45, 123456789, time.UTC),
			wantContains: []string{
				"2024-01-01T10:30:45.123456789Z",
				"session123",
				"_", // separator
			},
		},
		{
			name:      "handles different session ID",
			sessionID: "different-session-456",
			timestamp: time.Date(2023, 12, 25, 15, 45, 30, 987654321, time.UTC),
			wantContains: []string{
				"2023-12-25T15:45:30.987654321Z",
				"different-session-456",
				"_",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create test tokens and cost
			tokens := NewToken(100, 50, 0, 0)
			cost := NewCost(0.001)

			// Create APIRequest
			req := NewAPIRequest(tt.sessionID, tt.timestamp, "claude-3-sonnet", tokens, cost, 1000)

			// Get ID
			got := req.ID()

			// Verify all expected parts are present
			for _, want := range tt.wantContains {
				if !strings.Contains(got, want) {
					t.Errorf("APIRequest.ID() = %v, expected to contain %v", got, want)
				}
			}

			// Verify format: timestamp_sessionID
			parts := strings.Split(got, "_")
			if len(parts) != 2 {
				t.Errorf("APIRequest.ID() = %v, expected format timestamp_sessionID", got)
			}

			// Verify timestamp part is RFC3339Nano format
			expectedTimestamp := tt.timestamp.Format(time.RFC3339Nano)
			if parts[0] != expectedTimestamp {
				t.Errorf("APIRequest.ID() timestamp part = %v, expected %v", parts[0], expectedTimestamp)
			}

			// Verify session ID part
			if parts[1] != tt.sessionID {
				t.Errorf("APIRequest.ID() session part = %v, expected %v", parts[1], tt.sessionID)
			}
		})
	}
}

func TestAPIRequest_ID_Uniqueness(t *testing.T) {
	t.Parallel()

	// Create test data
	tokens := NewToken(100, 50, 0, 0)
	cost := NewCost(0.001)
	baseTime := time.Date(2024, 1, 1, 10, 30, 45, 123456789, time.UTC)

	// Test different timestamps with same session ID
	req1 := NewAPIRequest("same-session", baseTime, "claude-3-sonnet", tokens, cost, 1000)
	req2 := NewAPIRequest("same-session", baseTime.Add(time.Nanosecond), "claude-3-sonnet", tokens, cost, 1000)

	id1 := req1.ID()
	id2 := req2.ID()

	if id1 == id2 {
		t.Errorf("Expected different IDs for different timestamps, got same ID: %v", id1)
	}

	// Test different session IDs with same timestamp
	req3 := NewAPIRequest("session1", baseTime, "claude-3-sonnet", tokens, cost, 1000)
	req4 := NewAPIRequest("session2", baseTime, "claude-3-sonnet", tokens, cost, 1000)

	id3 := req3.ID()
	id4 := req4.ID()

	if id3 == id4 {
		t.Errorf("Expected different IDs for different sessions, got same ID: %v", id3)
	}
}
