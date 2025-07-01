package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/elct9620/ccmon/entity"
	"github.com/elct9620/ccmon/usecase"
)

// StatsModel handles the rendering of usage statistics and owns its data
type StatsModel struct {
	// Data ownership
	stats      entity.Stats
	blockStats entity.Stats
	block      *entity.Block

	// Configuration
	timezone *time.Location
	width    int

	// Progress bar components
	progressModel progress.Model

	// Business logic dependencies
	calculateStatsQuery *usecase.CalculateStatsQuery
}

// NewStatsModel creates a new statistics model with usecase dependency
func NewStatsModel(calculateStatsQuery *usecase.CalculateStatsQuery, timezone *time.Location, block *entity.Block) *StatsModel {
	// Initialize progress model with green to red gradient
	progressModel := progress.New(
		progress.WithWidth(40),
		progress.WithScaledGradient("42", "196"), // Green to Red
		progress.WithoutPercentage(),
	)

	return &StatsModel{
		stats:               entity.Stats{},
		blockStats:          entity.Stats{},
		block:               block,
		timezone:            timezone,
		width:               120, // Default width
		progressModel:       progressModel,
		calculateStatsQuery: calculateStatsQuery,
	}
}

// Init initializes the stats model
func (m *StatsModel) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model
func (m *StatsModel) Update(msg tea.Msg) (ComponentModel, tea.Cmd) {
	switch msg := msg.(type) {
	case ResizeMsg:
		m.width = msg.Width
	case StatsRefreshMsg:
		return m, m.refreshStats(msg.Period)
	case StatsDataMsg:
		m.stats = msg.Stats
		m.blockStats = msg.BlockStats
		if msg.Block != nil {
			m.block = msg.Block
		}
	}
	return m, nil
}

