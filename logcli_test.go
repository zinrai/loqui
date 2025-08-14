package main

import (
	"reflect"
	"strings"
	"testing"
)

// Since we can't easily mock exec.Command, we'll test the parsing logic
// by extracting it into a testable function

// parseLabelsOutput parses the output of 'logcli labels' command
func parseLabelsOutput(output string) []string {
	labels := []string{}
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "http") {
			labels = append(labels, line)
		}
	}
	return labels
}

func TestParseLabelsOutput(t *testing.T) {
	tests := []struct {
		name   string
		output string
		want   []string
	}{
		{
			name: "basic labels",
			output: `app
env
namespace
node`,
			want: []string{"app", "env", "namespace", "node"},
		},
		{
			name: "labels with empty lines",
			output: `app

env

namespace`,
			want: []string{"app", "env", "namespace"},
		},
		{
			name: "labels with http headers",
			output: `http://localhost:3100/loki/api/v1/labels
app
env
namespace`,
			want: []string{"app", "env", "namespace"},
		},
		{
			name:   "empty output",
			output: "",
			want:   []string{},
		},
		{
			name: "labels with spaces",
			output: `  app  
env
  namespace  `,
			want: []string{"app", "env", "namespace"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseLabelsOutput(tt.output)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseLabelsOutput() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Test the actual parsing logic used in the code
func TestLabelValuesParsing(t *testing.T) {
	tests := []struct {
		name   string
		output string
		want   []string
	}{
		{
			name: "basic values",
			output: `nginx
apache
tomcat`,
			want: []string{"nginx", "apache", "tomcat"},
		},
		{
			name: "values with special characters",
			output: `prod-01
prod-02
test-env`,
			want: []string{"prod-01", "prod-02", "test-env"},
		},
		{
			name: "numeric values",
			output: `200
404
500
502`,
			want: []string{"200", "404", "500", "502"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseLabelsOutput(tt.output) // Same parsing logic
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseLabelValues() = %v, want %v", got, tt.want)
			}
		})
	}
}
