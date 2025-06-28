package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
)

// APIRequest represents a single Claude Code API request
type APIRequest struct {
	SessionID           string
	Timestamp           time.Time
	Model               string
	InputTokens         int64
	OutputTokens        int64
	CacheReadTokens     int64
	CacheCreationTokens int64
	TotalTokens         int64
	CostUSD             float64
	DurationMS          int64
}

// TimeFilter represents available time filter options
type TimeFilter int

const (
	FilterAll TimeFilter = iota
	FilterHour
	FilterDay
	FilterWeek
	FilterMonth
)

// Model represents the state of our TUI application
type Model struct {
	requests             []APIRequest
	table                table.Model
	width                int
	height               int
	ready                bool
	requestChan          chan APIRequest
	totalRequests        int
	totalTokens          int64
	totalLimitedTokens   int64
	totalCacheTokens     int64
	totalCost            float64
	baseRequests         int
	baseTokens           int64
	baseLimitedTokens    int64
	baseCacheTokens      int64
	baseCost             float64
	premiumRequests      int
	premiumTokens        int64
	premiumLimitedTokens int64
	premiumCacheTokens   int64
	premiumCost          float64
	serverStatus         string
	db                   *Database
	timeFilter           TimeFilter
}

// NewModel creates a new Model with initial state
func NewModel(requestChan chan APIRequest, db *Database) Model {
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
		requests:     make([]APIRequest, 0),
		table:        t,
		requestChan:  requestChan,
		serverStatus: "Starting...",
		db:           db,
		timeFilter:   FilterAll,
	}
}

// Init is the Bubble Tea initialization function
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		waitForRequest(m.requestChan),
		tea.EnterAltScreen,
		m.refreshStats, // Load initial data from database
	)
}

// waitForRequest returns a command that waits for new API requests
func waitForRequest(requestChan chan APIRequest) tea.Cmd {
	return func() tea.Msg {
		return <-requestChan
	}
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

	case APIRequest:
		// Add new request to the beginning of the list
		m.requests = append([]APIRequest{msg}, m.requests...)
		// Keep only the last 100 requests
		if len(m.requests) > 100 {
			m.requests = m.requests[:100]
		}

		// Calculate limited tokens (input + output only)
		limitedTokens := msg.InputTokens + msg.OutputTokens
		cacheTokens := msg.CacheReadTokens + msg.CacheCreationTokens

		// Update statistics
		m.totalRequests++
		m.totalTokens += msg.TotalTokens
		m.totalLimitedTokens += limitedTokens
		m.totalCacheTokens += cacheTokens
		m.totalCost += msg.CostUSD

		// Update base or premium statistics
		if isBaseModel(msg.Model) {
			m.baseRequests++
			m.baseTokens += msg.TotalTokens
			m.baseLimitedTokens += limitedTokens
			m.baseCacheTokens += cacheTokens
			m.baseCost += msg.CostUSD
		} else {
			m.premiumRequests++
			m.premiumTokens += msg.TotalTokens
			m.premiumLimitedTokens += limitedTokens
			m.premiumCacheTokens += cacheTokens
			m.premiumCost += msg.CostUSD
		}

		// Update table rows
		m.updateTableRows()

		// Continue waiting for more requests
		return m, waitForRequest(m.requestChan)

	case serverStartedMsg:
		m.serverStatus = "Running on port 4317"

	case refreshStatsMsg:
		// Recalculate stats from database
		if m.db != nil {
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
			req.Timestamp.Format("15:04:05 2006-01-02"),
			truncateString(req.Model, 25),
			formatNumber(req.InputTokens),
			formatNumber(req.OutputTokens),
			formatNumber(req.CacheReadTokens + req.CacheCreationTokens),
			formatNumber(req.TotalTokens),
			formatCost(req.CostUSD),
			formatDuration(req.DurationMS),
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

// isBaseModel checks if the model is a base (Haiku) model
func isBaseModel(model string) bool {
	return strings.Contains(strings.ToLower(model), "haiku")
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

// getTimeRange returns the start and end time for the current filter
func (m Model) getTimeRange() (start, end time.Time) {
	end = time.Now()
	switch m.timeFilter {
	case FilterHour:
		start = end.Add(-1 * time.Hour)
	case FilterDay:
		start = end.Add(-24 * time.Hour)
	case FilterWeek:
		start = end.Add(-7 * 24 * time.Hour)
	case FilterMonth:
		start = end.Add(-30 * 24 * time.Hour)
	default:
		start = time.Time{} // Zero time for all records
	}
	return
}

// refreshStats returns a command to refresh statistics
func (m Model) refreshStats() tea.Msg {
	return refreshStatsMsg{}
}

// recalculateStats recalculates statistics from the database
func (m *Model) recalculateStats() {
	// Reset all stats
	m.totalRequests = 0
	m.totalTokens = 0
	m.totalLimitedTokens = 0
	m.totalCacheTokens = 0
	m.totalCost = 0
	m.baseRequests = 0
	m.baseTokens = 0
	m.baseLimitedTokens = 0
	m.baseCacheTokens = 0
	m.baseCost = 0
	m.premiumRequests = 0
	m.premiumTokens = 0
	m.premiumLimitedTokens = 0
	m.premiumCacheTokens = 0
	m.premiumCost = 0

	// Get time range
	start, end := m.getTimeRange()

	// Query database
	var requests []APIRequest
	var err error

	if m.timeFilter == FilterAll {
		requests, err = m.db.GetAllRequests()
	} else {
		requests, err = m.db.QueryTimeRange(start, end)
	}

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
	baseReqs, premiumReqs, baseTokens, premiumTokens, baseLimited, premiumLimited, baseCache, premiumCache, baseCost, premiumCost := CalculateStats(requests)

	m.baseRequests = baseReqs
	m.premiumRequests = premiumReqs
	m.totalRequests = baseReqs + premiumReqs

	m.baseTokens = baseTokens
	m.premiumTokens = premiumTokens
	m.totalTokens = baseTokens + premiumTokens

	m.baseLimitedTokens = baseLimited
	m.premiumLimitedTokens = premiumLimited
	m.totalLimitedTokens = baseLimited + premiumLimited

	m.baseCacheTokens = baseCache
	m.premiumCacheTokens = premiumCache
	m.totalCacheTokens = baseCache + premiumCache

	m.baseCost = baseCost
	m.premiumCost = premiumCost
	m.totalCost = baseCost + premiumCost

	// Update table
	m.updateTableRows()
}

// Message types
type serverStartedMsg struct{}
type refreshStatsMsg struct{}
