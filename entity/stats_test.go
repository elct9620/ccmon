package entity

import (
	"testing"
	"time"
)

func TestNewStatsFromRequests(t *testing.T) {
	t.Parallel()

	baseTime := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
	period := NewPeriod(baseTime, baseTime.Add(24*time.Hour))

	tests := []struct {
		name     string
		requests []APIRequest
		period   Period
		want     Stats
	}{
		{
			name:     "empty requests",
			requests: []APIRequest{},
			period:   period,
			want: NewStats(
				0, 0,
				Token{}, Token{},
				Cost{}, Cost{},
				period,
			),
		},
		{
			name: "single base model request",
			requests: []APIRequest{
				NewAPIRequest("session1", baseTime, "claude-3-haiku", NewToken(100, 50, 0, 0), NewCost(0.001), 1000),
			},
			period: period,
			want: NewStats(
				1, 0,
				NewToken(100, 50, 0, 0), Token{},
				NewCost(0.001), Cost{},
				period,
			),
		},
		{
			name: "single premium model request",
			requests: []APIRequest{
				NewAPIRequest("session1", baseTime, "claude-3-sonnet", NewToken(200, 100, 10, 5), NewCost(0.002), 1500),
			},
			period: period,
			want: NewStats(
				0, 1,
				Token{}, NewToken(200, 100, 10, 5),
				Cost{}, NewCost(0.002),
				period,
			),
		},
		{
			name: "mixed base and premium requests",
			requests: []APIRequest{
				NewAPIRequest("session1", baseTime, "claude-3-haiku", NewToken(100, 50, 0, 0), NewCost(0.001), 1000),
				NewAPIRequest("session2", baseTime.Add(time.Hour), "claude-3-sonnet", NewToken(200, 100, 10, 5), NewCost(0.002), 1500),
				NewAPIRequest("session3", baseTime.Add(2*time.Hour), "claude-3-haiku", NewToken(80, 40, 0, 0), NewCost(0.0008), 800),
				NewAPIRequest("session4", baseTime.Add(3*time.Hour), "claude-3-opus", NewToken(300, 150, 20, 10), NewCost(0.003), 2000),
			},
			period: period,
			want: NewStats(
				2, 2,
				NewToken(180, 90, 0, 0), NewToken(500, 250, 30, 15),
				NewCost(0.0018), NewCost(0.005),
				period,
			),
		},
		{
			name: "multiple requests same session",
			requests: []APIRequest{
				NewAPIRequest("same-session", baseTime, "claude-3-haiku", NewToken(50, 25, 0, 0), NewCost(0.001), 500),
				NewAPIRequest("same-session", baseTime.Add(time.Minute), "claude-3-haiku", NewToken(60, 30, 0, 0), NewCost(0.002), 600),
			},
			period: period,
			want: NewStats(
				2, 0,
				NewToken(110, 55, 0, 0), Token{},
				NewCost(0.003), Cost{},
				period,
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := NewStatsFromRequests(tt.requests, tt.period)

			if got.BaseRequests() != tt.want.BaseRequests() {
				t.Errorf("BaseRequests() = %v, want %v", got.BaseRequests(), tt.want.BaseRequests())
			}
			if got.PremiumRequests() != tt.want.PremiumRequests() {
				t.Errorf("PremiumRequests() = %v, want %v", got.PremiumRequests(), tt.want.PremiumRequests())
			}
			if got.BaseTokens() != tt.want.BaseTokens() {
				t.Errorf("BaseTokens() = %v, want %v", got.BaseTokens(), tt.want.BaseTokens())
			}
			if got.PremiumTokens() != tt.want.PremiumTokens() {
				t.Errorf("PremiumTokens() = %v, want %v", got.PremiumTokens(), tt.want.PremiumTokens())
			}
			if got.BaseCost() != tt.want.BaseCost() {
				t.Errorf("BaseCost() = %v, want %v", got.BaseCost(), tt.want.BaseCost())
			}
			if got.PremiumCost() != tt.want.PremiumCost() {
				t.Errorf("PremiumCost() = %v, want %v", got.PremiumCost(), tt.want.PremiumCost())
			}
			if got.Period() != tt.want.Period() {
				t.Errorf("Period() = %v, want %v", got.Period(), tt.want.Period())
			}
		})
	}
}

func TestNewStatsFromRequests_ModelClassification(t *testing.T) {
	t.Parallel()

	baseTime := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
	period := NewPeriod(baseTime, baseTime.Add(24*time.Hour))

	tests := []struct {
		name          string
		model         string
		expectedBase  bool
		expectedCount int
	}{
		{"haiku model", "claude-3-haiku-20240307", true, 1},
		{"sonnet model", "claude-3-sonnet-20240229", false, 1},
		{"opus model", "claude-3-opus-20240229", false, 1},
		{"unknown model", "unknown-model", false, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			request := NewAPIRequest("session", baseTime, tt.model, NewToken(100, 50, 0, 0), NewCost(0.001), 1000)
			stats := NewStatsFromRequests([]APIRequest{request}, period)

			if tt.expectedBase {
				if stats.BaseRequests() != tt.expectedCount {
					t.Errorf("Expected %d base requests, got %d", tt.expectedCount, stats.BaseRequests())
				}
				if stats.PremiumRequests() != 0 {
					t.Errorf("Expected 0 premium requests, got %d", stats.PremiumRequests())
				}
			} else {
				if stats.PremiumRequests() != tt.expectedCount {
					t.Errorf("Expected %d premium requests, got %d", tt.expectedCount, stats.PremiumRequests())
				}
				if stats.BaseRequests() != 0 {
					t.Errorf("Expected 0 base requests, got %d", stats.BaseRequests())
				}
			}
		})
	}
}
