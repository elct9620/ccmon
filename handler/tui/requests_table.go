package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/elct9620/ccmon/entity"
	"github.com/elct9620/ccmon/usecase"
)

var (
	helpStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))
)

// RequestsTableModel handles the requests table display and interaction and owns its data
type RequestsTableModel struct {
	// Data ownership
	table    table.Model
	requests []entity.APIRequest
	
	// Configuration
	timezone *time.Location
	width    int
	height   int
	
	// Business logic dependencies
	getFilteredQuery *usecase.GetFilteredApiRequestsQuery
}

// NewRequestsTableModel creates a new requests table model with usecase dependency
func NewRequestsTableModel(getFilteredQuery *usecase.GetFilteredApiRequestsQuery, timezone *time.Location) *RequestsTableModel {
	// Start with basic columns, will be resized when size is set
	initialWidths := CalculateTableColumnWidths(120) // Assume medium width initially
	columns := []table.Column{
		{Title: "Time", Width: initialWidths[0]},
		{Title: "Model", Width: initialWidths[1]},
		{Title: "Input", Width: initialWidths[2]},
		{Title: "Output", Width: initialWidths[3]},
		{Title: "Cache", Width: initialWidths[4]},
		{Title: "Total", Width: initialWidths[5]},
		{Title: "Cost ($)", Width: initialWidths[6]},
		{Title: "Duration", Width: initialWidths[7]},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.Bold(true)
	s.Selected = s.Selected.Bold(false)
	t.SetStyles(s)

	return &RequestsTableModel{
		table:            t,
		requests:         []entity.APIRequest{},
		timezone:         timezone,
		width:            120,
		height:           10,
		getFilteredQuery: getFilteredQuery,
	}
}

// Init initializes the table model
func (m *RequestsTableModel) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the table model
func (m *RequestsTableModel) Update(msg tea.Msg) (ComponentModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case ResizeMsg:
		m.SetSize(msg.Width, msg.Height)
	case RequestsRefreshMsg:
		return m, m.refreshRequests(msg.Period, msg.SortOrder)
	case RequestsDataMsg:
		m.requests = msg.Requests
		m.updateTableRows()
	case tea.KeyMsg:
		// Handle table navigation
		m.table, cmd = m.table.Update(msg)
	}

	return m, cmd
}

// View renders the requests table
func (m *RequestsTableModel) View() string {
	if len(m.requests) == 0 {
		var b strings.Builder
		b.WriteString(helpStyle.Render("\n  Waiting for API requests...\n"))
		b.WriteString(helpStyle.Render("\n  Make sure to set these environment variables:\n"))
		b.WriteString(helpStyle.Render("    export CLAUDE_CODE_ENABLE_TELEMETRY=1\n"))
		b.WriteString(helpStyle.Render("    export OTEL_METRICS_EXPORTER=otlp\n"))
		b.WriteString(helpStyle.Render("    export OTEL_LOGS_EXPORTER=otlp\n"))
		b.WriteString(helpStyle.Render("    export OTEL_EXPORTER_OTLP_PROTOCOL=grpc\n"))
		b.WriteString(helpStyle.Render("    export OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317\n"))
		return b.String()
	}

	return m.table.View()
}

// SetSize updates the table size and recalculates column widths
func (m *RequestsTableModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.resizeTableColumns()
	m.adjustTableHeight()
}

// UpdateRequests updates the requests data
func (m *RequestsTableModel) UpdateRequests(requests []entity.APIRequest) {
	m.requests = requests
	m.updateTableRows()
}

// GetTable returns the underlying table model for integration with other components
func (m *RequestsTableModel) GetTable() table.Model {
	return m.table
}

// Blur removes focus from the table
func (m *RequestsTableModel) Blur() {
	m.table.Blur()
}

// Focus sets focus on the table
func (m *RequestsTableModel) Focus() {
	m.table.Focus()
}

// Focused returns whether the table is focused
func (m *RequestsTableModel) Focused() bool {
	return m.table.Focused()
}

// updateTableRows updates the table rows based on current requests data
func (m *RequestsTableModel) updateTableRows() {
	rows := make([]table.Row, 0, len(m.requests))
	for _, req := range m.requests {
		// Format timestamp in configured timezone
		timestamp := req.Timestamp().In(m.timezone).Format("15:04:05 2006-01-02")

		if m.width < 80 {
			// Compact mode: combine cache and total tokens
			cacheAndTotal := fmt.Sprintf("%s/%s",
				FormatNumber(req.Tokens().Cache()),
				FormatNumber(req.Tokens().Total()))

			rows = append(rows, table.Row{
				timestamp,
				req.Model().String(), // Don't truncate - let auto-width handle it
				FormatNumber(req.Tokens().Input()),
				FormatNumber(req.Tokens().Output()),
				cacheAndTotal,
				FormatCost(req.Cost().Amount()),
				FormatDuration(req.DurationMS()),
			})
		} else {
			// Normal mode: separate columns
			rows = append(rows, table.Row{
				timestamp,
				req.Model().String(), // Don't truncate - let auto-width handle it
				FormatNumber(req.Tokens().Input()),
				FormatNumber(req.Tokens().Output()),
				FormatNumber(req.Tokens().Cache()),
				FormatNumber(req.Tokens().Total()),
				FormatCost(req.Cost().Amount()),
				FormatDuration(req.DurationMS()),
			})
		}
	}
	m.table.SetRows(rows)
}

