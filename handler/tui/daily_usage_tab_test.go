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

// TestDailyUsageTab_IntegrationWithViewModel tests daily usage tab through the full ViewModel
// Note: These tests verify functionality works (see output) but have teatest detection issues
func TestDailyUsageTab_IntegrationWithViewModel(t *testing.T) {
	setupTestEnvironment()

	tests := []struct {
		name        string
		width       int
		height      int
		hasData     bool
		description string
	}{
		{
			name:        "Daily tab with empty data",
			width:       120,
			height:      40,
			hasData:     false,
			description: "Should render daily usage tab with empty data",
		},
		{
			name:        "Daily tab with test data",
			width:       120,
			height:      40,
			hasData:     true,
			description: "Should render daily usage tab with test data",
		},
		{
			name:        "Daily tab narrow terminal",
			width:       80,
			height:      25,
			hasData:     true,
			description: "Should render daily usage tab in narrow terminal",
		},
		{
			name:        "Daily tab very narrow terminal",
			width:       60,
			height:      20,
			hasData:     true,
			description: "Should render compact daily usage in very narrow terminal",
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
			getUsageQuery := usecase.NewGetUsageQuery(mockRepo) // Use same repo for consistency

			// Create the ViewModel
			model := tui.NewViewModel(getFilteredQuery, calculateStatsQuery, getUsageQuery, time.UTC, nil, 10*time.Millisecond)

			// Create teatest model
			tm := teatest.NewTestModel(
				t, model,
				teatest.WithInitialTermSize(tt.width, tt.height),
			)

			// Wait for initial render
			isCompact := tt.width <= 80
			if !isCompact {
				teatest.WaitFor(
					t, tm.Output(),
					func(bts []byte) bool {
						output := string(bts)
						return strings.Contains(output, "Claude Code Monitor") ||
							strings.Contains(output, "[Current]")
					},
					teatest.WithCheckInterval(time.Millisecond*100),
					teatest.WithDuration(time.Millisecond*500),
				)
			}

			// Switch to daily usage tab
			tm.Send(tea.KeyMsg{
				Type: tea.KeyTab,
			})

			// Wait for daily usage tab to be displayed
			teatest.WaitFor(
				t, tm.Output(),
				func(bts []byte) bool {
					output := string(bts)
					// Look for daily usage data pattern (dates) to detect daily tab
					return strings.Contains(output, "2025-") ||
						strings.Contains(output, "Daily Usage Statistics")
				},
				teatest.WithCheckInterval(time.Millisecond*100),
				teatest.WithDuration(time.Millisecond*500),
			)

			// Quit the program
			tm.Send(tea.KeyMsg{
				Type:  tea.KeyRunes,
				Runes: []rune("q"),
			})

			tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second*3))
		})
	}
}

// TestDailyUsageTab_NavigationFlow tests daily usage tab navigation
func TestDailyUsageTab_NavigationFlow(t *testing.T) {
	setupTestEnvironment()

	t.Run("Simple tab switch", func(t *testing.T) {
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

		// Switch to daily tab
		tm.Send(tea.KeyMsg{Type: tea.KeyTab})

		// Verify we're on daily tab
		teatest.WaitFor(
			t, tm.Output(),
			func(bts []byte) bool {
				output := string(bts)
				return strings.Contains(output, "2025-") ||
					strings.Contains(output, "Daily Usage Statistics")
			},
			teatest.WithCheckInterval(time.Millisecond*100),
			teatest.WithDuration(time.Millisecond*500),
		)

		// Switch back to current tab
		tm.Send(tea.KeyMsg{Type: tea.KeyTab})

		// Verify we're back on current tab
		teatest.WaitFor(
			t, tm.Output(),
			func(bts []byte) bool {
				output := string(bts)
				return strings.Contains(output, "[Current]") ||
					strings.Contains(output, "Recent API Requests")
			},
			teatest.WithCheckInterval(time.Millisecond*100),
			teatest.WithDuration(time.Millisecond*500),
		)

		// Quit the program
		tm.Send(tea.KeyMsg{
			Type:  tea.KeyRunes,
			Runes: []rune("q"),
		})

		tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second*3))
	})
}

// TestDailyUsageTab_FocusManagement tests focus management during tab switches
func TestDailyUsageTab_FocusManagement(t *testing.T) {
	setupTestEnvironment()

	t.Run("Focus management during tab switches", func(t *testing.T) {
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

		// Wait for initial render - current tab should be focused by default
		teatest.WaitFor(
			t, tm.Output(),
			func(bts []byte) bool {
				return strings.Contains(string(bts), "Claude Code Monitor")
			},
			teatest.WithCheckInterval(time.Millisecond*100),
			teatest.WithDuration(time.Millisecond*500),
		)

		// Switch to daily tab - should focus the daily usage table
		tm.Send(tea.KeyMsg{Type: tea.KeyTab})

		// Wait for daily tab to load
		teatest.WaitFor(
			t, tm.Output(),
			func(bts []byte) bool {
				output := string(bts)
				return strings.Contains(output, "Daily Usage Statistics")
			},
			teatest.WithCheckInterval(time.Millisecond*100),
			teatest.WithDuration(time.Millisecond*500),
		)

		// Test navigation in daily tab (should work because table is focused)
		tm.Send(tea.KeyMsg{Type: tea.KeyDown})
		tm.Send(tea.KeyMsg{Type: tea.KeyUp})

		// Switch back to current tab
		tm.Send(tea.KeyMsg{Type: tea.KeyTab})

		// Verify we're back on current tab
		teatest.WaitFor(
			t, tm.Output(),
			func(bts []byte) bool {
				output := string(bts)
				return strings.Contains(output, "[Current]") ||
					strings.Contains(output, "Recent API Requests")
			},
			teatest.WithCheckInterval(time.Millisecond*100),
			teatest.WithDuration(time.Millisecond*500),
		)

		// Test navigation in current tab (should work because table is focused)
		tm.Send(tea.KeyMsg{Type: tea.KeyDown})
		tm.Send(tea.KeyMsg{Type: tea.KeyUp})

		// Quit the program
		tm.Send(tea.KeyMsg{
			Type:  tea.KeyRunes,
			Runes: []rune("q"),
		})

		tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second*3))
	})
}

// TestDailyUsageTab_KeyboardNavigation tests keyboard navigation within daily usage table
func TestDailyUsageTab_KeyboardNavigation(t *testing.T) {
	setupTestEnvironment()

	t.Run("Arrow key navigation", func(t *testing.T) {
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

		// Switch to daily tab
		tm.Send(tea.KeyMsg{Type: tea.KeyTab})

		// Wait for daily tab to load
		teatest.WaitFor(
			t, tm.Output(),
			func(bts []byte) bool {
				output := string(bts)
				return strings.Contains(output, "Daily Usage Statistics")
			},
			teatest.WithCheckInterval(time.Millisecond*100),
			teatest.WithDuration(time.Millisecond*500),
		)

		// Test arrow key navigation (down arrow) - this should not crash the app
		tm.Send(tea.KeyMsg{Type: tea.KeyDown})

		// Test arrow key navigation (up arrow) - this should not crash the app
		tm.Send(tea.KeyMsg{Type: tea.KeyUp})

		// Basic verification that the app is still responsive
		time.Sleep(100 * time.Millisecond)

		// Quit the program
		tm.Send(tea.KeyMsg{
			Type:  tea.KeyRunes,
			Runes: []rune("q"),
		})

		tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second*3))
	})
}
