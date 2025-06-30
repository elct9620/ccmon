package tui_test

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
	"github.com/elct9620/ccmon/handler/tui"
	"github.com/elct9620/ccmon/usecase"
)

// TestOverviewTab_IntegrationWithViewModel tests overview tab through the full ViewModel
func TestOverviewTab_IntegrationWithViewModel(t *testing.T) {
	setupTestEnvironment()

	tests := []struct {
		name        string
		width       int
		height      int
		hasData     bool
		description string
	}{
		{
			name:        "Overview tab with empty data",
			width:       120,
			height:      40,
			hasData:     false,
			description: "Should render overview tab with empty data",
		},
		{
			name:        "Overview tab with test data",
			width:       120,
			height:      40,
			hasData:     true,
			description: "Should render overview tab with test data",
		},
		{
			name:        "Overview tab narrow terminal",
			width:       80,
			height:      25,
			hasData:     true,
			description: "Should render overview tab in narrow terminal",
		},
		{
			name:        "Overview tab very narrow terminal",
			width:       60,
			height:      20,
			hasData:     true,
			description: "Should render compact overview in very narrow terminal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Setup test data
			mockRepo := NewMockAPIRequestRepository()
			if tt.hasData {
				mockRepo.SetMockData(CreateTestRequestsSet(), CreateTestStats())
			}

			getFilteredQuery := usecase.NewGetFilteredApiRequestsQuery(mockRepo)
			calculateStatsQuery := usecase.NewCalculateStatsQuery(mockRepo)
			getUsageQuery := usecase.NewGetUsageQuery(mockRepo)

			// Create the ViewModel (starts on overview tab by default)
			model := tui.NewViewModel(getFilteredQuery, calculateStatsQuery, getUsageQuery, time.UTC, nil, 10*time.Millisecond)

			// Create teatest model
			tm := teatest.NewTestModel(
				t, model,
				teatest.WithInitialTermSize(tt.width, tt.height),
			)

			isCompact := tt.width <= 80
			if !isCompact {
				// Wait for initial render - should be on overview tab
				teatest.WaitFor(
					t, tm.Output(),
					func(bts []byte) bool {
						output := string(bts)
						return strings.Contains(output, "Claude Code Monitor") ||
							strings.Contains(output, "[Current]") ||
							strings.Contains(output, "Recent API Requests")
					},
					teatest.WithCheckInterval(time.Millisecond*100),
					teatest.WithDuration(time.Millisecond*500),
				)
			}

			// Test renders the overview tab (the default tab)
			// Since we're already on overview tab by default, just verify basic functionality

			// Quit the program
			tm.Send(tea.KeyMsg{
				Type:  tea.KeyRunes,
				Runes: []rune("q"),
			})

			tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second*3))
		})
	}
}

// TestOverviewTab_TableInteractions tests table interactions in overview tab
func TestOverviewTab_TableInteractions(t *testing.T) {
	setupTestEnvironment()

	t.Run("Table navigation and focus", func(t *testing.T) {
		t.Parallel()
		// Setup test data
		mockRepo := NewMockAPIRequestRepository()
		mockRepo.SetMockData(CreateTestRequestsSet(), CreateTestStats())
		getFilteredQuery := usecase.NewGetFilteredApiRequestsQuery(mockRepo)
		calculateStatsQuery := usecase.NewCalculateStatsQuery(mockRepo)
		getUsageQuery := usecase.NewGetUsageQuery(mockRepo)

		model := tui.NewViewModel(getFilteredQuery, calculateStatsQuery, getUsageQuery, time.UTC, nil, 10*time.Millisecond)

		tm := teatest.NewTestModel(
			t, model,
			teatest.WithInitialTermSize(120, 40),
		)

		// Wait for initial render
		teatest.WaitFor(
			t, tm.Output(),
			func(bts []byte) bool {
				return strings.Contains(string(bts), "Claude Code Monitor")
			},
			teatest.WithCheckInterval(time.Millisecond*100),
			teatest.WithDuration(time.Millisecond*500),
		)

		// Test table navigation - arrow down
		tm.Send(tea.KeyMsg{Type: tea.KeyDown})

		// Test table navigation - arrow up
		tm.Send(tea.KeyMsg{Type: tea.KeyUp})

		// Test escape key for focus management
		tm.Send(tea.KeyMsg{Type: tea.KeyEsc})

		// Test sorting toggle
		tm.Send(tea.KeyMsg{
			Type:  tea.KeyRunes,
			Runes: []rune("o"),
		})

		// Quit the program
		tm.Send(tea.KeyMsg{
			Type:  tea.KeyRunes,
			Runes: []rune("q"),
		})

		tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second*3))
	})
}

// TestOverviewTab_TimeFiltering tests time filtering functionality
func TestOverviewTab_TimeFiltering(t *testing.T) {
	setupTestEnvironment()

	t.Run("Time filter key presses", func(t *testing.T) {
		t.Parallel()
		// Setup test data
		mockRepo := NewMockAPIRequestRepository()
		mockRepo.SetMockData(CreateTestRequestsSet(), CreateTestStats())
		getFilteredQuery := usecase.NewGetFilteredApiRequestsQuery(mockRepo)
		calculateStatsQuery := usecase.NewCalculateStatsQuery(mockRepo)
		getUsageQuery := usecase.NewGetUsageQuery(mockRepo)

		model := tui.NewViewModel(getFilteredQuery, calculateStatsQuery, getUsageQuery, time.UTC, nil, 10*time.Millisecond)

		tm := teatest.NewTestModel(
			t, model,
			teatest.WithInitialTermSize(120, 40),
		)

		// Wait for initial render
		teatest.WaitFor(
			t, tm.Output(),
			func(bts []byte) bool {
				return strings.Contains(string(bts), "Claude Code Monitor")
			},
			teatest.WithCheckInterval(time.Millisecond*100),
			teatest.WithDuration(time.Millisecond*500),
		)

		// Test time filter keys
		filterKeys := []string{"1", "3", "6", "12", "24", "a"}
		for _, key := range filterKeys {
			tm.Send(tea.KeyMsg{
				Type:  tea.KeyRunes,
				Runes: []rune(key),
			})
		}

		// Quit the program
		tm.Send(tea.KeyMsg{
			Type:  tea.KeyRunes,
			Runes: []rune("q"),
		})

		tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second*3))
	})
}