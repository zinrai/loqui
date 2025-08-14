package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

const (
	version = "0.1.0"
	usage   = `loqui - Interactive Loki Query Builder

Usage:
  loqui [options]

Options:
  -help        Show this help message
  -version     Show version
  -cache       Enable cache usage (default: disabled)

Examples:
  # Interactive query building
  $(loqui)

  # Build query using cache
  $(loqui -cache)
`
)

type Config struct {
	UseCache  bool
	CacheDir  string
	LogCLICmd string
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

	config := &Config{
		UseCache:  useCache,
		CacheDir:  getDefaultCacheDir(),
		LogCLICmd: "logcli",
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
