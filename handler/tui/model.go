package tui

import (
	"context"
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/elct9620/ccmon/entity"
	"github.com/elct9620/ccmon/usecase"
)

// TimeFilter represents the available time filter options for UI
type TimeFilter int

const (
	FilterAll TimeFilter = iota
	FilterHour
	FilterDay
	FilterWeek
	FilterMonth
)

// Model represents the state of our TUI monitor application
type Model struct {
	requests         []entity.APIRequest
	table            table.Model
	width            int
	height           int
	ready            bool
	stats            entity.Stats
	getFilteredQuery *usecase.GetFilteredApiRequestsQuery
	timeFilter       TimeFilter
}

// NewModel creates a new Model with initial state
func NewModel(getFilteredQuery *usecase.GetFilteredApiRequestsQuery) Model {
	columns := []table.Column{
		{Title: "Time", Width: 20},
		{Title: "Model", Width: 25},
		{Title: "Input", Width: 8},
		{Title: "Output", Width: 8},
		{Title: "Cache", Width: 8},
		{Title: "Total", Width: 8},
		{Title: "Cost ($)", Width: 10},
		{Title: "Duration", Width: 10},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.Bold(true)
	s.Selected = s.Selected.Bold(false)
	t.SetStyles(s)

	return Model{
		requests:         []entity.APIRequest{},
		table:            t,
		getFilteredQuery: getFilteredQuery,
		timeFilter:       FilterAll,
		stats:            entity.Stats{},
	}
}

// Init is the Bubble Tea initialization function
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		tea.EnterAltScreen,
		m.refreshStats, // Load initial data from database
		tick(),         // Start periodic refresh
	)
}

// tick returns a command that sends a tick message every 5 seconds
func tick() tea.Cmd {
	return tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "esc":
			if m.table.Focused() {
				m.table.Blur()
			} else {
				m.table.Focus()
			}
		case "a":
			m.timeFilter = FilterAll
			return m, m.refreshStats
		case "h":
			m.timeFilter = FilterHour
			return m, m.refreshStats
		case "d":
			m.timeFilter = FilterDay
			return m, m.refreshStats
		case "w":
			m.timeFilter = FilterWeek
			return m, m.refreshStats
		case "m":
			m.timeFilter = FilterMonth
			return m, m.refreshStats
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		// Adjust table height based on window size
		tableHeight := m.height - 17 // Leave room for header, stats table, and footer
		if tableHeight > 0 {
			m.table.SetHeight(tableHeight)
		}

	case tickMsg:
		// Periodic refresh
		return m, tea.Batch(tick(), m.refreshStats)

	case refreshStatsMsg:
		// Recalculate stats via usecase
		if m.getFilteredQuery != nil {
			m.recalculateStats()
		}
	}

	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

// updateTableRows updates the table with the current requests
func (m *Model) updateTableRows() {
	rows := make([]table.Row, 0, len(m.requests))
	for _, req := range m.requests {
		rows = append(rows, table.Row{
			req.Timestamp().Format("15:04:05 2006-01-02"),
			truncateString(req.Model().String(), 25),
			formatNumber(req.Tokens().Input()),
			formatNumber(req.Tokens().Output()),
			formatNumber(req.Tokens().Cache()),
			formatNumber(req.Tokens().Total()),
			formatCost(req.Cost().Amount()),
			formatDuration(req.DurationMS()),
		})
	}
	m.table.SetRows(rows)
}

// Helper functions
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func formatNumber(n int64) string {
	if n == 0 {
		return "-"
	}
	return fmt.Sprintf("%d", n)
}

func formatCost(cost float64) string {
	if cost == 0 {
		return "-"
	}
	return fmt.Sprintf("%.6f", cost)
}

func formatDuration(ms int64) string {
	if ms < 1000 {
		return fmt.Sprintf("%dms", ms)
	}
	return fmt.Sprintf("%.1fs", float64(ms)/1000)
}

// getTimeFilterString returns a string representation of the current time filter
func (m Model) getTimeFilterString() string {
	switch m.timeFilter {
	case FilterHour:
		return "Last Hour"
	case FilterDay:
		return "Last 24 Hours"
	case FilterWeek:
		return "Last 7 Days"
	case FilterMonth:
		return "Last 30 Days"
	default:
		return "All Time"
	}
}

// getTimePeriod returns entity.Period for the current filter
func (m Model) getTimePeriod() entity.Period {
	switch m.timeFilter {
	case FilterHour:
		return entity.NewPeriodFromDuration(time.Hour)
	case FilterDay:
		return entity.NewPeriodFromDuration(24 * time.Hour)
	case FilterWeek:
		return entity.NewPeriodFromDuration(7 * 24 * time.Hour)
	case FilterMonth:
		return entity.NewPeriodFromDuration(30 * 24 * time.Hour)
	default:
		return entity.NewAllTimePeriod()
	}
}

// refreshStats returns a command to refresh statistics
func (m Model) refreshStats() tea.Msg {
	return refreshStatsMsg{}
}

// recalculateStats recalculates statistics via usecase
func (m *Model) recalculateStats() {
	// Query via usecase with entity.Period
	period := m.getTimePeriod()
	params := usecase.GetFilteredApiRequestsParams{Period: period}
	requests, err := m.getFilteredQuery.Execute(context.Background(), params)
	if err != nil {
		// Handle error silently for now
		return
	}

	// Update display requests (show latest 100)
	if len(requests) > 100 {
		m.requests = requests[len(requests)-100:]
	} else {
		m.requests = requests
	}

	// Calculate stats
	m.stats = entity.CalculateStats(requests)

	// Update table
	m.updateTableRows()
}

// Message types
type tickMsg time.Time
type refreshStatsMsg struct{}
