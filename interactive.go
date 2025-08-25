package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type LabelSelector struct {
	Label    string
	Operator string
	Value    string
}

type LineFilter struct {
	Operator string
	Text     string
}

func InteractiveQueryBuilder(config *Config) error {
	// 1. Select time range (FIRST - to use for label queries)
	timeArgs, err := selectTimeRange()
	if err != nil {
		return fmt.Errorf("time range selection failed: %w", err)
	}

	// Set timeArgs in config for use in label queries
	config.TimeArgs = timeArgs

	// 2. Select labels
	selectors, err := selectLabels(config)
	if err != nil {
		return fmt.Errorf("label selection failed: %w", err)
	}

	// 3. Select line filter
	lineFilter, err := selectLineFilter()
	if err != nil {
		return fmt.Errorf("line filter selection failed: %w", err)
	}

	// 4. Build command arguments
	args := buildLogCLIArgs(config.LogCLICmd, selectors, lineFilter, timeArgs)

	// 5. Execute or output command
	if config.Execute {
		// Execute mode
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("execution failed: %w", err)
		}
	} else {
		// Output mode (default)
		fmt.Println(formatAsShellCommand(args))
	}

	return nil
}

func selectLabels(config *Config) ([]LabelSelector, error) {
	selectors := []LabelSelector{}

	for {
		// Show current labels
		showCurrentLabels(selectors)

		// Get available labels
		availableLabels, err := getAvailableLabels(config, selectors)
		if err != nil {
			return nil, err
		}

		if len(availableLabels) == 0 {
			fmt.Println("No more labels available.")
			break
		}

		// Select one label with operator and value
		selector, err := selectLabelWithOperatorAndValue(config, availableLabels)
		if err != nil {
			return nil, err
		}

		selectors = append(selectors, selector)

		// Ask if more labels needed
		continueAdding, err := promptForMoreLabels()
		if err != nil {
			return nil, err
		}
		if !continueAdding {
			break
		}
	}

	return selectors, nil
}

func showCurrentLabels(selectors []LabelSelector) {
	if len(selectors) > 0 {
		fmt.Println("\n=== Current labels ===")
		for _, s := range selectors {
			fmt.Printf("[SET] %s%s\"%s\"\n", s.Label, s.Operator, s.Value)
		}
	}
}

func getAvailableLabels(config *Config, selectors []LabelSelector) ([]string, error) {
	// Get all labels
	labels, err := getLabels(config)
	if err != nil {
		return nil, fmt.Errorf("failed to get labels: %w", err)
	}

	// Filter out already selected labels
	selectedLabels := make(map[string]bool)
	for _, s := range selectors {
		selectedLabels[s.Label] = true
	}

	availableLabels := []string{}
	for _, label := range labels {
		if !selectedLabels[label] {
			availableLabels = append(availableLabels, label)
		}
	}

	return availableLabels, nil
}

func selectLabelWithOperatorAndValue(config *Config, availableLabels []string) (LabelSelector, error) {
	// Select label
	label, err := selectWithFzf(availableLabels, "Select label:")
	if err != nil {
		return LabelSelector{}, fmt.Errorf("label selection failed: %w", err)
	}

	// Select operator
	operator, err := selectOperator(label)
	if err != nil {
		return LabelSelector{}, fmt.Errorf("operator selection failed: %w", err)
	}

	// Select or input value
	value, err := selectOrInputValue(config, label, operator)
	if err != nil {
		return LabelSelector{}, fmt.Errorf("value selection failed: %w", err)
	}

	return LabelSelector{
		Label:    label,
		Operator: operator,
		Value:    value,
	}, nil
}

func selectOrInputValue(config *Config, label string, operator string) (string, error) {
	if operator == "=" || operator == "!=" {
		// For equality operators, select from existing values
		values, err := getLabelValues(config, label)
		if err != nil {
			return "", fmt.Errorf("failed to get label values: %w", err)
		}
		return selectWithFzf(values, fmt.Sprintf("Select value for '%s':", label))
	} else {
		// For regex operators, input pattern
		return inputText(fmt.Sprintf("Enter regex pattern for '%s': ", label))
	}
}

func promptForMoreLabels() (bool, error) {
	fmt.Print("\nAdd more labels? (y/N): ")
	answer, err := inputText("")
	if err != nil {
		return false, err
	}
	answer = strings.ToLower(strings.TrimSpace(answer))
	return answer == "y" || answer == "yes", nil
}

func selectLineFilter() (*LineFilter, error) {
	fmt.Print("\nAdd line filter? (y/N): ")
	answer, err := inputText("")
	if err != nil {
		return nil, err
	}
	answer = strings.ToLower(strings.TrimSpace(answer))

	if answer != "y" && answer != "yes" {
		return nil, nil
	}

	// Select line filter operator
	operator, err := selectLineFilterOperator()
	if err != nil {
		return nil, err
	}

	// Input filter text
	fmt.Print("Enter filter text: ")
	text, err := inputText("")
	if err != nil {
		return nil, err
	}

	return &LineFilter{
		Operator: operator,
		Text:     text,
	}, nil
}

