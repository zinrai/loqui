package main

import (
	"strings"
	"testing"
	"time"
)

func TestConvertToRFC3339(t *testing.T) {
	// Set a fixed timezone for consistent testing
	loc, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		t.Fatal(err)
	}
	time.Local = loc

	tests := []struct {
		name     string
		input    string
		isStart  bool
		wantErr  bool
		validate func(t *testing.T, got string)
	}{
		{
			name:    "date only - start time",
			input:   "2025-08-14",
			isStart: true,
			wantErr: false,
			validate: func(t *testing.T, got string) {
				// Should be 00:00:00
				if !strings.Contains(got, "T00:00:00") {
					t.Errorf("expected 00:00:00 for start time, got %s", got)
				}
				if !strings.Contains(got, "+09:00") {
					t.Errorf("expected JST timezone, got %s", got)
				}
			},
		},
		{
			name:    "date only - end time",
			input:   "2025-08-14",
			isStart: false,
			wantErr: false,
			validate: func(t *testing.T, got string) {
				// Should be 23:59:59
				if !strings.Contains(got, "T23:59:59") {
					t.Errorf("expected 23:59:59 for end time, got %s", got)
				}
			},
		},
		{
			name:    "date and time",
			input:   "2025-08-14 15:30",
			isStart: true,
			wantErr: false,
			validate: func(t *testing.T, got string) {
				if !strings.Contains(got, "T15:30:00") {
					t.Errorf("expected 15:30:00, got %s", got)
				}
			},
		},
		{
			name:    "invalid format",
			input:   "invalid-date",
			isStart: true,
			wantErr: true,
		},
		{
			name:    "empty input",
			input:   "",
			isStart: true,
			wantErr: true,
		},
		{
			name:    "date with extra spaces",
			input:   "  2025-08-14  ",
			isStart: true,
			wantErr: false,
			validate: func(t *testing.T, got string) {
				if !strings.HasPrefix(got, "2025-08-14T") {
					t.Errorf("expected date to be parsed correctly, got %s", got)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := convertToRFC3339(tt.input, tt.isStart)
			if (err != nil) != tt.wantErr {
				t.Errorf("convertToRFC3339() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.validate != nil {
				// Validate it's proper RFC3339 format
				if _, err := time.Parse(time.RFC3339, got); err != nil {
					t.Errorf("output is not valid RFC3339: %s, error: %v", got, err)
				}
				tt.validate(t, got)
			}
		})
	}
}
