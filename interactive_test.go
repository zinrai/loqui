package main

import (
	"reflect"
	"testing"
)

func TestBuildLogCLIArgs(t *testing.T) {
	tests := []struct {
		name       string
		logcliCmd  string
		selectors  []LabelSelector
		lineFilter *LineFilter
		timeArgs   []string
		want       []string
	}{
		{
			name:      "single label with equals",
			logcliCmd: "logcli",
			selectors: []LabelSelector{
				{Label: "app", Operator: "=", Value: "nginx"},
			},
			lineFilter: nil,
			timeArgs:   []string{"--since", "1h"},
			want:       []string{"logcli", "query", `{app="nginx"}`, "--since", "1h"},
		},
		{
			name:      "multiple labels",
			logcliCmd: "logcli",
			selectors: []LabelSelector{
				{Label: "app", Operator: "=", Value: "nginx"},
				{Label: "env", Operator: "!=", Value: "test"},
			},
			lineFilter: nil,
			timeArgs:   []string{"--since", "2h"},
			want:       []string{"logcli", "query", `{app="nginx",env!="test"}`, "--since", "2h"},
		},
		{
			name:      "with line filter contains",
			logcliCmd: "logcli",
			selectors: []LabelSelector{
				{Label: "app", Operator: "=", Value: "nginx"},
			},
			lineFilter: &LineFilter{Operator: "|=", Text: "error"},
			timeArgs:   []string{"--since", "1h"},
			want:       []string{"logcli", "query", `{app="nginx"} |= "error"`, "--since", "1h"},
		},
		{
			name:      "with line filter not contains",
			logcliCmd: "logcli",
			selectors: []LabelSelector{
				{Label: "app", Operator: "=", Value: "nginx"},
			},
			lineFilter: &LineFilter{Operator: "!=", Text: "debug"},
			timeArgs:   []string{"--since", "1h"},
			want:       []string{"logcli", "query", `{app="nginx"} != "debug"`, "--since", "1h"},
		},
		{
			name:      "with line filter regex match",
			logcliCmd: "logcli",
			selectors: []LabelSelector{
				{Label: "app", Operator: "=", Value: "nginx"},
			},
			lineFilter: &LineFilter{Operator: "|~", Text: `error|warn`},
			timeArgs:   []string{"--since", "1h"},
			want:       []string{"logcli", "query", `{app="nginx"} |~ "error|warn"`, "--since", "1h"},
		},
		{
			name:      "with line filter regex not match",
			logcliCmd: "logcli",
			selectors: []LabelSelector{
				{Label: "app", Operator: "=", Value: "nginx"},
			},
			lineFilter: &LineFilter{Operator: "!~", Text: `\.(jpg|png|gif)$`},
			timeArgs:   []string{"--since", "1h"},
			want:       []string{"logcli", "query", `{app="nginx"} !~ "\.(jpg|png|gif)$"`, "--since", "1h"},
		},
		{
			name:      "regex operator",
			logcliCmd: "logcli",
			selectors: []LabelSelector{
				{Label: "status", Operator: "=~", Value: `5\d{2}`},
			},
			lineFilter: nil,
			timeArgs:   []string{"--since", "1h"},
			want:       []string{"logcli", "query", `{status=~"5\d{2}"}`, "--since", "1h"},
		},
		{
			name:      "regex not match operator",
			logcliCmd: "logcli",
			selectors: []LabelSelector{
				{Label: "path", Operator: "!~", Value: `\.(jpg|png|gif)$`},
			},
			lineFilter: nil,
			timeArgs:   []string{"--since", "1h"},
			want:       []string{"logcli", "query", `{path!~"\.(jpg|png|gif)$"}`, "--since", "1h"},
		},
		{
			name:      "absolute time range",
			logcliCmd: "logcli",
			selectors: []LabelSelector{
				{Label: "app", Operator: "=", Value: "nginx"},
			},
			lineFilter: nil,
			timeArgs:   []string{"--from", "2025-08-14T00:00:00+09:00", "--to", "2025-08-14T23:59:59+09:00"},
			want:       []string{"logcli", "query", `{app="nginx"}`, "--from", "2025-08-14T00:00:00+09:00", "--to", "2025-08-14T23:59:59+09:00"},
		},
		{
			name:      "custom logcli command",
			logcliCmd: "/usr/local/bin/logcli",
			selectors: []LabelSelector{
				{Label: "app", Operator: "=", Value: "nginx"},
			},
			lineFilter: nil,
			timeArgs:   []string{"--since", "1h"},
			want:       []string{"/usr/local/bin/logcli", "query", `{app="nginx"}`, "--since", "1h"},
		},
		{
			name:      "no line filter when nil",
			logcliCmd: "logcli",
			selectors: []LabelSelector{
				{Label: "app", Operator: "=", Value: "test"},
			},
			lineFilter: nil,
			timeArgs:   []string{"--since", "30m"},
			want:       []string{"logcli", "query", `{app="test"}`, "--since", "30m"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildLogCLIArgs(tt.logcliCmd, tt.selectors, tt.lineFilter, tt.timeArgs)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("buildLogCLIArgs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatAsShellCommand(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "basic query",
			args: []string{"logcli", "query", `{app="nginx"}`, "--since", "1h"},
			want: `logcli query '{app="nginx"}' --since 1h`,
		},
		{
			name: "query with line filter",
			args: []string{"logcli", "query", `{app="nginx"} |= "error"`, "--since", "1h"},
			want: `logcli query '{app="nginx"} |= "error"' --since 1h`,
		},
		{
			name: "complex query",
			args: []string{"logcli", "query", `{app="nginx",env!="test"} |~ "error|warn"`, "--from", "2025-08-14T00:00:00+09:00", "--to", "2025-08-14T23:59:59+09:00"},
			want: `logcli query '{app="nginx",env!="test"} |~ "error|warn"' --from 2025-08-14T00:00:00+09:00 --to 2025-08-14T23:59:59+09:00`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatAsShellCommand(tt.args)
			if got != tt.want {
				t.Errorf("formatAsShellCommand() = %v, want %v", got, tt.want)
			}
		})
	}
}
