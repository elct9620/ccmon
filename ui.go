package main

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

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(0, 1)
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
	status := statusStyle.Render(fmt.Sprintf("Server: %s | Press 'q' to quit", m.serverStatus))
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
	help := helpStyle.Render("\n  ↑/↓: Navigate • q: Quit")
	b.WriteString(help)

	return b.String()
}

// renderStats renders the statistics section
func (m Model) renderStats() string {
	var b strings.Builder

	// Current session stats
	b.WriteString(headerStyle.Render("Session Statistics") + "\n\n")

	statsLine1 := fmt.Sprintf(
		"%s %s  %s %s  %s %s",
		statStyle.Render("Requests:"),
		fmt.Sprintf("%d", m.totalRequests),
		statStyle.Render("Total Tokens:"),
		formatTokenCount(m.totalTokens),
		statStyle.Render("Total Cost:"),
		fmt.Sprintf("$%.6f", m.totalCost),
	)
	b.WriteString(statsLine1)

	// Latest request info
	if len(m.requests) > 0 {
		latest := m.requests[0]
		b.WriteString("\n\n")
		b.WriteString(headerStyle.Render("Latest Request") + "\n\n")

		latestInfo := fmt.Sprintf(
			"%s %s  %s %s  %s %dms",
			statStyle.Render("Model:"),
			latest.Model,
			statStyle.Render("Tokens:"),
			formatTokenCount(latest.TotalTokens),
			statStyle.Render("Duration:"),
			latest.DurationMS,
		)
		b.WriteString(latestInfo)
	}

	return b.String()
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
