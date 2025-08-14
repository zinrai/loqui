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

	// Get from logcli
	return getLabelsFromLogCLI(config.LogCLICmd)
}

// getLabelsFromLogCLI executes logcli to get labels
func getLabelsFromLogCLI(logcliCmd string) ([]string, error) {
	cmd := exec.Command(logcliCmd, "labels")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("logcli labels failed: %w", err)
	}

	labels := []string{}
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "http") {
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

	// Get from logcli
	return getLabelValuesFromLogCLI(config.LogCLICmd, label)
}

// getLabelValuesFromLogCLI executes logcli to get label values
func getLabelValuesFromLogCLI(logcliCmd string, label string) ([]string, error) {
	cmd := exec.Command(logcliCmd, "labels", label)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("logcli labels %s failed: %w", label, err)
	}

	values := []string{}
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "http") {
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
