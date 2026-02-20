package util

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestPrintSummary(t *testing.T) {
	tests := []struct {
		name             string
		plan             Plan
		expectedContains []string
	}{
		{
			name: "with changes and warnings",
			plan: Plan{
				Changes: []Change{
					{Scope: "team", Target: "team1", Action: "create", Details: nil},
					{Scope: "team", Target: "team2", Action: "create", Details: nil},
					{Scope: "team-member", Target: "user1", Action: "ensure", Details: nil},
					{Scope: "team-repo", Target: "repo1", Action: "grant", Details: nil},
					{Scope: "repo-pin", Target: "repo1", Action: "ensure", Details: nil},
				},
				Warnings: []string{"Test warning 1", "Test warning 2"},
			},
			expectedContains: []string{
				"Summary of Proposed Changes",
				"Total changes: 5",
				"Changes by scope:",
				"team:",
				"team-member:",
				"team-repo:",
				"repo-pin:",
				"Changes by action:",
				"create:",
				"ensure:",
				"grant:",
				"Warnings: 2",
				"Test warning 1",
				"Test warning 2",
			},
		},
		{
			name: "no changes",
			plan: Plan{
				Changes:  []Change{},
				Warnings: nil,
			},
			expectedContains: []string{
				"Summary of Proposed Changes",
				"No changes required - configuration is in sync",
			},
		},
		{
			name: "changes without warnings",
			plan: Plan{
				Changes: []Change{
					{Scope: "team", Target: "team1", Action: "create", Details: nil},
				},
				Warnings: nil,
			},
			expectedContains: []string{
				"Summary of Proposed Changes",
				"Total changes: 1",
				"Changes by scope:",
				"team:",
				"Changes by action:",
				"create:",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			PrintSummary(tt.plan)

			w.Close()
			os.Stdout = old

			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := buf.String()

			// Verify expected strings are in output
			for _, expected := range tt.expectedContains {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected output to contain %q, but it didn't.\nFull output:\n%s", expected, output)
				}
			}
		})
	}
}
