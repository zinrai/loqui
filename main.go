package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

const (
	version = "0.2.0"
	usage   = `loqui - Interactive Loki Query Builder

Usage:
  loqui [options]

Options:
  -help        Show this help message
  -version     Show version
  -cache       Enable cache usage (default: disabled)

Environment:
  LOKI_ADDR    Loki server address (required)
               Example: htp://localhost:3100

Examples:
  # Set Loki address and run interactive query building
  export LOKI_ADDR=http://localhost:3100
  $(loqui)

  # Build query using cache
  $(loqui -cache)
`
)

type Config struct {
	UseCache  bool
	CacheDir  string
	LogCLICmd string
	TimeArgs  []string // Added to store time range arguments
}

func main() {
	var (
		showHelp    bool
		showVersion bool
		useCache    bool
	)

	flag.BoolVar(&showHelp, "help", false, "Show help")
	flag.BoolVar(&showVersion, "version", false, "Show version")
	flag.BoolVar(&useCache, "cache", false, "Enable cache usage")

	flag.Usage = func() {
		fmt.Fprint(os.Stderr, usage)
	}

	flag.Parse()

	if showHelp {
		fmt.Print(usage)
		os.Exit(0)
	}

	if showVersion {
		fmt.Printf("loqui version %s\n", version)
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
		UseCache:  useCache,
		CacheDir:  getDefaultCacheDir(),
		LogCLICmd: "logcli",
		TimeArgs:  []string{}, // Initialize as empty, will be set in InteractiveQueryBuilder
	}

	// Run interactive mode
	if err := InteractiveQueryBuilder(config); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// getDefaultCacheDir returns the default cache directory following XDG Base Directory specification
func getDefaultCacheDir() string {
	// First, check XDG_CACHE_HOME
	if xdgCache := os.Getenv("XDG_CACHE_HOME"); xdgCache != "" {
		return filepath.Join(xdgCache, "loqui")
	}

	// Fall back to ~/.cache
	if homeDir, err := os.UserHomeDir(); err == nil {
		return filepath.Join(homeDir, ".cache", "loqui")
	}

	// Last resort fallback
	return filepath.Join(os.TempDir(), "loqui-cache")
}
