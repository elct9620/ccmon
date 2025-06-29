package main

import (
	"fmt"
	"os"
	"time"

	"github.com/elct9620/ccmon/entity"
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
		defer db.Close()

		repo := repository.NewBoltDBAPIRequestRepository(db)

		// Create usecases
		appendCommand := usecase.NewAppendApiRequestCommand(repo)
		getFilteredQuery := usecase.NewGetFilteredApiRequestsQuery(repo)
		calculateStatsQuery := usecase.NewCalculateStatsQuery(repo)

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
		defer repo.Close()

		// Create query usecases (no append command needed for monitor)
		getFilteredQuery := usecase.NewGetFilteredApiRequestsQuery(repo)
		calculateStatsQuery := usecase.NewCalculateStatsQuery(repo)

		// Load timezone for monitor mode
		timezone, err := time.LoadLocation(config.Monitor.Timezone)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to load timezone %s: %v\n", config.Monitor.Timezone, err)
			os.Exit(1)
		}

		// Parse block configuration if provided
		var block *entity.Block
		var tokenLimit int
		if blockTime != "" {
			startHour, err := entity.ParseBlockTime(blockTime)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Invalid block time format %s: %v\n", blockTime, err)
				os.Exit(1)
			}

			blockEntity := entity.NewBlock(startHour, timezone)
			block = &blockEntity
			tokenLimit = config.Claude.GetTokenLimit()

			if tokenLimit == 0 {
				fmt.Fprintf(os.Stderr, "Warning: No token limit configured. Set claude.plan or claude.max_tokens in config.\n")
			}
		}

		// Run monitor with usecases, timezone, and block config
		if err := tui.RunMonitor(getFilteredQuery, calculateStatsQuery, timezone, block, tokenLimit); err != nil {
			fmt.Fprintf(os.Stderr, "Monitor error: %v\n", err)
			os.Exit(1)
		}
	}
}
