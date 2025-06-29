package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	// Styles
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("86")).
			MarginBottom(1)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205"))

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	statStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86"))

	baseStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42"))

	premiumStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(0, 1)

	tableHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("86"))

	tableCellStyle = lipgloss.NewStyle().
			PaddingRight(2)
)

// View renders the entire UI
func (m Model) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}

	var b strings.Builder

	// Title
	title := titleStyle.Render("🖥️  Claude Code Monitor")
	b.WriteString(title + "\n")

	// Status line
	status := statusStyle.Render(fmt.Sprintf("Monitor Mode | Filter: %s | Sort: %s | Press 'q' to quit", m.getTimeFilterString(), m.getSortOrderString()))
	b.WriteString(status + "\n\n")

	// Statistics box
	statsContent := m.renderStats()
	statsBox := boxStyle.Width(m.width - 4).Render(statsContent)
	b.WriteString(statsBox + "\n\n")

	// Recent requests header
	requestsHeader := headerStyle.Render("Recent API Requests")
	b.WriteString(requestsHeader + "\n")

	// Table
	if len(m.requests) == 0 {
		b.WriteString(helpStyle.Render("\n  Waiting for API requests...\n"))
		b.WriteString(helpStyle.Render("\n  Make sure to set these environment variables:\n"))
		b.WriteString(helpStyle.Render("    export CLAUDE_CODE_ENABLE_TELEMETRY=1\n"))
		b.WriteString(helpStyle.Render("    export OTEL_METRICS_EXPORTER=otlp\n"))
		b.WriteString(helpStyle.Render("    export OTEL_LOGS_EXPORTER=otlp\n"))
		b.WriteString(helpStyle.Render("    export OTEL_EXPORTER_OTLP_PROTOCOL=grpc\n"))
		b.WriteString(helpStyle.Render("    export OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317\n"))
	} else {
		b.WriteString(m.table.View())
	}

	// Help text at bottom
	help := helpStyle.Render("\n  ↑/↓: Navigate • Time: h=hour d=day w=week m=month a=all • o=sort • q: Quit")
	b.WriteString(help)

	return b.String()
}

// renderStats renders the statistics section
func (m Model) renderStats() string {
	var b strings.Builder

	// Header
	b.WriteString(headerStyle.Render("Usage Statistics") + "\n\n")

	// Create table headers
	headers := []string{"Model Tier", "Requests", "Limited Tokens", "Cache Tokens", "Total Tokens", "Cost ($)"}

	// Calculate column widths
	colWidths := []int{15, 10, 15, 15, 15, 12}

	// Render header row
	for i, header := range headers {
		cell := tableHeaderStyle.Render(padRight(header, colWidths[i]))
		b.WriteString(cell)
	}
	b.WriteString("\n")

	// Separator line
	for _, width := range colWidths {
		b.WriteString(strings.Repeat("─", width))
	}
	b.WriteString("\n")

	// Base (Haiku) row
	baseRow := []string{
		baseStyle.Bold(true).Render("Base (Haiku)"),
		fmt.Sprintf("%d", m.stats.BaseRequests()),
		formatTokenCount(m.stats.BaseTokens().Limited()),
		formatTokenCount(m.stats.BaseTokens().Cache()),
		formatTokenCount(m.stats.BaseTokens().Total()),
		fmt.Sprintf("%.6f", m.stats.BaseCost().Amount()),
	}
	for i, cell := range baseRow {
		if i == 0 {
			b.WriteString(padRight(cell, colWidths[i]))
		} else {
			b.WriteString(baseStyle.Render(padRight(cell, colWidths[i])))
		}
	}
	b.WriteString("\n")

	// Premium (S/O) row
	premiumRow := []string{
		premiumStyle.Bold(true).Render("Premium (S/O)"),
		fmt.Sprintf("%d", m.stats.PremiumRequests()),
		formatTokenCount(m.stats.PremiumTokens().Limited()),
		formatTokenCount(m.stats.PremiumTokens().Cache()),
		formatTokenCount(m.stats.PremiumTokens().Total()),
		fmt.Sprintf("%.6f", m.stats.PremiumCost().Amount()),
	}
	for i, cell := range premiumRow {
		if i == 0 {
			b.WriteString(padRight(cell, colWidths[i]))
		} else {
			b.WriteString(premiumStyle.Render(padRight(cell, colWidths[i])))
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
		statStyle.Bold(true).Render("Total"),
		fmt.Sprintf("%d", m.stats.TotalRequests()),
		formatTokenCount(m.stats.TotalTokens().Limited()),
		formatTokenCount(m.stats.TotalTokens().Cache()),
		formatTokenCount(m.stats.TotalTokens().Total()),
		fmt.Sprintf("%.6f", m.stats.TotalCost().Amount()),
	}
	for i, cell := range totalRow {
		if i == 0 {
			b.WriteString(padRight(cell, colWidths[i]))
		} else {
			b.WriteString(statStyle.Render(padRight(cell, colWidths[i])))
		}
	}

	return b.String()
}

// padRight pads a string to the specified width
func padRight(s string, width int) string {
	// Account for ANSI escape codes when calculating padding
	visualLen := lipgloss.Width(s)
	if visualLen >= width {
		return s
	}
	return s + strings.Repeat(" ", width-visualLen)
}

// formatTokenCount formats large token counts with K/M suffixes
func formatTokenCount(tokens int64) string {
	if tokens < 1000 {
		return fmt.Sprintf("%d", tokens)
	} else if tokens < 1000000 {
		return fmt.Sprintf("%.1fK", float64(tokens)/1000)
	} else {
		return fmt.Sprintf("%.2fM", float64(tokens)/1000000)
	}
}
