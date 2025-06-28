package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/elct9620/ccmon/db"
	"github.com/elct9620/ccmon/monitor"
	"github.com/elct9620/ccmon/receiver"
)

func main() {
	// Parse command line flags
	var serverMode bool
	flag.BoolVar(&serverMode, "s", false, "Run as OTLP server (headless mode)")
	flag.BoolVar(&serverMode, "server", false, "Run as OTLP server (headless mode)")
	flag.Parse()

	if serverMode {
		if err := receiver.RunServer(func() (receiver.Database, error) {
			db, err := db.NewDatabase()
			return db, err
		}); err != nil {
			fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
			os.Exit(1)
		}
	} else {
		if err := monitor.RunMonitor(func() (monitor.Database, error) {
			db, err := db.NewDatabaseReadOnly()
			return db, err
		}); err != nil {
			fmt.Fprintf(os.Stderr, "Monitor error: %v\n", err)
			os.Exit(1)
		}
	}
}
