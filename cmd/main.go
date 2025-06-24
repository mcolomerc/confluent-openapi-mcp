package main

import (
	"context"
	"flag"
	"fmt"
	"mcolomerc/mcp-server/internal/config"
	"mcolomerc/mcp-server/internal/monitoring"
	"mcolomerc/mcp-server/internal/openapi"
	"mcolomerc/mcp-server/internal/server"
	"mcolomerc/mcp-server/internal/tools"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	fmt.Fprintf(os.Stderr, "Starting server...v3 \n")

	// Parse command line arguments
	envFile := flag.String("env", "", "Path to environment file")
	mode := flag.String("mode", "both", "Server mode: 'stdio', 'http', or 'both'")
	monitorInterval := flag.String("monitor", "30s", "Resource monitoring interval (e.g., 30s, 1m, 5m). Set to 'off' to disable")
	flag.Parse()

	// Setup context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Setup resource monitoring if enabled
	var monitor *monitoring.Monitor
	if *monitorInterval != "off" {
		interval, err := time.ParseDuration(*monitorInterval)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid monitor interval '%s', using default 30s: %v\n", *monitorInterval, err)
			interval = 30 * time.Second
		}

		monitor = monitoring.NewMonitor(interval)
		fmt.Fprintf(os.Stderr, "Resource monitoring enabled with %v interval\n", interval)

		// Start monitoring in a separate goroutine
		go monitor.StartPeriodicLogging(ctx)

		// Log initial metrics
		fmt.Fprintf(os.Stderr, "Initial resource metrics:\n")
		if metricsJSON, err := monitor.GetMetricsJSON(); err == nil {
			fmt.Fprintf(os.Stderr, "%s\n", metricsJSON)
		}
	}

	// Load environment configuration
	envPath := ".env"
	if *envFile != "" {
		envPath = *envFile
	}
	cfg, err := config.LoadConfig(envPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Load and parse OpenAPI spec
	spec, err := openapi.LoadSpec()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load OpenAPI spec: %v\n", err)
		os.Exit(1)
	}

	// Generate semantic tools from OpenAPI
	semanticTools, err := tools.GenerateSemanticTools(*spec)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to generate semantic tools: %v\n", err)
		os.Exit(1)
	}

	// Create the composite MCPServer instance with config, spec and semanticTools
	mcpServer := server.NewCompositeServer(cfg, spec, semanticTools)

	// Connect monitor to server if monitoring is enabled
	if monitor != nil {
		mcpServer.SetMonitor(monitor)
	}

	// Start server in a separate goroutine
	serverErrCh := make(chan error, 1)
	go func() {
		err := mcpServer.StartWithMode(*mode, ":8080")
		if err != nil {
			serverErrCh <- err
		}
	}()

	// Wait for shutdown signal or server error
	select {
	case sig := <-sigChan:
		fmt.Fprintf(os.Stderr, "Received signal %v, shutting down gracefully...\n", sig)
		cancel() // Cancel context to stop monitoring

		// Log final metrics if monitoring is enabled
		if monitor != nil {
			fmt.Fprintf(os.Stderr, "Final resource metrics:\n")
			if metricsJSON, err := monitor.GetMetricsJSON(); err == nil {
				fmt.Fprintf(os.Stderr, "%s\n", metricsJSON)
			}
			monitor.Stop()
		}

	case err := <-serverErrCh:
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		cancel()
		if monitor != nil {
			monitor.Stop()
		}
		os.Exit(1)
	}
}
