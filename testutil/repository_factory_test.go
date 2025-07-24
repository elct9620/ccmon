package testutil

import (
	"errors"
	"testing"
	"time"

	"github.com/elct9620/ccmon/entity"
)

func TestNewMockAPIRequestRepository(t *testing.T) {
	repo := NewMockAPIRequestRepository()

	if repo == nil {
		t.Fatal("Expected repository to be created")
	}

	if len(repo.requests) != 0 {
		t.Errorf("Expected empty requests slice, got %d items", len(repo.requests))
	}

	if repo.err != nil {
		t.Errorf("Expected no error, got %v", repo.err)
	}
}

func TestMockAPIRequestRepository_Save(t *testing.T) {
	repo := NewMockAPIRequestRepository()

	req := CreateTestAPIRequest("test", time.Now(), "claude-3-haiku-20240307", 100, 50, 0.01)

	err := repo.Save(req)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(repo.requests) != 1 {
		t.Errorf("Expected 1 request, got %d", len(repo.requests))
	}

	if repo.requests[0].SessionID() != "test" {
		t.Errorf("Expected session ID 'test', got %s", repo.requests[0].SessionID())
	}
}

func TestMockAPIRequestRepository_SaveWithError(t *testing.T) {
	expectedErr := errors.New("save error")
	repo := NewMockAPIRequestRepositoryWithError(expectedErr)

	req := CreateTestAPIRequest("test", time.Now(), "claude-3-haiku-20240307", 100, 50, 0.01)

	err := repo.Save(req)
	if err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}
}

func TestMockAPIRequestRepository_FindByPeriodWithLimit(t *testing.T) {
	repo := NewMockAPIRequestRepository()
	now := time.Now()

	// Add test requests
	req1 := CreateTestAPIRequest("session1", now.Add(-2*time.Hour), "claude-3-haiku-20240307", 100, 50, 0.01)
	req2 := CreateTestAPIRequest("session2", now.Add(-1*time.Hour), "claude-3-sonnet-20240229", 200, 100, 0.02)
	req3 := CreateTestAPIRequest("session3", now.Add(-30*time.Minute), "claude-3-haiku-20240307", 150, 75, 0.015)

	if err := repo.Save(req1); err != nil {
		t.Fatalf("Failed to save req1: %v", err)
	}
	if err := repo.Save(req2); err != nil {
		t.Fatalf("Failed to save req2: %v", err)
	}
	if err := repo.Save(req3); err != nil {
		t.Fatalf("Failed to save req3: %v", err)
	}

	t.Run("all time period", func(t *testing.T) {
		period := entity.NewAllTimePeriod(now)
		requests, err := repo.FindByPeriodWithLimit(period, 0, 0)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if len(requests) != 3 {
			t.Errorf("Expected 3 requests, got %d", len(requests))
		}
	})

	t.Run("time filtered period", func(t *testing.T) {
		period := entity.NewPeriod(now.Add(-90*time.Minute), now)
		requests, err := repo.FindByPeriodWithLimit(period, 0, 0)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if len(requests) != 2 {
			t.Errorf("Expected 2 requests, got %d", len(requests))
		}
	})

	t.Run("with limit", func(t *testing.T) {
		period := entity.NewAllTimePeriod(now)
		requests, err := repo.FindByPeriodWithLimit(period, 2, 0)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if len(requests) != 2 {
			t.Errorf("Expected 2 requests due to limit, got %d", len(requests))
		}
	})

	t.Run("with offset", func(t *testing.T) {
		period := entity.NewAllTimePeriod(now)
		requests, err := repo.FindByPeriodWithLimit(period, 0, 1)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if len(requests) != 2 {
			t.Errorf("Expected 2 requests after offset, got %d", len(requests))
		}
	})
}

