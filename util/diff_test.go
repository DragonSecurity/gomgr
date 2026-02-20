package util

import (
	"testing"
)

func TestPrintSummary(t *testing.T) {
	// Test with changes
	plan := Plan{
		Changes: []Change{
			{Scope: "team", Target: "team1", Action: "create", Details: nil},
			{Scope: "team", Target: "team2", Action: "create", Details: nil},
			{Scope: "team-member", Target: "user1", Action: "ensure", Details: nil},
			{Scope: "team-repo", Target: "repo1", Action: "grant", Details: nil},
			{Scope: "repo-pin", Target: "repo1", Action: "ensure", Details: nil},
		},
		Warnings: []string{"Test warning 1", "Test warning 2"},
	}

	// Just make sure it doesn't panic
	PrintSummary(plan)

	// Test with no changes
	emptyPlan := Plan{
		Changes:  []Change{},
		Warnings: nil,
	}
	PrintSummary(emptyPlan)
}
