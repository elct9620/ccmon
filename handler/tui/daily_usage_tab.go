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

	// Display mode configuration
	displayMode DailyDisplayMode

	// Business logic dependencies
	getUsageQuery *usecase.GetUsageQuery
}

// DailyDisplayMode defines the table display mode based on available width
type DailyDisplayMode int

const (
	// FullMode shows all 9 columns with complete token breakdown
	FullMode DailyDisplayMode = iota
	// GroupedMode shows 4 main columns with grouped token details
	GroupedMode
	// CompactMode shows 4 main columns with simplified token display
	CompactMode
)

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
		{Title: "Burn Rate", Width: 10},
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
		displayMode:   FullMode,
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
	tableOverhead := 9 * 3 // 9 columns * 3 chars overhead each
	usableWidth := availableWidth - tableOverhead

	// Ensure we have minimum usable width for full mode
	if usableWidth < 100 {
		usableWidth = 100
	}

	// Calculate proportional widths with better distribution
	// Base widths: Date: 10, Requests: 10, Input: 8, Output: 8, Read Cache: 9, Creation Cache: 11, Total: 8, Burn Rate: 10, Premium Cost: remaining
	baseWidths := []int{10, 10, 8, 8, 9, 11, 8, 10, 14} // Total base: 88
	totalBaseWidth := 0
	for _, w := range baseWidths {
		totalBaseWidth += w
	}

	// If we have extra space, distribute it proportionally
	if usableWidth > totalBaseWidth {
		extraSpace := usableWidth - totalBaseWidth
		// Distribute extra space: favor Premium Cost (40%), Date (15%), Burn Rate (15%), others get smaller amounts
		distribution := []float64{0.15, 0.08, 0.06, 0.06, 0.06, 0.08, 0.06, 0.15, 0.30}

		for i := range baseWidths {
			extra := int(float64(extraSpace) * distribution[i])
			baseWidths[i] += extra
		}
	}

	// Ensure minimum widths
	if baseWidths[8] < 12 { // Premium Cost minimum
		baseWidths[8] = 12
	}

	return baseWidths
}

// calculateGroupedTableWidths calculates column widths for grouped display mode
func (m *DailyUsageTabModel) calculateGroupedTableWidths(availableWidth int) []int {
	// Account for table internal spacing
	tableOverhead := 4 * 3 // 4 columns * 3 chars overhead each
	usableWidth := availableWidth - tableOverhead

	// Ensure minimum usable width
	if usableWidth < 40 {
		usableWidth = 40
	}

	// Base widths for grouped mode: Date, B/P Reqs, Burn Rate, Cost
	baseWidths := []int{12, 10, 12, 12} // Total: 46
	totalBaseWidth := 0
	for _, w := range baseWidths {
		totalBaseWidth += w
	}

	// Distribute extra space if available
	if usableWidth > totalBaseWidth {
		extraSpace := usableWidth - totalBaseWidth
		// Distribute: Date 30%, B/P Reqs 20%, Burn Rate 25%, Cost 25%
		distribution := []float64{0.30, 0.20, 0.25, 0.25}

		for i := range baseWidths {
			extra := int(float64(extraSpace) * distribution[i])
			baseWidths[i] += extra
		}
	}

	// Ensure minimum widths
	if baseWidths[0] < 10 { // Date minimum
		baseWidths[0] = 10
	}
	if baseWidths[1] < 8 { // B/P Reqs minimum
		baseWidths[1] = 8
	}
	if baseWidths[2] < 10 { // Burn Rate minimum
		baseWidths[2] = 10
	}
	if baseWidths[3] < 10 { // Cost minimum
		baseWidths[3] = 10
	}

	return baseWidths
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
	// - Safety margin: 3 lines (increased for better header visibility)
	fixedHeight := 10

	// Calculate remaining height for table
	tableHeight := m.height - fixedHeight

	// More conservative minimum to ensure headers stay visible
	if tableHeight < 3 {
		tableHeight = 3
	} else if tableHeight > 25 {
		tableHeight = 25 // Reduced max to leave more space for headers
	}

	// For very small screens, be even more conservative
	if m.height < 20 {
		tableHeight = max(2, m.height-12) // Leave even more space for headers
	}

	m.table.SetHeight(tableHeight)
}

