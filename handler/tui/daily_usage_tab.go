package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/elct9620/ccmon/entity"
	"github.com/elct9620/ccmon/usecase"
)

// DailyUsageTabModel handles the daily usage tab that shows usage statistics over time and owns its data
type DailyUsageTabModel struct {
	// Data ownership
	usage entity.Usage
	table table.Model

	// Configuration
	timezone *time.Location
	width    int
	height   int

	// Business logic dependencies
	getUsageQuery *usecase.GetUsageQuery
}

// NewDailyUsageTabModel creates a new daily usage tab model with usecase dependency
func NewDailyUsageTabModel(getUsageQuery *usecase.GetUsageQuery, timezone *time.Location) *DailyUsageTabModel {
	// Initialize table columns
	columns := []table.Column{
		{Title: "Date", Width: 10},
		{Title: "Requests", Width: 10},
		{Title: "Input", Width: 8},
		{Title: "Output", Width: 8},
		{Title: "Read Cache", Width: 10},
		{Title: "Creation Cache", Width: 12},
		{Title: "Total", Width: 8},
		{Title: "Premium Cost ($)", Width: 16},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(false), // Daily tab doesn't need focus by default
		table.WithHeight(10),     // Will be adjusted based on terminal size
	)

	// Set table styles
	s := table.DefaultStyles()
	s.Header = s.Header.Bold(true)
	s.Selected = s.Selected.Bold(false)
	t.SetStyles(s)

	return &DailyUsageTabModel{
		usage:         entity.Usage{},
		table:         t,
		timezone:      timezone,
		width:         120,
		height:        30,
		getUsageQuery: getUsageQuery,
	}
}

// Init initializes the daily usage tab model
func (m *DailyUsageTabModel) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model
func (m *DailyUsageTabModel) Update(msg tea.Msg) (ComponentModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case ResizeMsg:
		m.SetSize(msg.Width, msg.Height)
	case UsageRefreshMsg:
		return m, m.refreshUsage()
	case UsageDataMsg:
		m.usage = msg.Usage
		m.updateTableRows()
	case tea.KeyMsg:
		// Handle table navigation
		m.table, cmd = m.table.Update(msg)
	}
	return m, cmd
}

// View renders the daily usage tab
func (m *DailyUsageTabModel) View() string {
	var b strings.Builder

	// Daily usage header
	dailyHeader := HeaderStyle.Render("Daily Usage Statistics (Last 30 Days)")
	b.WriteString(dailyHeader + "\n")

	// Subtitle explaining premium token focus
	subtitle := HelpStyle.Render("Premium Token Breakdown (Base tokens are free and not shown)")
	b.WriteString(subtitle + "\n")

	// Legend explaining column meanings
	legend := HelpStyle.Render("Requests: Base/Premium • Tokens: Premium only (Sonnet/Opus)")
	b.WriteString(legend + "\n\n")

	// Check if we have data
	if len(m.usage.GetStats()) == 0 {
		emptyContent := HelpStyle.Render("No usage data available")
		dailyBox := BoxStyle.Width(m.width - 4).Render(emptyContent)
		b.WriteString(dailyBox + "\n")
		return b.String()
	}

	// Daily usage table - now using table.Model
	dailyBox := BoxStyle.Width(m.width - 4).Render(m.table.View())
	b.WriteString(dailyBox + "\n")

	return b.String()
}

// SetSize updates the size of the daily usage tab
func (m *DailyUsageTabModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.resizeTableColumns()
	m.adjustTableHeight()
}

// UpdateUsage updates the usage data
func (m *DailyUsageTabModel) UpdateUsage(usage entity.Usage) {
	m.usage = usage
	m.updateTableRows()
}

// calculateDailyTableWidths calculates column widths for daily usage table
func (m *DailyUsageTabModel) calculateDailyTableWidths(availableWidth int) []int {
	// Account for table internal spacing - Bubble Tea table adds padding/borders
	// Estimate ~2-3 chars per column for internal spacing/borders
	tableOverhead := 8 * 3 // 8 columns * 3 chars overhead each
	usableWidth := availableWidth - tableOverhead

	// Ensure we have minimum usable width
	if usableWidth < 60 {
		usableWidth = 60
	}

	// Date: 10 (2025-06-30), Requests: 10 (999/999), Input: 8, Output: 8, Read Cache: 10, Creation Cache: 12, Total: 8, Premium Cost: remaining
	dateWidth := 10
	requestsWidth := 10
	inputWidth := 8
	outputWidth := 8
	readCacheWidth := 10
	creationCacheWidth := 12
	totalWidth := 8
	costWidth := usableWidth - dateWidth - requestsWidth - inputWidth - outputWidth - readCacheWidth - creationCacheWidth - totalWidth

	if costWidth < 12 {
		costWidth = 12
	}

	return []int{dateWidth, requestsWidth, inputWidth, outputWidth, readCacheWidth, creationCacheWidth, totalWidth, costWidth}
}

