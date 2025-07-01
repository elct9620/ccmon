package tui

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/elct9620/ccmon/entity"
	"github.com/elct9620/ccmon/usecase"
)

// OverviewTabModel handles the current/overview tab that shows stats and requests table
type OverviewTabModel struct {
	statsModel         *StatsModel
	requestsTableModel *RequestsTableModel
	width              int
	height             int
}

// NewOverviewTabModel creates a new overview tab model
func NewOverviewTabModel(calculateStatsQuery *usecase.CalculateStatsQuery, getFilteredQuery *usecase.GetFilteredApiRequestsQuery, timezone *time.Location, block *entity.Block) *OverviewTabModel {
	return &OverviewTabModel{
		statsModel:         NewStatsModel(calculateStatsQuery, timezone, block),
		requestsTableModel: NewRequestsTableModel(getFilteredQuery, timezone),
		width:              120,
		height:             30,
	}
}

// Init initializes the overview tab model
func (m *OverviewTabModel) Init() tea.Cmd {
	return tea.Batch(
		m.statsModel.Init(),
		m.requestsTableModel.Init(),
	)
}

// Update handles messages and updates the model
func (m *OverviewTabModel) Update(msg tea.Msg) (ComponentModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case ResizeMsg:
		m.SetSize(msg.Width, msg.Height)

	case StatsRefreshMsg:
		// Forward stats refresh to stats model
		_, cmd := m.statsModel.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

	case StatsDataMsg:
		// Forward stats data to stats model
		_, cmd := m.statsModel.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

	case RequestsRefreshMsg:
		// Forward requests refresh to table model
		_, cmd := m.requestsTableModel.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

	case RequestsDataMsg:
		// Forward requests data to table model
		_, cmd := m.requestsTableModel.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

	case tea.KeyMsg:
		// Handle keyboard input - mainly for table navigation
		switch msg.String() {
		case "esc":
			// Toggle table focus
			if m.requestsTableModel.Focused() {
				m.requestsTableModel.Blur()
			} else {
				m.requestsTableModel.Focus()
			}
		default:
			// Forward other key messages to table model
			_, cmd := m.requestsTableModel.Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
	}

	return m, tea.Batch(cmds...)
}

// View renders the overview tab
func (m *OverviewTabModel) View() string {
	var b strings.Builder

	// Statistics box
	statsContent := m.statsModel.View()
	statsBox := BoxStyle.Width(m.width - 4).Render(statsContent)
	b.WriteString(statsBox + "\n\n")

	// Recent requests header
	requestsHeader := HeaderStyle.Render("Recent API Requests")
	b.WriteString(requestsHeader + "\n")

	// Table
	tableView := m.requestsTableModel.View()
	b.WriteString(tableView + "\n")

	return b.String()
}

// SetSize updates the size of the overview tab and its components
func (m *OverviewTabModel) SetSize(width, height int) {
	m.width = width
	m.height = height

	// Update sub-component sizes
	m.statsModel.SetSize(width, height)
	m.requestsTableModel.SetSize(width, height)
}

// RefreshStats triggers a stats refresh with the given period
func (m *OverviewTabModel) RefreshStats(period entity.Period) tea.Cmd {
	msg := StatsRefreshMsg{Period: period}
	_, cmd := m.statsModel.Update(msg)
	return cmd
}

// RefreshRequests triggers a requests refresh with the given period and sort order
func (m *OverviewTabModel) RefreshRequests(period entity.Period, sortOrder SortOrder) tea.Cmd {
	msg := RequestsRefreshMsg{Period: period, SortOrder: sortOrder}
	_, cmd := m.requestsTableModel.Update(msg)
	return cmd
}

// GetRequestsTable returns the requests table model for external access
func (m *OverviewTabModel) GetRequestsTable() *RequestsTableModel {
	return m.requestsTableModel
}

// Focus sets focus on the requests table
func (m *OverviewTabModel) Focus() {
	m.requestsTableModel.Focus()
}

// Blur removes focus from the requests table
func (m *OverviewTabModel) Blur() {
	m.requestsTableModel.Blur()
}

// Focused returns whether the requests table is focused
func (m *OverviewTabModel) Focused() bool {
	return m.requestsTableModel.Focused()
}
