package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Initialize database
	db, err := NewDatabase()
	if err != nil {
		fmt.Println("Failed to initialize database:", err)
		os.Exit(1)
	}
	defer db.Close()

	// Create a channel for API requests
	requestChan := make(chan APIRequest, 100)

	// Create the Bubble Tea model
	model := NewModel(requestChan, db)

	// Create the Bubble Tea program
	p := tea.NewProgram(model, tea.WithAltScreen())

	// Create and start the OTLP receiver
	receiver := NewReceiver(requestChan, p, db)

	// Create a context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the receiver in a goroutine
	go func() {
		if err := receiver.Start(ctx); err != nil {
			log.Printf("Failed to start receiver: %v", err)
			p.Quit()
		}
	}()

	// Handle interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		cancel()
		p.Quit()
	}()

	// Run the Bubble Tea program
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}

	// Cleanup
	receiver.Stop()
}
