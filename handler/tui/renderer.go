package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/elct9620/ccmon/entity"
	"github.com/elct9620/ccmon/handler/tui/components"
)

// Renderer handles the main view rendering logic
type Renderer struct {
	tableComponent *components.TableComponent
}

// NewRenderer creates a new renderer with components
func NewRenderer(tableComponent *components.TableComponent) *Renderer {
	return &Renderer{
		tableComponent: tableComponent,
	}
}

// View renders the entire UI
func (r *Renderer) View(vm *ViewModel, width int) string {
	if !vm.Ready() {
		return "\n  Initializing..."
	}

	var b strings.Builder

	// Title
	title := TitleStyle.Render("🖥️  Claude Code Monitor")
	b.WriteString(title + "\n")

	// Tab navigation
	tabNav := r.renderTabNavigation(vm, width)
	b.WriteString(tabNav + "\n")

	// Status line (only for current tab)
	if vm.CurrentTab() == TabCurrent {
		status := StatusStyle.Render(fmt.Sprintf("Monitor Mode | Filter: %s | Sort: %s", vm.GetTimeFilterString(), vm.GetSortOrderString()))
		b.WriteString(status + "\n\n")
	} else {
		b.WriteString("\n")
	}

	// Content based on current tab
	switch vm.CurrentTab() {
	case TabCurrent:
		content := r.renderCurrentTab(vm, width)
		b.WriteString(content)
	case TabDaily:
		content := r.renderDailyTab(vm, width)
		b.WriteString(content)
	}

	// Help text at bottom
	helpText := r.renderHelpText(vm)
	help := HelpStyle.Render(helpText)
	b.WriteString(help)

	return b.String()
}

// renderStats renders the statistics section
func (r *Renderer) renderStats(vm *ViewModel, width int) string {
	var b strings.Builder

	// Header
	b.WriteString(HeaderStyle.Render("Usage Statistics") + "\n\n")

	// Calculate available width for stats table (account for box padding)
	availableWidth := width - 6 // Leave margin for box borders and padding
	if availableWidth < 60 {
		// Render compact stats for narrow terminals
		return r.renderCompactStats(vm)
	}

	// Create table headers
	headers := []string{"Model Tier", "Reqs", "Limited", "Cache", "Total", "Cost ($)"}

	// Calculate dynamic column widths based on available space
	colWidths := CalculateStatsColumnWidths(availableWidth)

	// Render header row
	for i, header := range headers {
		cell := TableHeaderStyle.Render(PadRight(header, colWidths[i]))
		b.WriteString(cell)
	}
	b.WriteString("\n")

	// Separator line
	for _, width := range colWidths {
		b.WriteString(strings.Repeat("─", width))
	}
	b.WriteString("\n")

	stats := vm.Stats()

	// Base (Haiku) row
	baseRow := []string{
		BaseStyle.Bold(true).Render("Base (Haiku)"),
		fmt.Sprintf("%d", stats.BaseRequests()),
		FormatTokenCount(stats.BaseTokens().Limited()),
		FormatTokenCount(stats.BaseTokens().Cache()),
		FormatTokenCount(stats.BaseTokens().Total()),
		fmt.Sprintf("%.6f", stats.BaseCost().Amount()),
	}
	for i, cell := range baseRow {
		if i == 0 {
			b.WriteString(PadRight(cell, colWidths[i]))
		} else {
			b.WriteString(BaseStyle.Render(PadRight(cell, colWidths[i])))
		}
	}
	b.WriteString("\n")

	// Premium (S/O) row
	premiumRow := []string{
		PremiumStyle.Bold(true).Render("Premium (S/O)"),
		fmt.Sprintf("%d", stats.PremiumRequests()),
		FormatTokenCount(stats.PremiumTokens().Limited()),
		FormatTokenCount(stats.PremiumTokens().Cache()),
		FormatTokenCount(stats.PremiumTokens().Total()),
		fmt.Sprintf("%.6f", stats.PremiumCost().Amount()),
	}
	for i, cell := range premiumRow {
		if i == 0 {
			b.WriteString(PadRight(cell, colWidths[i]))
		} else {
			b.WriteString(PremiumStyle.Render(PadRight(cell, colWidths[i])))
		}
	}
	b.WriteString("\n")

	// Separator before total
	for _, width := range colWidths {
		b.WriteString(strings.Repeat("─", width))
	}
	b.WriteString("\n")

	// Total row
	totalRow := []string{
		StatStyle.Bold(true).Render("Total"),
		fmt.Sprintf("%d", stats.TotalRequests()),
		FormatTokenCount(stats.TotalTokens().Limited()),
		FormatTokenCount(stats.TotalTokens().Cache()),
		FormatTokenCount(stats.TotalTokens().Total()),
		fmt.Sprintf("%.6f", stats.TotalCost().Amount()),
	}
	for i, cell := range totalRow {
		if i == 0 {
			b.WriteString(PadRight(cell, colWidths[i]))
		} else {
			b.WriteString(StatStyle.Render(PadRight(cell, colWidths[i])))
		}
	}

	// Add progress bar section if block is configured with limit
	if vm.Block() != nil && vm.Block().HasLimit() {
		b.WriteString("\n\n")
		b.WriteString(r.renderBlockProgress(vm))
	} else if vm.Block() == nil {
		// Show help message if no block is configured
		b.WriteString("\n\n")
		b.WriteString(HelpStyle.Render("Use -b 5am to track token limits"))
	}

	return b.String()
}

