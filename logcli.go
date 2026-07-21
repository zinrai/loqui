package main

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"
)

// getLabels retrieves available labels from Loki via logcli
func getLabels(config *Config) ([]string, error) {
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