// Helper function for max
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// resizeTableColumns resizes table columns based on available width
func (m *DailyUsageTabModel) resizeTableColumns() {
	// Calculate available width for table (accounting for box padding)
	availableWidth := m.width - 6

	// Determine display mode based on available width
	var newDisplayMode DailyDisplayMode
	var columns []table.Column

	if availableWidth >= 140 {
		// Full mode: traditional 9-column layout
		newDisplayMode = FullMode
		colWidths := m.calculateDailyTableWidths(availableWidth)
		columns = []table.Column{
			{Title: "Date", Width: colWidths[0]},
			{Title: "Requests", Width: colWidths[1]},
			{Title: "Input", Width: colWidths[2]},
			{Title: "Output", Width: colWidths[3]},
			{Title: "Read Cache", Width: colWidths[4]},
			{Title: "Creation Cache", Width: colWidths[5]},
			{Title: "Total", Width: colWidths[6]},
			{Title: "Burn Rate", Width: colWidths[7]},
			{Title: "Premium Cost ($)", Width: colWidths[8]},
		}
	} else if availableWidth >= 80 {
		// Grouped mode: 4 main columns with token details in sub-rows
		newDisplayMode = GroupedMode
		colWidths := m.calculateGroupedTableWidths(availableWidth)
		columns = []table.Column{
			{Title: "Date", Width: colWidths[0]},
			{Title: "B/P Reqs", Width: colWidths[1]},
			{Title: "Burn Rate", Width: colWidths[2]},
			{Title: "Cost ($)", Width: colWidths[3]},
		}
	} else {
		// Compact mode: 4 simplified columns
		newDisplayMode = CompactMode
		columns = []table.Column{
			{Title: "Date", Width: 10},
			{Title: "Reqs", Width: 8},
			{Title: "Rate/min", Width: 12},
			{Title: "Cost", Width: 10},
		}
	}

	// Update display mode if changed
	m.displayMode = newDisplayMode

	// Clear rows before setting new columns to avoid index out of range
	m.table.SetRows([]table.Row{})
	m.table.SetColumns(columns)
	m.updateTableRows() // Update rows to match new column structure
}

// updateTableRows updates the table rows based on current usage data
func (m *DailyUsageTabModel) updateTableRows() {
	stats := m.usage.GetStats()
	rows := make([]table.Row, 0, len(stats)*2) // Pre-allocate for potential sub-rows

	for _, stat := range stats {
		period := stat.Period()
		if period.IsAllTime() {
			continue // Skip all-time periods
		}

		date := period.StartAt().In(m.timezone).Format("2006-01-02")
		rows = append(rows, m.createRowsForStat(stat, date)...)
	}

	m.table.SetRows(rows)
}

// createRowsForStat creates table rows for a single stat based on display mode
func (m *DailyUsageTabModel) createRowsForStat(stat entity.Stats, date string) []table.Row {
	switch m.displayMode {
	case FullMode:
		// Traditional 9-column layout
		requests := fmt.Sprintf("%d/%d", stat.BaseRequests(), stat.PremiumRequests())
		input := FormatTokenCount(stat.PremiumTokens().Input())
		output := FormatTokenCount(stat.PremiumTokens().Output())
		readCache := FormatTokenCount(stat.PremiumTokens().CacheRead())
		creationCache := FormatTokenCount(stat.PremiumTokens().CacheCreation())
		total := FormatTokenCount(stat.PremiumTokens().Total())
		burnRate := FormatBurnRate(stat.PremiumTokenBurnRate())
		cost := fmt.Sprintf("%.6f", stat.PremiumCost().Amount())
		return []table.Row{{date, requests, input, output, readCache, creationCache, total, burnRate, cost}}

	case GroupedMode:
		// 4 main columns with token details in sub-rows
		requests := fmt.Sprintf("%d/%d", stat.BaseRequests(), stat.PremiumRequests())
		burnRate := FormatBurnRate(stat.PremiumTokenBurnRate())
		cost := fmt.Sprintf("%.4f", stat.PremiumCost().Amount())

		// Main row
		mainRow := table.Row{date, requests, burnRate, cost}

		// Token detail sub-rows (formatted to show grouping)
		input := FormatTokenCount(stat.PremiumTokens().Input())
		output := FormatTokenCount(stat.PremiumTokens().Output())
		readCache := FormatTokenCount(stat.PremiumTokens().CacheRead())
		creationCache := FormatTokenCount(stat.PremiumTokens().CacheCreation())

		// Create grouped token display in second column
		tokenDetails := fmt.Sprintf("├─I:%s O:%s", input, output)
		cacheDetails := fmt.Sprintf("└─CR:%s CC:%s", readCache, creationCache)

		subRow1 := table.Row{"", tokenDetails, "", ""}
		subRow2 := table.Row{"", cacheDetails, "", ""}

		return []table.Row{mainRow, subRow1, subRow2}

	case CompactMode:
		// 4 simplified columns
		requests := fmt.Sprintf("%d/%d", stat.BaseRequests(), stat.PremiumRequests())
		burnRate := FormatBurnRate(stat.PremiumTokenBurnRate())
		cost := fmt.Sprintf("%.3f", stat.PremiumCost().Amount())
		return []table.Row{{date, requests, burnRate, cost}}

	default:
		// Fallback
		requests := fmt.Sprintf("%d/%d", stat.BaseRequests(), stat.PremiumRequests())
		burnRate := FormatBurnRate(stat.PremiumTokenBurnRate())
		return []table.Row{{date, requests, burnRate, "-"}}
	}
}

// Message types for DailyUsageTabModel
type UsageRefreshMsg struct{}

type UsageDataMsg struct {
	Usage entity.Usage
}