// renderCompactStats renders a compact version of stats for narrow terminals
func (r *Renderer) renderCompactStats(vm *ViewModel) string {
	var b strings.Builder

	// Header
	b.WriteString(HeaderStyle.Render("Usage Statistics") + "\n\n")

	stats := vm.Stats()

	// Compact format for narrow terminals
	b.WriteString(StatStyle.Render("Total Requests: "))
	b.WriteString(fmt.Sprintf("%d\n", stats.TotalRequests()))

	b.WriteString(StatStyle.Render("Total Tokens: "))
	b.WriteString(fmt.Sprintf("%s\n", FormatTokenCount(stats.TotalTokens().Total())))

	b.WriteString(StatStyle.Render("Total Cost: "))
	b.WriteString(fmt.Sprintf("$%.6f\n", stats.TotalCost().Amount()))

	b.WriteString("\n")
	b.WriteString(BaseStyle.Render("Base: "))
	b.WriteString(fmt.Sprintf("%d reqs, %s tokens, $%.6f\n",
		stats.BaseRequests(),
		FormatTokenCount(stats.BaseTokens().Total()),
		stats.BaseCost().Amount()))

	b.WriteString(PremiumStyle.Render("Premium: "))
	b.WriteString(fmt.Sprintf("%d reqs, %s tokens, $%.6f",
		stats.PremiumRequests(),
		FormatTokenCount(stats.PremiumTokens().Total()),
		stats.PremiumCost().Amount()))

	// Add progress bar section if block is configured with limit
	if vm.Block() != nil && vm.Block().HasLimit() {
		b.WriteString("\n\n")
		b.WriteString(r.renderBlockProgress(vm))
	} else if vm.Block() == nil {
		// Show help message if no block is configured
		b.WriteString("\n\n")
		b.WriteString(HelpStyle.Render("Use -b 5am to track token limits"))
	}

	return b.String()
}