func selectLineFilterOperator() (string, error) {
	fmt.Println("\nSelect line filter operator (default: 1):")
	fmt.Println("1. |= (contains)")
	fmt.Println("2. != (does not contain)")
	fmt.Println("3. |~ (matches regex)")
	fmt.Println("4. !~ (does not match regex)")
	fmt.Print("Enter number (1-4) or press Enter for default: ")

	choice, err := inputText("")
	if err != nil {
		return "", err
	}

	if choice == "" {
		return "|=", nil
	}

	num, err := strconv.Atoi(choice)
	if err != nil || num < 1 || num > 4 {
		return "", fmt.Errorf("invalid choice: %s", choice)
	}

	operators := []string{"|=", "!=", "|~", "!~"}
	return operators[num-1], nil
}

func inputText(prompt string) (string, error) {
	if prompt != "" {
		fmt.Print(prompt)
	}
	reader := bufio.NewReader(os.Stdin)
	text, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(text), nil
}

func selectTimeRange() ([]string, error) {
	fmt.Println("Select time range type:")
	fmt.Println("1. Relative (e.g., 1h, 24h)")
	fmt.Println("2. Absolute (specific dates)")
	fmt.Print("Enter choice (1-2): ")

	choice, err := inputText("")
	if err != nil {
		return nil, err
	}

	switch choice {
	case "1":
		return selectRelativeTime()
	case "2":
		return selectAbsoluteTime()
	default:
		return nil, fmt.Errorf("invalid choice: %s", choice)
	}
}

func selectRelativeTime() ([]string, error) {
	fmt.Print("Enter relative time (e.g., 1h, 24h, 7d): ")
	duration, err := inputText("")
	if err != nil {
		return nil, err
	}
	return []string{"--since", duration}, nil
}

func selectAbsoluteTime() ([]string, error) {
	fmt.Print("Enter start time (YYYY-MM-DD HH:MM or YYYY-MM-DD): ")
	start, err := inputText("")
	if err != nil {
		return nil, err
	}

	fmt.Print("Enter end time (YYYY-MM-DD HH:MM or YYYY-MM-DD): ")
	end, err := inputText("")
	if err != nil {
		return nil, err
	}

	startRFC, err := convertToRFC3339(start, true)
	if err != nil {
		return nil, fmt.Errorf("invalid start time: %w", err)
	}

	endRFC, err := convertToRFC3339(end, false)
	if err != nil {
		return nil, fmt.Errorf("invalid end time: %w", err)
	}

	return []string{"--from", startRFC, "--to", endRFC}, nil
}

func selectOperator(label string) (string, error) {
	fmt.Printf("\nSelect operator for '%s' (default: 1):\n", label)
	fmt.Println("1. = (equals)")
	fmt.Println("2. != (not equals)")
	fmt.Println("3. =~ (regex match)")
	fmt.Println("4. !~ (regex not match)")
	fmt.Print("Enter number (1-4) or press Enter for default: ")

	choice, err := inputText("")
	if err != nil {
		return "", err
	}

	if choice == "" {
		return "=", nil
	}

	num, err := strconv.Atoi(choice)
	if err != nil || num < 1 || num > 4 {
		return "", fmt.Errorf("invalid choice: %s", choice)
	}

	operators := []string{"=", "!=", "=~", "!~"}
	return operators[num-1], nil
}

func selectWithFzf(items []string, prompt string) (string, error) {
	if len(items) == 0 {
		return "", fmt.Errorf("no items to select")
	}

	cmd := exec.Command("fzf", "--prompt", prompt)
	cmd.Stdin = strings.NewReader(strings.Join(items, "\n"))
	cmd.Stderr = os.Stderr

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("fzf failed: %w", err)
	}

	selected := strings.TrimSpace(string(output))
	if selected == "" {
		return "", fmt.Errorf("no selection made")
	}

	return selected, nil
}

func buildLogCLIArgs(logcliCmd string, selectors []LabelSelector, lineFilter *LineFilter, timeArgs []string) []string {
	// Build LogQL query
	query := "{"
	for i, s := range selectors {
		if i > 0 {
			query += ","
		}
		query += fmt.Sprintf("%s%s\"%s\"", s.Label, s.Operator, s.Value)
	}
	query += "}"

	if lineFilter != nil {
		query += fmt.Sprintf(" %s \"%s\"", lineFilter.Operator, lineFilter.Text)
	}

	// Build command arguments
	args := []string{logcliCmd, "query", query}
	args = append(args, timeArgs...)

	return args
}

func formatAsShellCommand(args []string) string {
	// Create a copy to avoid modifying the original
	quotedArgs := make([]string, len(args))
	copy(quotedArgs, args)

	// Add single quotes around the query (3rd argument)
	if len(quotedArgs) > 2 {
		quotedArgs[2] = "'" + quotedArgs[2] + "'"
	}

	return strings.Join(quotedArgs, " ")
}
