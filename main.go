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
		if err := grpcserver.RunServer(config.Server.Address, func() (grpcserver.Database, error) {
			db, err := db.NewDatabase(config.Database.Path)
			return db, err
		}); err != nil {
			fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
			os.Exit(1)
		}
	} else {
		if err := tui.RunMonitor(func() (tui.Database, error) {
			db, err := db.NewDatabaseReadOnly(config.Database.Path)
			return db, err
		}); err != nil {
			fmt.Fprintf(os.Stderr, "Monitor error: %v\n", err)
			os.Exit(1)
		}
	}
}
