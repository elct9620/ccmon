package tui

import (
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/elct9620/ccmon/entity"
	"github.com/elct9620/ccmon/usecase"
)

// TimeFilter represents the available time filter options for UI
type TimeFilter int

const (
	FilterAll TimeFilter = iota
	FilterHour
	FilterDay
	FilterWeek
	FilterMonth
	FilterBlock // Current block timeframe
)

// Tab represents the available tabs in the UI
type Tab int

const (
	TabCurrent Tab = iota // Current view (requests and stats)
	TabDaily              // Daily usage view
)

// ViewModel represents the refactored state of our TUI monitor application using component models
type ViewModel struct {
	// Tab models
	overviewTab   *OverviewTabModel
	dailyUsageTab *DailyUsageTabModel

	// Application state
	currentTab      Tab
	width           int
	height          int
	ready           bool
	timeFilter      TimeFilter
	sortOrder       SortOrder
	timezone        *time.Location
	refreshInterval time.Duration
}

// NewViewModel creates a new refactored ViewModel with component models
func NewViewModel(getFilteredQuery *usecase.GetFilteredApiRequestsQuery, calculateStatsQuery *usecase.CalculateStatsQuery, getUsageQuery *usecase.GetUsageQuery, timezone *time.Location, block *entity.Block, refreshInterval time.Duration) *ViewModel {
	return &ViewModel{
		overviewTab:     NewOverviewTabModel(calculateStatsQuery, getFilteredQuery, timezone, block),
		dailyUsageTab:   NewDailyUsageTabModel(getUsageQuery, timezone),
		currentTab:      TabCurrent,
		timeFilter:      FilterAll,
		sortOrder:       SortDescending,
		timezone:        timezone,
		refreshInterval: refreshInterval,
	}
}

// Init is the Bubble Tea initialization function
func (vm *ViewModel) Init() tea.Cmd {
	// Ensure the current tab is focused on startup
	vm.overviewTab.Focus()
	vm.dailyUsageTab.Blur()

	return tea.Batch(
		tea.EnterAltScreen,
		vm.overviewTab.Init(),
		vm.dailyUsageTab.Init(),
		vm.refreshStats, // Load initial data from database
		vm.tick(),       // Start periodic refresh
	)
}

// Update handles messages and updates the model
func (vm *ViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return vm, tea.Quit
		case "a":
			vm.timeFilter = FilterAll
			return vm, vm.refreshStats
		case "h":
			vm.timeFilter = FilterHour
			return vm, vm.refreshStats
		case "d":
			vm.timeFilter = FilterDay
			return vm, vm.refreshStats
		case "w":
			vm.timeFilter = FilterWeek
			return vm, vm.refreshStats
		case "m":
			vm.timeFilter = FilterMonth
			return vm, vm.refreshStats
		case "b":
			if vm.Block() != nil {
				vm.timeFilter = FilterBlock
				return vm, vm.refreshStats
			}
		case "o":
			// Toggle sort order
			if vm.sortOrder == SortDescending {
				vm.sortOrder = SortAscending
			} else {
				vm.sortOrder = SortDescending
			}
			return vm, vm.refreshStats
		case "tab":
			// Switch tabs
			if vm.currentTab == TabCurrent {
				// Blur current tab and focus daily tab
				vm.overviewTab.Blur()
				vm.currentTab = TabDaily
				vm.dailyUsageTab.Focus()
				return vm, vm.refreshUsage
			} else {
				// Blur daily tab and focus current tab
				vm.dailyUsageTab.Blur()
				vm.currentTab = TabCurrent
				vm.overviewTab.Focus()
				return vm, vm.refreshStats
			}
		default:
			// Forward key messages to active tab
			switch vm.currentTab {
			case TabCurrent:
				_, cmd := vm.overviewTab.Update(msg)
				if cmd != nil {
					cmds = append(cmds, cmd)
				}
			case TabDaily:
				_, cmd := vm.dailyUsageTab.Update(msg)
				if cmd != nil {
					cmds = append(cmds, cmd)
				}
			}
		}

	case tea.WindowSizeMsg:
		vm.width = msg.Width
		vm.height = msg.Height
		vm.ready = true

		// Update tab sizes
		resizeMsg := ResizeMsg{Width: msg.Width, Height: msg.Height}
		_, cmd1 := vm.overviewTab.Update(resizeMsg)
		_, cmd2 := vm.dailyUsageTab.Update(resizeMsg)

		if cmd1 != nil {
			cmds = append(cmds, cmd1)
		}
		if cmd2 != nil {
			cmds = append(cmds, cmd2)
		}

	case tickMsg:
		// Periodic refresh - refresh based on current tab
		if vm.currentTab == TabDaily {
			return vm, tea.Batch(vm.tick(), vm.refreshUsage)
		} else {
			return vm, tea.Batch(vm.tick(), vm.refreshStats)
		}

	case refreshStatsMsg:
		// Send refresh messages to overview tab with current period
		if vm.currentTab == TabCurrent {
			period := vm.getTimePeriod()
			// Refresh both stats and requests
			statsCmd := vm.overviewTab.RefreshStats(period)
			requestsCmd := vm.overviewTab.RefreshRequests(period, vm.sortOrder)
			if statsCmd != nil {
				cmds = append(cmds, statsCmd)
			}
			if requestsCmd != nil {
				cmds = append(cmds, requestsCmd)
			}
		}
	case refreshUsageMsg:
		// Send refresh message to daily usage tab
		if vm.currentTab == TabDaily {
			_, cmd := vm.dailyUsageTab.Update(UsageRefreshMsg{})
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}

	case StatsDataMsg:
		// Forward stats data to overview tab
		_, cmd := vm.overviewTab.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

	case RequestsDataMsg:
		// Forward requests data to overview tab
		_, cmd := vm.overviewTab.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

	case UsageDataMsg:
		// Forward usage data to daily usage tab
		_, cmd := vm.dailyUsageTab.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return vm, tea.Batch(cmds...)
}

