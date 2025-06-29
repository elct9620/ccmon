package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	grpcserver "github.com/elct9620/ccmon/handler/grpc"
	"github.com/elct9620/ccmon/handler/tui"
	"github.com/elct9620/ccmon/repository"
	"github.com/elct9620/ccmon/usecase"
)

func main() {
	// Parse command line flags
	var serverMode bool
	flag.BoolVar(&serverMode, "s", false, "Run as OTLP server (headless mode)")
	flag.BoolVar(&serverMode, "server", false, "Run as OTLP server (headless mode)")
	flag.Parse()

	// Load configuration
	config, err := LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
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
		getStatsQuery := usecase.NewGetStatsQuery(repo)

		// Run server with usecases
		if err := grpcserver.RunServer(config.Server.Address, appendCommand, getFilteredQuery, getStatsQuery); err != nil {
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

		// Create query usecase (no append command needed for monitor)
		getFilteredQuery := usecase.NewGetFilteredApiRequestsQuery(repo)

		// Load timezone for monitor mode
		timezone, err := time.LoadLocation(config.Monitor.Timezone)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to load timezone %s: %v\n", config.Monitor.Timezone, err)
			os.Exit(1)
		}

		// Run monitor with usecase and timezone
		if err := tui.RunMonitor(getFilteredQuery, timezone); err != nil {
			fmt.Fprintf(os.Stderr, "Monitor error: %v\n", err)
			os.Exit(1)
		}
	}
}
