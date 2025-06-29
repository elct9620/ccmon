package tui

import (
	"context"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/elct9620/ccmon/entity"
	"github.com/elct9620/ccmon/handler/tui/components"
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

// SortOrder represents the sorting order for requests
type SortOrder int

const (
	SortDescending SortOrder = iota // Latest first (default)
	SortAscending                   // Oldest first
)

// ViewModel represents the state of our TUI monitor application
type ViewModel struct {
	requests            []entity.APIRequest
	table               table.Model
	width               int
	height              int
	ready               bool
	stats               entity.Stats // Stats for the current filter (displayed in statistics table)
	blockStats          entity.Stats // Stats for the current block (used for progress bar)
	getFilteredQuery    *usecase.GetFilteredApiRequestsQuery
	calculateStatsQuery *usecase.CalculateStatsQuery
	timeFilter          TimeFilter
	sortOrder           SortOrder
	timezone            *time.Location
	block               *entity.Block // nil if no block configured
	tokenLimit          int           // token limit for current block
	renderer            *Renderer     // renderer for the view
}

// NewViewModel creates a new ViewModel with initial state
func NewViewModel(getFilteredQuery *usecase.GetFilteredApiRequestsQuery, calculateStatsQuery *usecase.CalculateStatsQuery, timezone *time.Location, block *entity.Block, tokenLimit int) *ViewModel {
	// Start with basic columns, will be resized on first window size message
	columns := []table.Column{
		{Title: "Time", Width: 16},
		{Title: "Model", Width: 20},
		{Title: "Input", Width: 6},
		{Title: "Output", Width: 6},
		{Title: "Cache", Width: 6},
		{Title: "Total", Width: 6},
		{Title: "Cost ($)", Width: 8},
		{Title: "Duration", Width: 8},
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

	vm := &ViewModel{
		requests:            []entity.APIRequest{},
		table:               t,
		getFilteredQuery:    getFilteredQuery,
		calculateStatsQuery: calculateStatsQuery,
		timeFilter:          FilterAll,
		sortOrder:           SortDescending, // Default to latest first
		stats:               entity.Stats{},
		blockStats:          entity.Stats{},
		timezone:            timezone,
		block:               block,
		tokenLimit:          tokenLimit,
	}

	// Create components and renderer
	tableComponent := components.NewTableComponent(t)
	vm.renderer = NewRenderer(tableComponent)

	return vm
}

// Init is the Bubble Tea initialization function
func (vm *ViewModel) Init() tea.Cmd {
	return tea.Batch(
		tea.EnterAltScreen,
		vm.refreshStats, // Load initial data from database
		tick(),          // Start periodic refresh
	)
}

// Update handles messages and updates the model
func (vm *ViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return vm, tea.Quit
		case "esc":
			if vm.table.Focused() {
				vm.table.Blur()
			} else {
				vm.table.Focus()
			}
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
			if vm.block != nil {
				vm.timeFilter = FilterBlock
				return vm, vm.refreshStats
			}
			// If no block configured, ignore the key press
		case "o":
			// Toggle sort order
			if vm.sortOrder == SortDescending {
				vm.sortOrder = SortAscending
			} else {
				vm.sortOrder = SortDescending
			}
			return vm, vm.refreshStats
		}

	case tea.WindowSizeMsg:
		vm.width = msg.Width
		vm.height = msg.Height
		vm.ready = true
		// Resize table columns based on available width
		vm.resizeTableColumns()
		// Calculate dynamic table height based on content
		vm.adjustTableHeight()

	case tickMsg:
		// Periodic refresh
		return vm, tea.Batch(tick(), vm.refreshStats)

	case refreshStatsMsg:
		// Recalculate stats via usecase
		if vm.getFilteredQuery != nil {
			vm.recalculateStats()
		}
	}

	vm.table, cmd = vm.table.Update(msg)
	return vm, cmd
}

// Getters for accessing view model state
func (vm *ViewModel) Requests() []entity.APIRequest {
	return vm.requests
}

func (vm *ViewModel) Table() table.Model {
	return vm.table
}

func (vm *ViewModel) Ready() bool {
	return vm.ready
}

func (vm *ViewModel) Stats() entity.Stats {
	return vm.stats
}

func (vm *ViewModel) BlockStats() entity.Stats {
	return vm.blockStats
}

func (vm *ViewModel) Block() *entity.Block {
	return vm.block
}

func (vm *ViewModel) TokenLimit() int {
	return vm.tokenLimit
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
		if vm.block != nil {
			return "Current Block (" + vm.block.FormatBlockTime(time.Now()) + ")"
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
		return entity.NewPeriodFromDurationWithTimezone(time.Hour, vm.timezone)
	case FilterDay:
		return entity.NewPeriodFromDurationWithTimezone(24*time.Hour, vm.timezone)
	case FilterWeek:
		return entity.NewPeriodFromDurationWithTimezone(7*24*time.Hour, vm.timezone)
	case FilterMonth:
		return entity.NewPeriodFromDurationWithTimezone(30*24*time.Hour, vm.timezone)
	case FilterBlock:
		if vm.block != nil {
			return vm.block.CurrentBlock(time.Now())
		}
		return entity.NewAllTimePeriod()
	default:
		return entity.NewAllTimePeriod()
	}
}

func (vm *ViewModel) refreshStats() tea.Msg {
	return refreshStatsMsg{}
}