// refreshUsage handles data fetching for the daily usage model
func (m *DailyUsageTabModel) refreshUsage() tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		if m.getUsageQuery == nil {
			return UsageDataMsg{Usage: entity.Usage{}}
		}

		// Fetch daily usage statistics (last 30 days)
		usage, err := m.getUsageQuery.ListByDay(context.Background(), 30, m.timezone)
		if err != nil {
			usage = entity.Usage{}
		}

		return UsageDataMsg{Usage: usage}
	})
}

// Usage returns the current usage (for compatibility)
func (m *DailyUsageTabModel) Usage() entity.Usage {
	return m.usage
}

// Focus sets focus on the table
func (m *DailyUsageTabModel) Focus() {
	m.table.Focus()
}

// Blur removes focus from the table
func (m *DailyUsageTabModel) Blur() {
	m.table.Blur()
}

// Focused returns whether the table is focused
func (m *DailyUsageTabModel) Focused() bool {
	return m.table.Focused()
}

// adjustTableHeight calculates and sets appropriate table height
func (m *DailyUsageTabModel) adjustTableHeight() {
	// Fixed height components:
	// - Header: 1 line (Daily Usage Statistics)
	// - Subtitle: 1 line (Premium Token Breakdown)
	// - Legend: 1 line (Requests: Base/Premium...)
	// - Empty lines: 2 lines
	// - Box borders: 2 lines
	// - Safety margin: 2 lines
	fixedHeight := 9

	// Calculate remaining height for table
	tableHeight := m.height - fixedHeight

	// Ensure reasonable minimum and maximum
	if tableHeight < 5 {
		tableHeight = 5
	} else if tableHeight > 30 {
		tableHeight = 30 // Cap maximum table height to show all 30 days
	}

	m.table.SetHeight(tableHeight)
}

// resizeTableColumns resizes table columns based on available width
func (m *DailyUsageTabModel) resizeTableColumns() {
	// Calculate available width for table (accounting for box padding)
	availableWidth := m.width - 6

	var columns []table.Column
	if availableWidth < 60 {
		// Compact layout for very narrow terminals
		columns = []table.Column{
			{Title: "Date", Width: 10},
			{Title: "Reqs", Width: 8},
			{Title: "Tokens", Width: 12},
			{Title: "Cost", Width: 12},
		}
	} else {
		// Calculate column widths
		colWidths := m.calculateDailyTableWidths(availableWidth)
		columns = []table.Column{
			{Title: "Date", Width: colWidths[0]},
			{Title: "Requests", Width: colWidths[1]},
			{Title: "Input", Width: colWidths[2]},
			{Title: "Output", Width: colWidths[3]},
			{Title: "Read Cache", Width: colWidths[4]},
			{Title: "Creation Cache", Width: colWidths[5]},
			{Title: "Total", Width: colWidths[6]},
			{Title: "Premium Cost ($)", Width: colWidths[7]},
		}
	}

	m.table.SetColumns(columns)
	m.updateTableRows() // Update rows to match new column structure
}

// updateTableRows updates the table rows based on current usage data
func (m *DailyUsageTabModel) updateTableRows() {
	stats := m.usage.GetStats()
	rows := make([]table.Row, 0, len(stats))

	for _, stat := range stats {
		period := stat.Period()
		if period.IsAllTime() {
			continue // Skip all-time periods
		}

		date := period.StartAt().In(m.timezone).Format("2006-01-02")

		// Check if we're in compact mode
		availableWidth := m.width - 6
		if availableWidth < 60 {
			// Compact mode: combine data
			requests := fmt.Sprintf("%d/%d", stat.BaseRequests(), stat.PremiumRequests())
			totalTokens := FormatTokenCount(stat.PremiumTokens().Total())
			cost := fmt.Sprintf("%.4f", stat.PremiumCost().Amount())

			rows = append(rows, table.Row{date, requests, totalTokens, cost})
		} else {
			// Normal mode: full columns
			requests := fmt.Sprintf("%d/%d", stat.BaseRequests(), stat.PremiumRequests())
			input := FormatTokenCount(stat.PremiumTokens().Input())
			output := FormatTokenCount(stat.PremiumTokens().Output())
			readCache := FormatTokenCount(stat.PremiumTokens().CacheRead())
			creationCache := FormatTokenCount(stat.PremiumTokens().CacheCreation())
			total := FormatTokenCount(stat.PremiumTokens().Total())
			cost := fmt.Sprintf("%.6f", stat.PremiumCost().Amount())

			rows = append(rows, table.Row{date, requests, input, output, readCache, creationCache, total, cost})
		}
	}

	m.table.SetRows(rows)
}

// Message types for DailyUsageTabModel
type UsageRefreshMsg struct{}

type UsageDataMsg struct {
	Usage entity.Usage
}
