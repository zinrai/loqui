package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Simple cache structure for reading
type LabelCache struct {
	Labels []string            `json:"labels"`
	Values map[string][]string `json:"values"`
}

// getLabels retrieves available labels from Loki via logcli
func getLabels(config *Config) ([]string, error) {
	// Check cache first if enabled
	if config.UseCache {
		cached, err := loadCachedLabels(config.CacheDir)
		if err == nil && cached != nil {
			return cached.Labels, nil
		}
		// If cache fails, fall back to logcli
	}

	// Get from logcli with time range
	return getLabelsFromLogCLI(config.LogCLICmd, config.TimeArgs)
}

// getLabelsFromLogCLI executes logcli to get labels
func getLabelsFromLogCLI(logcliCmd string, timeArgs []string) ([]string, error) {
	args := []string{"labels", "--quiet"}
	args = append(args, timeArgs...)

	cmd := exec.Command(logcliCmd, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("logcli labels failed: %w\nOutput: %s", err, string(output))
	}

	return parseLabelsOutput(string(output))
}

// parseLabelsOutput parses the output of 'logcli labels' command
func parseLabelsOutput(output string) ([]string, error) {
	labels := []string{}
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			labels = append(labels, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to parse labels: %w", err)
	}

	return labels, nil
}

// getLabelValues retrieves values for a specific label
func getLabelValues(config *Config, label string) ([]string, error) {
	// Check cache first if enabled
	if config.UseCache {
		cached, err := loadCachedLabels(config.CacheDir)
		if err == nil && cached != nil {
			if values, ok := cached.Values[label]; ok {
				return values, nil
			}
		}
		// If cache fails, fall back to logcli
	}

	// Get from logcli with time range
	return getLabelValuesFromLogCLI(config.LogCLICmd, label, config.TimeArgs)
}

// getLabelValuesFromLogCLI executes logcli to get label values
func getLabelValuesFromLogCLI(logcliCmd string, label string, timeArgs []string) ([]string, error) {
	args := []string{"labels", label, "--quiet"}
	args = append(args, timeArgs...)

	cmd := exec.Command(logcliCmd, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("logcli labels %s failed: %w\nOutput: %s", label, err, string(output))
	}

	return parseLabelValuesOutput(string(output))
}

// parseLabelValuesOutput parses the output of 'logcli labels <label>' command
func parseLabelValuesOutput(output string) ([]string, error) {
	values := []string{}
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			values = append(values, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to parse label values: %w", err)
	}

	return values, nil
}

// loadCachedLabels loads labels from cache if available
func loadCachedLabels(cacheDir string) (*LabelCache, error) {
	cacheFile := filepath.Join(cacheDir, "labels.json")

	// Check if cache file exists
	if _, err := os.Stat(cacheFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("cache file not found")
	}

	// Read cache file
	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read cache file: %w", err)
	}

	// Parse cache
	var cache LabelCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, fmt.Errorf("failed to parse cache: %w", err)
	}

	return &cache, nil
}
