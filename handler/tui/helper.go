package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// Styles
var (
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("86")).
			MarginBottom(1)

	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205"))

	StatusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	StatStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86"))

	BaseStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42"))

	PremiumStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214"))

	HelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(0, 1)

	TableHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("86"))
)

// String formatting functions
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func FormatNumber(n int64) string {
	if n == 0 {
		return "-"
	}
	return fmt.Sprintf("%d", n)
}

func FormatCost(cost float64) string {
	if cost == 0 {
		return "-"
	}
	return fmt.Sprintf("%.6f", cost)
}

func FormatDuration(ms int64) string {
	if ms < 1000 {
		return fmt.Sprintf("%dms", ms)
	}
	return fmt.Sprintf("%.1fs", float64(ms)/1000)
}

func FormatTokenCount(tokens int64) string {
	if tokens < 1000 {
		return fmt.Sprintf("%d", tokens)
	} else if tokens < 1000000 {
		return fmt.Sprintf("%.1fK", float64(tokens)/1000)
	} else {
		return fmt.Sprintf("%.2fM", float64(tokens)/1000000)
	}
}

func FormatDurationFromTime(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	} else if d < time.Hour {
		return fmt.Sprintf("%dm %ds", int(d.Minutes()), int(d.Seconds())%60)
	} else {
		hours := int(d.Hours())
		minutes := int(d.Minutes()) % 60
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
}

// Layout helper functions
func PadRight(s string, width int) string {
	// Account for ANSI escape codes when calculating padding
	visualLen := lipgloss.Width(s)
	if visualLen >= width {
		return s
	}
	return s + strings.Repeat(" ", width-visualLen)
}

func CalculateStatsColumnWidths(availableWidth int) []int {
	// Base minimum widths for each column
	minWidths := []int{12, 5, 8, 6, 8, 10} // Model Tier, Reqs, Limited, Cache, Total, Cost

	// Calculate total minimum width
	totalMinWidth := 0
	for _, w := range minWidths {
		totalMinWidth += w
	}

	// If we have extra space, distribute it proportionally
	if availableWidth > totalMinWidth {
		extraSpace := availableWidth - totalMinWidth
		// Distribute extra space: favor first and last columns
		distribution := []float64{0.3, 0.1, 0.2, 0.1, 0.2, 0.1}

		for i := range minWidths {
			extra := int(float64(extraSpace) * distribution[i])
			minWidths[i] += extra
		}
	}

	return minWidths
}

// Progress bar rendering
func RenderProgressBar(percentage float64, width int) string {
	filled := int(percentage / 100 * float64(width))
	if filled > width {
		filled = width
	}

	var color lipgloss.Color
	if percentage >= 90 {
		color = lipgloss.Color("196") // Red
	} else if percentage >= 75 {
		color = lipgloss.Color("214") // Orange
	} else {
		color = lipgloss.Color("42") // Green
	}

	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	style := lipgloss.NewStyle().Foreground(color)
	return "[" + style.Render(bar) + "]"
}
