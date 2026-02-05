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

const (
	defaultConfigPath = "/etc/archon/node-config.toml"
)

func main() {
	// Check if we have a subcommand
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "configure":
			handleConfigure(os.Args[2:])
			return
		case "help", "-h", "--help":
			printUsage()
			return
		case "version", "-v", "--version":
			fmt.Println("archon-node version 1.0.0")
			return
		}
	}

	// Default: run server
	runServer()
}

func printUsage() {
	fmt.Print(`Usage: archon-node [command] [options]

Commands:
  configure   Fetch configuration from a remote URL
  help        Show this help message
  version     Show version information

Server Mode (default):
  archon-node [options]
    -config string  Path to configuration file (default "/etc/archon/node-config.toml")

Configure Mode:
  archon-node configure [options]
    --from-url string   URL to fetch configuration from (pastebin-style)
    --from-file string  Path to local configuration file
    --output string     Path to save configuration (default "/etc/archon/node-config.toml")

Examples:
  # Start the node server
  archon-node

  # Start with custom config path
  archon-node -config /path/to/config.toml

  # Configure from remote URL (quick configure)
  archon-node configure --from-url https://master-node:8080/api/v1/configs/abc123

  # Configure from local file
  archon-node configure --from-file /tmp/config.toml
`)
}

func handleConfigure(args []string) {
	configureCmd := flag.NewFlagSet("configure", flag.ExitOnError)
	fromURL := configureCmd.String("from-url", "", "URL to fetch configuration from")
	fromFile := configureCmd.String("from-file", "", "Path to local configuration file")
	outputPath := configureCmd.String("output", defaultConfigPath, "Path to save configuration")

	if err := configureCmd.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		os.Exit(1)
	}

	// Validate flags
	if *fromURL == "" && *fromFile == "" {
		fmt.Fprintf(os.Stderr, "Error: Either --from-url or --from-file must be specified\n")
		configureCmd.Usage()
		os.Exit(1)
	}

	if *fromURL != "" && *fromFile != "" {
		fmt.Fprintf(os.Stderr, "Error: Cannot specify both --from-url and --from-file\n")
		os.Exit(1)
	}

	if *fromURL != "" {
		// Fetch config from URL
		fmt.Printf("Fetching configuration from: %s\n", *fromURL)
		cfg, err := config.FetchConfigFromURL(*fromURL, *outputPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to fetch configuration: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Configuration saved to: %s\n", *outputPath)
		fmt.Printf("Server will listen on %s:%d\n", cfg.Server.Host, cfg.Server.Port)
	} else {
		// Copy from local file
		fmt.Printf("Copying configuration from: %s\n", *fromFile)

		// Load and validate the config
		cfg, err := config.Load(*fromFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
			os.Exit(1)
		}

		// Save to output path
		if err := config.Save(*outputPath, cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to save configuration: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Configuration saved to: %s\n", *outputPath)
		fmt.Printf("Server will listen on %s:%d\n", cfg.Server.Host, cfg.Server.Port)
	}

	fmt.Println("\nTo start the node, run:")
	fmt.Printf("  archon-node -config %s\n", *outputPath)
}

func runServer() {
	// Parse command line flags
	configPath := flag.String("config", defaultConfigPath, "Path to configuration file")
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
