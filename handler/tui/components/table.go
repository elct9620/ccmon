package components

import (
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	"github.com/elct9620/ccmon/entity"
)

var (
	helpStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))
)

// TableComponent handles table-specific rendering and logic
type TableComponent struct {
	model table.Model
}

// NewTableComponent creates a new table component
func NewTableComponent(tableModel table.Model) *TableComponent {
	return &TableComponent{
		model: tableModel,
	}
}

// RenderTable renders the table view
func (tc *TableComponent) RenderTable(requests []entity.APIRequest, tableModel table.Model) string {
	if len(requests) == 0 {
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

	return tableModel.View()
}
