package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/BlueBeard63/archon-node/internal/api"
	"github.com/BlueBeard63/archon-node/internal/config"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "/etc/archon/node-config.toml", "Path to configuration file")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Validate API key
	if cfg.Server.APIKey == "" {
		fmt.Fprintf(os.Stderr, "ERROR: API key is not set in configuration\n")
		fmt.Fprintf(os.Stderr, "Please edit %s and set server.api_key\n", *configPath)
		os.Exit(1)
	}

	// Create API server
	server, err := api.NewServer(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create server: %v\n", err)
		os.Exit(1)
	}

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start server in goroutine
	errChan := make(chan error, 1)
	go func() {
		if err := server.Start(); err != nil {
			errChan <- err
		}
	}()

	// Wait for signal or error
	select {
	case <-sigChan:
		fmt.Println("\nReceived shutdown signal")
	case err := <-errChan:
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
	}

	// Shutdown server
	if err := server.Shutdown(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error during shutdown: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Server shut down successfully")
}