func (vm *ViewModel) recalculateStats() {
	period := vm.getTimePeriod()

	// Query for display requests (limit to 100 for TUI display)
	displayParams := usecase.GetFilteredApiRequestsParams{
		Period: period,
		Limit:  100,
		Offset: 0,
	}
	requests, err := vm.getFilteredQuery.Execute(context.Background(), displayParams)
	if err != nil {
		// Handle error silently for now
		return
	}
	vm.requests = requests

	// Apply sorting based on user preference
	if vm.sortOrder == SortDescending {
		// Reverse to show latest first (since DB returns chronological order)
		vm.reverseRequests()
	}
	// For SortAscending, keep the original order (oldest first)

	// Calculate filtered stats for display (always based on current filter)
	if vm.calculateStatsQuery != nil {
		statsParams := usecase.CalculateStatsParams{Period: period}
		stats, err := vm.calculateStatsQuery.Execute(context.Background(), statsParams)
		if err != nil {
			// Handle error silently for now, stats will remain empty
			vm.stats = entity.Stats{}
		} else {
			vm.stats = stats
		}
	}

	// Calculate block stats for progress bar (only when block tracking is enabled)
	if vm.block != nil && vm.calculateStatsQuery != nil {
		currentBlock := vm.block.CurrentBlock(time.Now())
		blockStatsParams := usecase.CalculateStatsParams{
			Period:          currentBlock,
			BlockTokenLimit: vm.tokenLimit,
			BlockStartTime:  currentBlock.StartAt(),
			BlockEndTime:    currentBlock.EndAt(),
		}
		blockStats, err := vm.calculateStatsQuery.Execute(context.Background(), blockStatsParams)
		if err != nil {
			// Keep previous block stats or use empty stats
			vm.blockStats = entity.Stats{}
		} else {
			vm.blockStats = blockStats
		}
	}

	// Update table
	vm.updateTableRows()
}

func (vm *ViewModel) reverseRequests() {
	for i, j := 0, len(vm.requests)-1; i < j; i, j = i+1, j-1 {
		vm.requests[i], vm.requests[j] = vm.requests[j], vm.requests[i]
	}
}

func (vm *ViewModel) updateTableRows() {
	rows := make([]table.Row, 0, len(vm.requests))
	for _, req := range vm.requests {
		// Format timestamp in configured timezone
		timestamp := req.Timestamp().In(vm.timezone).Format("15:04:05 2006-01-02")
		rows = append(rows, table.Row{
			timestamp,
			TruncateString(req.Model().String(), 25),
			FormatNumber(req.Tokens().Input()),
			FormatNumber(req.Tokens().Output()),
			FormatNumber(req.Tokens().Cache()),
			FormatNumber(req.Tokens().Total()),
			FormatCost(req.Cost().Amount()),
			FormatDuration(req.DurationMS()),
		})
	}
	vm.table.SetRows(rows)
}

func (vm *ViewModel) resizeTableColumns() {
	if vm.width < 80 {
		// Compact layout for narrow terminals
		columns := []table.Column{
			{Title: "Time", Width: 11},  // HH:MM:SS
			{Title: "Model", Width: 10}, // Shortened
			{Title: "In", Width: 4},     // Input tokens
			{Title: "Out", Width: 4},    // Output tokens
			{Title: "Tot", Width: 6},    // Total tokens
			{Title: "Cost", Width: 8},   // Cost
			{Title: "Dur", Width: 6},    // Duration
		}
		vm.table.SetColumns(columns)
	} else if vm.width < 120 {
		// Medium layout for normal terminals
		columns := []table.Column{
			{Title: "Time", Width: 16},
			{Title: "Model", Width: 18},
			{Title: "Input", Width: 6},
			{Title: "Output", Width: 6},
			{Title: "Cache", Width: 6},
			{Title: "Total", Width: 8},
			{Title: "Cost ($)", Width: 8},
			{Title: "Duration", Width: 8},
		}
		vm.table.SetColumns(columns)
	} else {
		// Full layout for wide terminals
		columns := []table.Column{
			{Title: "Time", Width: 20},
			{Title: "Model", Width: 25},
			{Title: "Input", Width: 8},
			{Title: "Output", Width: 8},
			{Title: "Cache", Width: 8},
			{Title: "Total", Width: 8},
			{Title: "Cost ($)", Width: 10},
			{Title: "Duration", Width: 10},
		}
		vm.table.SetColumns(columns)
	}

	// Update table rows to match new column layout
	vm.updateTableRows()
}

func (vm *ViewModel) adjustTableHeight() {
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

	if vm.block != nil && vm.tokenLimit > 0 {
		statsHeight += 4 // Progress bar section
	} else if vm.block == nil {
		statsHeight += 2 // Help message
	}

	// For compact stats, reduce height
	if vm.width < 60 {
		statsHeight = 8 // Compact stats are shorter
	}

	// Calculate remaining height for table with safety margin
	tableHeight := vm.height - fixedHeight - statsHeight - 2 // Extra 2 lines safety margin

	// Ensure reasonable minimum and maximum
	if tableHeight < 3 {
		tableHeight = 3
	} else if tableHeight > 20 {
		tableHeight = 20 // Cap maximum table height
	}

	vm.table.SetHeight(tableHeight)
}

// Message types
type tickMsg time.Time
type refreshStatsMsg struct{}

// View renders the UI using the renderer
func (vm *ViewModel) View() string {
	return vm.renderer.View(vm, vm.width)
}

// tick returns a command that sends a tick message every 5 seconds
func tick() tea.Cmd {
	return tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}