// renderBlockProgress renders the block progress bar section
func (r *Renderer) renderBlockProgress(vm *ViewModel) string {
	var b strings.Builder

	blockStats := vm.BlockStats()

	// Calculate progress using Block entity method
	percentage := vm.Block().CalculateProgress(blockStats.PremiumTokens())

	if percentage > 100 {
		percentage = 100
	}

	// Calculate time remaining until next block
	now := time.Now().UTC()
	var timeRemaining time.Duration
	if now.Before(vm.Block().EndAt()) {
		timeRemaining = vm.Block().EndAt().Sub(now)
	}

	// Block header
	blockTime := ""
	if vm.Block() != nil {
		blockTime = vm.Block().FormatBlockTime(vm.Timezone())
	}
	b.WriteString(HeaderStyle.Render(fmt.Sprintf("Block Progress (%s)", blockTime)))
	b.WriteString("\n\n")

	// Progress bar
	progressBar := RenderProgressBar(percentage, 40)
	b.WriteString(progressBar)
	b.WriteString(" ")
	used := blockStats.PremiumTokens().Limited()
	limit := int64(vm.Block().TokenLimit())
	b.WriteString(StatStyle.Render(fmt.Sprintf("%.1f%% (%s/%s tokens)", percentage, FormatTokenCount(used), FormatTokenCount(limit))))
	b.WriteString("\n")

	// Time remaining
	if timeRemaining > 0 {
		b.WriteString(HelpStyle.Render(fmt.Sprintf("Time remaining: %s", FormatDurationFromTime(timeRemaining))))
	} else {
		b.WriteString(HelpStyle.Render("Block expired"))
	}

	return b.String()
}

// renderTabNavigation renders the tab navigation bar
func (r *Renderer) renderTabNavigation(vm *ViewModel, width int) string {
	var b strings.Builder

	// Tab buttons
	currentTabStyle := StatStyle.Bold(true)
	inactiveTabStyle := HelpStyle

	if vm.CurrentTab() == TabCurrent {
		b.WriteString(currentTabStyle.Render("[Current]"))
	} else {
		b.WriteString(inactiveTabStyle.Render(" Current "))
	}

	b.WriteString("  ")

	if vm.CurrentTab() == TabDaily {
		b.WriteString(currentTabStyle.Render("[Daily Usage]"))
	} else {
		b.WriteString(inactiveTabStyle.Render(" Daily Usage "))
	}

	return b.String()
}

// renderCurrentTab renders the current tab content
func (r *Renderer) renderCurrentTab(vm *ViewModel, width int) string {
	var b strings.Builder

	// Statistics box
	statsContent := r.renderStats(vm, width)
	statsBox := BoxStyle.Width(width - 4).Render(statsContent)
	b.WriteString(statsBox + "\n\n")

	// Recent requests header
	requestsHeader := HeaderStyle.Render("Recent API Requests")
	b.WriteString(requestsHeader + "\n")

	// Table
	tableView := r.tableComponent.RenderTable(vm.Requests(), vm.Table())
	b.WriteString(tableView + "\n")

	return b.String()
}

// renderDailyTab renders the daily usage tab content
func (r *Renderer) renderDailyTab(vm *ViewModel, width int) string {
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
	dailyContent := r.renderDailyUsageTable(vm, width)
	dailyBox := BoxStyle.Width(width - 4).Render(dailyContent)
	b.WriteString(dailyBox + "\n")

	return b.String()
}

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
		return r.renderCompactDailyUsage(stats)
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
func (r *Renderer) renderCompactDailyUsage(stats []entity.Stats) string {
	var b strings.Builder

	for _, stat := range stats {
		period := stat.Period()
		if period.IsAllTime() {
			continue
		}

		date := period.StartAt().Format("2006-01-02")
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

// renderHelpText renders the help text based on current tab
func (r *Renderer) renderHelpText(vm *ViewModel) string {
	var helpText string

	switch vm.CurrentTab() {
	case TabCurrent:
		helpText = "\n  ↑/↓: Navigate • Time: h=hour d=day w=week m=month a=all"
		if vm.Block() != nil {
			helpText += " b=block"
		}
		helpText += " • o=sort • Tab: Switch tabs • q: Quit"
	case TabDaily:
		helpText = "\n  ↑/↓: Navigate • Tab: Switch tabs • q: Quit"
	}

	return helpText
}
