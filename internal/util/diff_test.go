package util

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"
)

func TestPrintPlan(t *testing.T) {
	plan := Plan{
		Changes: []Change{
			{Scope: "team", Target: "backend", Action: "create", Details: map[string]any{"org": "myorg"}},
		},
		Warnings: []string{"test warning"},
	}

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := PrintPlan(plan)

	w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	// Verify it's valid JSON
	var parsed Plan
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v\nOutput: %s", err, output)
	}
	if len(parsed.Changes) != 1 {
		t.Errorf("expected 1 change, got %d", len(parsed.Changes))
	}
}

func TestPrintStatePair(t *testing.T) {
	tests := []struct {
		name     string
		label    string
		pair     StatePair
		contains string
		empty    bool
	}{
		{name: "increase", label: "Teams:", pair: StatePair{Current: 2, Desired: 5}, contains: "(+3)"},
		{name: "decrease", label: "Teams:", pair: StatePair{Current: 5, Desired: 2}, contains: "(-3)"},
		{name: "no change", label: "Teams:", pair: StatePair{Current: 3, Desired: 3}, contains: "(no change)"},
		{name: "both zero", label: "Teams:", pair: StatePair{Current: 0, Desired: 0}, empty: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			printStatePair(tt.label, tt.pair)

			w.Close()
			os.Stdout = old

			var buf bytes.Buffer
			_, _ = io.Copy(&buf, r)
			output := buf.String()

			if tt.empty {
				if output != "" {
					t.Errorf("expected no output for zero pair, got %q", output)
				}
				return
			}
			if !strings.Contains(output, tt.contains) {
				t.Errorf("expected output to contain %q, got %q", tt.contains, output)
			}
		})
	}
}

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
		{
			name: "with state statistics",
			plan: Plan{
				Changes: []Change{
					{Scope: "team", Target: "team1", Action: "create", Details: nil},
					{Scope: "team-member", Target: "user1", Action: "ensure", Details: nil},
				},
				Warnings: nil,
				Stats: &StateStats{
					Teams: StatePair{
						Current: 2,
						Desired: 3,
					},
					TeamMembers: StatePair{
						Current: 5,
						Desired: 7,
					},
					Repositories: StatePair{
						Current: 10,
						Desired: 12,
					},
					RepoPermissions: StatePair{
						Current: 15,
						Desired: 18,
					},
				},
			},
			expectedContains: []string{
				"Summary of Proposed Changes",
				"Current State vs Desired State:",
				"Teams:",
				"2 → 3",
				"Team Members:",
				"5 → 7",
				"Repositories:",
				"10 → 12",
				"Repo Permissions:",
				"15 → 18",
				"Total changes: 2",
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
			_, _ = io.Copy(&buf, r) // explicitly discard error
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