// View renders the statistics section
func (m *StatsModel) View() string {
	var b strings.Builder

	// Header
	b.WriteString(HeaderStyle.Render("Usage Statistics") + "\n\n")

	// Calculate available width for stats table (account for box padding)
	availableWidth := m.width - 6 // Leave margin for box borders and padding
	if availableWidth < 60 {
		// Render compact stats for narrow terminals
		return m.renderCompact()
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

	// Base (Haiku) row
	baseRow := []string{
		BaseStyle.Bold(true).Render("Base (Haiku)"),
		fmt.Sprintf("%d", m.stats.BaseRequests()),
		FormatTokenCount(m.stats.BaseTokens().Limited()),
		FormatTokenCount(m.stats.BaseTokens().Cache()),
		FormatTokenCount(m.stats.BaseTokens().Total()),
		fmt.Sprintf("%.6f", m.stats.BaseCost().Amount()),
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
		fmt.Sprintf("%d", m.stats.PremiumRequests()),
		FormatTokenCount(m.stats.PremiumTokens().Limited()),
		FormatTokenCount(m.stats.PremiumTokens().Cache()),
		FormatTokenCount(m.stats.PremiumTokens().Total()),
		fmt.Sprintf("%.6f", m.stats.PremiumCost().Amount()),
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
		fmt.Sprintf("%d", m.stats.TotalRequests()),
		FormatTokenCount(m.stats.TotalTokens().Limited()),
		FormatTokenCount(m.stats.TotalTokens().Cache()),
		FormatTokenCount(m.stats.TotalTokens().Total()),
		fmt.Sprintf("%.6f", m.stats.TotalCost().Amount()),
	}
	for i, cell := range totalRow {
		if i == 0 {
			b.WriteString(PadRight(cell, colWidths[i]))
		} else {
			b.WriteString(StatStyle.Render(PadRight(cell, colWidths[i])))
		}
	}

	// Add progress bar section if block is configured with limit
	if m.block != nil && m.block.HasLimit() {
		b.WriteString("\n\n")
		b.WriteString(m.renderBlockProgress())
	} else if m.block == nil {
		// Show help message if no block is configured
		b.WriteString("\n\n")
		b.WriteString(HelpStyle.Render("Use -b 5am to track token limits"))
	}

	return b.String()
}

// renderCompact renders a compact version of stats for narrow terminals
func (m *StatsModel) renderCompact() string {
	var b strings.Builder

	// Header
	b.WriteString(HeaderStyle.Render("Usage Statistics") + "\n\n")

	// Compact format for narrow terminals
	b.WriteString(StatStyle.Render("Total Requests: "))
	b.WriteString(fmt.Sprintf("%d\n", m.stats.TotalRequests()))

	b.WriteString(StatStyle.Render("Total Tokens: "))
	b.WriteString(fmt.Sprintf("%s\n", FormatTokenCount(m.stats.TotalTokens().Total())))

	b.WriteString(StatStyle.Render("Total Cost: "))
	b.WriteString(fmt.Sprintf("$%.6f\n", m.stats.TotalCost().Amount()))

	b.WriteString("\n")
	b.WriteString(BaseStyle.Render("Base: "))
	b.WriteString(fmt.Sprintf("%d reqs, %s tokens, $%.6f\n",
		m.stats.BaseRequests(),
		FormatTokenCount(m.stats.BaseTokens().Total()),
		m.stats.BaseCost().Amount()))

	b.WriteString(PremiumStyle.Render("Premium: "))
	b.WriteString(fmt.Sprintf("%d reqs, %s tokens, $%.6f",
		m.stats.PremiumRequests(),
		FormatTokenCount(m.stats.PremiumTokens().Total()),
		m.stats.PremiumCost().Amount()))

	// Add progress bar section if block is configured with limit
	if m.block != nil && m.block.HasLimit() {
		b.WriteString("\n\n")
		b.WriteString(m.renderBlockProgress())
	} else if m.block == nil {
		// Show help message if no block is configured
		b.WriteString("\n\n")
		b.WriteString(HelpStyle.Render("Use -b 5am to track token limits"))
	}

	return b.String()
}

// renderBlockProgress renders the block progress bar section
func (m *StatsModel) renderBlockProgress() string {
	var b strings.Builder

	// Calculate progress using Block entity method
	percentage := m.block.CalculateProgress(m.blockStats.PremiumTokens())

	if percentage > 100 {
		percentage = 100
	}

	// Calculate time remaining until next block
	now := time.Now().UTC()
	var timeRemaining time.Duration
	if now.Before(m.block.EndAt()) {
		timeRemaining = m.block.EndAt().Sub(now)
	}

	// Block header
	blockTime := ""
	if m.block != nil {
		blockTime = m.block.FormatBlockTime(m.timezone)
	}
	b.WriteString(HeaderStyle.Render(fmt.Sprintf("Block Progress (%s)", blockTime)))
	b.WriteString("\n\n")

	// Progress bar using calculated percentage

	progressBar := "[" + m.progressModel.ViewAs(percentage/100) + "]"
	b.WriteString(progressBar)
	b.WriteString(" ")
	used := m.blockStats.PremiumTokens().Limited()
	limit := int64(m.block.TokenLimit())
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

// SetSize updates the model size
func (m *StatsModel) SetSize(width, height int) {
	m.width = width
}

// refreshStats handles data fetching for the stats model
func (m *StatsModel) refreshStats(period entity.Period) tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		if m.calculateStatsQuery == nil {
			return StatsDataMsg{Stats: entity.Stats{}, BlockStats: entity.Stats{}, Block: m.block}
		}

		// Calculate filtered stats for display
		statsParams := usecase.CalculateStatsParams{Period: period}
		stats, err := m.calculateStatsQuery.Execute(context.Background(), statsParams)
		if err != nil {
			stats = entity.Stats{}
		}

		// Update block to current time (may advance to next block automatically)
		var currentBlock *entity.Block
		if m.block != nil {
			nextBlock := m.block.NextBlock(time.Now())
			currentBlock = &nextBlock
		}

		// Calculate block stats for progress bar (only when block tracking is enabled)
		var blockStats entity.Stats
		if currentBlock != nil && m.calculateStatsQuery != nil {
			blockStatsParams := usecase.CalculateStatsParams{
				Period: currentBlock.Period(),
			}
			calculatedBlockStats, err := m.calculateStatsQuery.Execute(context.Background(), blockStatsParams)
			if err == nil {
				blockStats = calculatedBlockStats
			}
		}

		return StatsDataMsg{
			Stats:      stats,
			BlockStats: blockStats,
			Block:      currentBlock,
		}
	})
}

// Stats returns the current stats (for compatibility)
func (m *StatsModel) Stats() entity.Stats {
	return m.stats
}

// BlockStats returns the current block stats (for compatibility)
func (m *StatsModel) BlockStats() entity.Stats {
	return m.blockStats
}

// Block returns the current block (for compatibility)
func (m *StatsModel) Block() *entity.Block {
	return m.block
}

// Message types for StatsModel
type StatsRefreshMsg struct {
	Period entity.Period
}

type StatsDataMsg struct {
	Stats      entity.Stats
	BlockStats entity.Stats
	Block      *entity.Block
}
