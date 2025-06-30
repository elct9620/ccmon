package main

import (
	"fmt"
	"log"
	"os"

	grpcserver "github.com/elct9620/ccmon/handler/grpc"
	"github.com/elct9620/ccmon/handler/tui"
	"github.com/elct9620/ccmon/repository"
	"github.com/elct9620/ccmon/usecase"
	"github.com/spf13/pflag"
)

func main() {
	// Parse command line flags using pflag
	var serverMode bool
	var blockTime string
	pflag.BoolVarP(&serverMode, "server", "s", false, "Run as OTLP server (headless mode)")
	pflag.StringVarP(&blockTime, "block", "b", "", "Set block start time for token tracking (e.g., '5am', '11pm')")

	// Add help flag
	pflag.BoolP("help", "h", false, "Show help")

	// Load configuration (this will parse flags internally)
	config, err := LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
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

		// Create usecases
		appendCommand := usecase.NewAppendApiRequestCommand(repo)
		getFilteredQuery := usecase.NewGetFilteredApiRequestsQuery(repo)
		calculateStatsQuery := usecase.NewCalculateStatsQuery(repo)
		// Note: getUsageQuery would be used if we add usage endpoints to gRPC server
		_ = usecase.NewGetUsageQuery(repo) // Avoid unused variable

		// Run server with usecases
		if err := grpcserver.RunServer(config.Server.Address, appendCommand, getFilteredQuery, calculateStatsQuery); err != nil {
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

		// Create query usecases (no append command needed for monitor)
		getFilteredQuery := usecase.NewGetFilteredApiRequestsQuery(repo)
		calculateStatsQuery := usecase.NewCalculateStatsQuery(repo)
		getUsageQuery := usecase.NewGetUsageQuery(repo)

		// Convert config to TUI-specific struct
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