// resizeTableColumns resizes table columns based on available width
func (m *RequestsTableModel) resizeTableColumns() {
	// Calculate auto-width columns based on available terminal width
	widths := CalculateTableColumnWidths(m.width)

	// Define column titles based on available width
	var columns []table.Column
	if m.width < 80 {
		// Compact layout for narrow terminals - shorter titles
		columns = []table.Column{
			{Title: "Time", Width: widths[0]},
			{Title: "Model", Width: widths[1]},
			{Title: "In", Width: widths[2]},
			{Title: "Out", Width: widths[3]},
			{Title: "Tot", Width: widths[4] + widths[5]}, // Combine Cache+Total for space
			{Title: "Cost", Width: widths[6]},
			{Title: "Dur", Width: widths[7]},
		}
		// For compact mode, merge cache and total columns
		m.setCompactColumns(columns)
	} else {
		// Normal layout - full column titles
		columns = []table.Column{
			{Title: "Time", Width: widths[0]},
			{Title: "Model", Width: widths[1]},
			{Title: "Input", Width: widths[2]},
			{Title: "Output", Width: widths[3]},
			{Title: "Cache", Width: widths[4]},
			{Title: "Total", Width: widths[5]},
			{Title: "Cost ($)", Width: widths[6]},
			{Title: "Duration", Width: widths[7]},
		}
		m.table.SetColumns(columns)
	}

	// Update table rows to match new column layout
	m.updateTableRows()
}

// setCompactColumns sets compact column layout
func (m *RequestsTableModel) setCompactColumns(columns []table.Column) {
	// Set the compact columns (6 columns instead of 8)
	compactColumns := []table.Column{
		columns[0], // Time
		columns[1], // Model
		columns[2], // Input
		columns[3], // Output
		columns[4], // Combined Total
		columns[5], // Cost
		columns[6], // Duration
	}
	m.table.SetColumns(compactColumns)
}

// adjustTableHeight calculates and sets appropriate table height
func (m *RequestsTableModel) adjustTableHeight() {
	// Table height will be set by the parent component based on available space
	// For now, keep the current height logic from the original implementation

	// Be more conservative with height calculations to prevent overflow
	// Components breakdown:
	// - Title: 2 lines (title + newline)
	// - Status: 2 lines (status + newline)
	// - Stats box: varies (8-12 lines with borders and content)
	// - Table header: 1 line
	// - Help text: 2 lines (newline + help)
	// - Safety margin: 2 lines

	fixedHeight := 9 // Title, status, table header, help, margins

	// Calculate stats section height more accurately
	statsHeight := 10 // Conservative estimate for stats box with borders

	// For compact stats, reduce height
	if m.width < 60 {
		statsHeight = 8 // Compact stats are shorter
	}

	// Calculate remaining height for table with safety margin
	tableHeight := m.height - fixedHeight - statsHeight - 2 // Extra 2 lines safety margin

	// Ensure reasonable minimum and maximum
	if tableHeight < 3 {
		tableHeight = 3
	} else if tableHeight > 20 {
		tableHeight = 20 // Cap maximum table height
	}

	m.table.SetHeight(tableHeight)
}

// refreshRequests handles data fetching for the requests table model
func (m *RequestsTableModel) refreshRequests(period entity.Period, sortOrder SortOrder) tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		if m.getFilteredQuery == nil {
			return RequestsDataMsg{Requests: []entity.APIRequest{}}
		}

		// Query for display requests (limit to 100 for TUI display)
		displayParams := usecase.GetFilteredApiRequestsParams{
			Period: period,
			Limit:  100,
			Offset: 0,
		}
		requests, err := m.getFilteredQuery.Execute(context.Background(), displayParams)
		if err != nil {
			return RequestsDataMsg{Requests: []entity.APIRequest{}}
		}

		// Apply sorting based on user preference
		if sortOrder == SortDescending {
			// Reverse to show latest first (since DB returns chronological order)
			m.reverseRequests(requests)
		}

		return RequestsDataMsg{Requests: requests}
	})
}

// reverseRequests reverses the order of requests slice
func (m *RequestsTableModel) reverseRequests(requests []entity.APIRequest) {
	for i, j := 0, len(requests)-1; i < j; i, j = i+1, j-1 {
		requests[i], requests[j] = requests[j], requests[i]
	}
}

// Requests returns the current requests (for compatibility)
func (m *RequestsTableModel) Requests() []entity.APIRequest {
	return m.requests
}

// Message types for RequestsTableModel
type RequestsRefreshMsg struct {
	Period    entity.Period
	SortOrder SortOrder
}

type RequestsDataMsg struct {
	Requests []entity.APIRequest
}
