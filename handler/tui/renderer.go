package tui

import (
	"fmt"
	"strings"

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
