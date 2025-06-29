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
	FilterBlock // Current block timeframe
)

// SortOrder represents the sorting order for requests
type SortOrder int

const (
	SortDescending SortOrder = iota // Latest first (default)
	SortAscending                   // Oldest first
)

// Model represents the state of our TUI monitor application
type Model struct {
	requests           []entity.APIRequest
	table              table.Model
	width              int
	height             int
	ready              bool
	stats              entity.Stats
	getFilteredQuery   *usecase.GetFilteredApiRequestsQuery
	getBlockStatsQuery *usecase.GetBlockStatsQuery
	timeFilter         TimeFilter
	sortOrder          SortOrder
	timezone           *time.Location
	block              *entity.Block // nil if no block configured
	tokenLimit         int           // token limit for current block
}

// NewModel creates a new Model with initial state
func NewModel(getFilteredQuery *usecase.GetFilteredApiRequestsQuery, getBlockStatsQuery *usecase.GetBlockStatsQuery, timezone *time.Location, block *entity.Block, tokenLimit int) Model {
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
		requests:           []entity.APIRequest{},
		table:              t,
		getFilteredQuery:   getFilteredQuery,
		getBlockStatsQuery: getBlockStatsQuery,
		timeFilter:         FilterAll,
		sortOrder:          SortDescending, // Default to latest first
		stats:              entity.Stats{},
		timezone:           timezone,
		block:              block,
		tokenLimit:         tokenLimit,
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
		case "b":
			if m.block != nil {
				m.timeFilter = FilterBlock
				return m, m.refreshStats
			}
			// If no block configured, ignore the key press
		case "o":
			// Toggle sort order
			if m.sortOrder == SortDescending {
				m.sortOrder = SortAscending
			} else {
				m.sortOrder = SortDescending
			}
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
		// Format timestamp in configured timezone
		timestamp := req.Timestamp().In(m.timezone).Format("15:04:05 2006-01-02")
		rows = append(rows, table.Row{
			timestamp,
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
	case FilterBlock:
		if m.block != nil {
			return "Current Block (" + m.block.FormatBlockTime(time.Now()) + ")"
		}
		return "Block (not configured)"
	default:
		return "All Time"
	}
}

// getSortOrderString returns a string representation of the current sort order
func (m Model) getSortOrderString() string {
	switch m.sortOrder {
	case SortDescending:
		return "Latest First"
	case SortAscending:
		return "Oldest First"
	default:
		return "Latest First"
	}
}

// getTimePeriod returns entity.Period for the current filter using configured timezone
func (m Model) getTimePeriod() entity.Period {
	switch m.timeFilter {
	case FilterHour:
		return entity.NewPeriodFromDurationWithTimezone(time.Hour, m.timezone)
	case FilterDay:
		return entity.NewPeriodFromDurationWithTimezone(24*time.Hour, m.timezone)
	case FilterWeek:
		return entity.NewPeriodFromDurationWithTimezone(7*24*time.Hour, m.timezone)
	case FilterMonth:
		return entity.NewPeriodFromDurationWithTimezone(30*24*time.Hour, m.timezone)
	case FilterBlock:
		if m.block != nil {
			return m.block.CurrentBlock(time.Now())
		}
		return entity.NewAllTimePeriod()
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
	period := m.getTimePeriod()

	// Query for display requests (limit to 100 for TUI display)
	displayParams := usecase.GetFilteredApiRequestsParams{
		Period: period,
		Limit:  100,
		Offset: 0,
	}
	requests, err := m.getFilteredQuery.Execute(context.Background(), displayParams)
	if err != nil {
		// Handle error silently for now
		return
	}
	m.requests = requests

	// Apply sorting based on user preference
	if m.sortOrder == SortDescending {
		// Reverse to show latest first (since DB returns chronological order)
		reverseRequests(m.requests)
	}
	// For SortAscending, keep the original order (oldest first)

	// Calculate stats: Use block stats for progress bars, or filtered stats for display
	if m.block != nil && m.getBlockStatsQuery != nil {
		// For block tracking: Always use current block stats regardless of display filter
		blockStatsParams := usecase.GetBlockStatsParams{
			Block: *m.block,
		}
		blockStats, err := m.getBlockStatsQuery.Execute(context.Background(), blockStatsParams)
		if err != nil {
			// Fallback to filtered stats calculation
			m.stats = entity.CalculateStats(m.requests)
		} else {
			// Use block stats directly - progress calculation will be handled in UI layer
			m.stats = blockStats
		}
	} else {
		// For non-block mode: Calculate stats from filtered requests
		statsParams := usecase.GetFilteredApiRequestsParams{
			Period: period,
			Limit:  0, // No limit for stats calculation
			Offset: 0,
		}
		allRequests, err := m.getFilteredQuery.Execute(context.Background(), statsParams)
		if err != nil {
			// Handle error silently for now, use display requests for stats
			m.stats = entity.CalculateStats(m.requests)
		} else {
			m.stats = entity.CalculateStats(allRequests)
		}
	}

	// Update table
	m.updateTableRows()
}

// reverseRequests reverses the order of API requests slice
func reverseRequests(requests []entity.APIRequest) {
	for i, j := 0, len(requests)-1; i < j; i, j = i+1, j-1 {
		requests[i], requests[j] = requests[j], requests[i]
	}
}

// Message types
type tickMsg time.Time
type refreshStatsMsg struct{}
