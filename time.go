package main

import (
	"fmt"
	"strings"
	"time"
)

// convertToRFC3339 converts user-friendly time format to RFC3339
// isStart determines whether to use 00:00:00 or 23:59:59 for date-only input
func convertToRFC3339(input string, isStart bool) (string, error) {
	input = strings.TrimSpace(input)

	// Get local timezone
	loc := time.Local

	// Try parsing with time
	if t, err := time.ParseInLocation("2006-01-02 15:04", input, loc); err == nil {
		return t.Format(time.RFC3339), nil
	}

	// Try parsing date only
	if t, err := time.ParseInLocation("2006-01-02", input, loc); err == nil {
		if isStart {
			// For start time, use 00:00:00
			return t.Format(time.RFC3339), nil
		} else {
			// For end time, use 23:59:59
			// Create end of day properly
			endOfDay := time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, loc)
			return endOfDay.Format(time.RFC3339), nil
		}
	}

	return "", fmt.Errorf("invalid time format: %s (expected YYYY-MM-DD HH:MM or YYYY-MM-DD)", input)
}