func TestMockAPIRequestRepository_DeleteOlderThan(t *testing.T) {
	repo := NewMockAPIRequestRepository()
	now := time.Now()
	cutoff := now.Add(-1 * time.Hour)

	// Add requests before and after cutoff
	oldReq := CreateTestAPIRequest("old", now.Add(-2*time.Hour), "claude-3-haiku-20240307", 100, 50, 0.01)
	newReq := CreateTestAPIRequest("new", now.Add(-30*time.Minute), "claude-3-sonnet-20240229", 200, 100, 0.02)

	if err := repo.Save(oldReq); err != nil {
		t.Fatalf("Failed to save oldReq: %v", err)
	}
	if err := repo.Save(newReq); err != nil {
		t.Fatalf("Failed to save newReq: %v", err)
	}

	deletedCount, err := repo.DeleteOlderThan(cutoff)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if deletedCount != 1 {
		t.Errorf("Expected 1 deleted record, got %d", deletedCount)
	}

	remaining, _ := repo.FindAll()
	if len(remaining) != 1 {
		t.Errorf("Expected 1 remaining record, got %d", len(remaining))
	}

	if remaining[0].SessionID() != "new" {
		t.Errorf("Expected remaining request to be 'new', got %s", remaining[0].SessionID())
	}
}

func TestNewMockStatsRepository(t *testing.T) {
	apiRepo := NewMockAPIRequestRepository()
	statsRepo := NewMockStatsRepository(apiRepo)

	if statsRepo == nil {
		t.Fatal("Expected stats repository to be created")
	}

	if statsRepo.apiRepo != apiRepo {
		t.Error("Expected stats repository to wrap the API repository")
	}
}

func TestMockStatsRepository_GetStatsByPeriod(t *testing.T) {
	apiRepo, statsRepo := NewMockRepositoryPair()

	now := time.Now()
	req1 := CreateTestAPIRequest("session1", now, "claude-3-haiku-20240307", 100, 50, 0.01)
	req2 := CreateTestAPIRequest("session2", now, "claude-3-sonnet-20240229", 200, 100, 0.02)

	if err := apiRepo.Save(req1); err != nil {
		t.Fatalf("Failed to save req1: %v", err)
	}
	if err := apiRepo.Save(req2); err != nil {
		t.Fatalf("Failed to save req2: %v", err)
	}

	period := entity.NewAllTimePeriod(now)
	stats, err := statsRepo.GetStatsByPeriod(period)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if stats.TotalRequests() != 2 {
		t.Errorf("Expected 2 total requests, got %d", stats.TotalRequests())
	}

	if stats.BaseRequests() != 1 {
		t.Errorf("Expected 1 base request, got %d", stats.BaseRequests())
	}

	if stats.PremiumRequests() != 1 {
		t.Errorf("Expected 1 premium request, got %d", stats.PremiumRequests())
	}
}

func TestNewMockRepositoryPair(t *testing.T) {
	apiRepo, statsRepo := NewMockRepositoryPair()

	if apiRepo == nil {
		t.Fatal("Expected API repository to be created")
	}

	if statsRepo == nil {
		t.Fatal("Expected stats repository to be created")
	}

	if statsRepo.apiRepo != apiRepo {
		t.Error("Expected stats repository to wrap the API repository")
	}
}

func TestNewInstrumentedRepositoryPair(t *testing.T) {
	apiRepo, statsRepo, callCount := NewInstrumentedRepositoryPair()

	if apiRepo == nil {
		t.Fatal("Expected API repository to be created")
	}

	if statsRepo == nil {
		t.Fatal("Expected stats repository to be created")
	}

	if callCount == nil {
		t.Fatal("Expected call count pointer to be created")
	}

	if *callCount != 0 {
		t.Errorf("Expected initial call count to be 0, got %d", *callCount)
	}

	// Test call counting
	now := time.Now()
	req := CreateTestAPIRequest("test", now, "claude-3-haiku-20240307", 100, 50, 0.01)
	if err := apiRepo.Save(req); err != nil {
		t.Fatalf("Failed to save req: %v", err)
	}

	period := entity.NewAllTimePeriod(now)
	_, err := statsRepo.GetStatsByPeriod(period)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if *callCount != 1 {
		t.Errorf("Expected call count to be 1 after stats query, got %d", *callCount)
	}
}

