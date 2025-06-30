package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/elct9620/ccmon/entity"
)

// renderDailyUsageTable renders the daily usage statistics as a table
func (r *Renderer) renderDailyUsageTable(vm *ViewModel, width int) string {
	var b strings.Builder

	usage := vm.Usage()
	stats := usage.GetStats()

	if len(stats) == 0 {
		b.WriteString(HelpStyle.Render("No usage data available"))
		return b.String()
	}

	// Calculate available width for table
	availableWidth := width - 6 // Account for box padding
	if availableWidth < 60 {
		return r.renderCompactDailyUsage(stats, vm.Timezone())
	}

	// Table headers
	headers := []string{"Date", "Requests", "Input", "Output", "Read Cache", "Creation Cache", "Total", "Premium Cost ($)"}
	colWidths := r.calculateDailyTableWidths(availableWidth)

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

		date := period.StartAt().In(vm.Timezone()).Format("2006-01-02")
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
func (r *Renderer) renderCompactDailyUsage(stats []entity.Stats, timezone *time.Location) string {
	var b strings.Builder

	for _, stat := range stats {
		period := stat.Period()
		if period.IsAllTime() {
			continue
		}

		// Convert UTC time back to user's timezone for display
		date := period.StartAt().In(timezone).Format("2006-01-02")
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
func (r *Renderer) calculateDailyTableWidths(availableWidth int) []int {
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
