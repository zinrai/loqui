package main

import (
	"flag"
	"fmt"
	"os"
)

// Populated at build time via goreleaser ldflags (-X main.version, etc.)
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

const usage = `loqui - Interactive Loki Query Builder

Usage:
  loqui [options]

Options:
  -help        Show this help message
  -version     Show version
  -exec        Execute the command immediately

Environment:
  LOKI_ADDR    Loki server address (required)
               Example: http://localhost:3100

Examples:
  # Set Loki address and run interactive query building
  export LOKI_ADDR=http://localhost:3100
  $(loqui)

  # Execute query immediately
  loqui -exec
`

type Config struct {
	LogCLICmd string
	TimeArgs  []string // Added to store time range arguments
	Execute   bool     // Added for -exec option
}

func main() {
	var (
		showHelp    bool
		showVersion bool
		execute     bool
	)

	flag.BoolVar(&showHelp, "help", false, "Show help")
	flag.BoolVar(&showVersion, "version", false, "Show version")
	flag.BoolVar(&execute, "exec", false, "Execute the command immediately")

	flag.Usage = func() {
		fmt.Fprint(os.Stderr, usage)
	}

	flag.Parse()

	if showHelp {
		fmt.Print(usage)
		os.Exit(0)
	}

	if showVersion {
		fmt.Printf("loqui version %s (commit %s, built %s)\n", version, commit, date)
		os.Exit(0)
	}

	// Check LOKI_ADDR environment variable
	lokiAddr := os.Getenv("LOKI_ADDR")
	if lokiAddr == "" {
		fmt.Fprintf(os.Stderr, "Error: LOKI_ADDR environment variable is not set\n")
		fmt.Fprintf(os.Stderr, "Please set it to your Loki server address\n")
		fmt.Fprintf(os.Stderr, "Example: export LOKI_ADDR=http://localhost:3100\n")
		os.Exit(1)
	}

	config := &Config{
		LogCLICmd: "logcli",
		TimeArgs:  []string{}, // Initialize as empty, will be set in InteractiveQueryBuilder
		Execute:   execute,
	}

	// Run interactive mode
	if err := InteractiveQueryBuilder(config); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
