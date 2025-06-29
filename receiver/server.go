package receiver

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	dbpkg "github.com/elct9620/ccmon/db"
)

// RunServer runs the headless OTLP server mode
func RunServer(address string, newDB func() (Database, error)) error {
	log.Println("Starting ccmon in server mode...")

	// Initialize database
	db, err := newDB()
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer db.Close()

	// Create and start the OTLP receiver
	receiver := NewReceiver(nil, nil, db) // No channel or TUI program needed

	// Create a context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("Shutting down server...")
		cancel()
	}()

	// Start request counter
	go logRequestStats(ctx, db)

	// Start the receiver
	log.Printf("OTLP receiver listening on %s\n", address)
	if err := receiver.Start(ctx, address); err != nil {
		return fmt.Errorf("failed to start receiver: %w", err)
	}

	// Cleanup
	receiver.Stop()
	log.Println("Server stopped")
	return nil
}

// logRequestStats periodically logs request statistics
func logRequestStats(ctx context.Context, db Database) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Get all requests to calculate stats
			requests, err := db.GetAllRequests()
			if err != nil {
				log.Printf("Error reading stats: %v", err)
				continue
			}

			// Calculate stats
			baseReqs, premiumReqs, baseTokens, premiumTokens, _, _, _, _, baseCost, premiumCost := dbpkg.CalculateStats(requests)
			totalReqs := baseReqs + premiumReqs
			totalTokens := baseTokens + premiumTokens
			totalCost := baseCost + premiumCost

			log.Printf("Stats - Requests: %d (Base: %d, Premium: %d) | Tokens: %d | Cost: $%.6f",
				totalReqs, baseReqs, premiumReqs, totalTokens, totalCost)
		}
	}
}
