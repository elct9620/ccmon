package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/elct9620/ccmon/entity"
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

func FormatBurnRate(tokensPerMinute float64) string {
	if tokensPerMinute <= 0 {
		return "-"
	}

	if tokensPerMinute < 1000 {
		return fmt.Sprintf("%.1f/min", tokensPerMinute)
	} else if tokensPerMinute < 1000000 {
		return fmt.Sprintf("%.1fK/min", tokensPerMinute/1000)
	} else {
		return fmt.Sprintf("%.2fM/min", tokensPerMinute/1000000)
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
	minWidths := []int{12, 5, 8, 6, 8, 10, 10} // Model Tier, Reqs, Limited, Cache, Total, Cost, Burn Rate

	// Calculate total minimum width
	totalMinWidth := 0
	for _, w := range minWidths {
		totalMinWidth += w
	}

	// If we have extra space, distribute it proportionally
	if availableWidth > totalMinWidth {
		extraSpace := availableWidth - totalMinWidth
		// Distribute extra space: favor first column and burn rate column
		distribution := []float64{0.25, 0.1, 0.15, 0.1, 0.15, 0.1, 0.15}

		for i := range minWidths {
			extra := int(float64(extraSpace) * distribution[i])
			minWidths[i] += extra
		}
	}

	return minWidths
}

func CalculateTableColumnWidths(availableWidth int) []int {
	// Base minimum widths for each column
	// Time, Model, Input, Output, Cache, Total, Cost, Duration
	minWidths := []int{16, 20, 6, 6, 6, 6, 8, 8}

	// Account for borders, padding, and separators (approximately 2 chars per column)
	overhead := len(minWidths) * 2
	usableWidth := availableWidth - overhead
	if usableWidth < 0 {
		usableWidth = availableWidth
	}

	// Calculate total minimum width
	totalMinWidth := 0
	for _, w := range minWidths {
		totalMinWidth += w
	}

	// If we have extra space, distribute it proportionally
	if usableWidth > totalMinWidth {
		extraSpace := usableWidth - totalMinWidth
		// Distribute extra space: favor Model column most, then Time
		// Model gets 50% of extra space, Time gets 20%, others get smaller amounts
		distribution := []float64{0.2, 0.5, 0.05, 0.05, 0.05, 0.05, 0.05, 0.05}

		for i := range minWidths {
			extra := int(float64(extraSpace) * distribution[i])
			minWidths[i] += extra
		}
	}

	return minWidths
}

// FormatBlockTime formats the block period for display in the given timezone
func FormatBlockTime(block entity.Block, timezone *time.Location) string {
	startLocal := block.StartAt().In(timezone)
	endLocal := block.EndAt().In(timezone)

	startStr := formatHour(startLocal.Hour())
	endStr := formatHour(endLocal.Hour())

	return fmt.Sprintf("%s - %s", startStr, endStr)
}

// formatHour formats hour (0-23) into 12-hour format with am/pm
func formatHour(hour int) string {
	if hour == 0 {
		return "12am"
	} else if hour < 12 {
		return fmt.Sprintf("%dam", hour)
	} else if hour == 12 {
		return "12pm"
	} else {
		return fmt.Sprintf("%dpm", hour-12)
	}
}
