package main

import (
	"flag"
	"fmt"
	"os"

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
		repo, err := repository.NewBoltDBAPIRequestRepository(config.Database.Path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to initialize repository: %v\n", err)
			os.Exit(1)
		}
		defer repo.Close()

		// Create usecases
		appendCommand := usecase.NewAppendApiRequestCommand(repo)
		getAllQuery := usecase.NewGetAllApiRequestsQuery(repo)
		getFilteredQuery := usecase.NewGetFilteredApiRequestsQuery(repo)
		getStatsQuery := usecase.NewGetStatsQuery(repo)

		// Run server with usecases
		if err := grpcserver.RunServer(config.Server.Address, appendCommand, getAllQuery, getFilteredQuery, getStatsQuery); err != nil {
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

		// Run monitor with usecase
		if err := tui.RunMonitor(getFilteredQuery); err != nil {
			fmt.Fprintf(os.Stderr, "Monitor error: %v\n", err)
			os.Exit(1)
		}
	}
}