func TestNewMockRepositoryWithData(t *testing.T) {
	testRequests := CreateTestRequestsSet()
	apiRepo, statsRepo := NewMockRepositoryWithData(testRequests)

	if apiRepo == nil {
		t.Fatal("Expected API repository to be created")
	}

	if statsRepo == nil {
		t.Fatal("Expected stats repository to be created")
	}

	// Verify data was set
	requests, err := apiRepo.FindAll()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(requests) != len(testRequests) {
		t.Errorf("Expected %d requests, got %d", len(testRequests), len(requests))
	}
}

func TestNewMockRepositoryWithTestData(t *testing.T) {
	apiRepo, statsRepo := NewMockRepositoryWithTestData()

	if apiRepo == nil {
		t.Fatal("Expected API repository to be created")
	}

	if statsRepo == nil {
		t.Fatal("Expected stats repository to be created")
	}

	// Verify test data was loaded
	requests, err := apiRepo.FindAll()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(requests) == 0 {
		t.Error("Expected test data to be loaded")
	}

	// Verify we can get stats
	period := entity.NewAllTimePeriod(time.Now())
	stats, err := statsRepo.GetStatsByPeriod(period)
	if err != nil {
		t.Errorf("Unexpected error getting stats: %v", err)
	}

	if stats.TotalRequests() == 0 {
		t.Error("Expected stats to show requests from test data")
	}
}

func TestCreateTestAPIRequest(t *testing.T) {
	sessionID := "test-session"
	timestamp := time.Now()
	model := "claude-3-haiku-20240307"
	inputTokens := int64(100)
	outputTokens := int64(50)
	cost := 0.01

	req := CreateTestAPIRequest(sessionID, timestamp, model, inputTokens, outputTokens, cost)

	if req.SessionID() != sessionID {
		t.Errorf("Expected session ID %s, got %s", sessionID, req.SessionID())
	}

	if !req.Timestamp().Equal(timestamp) {
		t.Errorf("Expected timestamp %v, got %v", timestamp, req.Timestamp())
	}

	if string(req.Model()) != model {
		t.Errorf("Expected model %s, got %s", model, string(req.Model()))
	}

	if req.Tokens().Input() != inputTokens {
		t.Errorf("Expected input tokens %d, got %d", inputTokens, req.Tokens().Input())
	}

	if req.Tokens().Output() != outputTokens {
		t.Errorf("Expected output tokens %d, got %d", outputTokens, req.Tokens().Output())
	}

	if req.Cost().Amount() != cost {
		t.Errorf("Expected cost %f, got %f", cost, req.Cost().Amount())
	}
}

func TestCreateTestRequestsSet(t *testing.T) {
	requests := CreateTestRequestsSet()

	if len(requests) != 4 {
		t.Errorf("Expected 4 test requests, got %d", len(requests))
	}

	// Verify we have both base and premium models
	hasBase := false
	hasPremium := false

	for _, req := range requests {
		if req.Model().IsBase() {
			hasBase = true
		} else {
			hasPremium = true
		}
	}

	if !hasBase {
		t.Error("Expected at least one base model request")
	}

	if !hasPremium {
		t.Error("Expected at least one premium model request")
	}
}

func TestCreateTestStats(t *testing.T) {
	stats := CreateTestStats()

	if stats.BaseRequests() != 2 {
		t.Errorf("Expected 2 base requests, got %d", stats.BaseRequests())
	}

	if stats.PremiumRequests() != 2 {
		t.Errorf("Expected 2 premium requests, got %d", stats.PremiumRequests())
	}

	if stats.BaseTokens().Total() == 0 {
		t.Error("Expected non-zero base tokens")
	}

	if stats.PremiumTokens().Total() == 0 {
		t.Error("Expected non-zero premium tokens")
	}

	if stats.BaseCost().Amount() == 0 {
		t.Error("Expected non-zero base cost")
	}

	if stats.PremiumCost().Amount() == 0 {
		t.Error("Expected non-zero premium cost")
	}
}
