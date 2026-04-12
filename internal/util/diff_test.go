package util

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

// capturePrint runs fn with os.Stdout redirected to a pipe and returns the
// captured output.
func capturePrint(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	fn()
	_ = w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	return buf.String()
}

func TestPrintPlan_HumanReadable(t *testing.T) {
	plan := Plan{
		Changes: []Change{
			{Scope: "repo-file", Target: "infra:README.md", Action: "ensure", Details: map[string]any{"org": "KaMuses"}},
			{Scope: "repo", Target: "api", Action: "delete", Details: map[string]any{"org": "KaMuses"}},
			{Scope: "team", Target: "backend", Action: "update", Details: map[string]any{"org": "KaMuses"}},
		},
	}

	out := capturePrint(t, func() {
		if err := PrintPlan(plan); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	want := []string{
		"Plan (3 changes):",
		"+ repo-file:ensure",
		"KaMuses/infra:README.md",
		"- repo:delete",
		"KaMuses/api",
		"~ team:update",
		"KaMuses/backend",
	}
	for _, s := range want {
		if !strings.Contains(out, s) {
			t.Errorf("expected output to contain %q, got:\n%s", s, out)
		}
	}
	// Ensure the noisy JSON details are gone.
	if strings.Contains(out, `"details"`) {
		t.Errorf("expected no JSON details in output, got:\n%s", out)
	}
}

func TestPrintPlan_Empty(t *testing.T) {
	out := capturePrint(t, func() {
		_ = PrintPlan(Plan{})
	})
	if !strings.Contains(out, "no changes") {
		t.Errorf("expected 'no changes' message, got %q", out)
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
