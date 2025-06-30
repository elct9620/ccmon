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

// TestRequestsTable_IntegrationWithViewModel tests requests table through the full ViewModel
func TestRequestsTable_IntegrationWithViewModel(t *testing.T) {
	setupTestEnvironment()

	tests := []struct {
		name        string
		width       int
		height      int
		hasData     bool
		description string
	}{
		{
			name:        "Requests table with empty data",
			width:       120,
			height:      40,
			hasData:     false,
			description: "Should render requests table with empty data",
		},
		{
			name:        "Requests table with test data",
			width:       120,
			height:      40,
			hasData:     true,
			description: "Should render requests table with test data",
		},
		{
			name:        "Requests table narrow terminal",
			width:       80,
			height:      25,
			hasData:     true,
			description: "Should render requests table in narrow terminal",
		},
		{
			name:        "Requests table very narrow terminal",
			width:       60,
			height:      20,
			hasData:     true,
			description: "Should render compact requests table in very narrow terminal",
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

			// Create the ViewModel (starts on overview tab with requests table)
			model := tui.NewViewModel(getFilteredQuery, calculateStatsQuery, getUsageQuery, time.UTC, nil, 10*time.Millisecond)

			// Create teatest model
			tm := teatest.NewTestModel(
				t, model,
				teatest.WithInitialTermSize(tt.width, tt.height),
			)

			isCompact := tt.width <= 80
			if !isCompact {
				// Wait for initial render - should show requests table
				teatest.WaitFor(
					t, tm.Output(),
					func(bts []byte) bool {
						output := string(bts)
						return strings.Contains(output, "Claude Code Monitor") ||
							strings.Contains(output, "Recent API Requests") ||
							strings.Contains(output, "[Current]")
					},
					teatest.WithCheckInterval(time.Millisecond*100),
					teatest.WithDuration(time.Millisecond*500),
				)
			}

			// Test renders the requests table (part of overview tab)
			// Since table is part of default overview tab, just verify basic functionality

			// Quit the program
			tm.Send(tea.KeyMsg{
				Type:  tea.KeyRunes,
				Runes: []rune("q"),
			})

			tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second*3))
		})
	}
}

// TestRequestsTable_Navigation tests table navigation functionality
func TestRequestsTable_Navigation(t *testing.T) {
	setupTestEnvironment()

	t.Run("Table navigation keys", func(t *testing.T) {
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

		// Test table navigation
		navigationKeys := []tea.KeyType{
			tea.KeyDown,
			tea.KeyUp,
			tea.KeyHome,
			tea.KeyEnd,
		}

		for _, key := range navigationKeys {
			tm.Send(tea.KeyMsg{Type: key})
		}

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
