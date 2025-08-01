package main

import (
	"embed"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/elct9620/ccmon/handler/cli"
	grpcserver "github.com/elct9620/ccmon/handler/grpc"
	"github.com/elct9620/ccmon/handler/tui"
	"github.com/elct9620/ccmon/repository"
	"github.com/elct9620/ccmon/service"
	"github.com/elct9620/ccmon/usecase"
	"github.com/spf13/pflag"
)

//go:embed data/*
var dataFS embed.FS

var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

// createStatsCache creates a stats cache implementation based on configuration
func createStatsCache(cacheConfig CacheStats) usecase.StatsCache {
	if !cacheConfig.Enabled {
		return &service.NoOpStatsCache{}
	}

	ttl, err := time.ParseDuration(cacheConfig.TTL)
	if err != nil {
		log.Printf("Invalid cache TTL '%s', using 1 minute default: %v", cacheConfig.TTL, err)
		ttl = time.Minute
	}

	return service.NewInMemoryStatsCache(ttl)
}

func main() {
	// Parse command line flags using pflag
	var serverMode bool
	var blockTime string
	var showVersion bool
	var formatString string
	pflag.BoolVarP(&serverMode, "server", "s", false, "Run as OTLP server (headless mode)")
	pflag.StringVarP(&blockTime, "block", "b", "", "Set block start time for token tracking (e.g., '5am', '11pm')")
	pflag.BoolVarP(&showVersion, "version", "v", false, "Show version information")
	pflag.StringVar(&formatString, "format", "", "Format string for quick query (e.g., '@daily_cost')")

	// Add help flag
	pflag.BoolP("help", "h", false, "Show help")

	// Load configuration (this will parse flags internally)
	config, err := LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Check for version flag after config is loaded
	if showVersion {
		if commit != "unknown" && commit != "" {
			shortCommit := commit
			if len(commit) > 7 {
				shortCommit = commit[:7]
			}
			fmt.Printf("ccmon version %s-%s (built %s)\n", version, shortCommit, date)
		} else {
			fmt.Printf("ccmon version %s\n", version)
		}
		os.Exit(0)
	}

	// Check for help flag after config is loaded
	if help, _ := pflag.CommandLine.GetBool("help"); help {
		pflag.Usage()
		os.Exit(0)
	}

	if serverMode {
		// Server mode: Use BoltDB repository
		db, err := NewDatabase(config.Database.Path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to initialize database: %v\n", err)
			os.Exit(1)
		}
		defer func() {
			if err := db.Close(); err != nil {
				log.Printf("Error closing database: %v", err)
			}
		}()

		repo := repository.NewBoltDBAPIRequestRepository(db)

		// Create cache
		statsCache := createStatsCache(config.Server.Cache.Stats)

		// Create stats repository for server side
		statsRepo := repository.NewBoltDBStatsRepository(repo)

		// Create usecases
		appendCommand := usecase.NewAppendApiRequestCommand(repo)
		getFilteredQuery := usecase.NewGetFilteredApiRequestsQuery(repo)
		calculateStatsQuery := usecase.NewCalculateStatsQuery(statsRepo, statsCache)
		cleanupCommand := usecase.NewCleanupOldRecordsCommand(repo)
		// Note: getUsageQuery would be used if we add usage endpoints to gRPC server
		// Server mode uses UTC timezone for consistency
		periodFactory := service.NewTimePeriodFactory(time.UTC)
		_ = usecase.NewGetUsageQuery(repo, periodFactory) // Avoid unused variable

		// Run server with usecases
		if err := grpcserver.RunServer(config.Server.Address, appendCommand, getFilteredQuery, calculateStatsQuery, cleanupCommand, &config.Server); err != nil {
			fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
			os.Exit(1)
		}
	} else {
		// Monitor mode: Use gRPC repository
		repo, err := repository.NewGRPCAPIRequestRepository(config.Monitor.Server)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to initialize gRPC repository: %v\n", err)
			os.Exit(1)
		}
		defer func() {
			if err := repo.Close(); err != nil {
				log.Printf("Error closing gRPC repository: %v", err)
			}
		}()

		// Create cache
		statsCache := createStatsCache(config.Server.Cache.Stats)

		// Create gRPC stats repository for TUI mode
		tuiStatsRepo, err := repository.NewGRPCStatsRepository(config.Monitor.Server)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to initialize gRPC stats repository: %v\n", err)
			os.Exit(1)
		}
		defer func() {
			if err := tuiStatsRepo.Close(); err != nil {
				log.Printf("Error closing TUI stats repository: %v", err)
			}
		}()

		// Create query usecases (no append command needed for monitor)
		getFilteredQuery := usecase.NewGetFilteredApiRequestsQuery(repo)
		calculateStatsQuery := usecase.NewCalculateStatsQuery(tuiStatsRepo, statsCache)
		timezone, err := time.LoadLocation(config.Monitor.Timezone)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid timezone: %v\n", err)
			os.Exit(1)
		}
		periodFactory := service.NewTimePeriodFactory(timezone)
		getUsageQuery := usecase.NewGetUsageQuery(repo, periodFactory)

		// Convert config to TUI-specific struct
		// Handle format query mode - bypass TUI and output directly to stdout
		if formatString != "" {
			// Create plan repository for usage percentage calculations
			planRepository, err := repository.NewEmbeddedPlanRepository(config, dataFS)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to initialize plan repository: %v\n", err)
				os.Exit(1)
			}

			// Create gRPC stats repository for efficient stats retrieval
			statsRepo, err := repository.NewGRPCStatsRepository(config.Monitor.Server)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to initialize stats repository: %v\n", err)
				os.Exit(1)
			}
			defer func() {
				if err := statsRepo.Close(); err != nil {
					log.Printf("Error closing stats repository: %v", err)
				}
			}()

			// Create CalculateStatsQuery that uses gRPC StatsRepository
			formatCalculateStatsQuery := usecase.NewCalculateStatsQuery(statsRepo, statsCache)

			// Create GetUsageVariablesQuery with format-optimized dependencies
			usageVariablesQuery := usecase.NewGetUsageVariablesQuery(
				formatCalculateStatsQuery,
				planRepository,
				periodFactory,
			)

			// Create format renderer and query handler
			renderer := cli.NewFormatRenderer(usageVariablesQuery)
			queryHandler := cli.NewQueryHandler(renderer)

			if err := queryHandler.HandleFormatQuery(formatString); err != nil {
				os.Exit(1)
			}
			os.Exit(0)
		}

		monitorConfig := tui.MonitorConfig{
			Server:          config.Monitor.Server,
			Timezone:        config.Monitor.Timezone,
			RefreshInterval: config.Monitor.RefreshInterval,
			TokenLimit:      config.Claude.GetTokenLimit(),
			BlockTime:       blockTime,
		}

		// Run monitor with usecases and config - TUI handler owns block logic
		if err := tui.RunMonitor(getFilteredQuery, calculateStatsQuery, getUsageQuery, monitorConfig); err != nil {
			fmt.Fprintf(os.Stderr, "Monitor error: %v\n", err)
			os.Exit(1)
		}
	}
}
