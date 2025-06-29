package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/elct9620/ccmon/db"
	grpcserver "github.com/elct9620/ccmon/handler/grpc"
	"github.com/elct9620/ccmon/handler/tui"
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
		// Initialize database for server mode (read-write)
		database, err := db.NewDatabase(config.Database.Path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to initialize database: %v\n", err)
			os.Exit(1)
		}
		defer database.Close()

		if err := grpcserver.RunServer(config.Server.Address, database); err != nil {
			fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
			os.Exit(1)
		}
	} else {
		// Run monitor mode with gRPC client
		if err := tui.RunMonitor(config.Monitor.Server); err != nil {
			fmt.Fprintf(os.Stderr, "Monitor error: %v\n", err)
			os.Exit(1)
		}
	}
}
