package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/elct9620/ccmon/entity"
	"github.com/elct9620/ccmon/usecase"
)

// DailyUsageTabModel handles the daily usage tab that shows usage statistics over time and owns its data
type DailyUsageTabModel struct {
	// Data ownership
	usage entity.Usage
	
	// Configuration
	timezone *time.Location
	width    int
	height   int
	
	// Business logic dependencies
	getUsageQuery *usecase.GetUsageQuery
}

// NewDailyUsageTabModel creates a new daily usage tab model with usecase dependency
func NewDailyUsageTabModel(getUsageQuery *usecase.GetUsageQuery, timezone *time.Location) *DailyUsageTabModel {
	return &DailyUsageTabModel{
		usage:         entity.Usage{},
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
	switch msg := msg.(type) {
	case ResizeMsg:
		m.SetSize(msg.Width, msg.Height)
	case UsageRefreshMsg:
		return m, m.refreshUsage()
	case UsageDataMsg:
		m.usage = msg.Usage
	}
	return m, nil
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

	// Daily usage table
	dailyContent := m.renderDailyUsageTable()
	dailyBox := BoxStyle.Width(m.width - 4).Render(dailyContent)
	b.WriteString(dailyBox + "\n")

	return b.String()
}

// SetSize updates the size of the daily usage tab
func (m *DailyUsageTabModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// UpdateUsage updates the usage data
func (m *DailyUsageTabModel) UpdateUsage(usage entity.Usage) {
	m.usage = usage
}

// renderDailyUsageTable renders the daily usage statistics as a table
func (m *DailyUsageTabModel) renderDailyUsageTable() string {
	var b strings.Builder

	stats := m.usage.GetStats()

	if len(stats) == 0 {
		b.WriteString(HelpStyle.Render("No usage data available"))
		return b.String()
	}

	// Calculate available width for table
	availableWidth := m.width - 6 // Account for box padding
	if availableWidth < 60 {
		return m.renderCompactDailyUsage(stats)
	}

	// Table headers
	headers := []string{"Date", "Requests", "Input", "Output", "Read Cache", "Creation Cache", "Total", "Premium Cost ($)"}
	colWidths := m.calculateDailyTableWidths(availableWidth)

	// Render header row
	for i, header := range headers {
		cell := TableHeaderStyle.Render(PadRight(header, colWidths[i]))
		b.WriteString(cell)
		if i < len(headers)-1 {
			b.WriteString(" ") // Add space between columns
		}
	}
	b.WriteString("\n")

	// Separator line
	for i, width := range colWidths {
		b.WriteString(strings.Repeat("─", width))
		if i < len(colWidths)-1 {
			b.WriteString(" ") // Add space between separator lines
		}
	}
	b.WriteString("\n")

	// Data rows
	for _, stat := range stats {
		period := stat.Period()
		if period.IsAllTime() {
			continue // Skip all-time periods
		}

		date := period.StartAt().In(m.timezone).Format("2006-01-02")
		requests := fmt.Sprintf("%d/%d", stat.BaseRequests(), stat.PremiumRequests())
		input := FormatTokenCount(stat.PremiumTokens().Input())
		output := FormatTokenCount(stat.PremiumTokens().Output())
		readCache := FormatTokenCount(stat.PremiumTokens().CacheRead())
		creationCache := FormatTokenCount(stat.PremiumTokens().CacheCreation())
		total := FormatTokenCount(stat.PremiumTokens().Total())
		cost := fmt.Sprintf("%.6f", stat.PremiumCost().Amount())

		row := []string{date, requests, input, output, readCache, creationCache, total, cost}
		for i, cell := range row {
			b.WriteString(PadRight(cell, colWidths[i]))
			if i < len(row)-1 {
				b.WriteString(" ") // Add space between columns
			}
		}
		b.WriteString("\n")
	}

	return b.String()
}

// renderCompactDailyUsage renders compact daily usage for narrow terminals
func (m *DailyUsageTabModel) renderCompactDailyUsage(stats []entity.Stats) string {
	var b strings.Builder

	for _, stat := range stats {
		period := stat.Period()
		if period.IsAllTime() {
			continue
		}

		// Convert UTC time back to user's timezone for display
		date := period.StartAt().In(m.timezone).Format("2006-01-02")
		b.WriteString(StatStyle.Render(date))
		b.WriteString(fmt.Sprintf(": %d/%d reqs, %s premium tokens, $%.6f\n",
			stat.BaseRequests(),
			stat.PremiumRequests(),
			FormatTokenCount(stat.PremiumTokens().Total()),
			stat.PremiumCost().Amount()))
	}

	return b.String()
}

// calculateDailyTableWidths calculates column widths for daily usage table
func (m *DailyUsageTabModel) calculateDailyTableWidths(availableWidth int) []int {
	// Account for spaces between columns (7 spaces for 8 columns)
	spaceBetweenColumns := 7
	usableWidth := availableWidth - spaceBetweenColumns

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

// Message types for DailyUsageTabModel
type UsageRefreshMsg struct{}

type UsageDataMsg struct {
	Usage entity.Usage
}
