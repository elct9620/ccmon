package tui

import (
	"strings"
)

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