// View renders the UI by delegating to the appropriate tab model
func (vm *ViewModel) View() string {
	if !vm.ready {
		return "\n  Initializing..."
	}

	// Common header
	content := TitleStyle.Render("🖥️  Claude Code Monitor") + "\n"
	content += vm.renderTabNavigation() + "\n"

	// Tab-specific content
	switch vm.currentTab {
	case TabCurrent:
		// Status line for current tab
		content += StatusStyle.Render("Monitor Mode | Filter: "+vm.GetTimeFilterString()+" | Sort: "+vm.GetSortOrderString()) + "\n\n"
		content += vm.overviewTab.View()
	case TabDaily:
		content += "\n" + vm.dailyUsageTab.View()
	}

	// Help text
	content += vm.renderHelpText()

	return content
}

// renderTabNavigation renders the tab navigation bar
func (vm *ViewModel) renderTabNavigation() string {
	currentTabStyle := StatStyle.Bold(true)
	inactiveTabStyle := HelpStyle

	var content string
	if vm.currentTab == TabCurrent {
		content += currentTabStyle.Render("[Current]")
	} else {
		content += inactiveTabStyle.Render(" Current ")
	}

	content += "  "

	if vm.currentTab == TabDaily {
		content += currentTabStyle.Render("[Daily Usage]")
	} else {
		content += inactiveTabStyle.Render(" Daily Usage ")
	}

	return content
}

// renderHelpText renders the help text based on current tab
func (vm *ViewModel) renderHelpText() string {
	var helpText string

	switch vm.currentTab {
	case TabCurrent:
		helpText = "\n  ↑/↓: Navigate • Time: h=hour d=day w=week m=month a=all"
		if vm.Block() != nil {
			helpText += " b=block"
		}
		helpText += " • o=sort • Tab: Switch tabs • q: Quit"
	case TabDaily:
		helpText = "\n  ↑/↓: Navigate • Tab: Switch tabs • q: Quit"
	}

	return HelpStyle.Render(helpText)
}

// Business logic methods
func (vm *ViewModel) GetTimeFilterString() string {
	switch vm.timeFilter {
	case FilterHour:
		return "Last Hour"
	case FilterDay:
		return "Last 24 Hours"
	case FilterWeek:
		return "Last 7 Days"
	case FilterMonth:
		return "Last 30 Days"
	case FilterBlock:
		if vm.Block() != nil {
			return "Current Block (" + vm.Block().FormatBlockTime(vm.timezone) + ")"
		}
		return "Block (not configured)"
	default:
		return "All Time"
	}
}

func (vm *ViewModel) GetSortOrderString() string {
	switch vm.sortOrder {
	case SortDescending:
		return "Latest First"
	case SortAscending:
		return "Oldest First"
	default:
		return "Latest First"
	}
}

func (vm *ViewModel) getTimePeriod() entity.Period {
	switch vm.timeFilter {
	case FilterHour:
		return entity.NewPeriodFromDuration(time.Now().UTC(), time.Hour)
	case FilterDay:
		return entity.NewPeriodFromDuration(time.Now().UTC(), 24*time.Hour)
	case FilterWeek:
		return entity.NewPeriodFromDuration(time.Now().UTC(), 7*24*time.Hour)
	case FilterMonth:
		return entity.NewPeriodFromDuration(time.Now().UTC(), 30*24*time.Hour)
	case FilterBlock:
		if vm.Block() != nil {
			return vm.Block().Period()
		}
		return entity.NewAllTimePeriod(time.Now().UTC())
	default:
		return entity.NewAllTimePeriod(time.Now().UTC())
	}
}

func (vm *ViewModel) refreshStats() tea.Msg {
	return refreshStatsMsg{}
}

func (vm *ViewModel) refreshUsage() tea.Msg {
	return refreshUsageMsg{}
}

// tick returns a command that sends a tick message using the configured refresh interval
func (vm *ViewModel) tick() tea.Cmd {
	return tea.Tick(vm.refreshInterval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Getter methods for compatibility with existing renderers
func (vm *ViewModel) Ready() bool {
	return vm.ready
}

func (vm *ViewModel) CurrentTab() Tab {
	return vm.currentTab
}

func (vm *ViewModel) Usage() entity.Usage {
	// Return usage from daily tab model
	return vm.dailyUsageTab.Usage()
}

func (vm *ViewModel) Timezone() *time.Location {
	return vm.timezone
}

func (vm *ViewModel) Block() *entity.Block {
	// Return block from overview tab stats model (it manages block state now)
	return vm.overviewTab.statsModel.Block()
}

func (vm *ViewModel) Stats() entity.Stats {
	// Return stats from overview tab stats model
	return vm.overviewTab.statsModel.Stats()
}

func (vm *ViewModel) BlockStats() entity.Stats {
	// Return block stats from overview tab stats model
	return vm.overviewTab.statsModel.BlockStats()
}

func (vm *ViewModel) Requests() []entity.APIRequest {
	// Return requests from overview tab requests table model
	return vm.overviewTab.requestsTableModel.Requests()
}

func (vm *ViewModel) Table() table.Model {
	// Return table from overview tab requests table model
	return vm.overviewTab.requestsTableModel.table
}

func (vm *ViewModel) TokenLimit() int {
	if vm.Block() != nil {
		return vm.Block().TokenLimit()
	}
	return 0
}

// Message types
type tickMsg time.Time
type refreshStatsMsg struct{}
type refreshUsageMsg struct{}
