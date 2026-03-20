package util

import (
	"encoding/json"
	"fmt"
	"sort"
)

type Change struct {
	Scope   string      `json:"scope"`
	Target  string      `json:"target"`
	Action  string      `json:"action"`
	Details interface{} `json:"details"`
}

type StateStats struct {
	Teams           StatePair `json:"teams"`
	TeamMembers     StatePair `json:"team_members"`
	Repositories    StatePair `json:"repositories"`
	RepoPermissions StatePair `json:"repo_permissions"`
	CustomRoles     StatePair `json:"custom_roles"`
}

type StatePair struct {
	Current int `json:"current"`
	Desired int `json:"desired"`
}

type Plan struct {
	Changes  []Change    `json:"changes"`
	Warnings []string    `json:"warnings"`
	Stats    *StateStats `json:"stats,omitempty"`
}

func PrintPlan(p Plan) error {
	b, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal plan: %w", err)
	}
	fmt.Println(string(b))
	return nil
}

// PrintSummary prints a human-readable summary of the plan
func PrintSummary(p Plan) {
	separator := "\n" + "================================================================"
	fmt.Println(separator)
	fmt.Println("Summary of Proposed Changes")
	fmt.Println("================================================================")

	// Show current vs desired state if available
	if p.Stats != nil {
		fmt.Println("\nCurrent State vs Desired State:")
		fmt.Println("--------------------------------")

		printStatePair("Teams:", p.Stats.Teams)
		printStatePair("Team Members:", p.Stats.TeamMembers)
		printStatePair("Repositories:", p.Stats.Repositories)
		printStatePair("Repo Permissions:", p.Stats.RepoPermissions)
		printStatePair("Custom Roles:", p.Stats.CustomRoles)
		fmt.Println()
	}

	if len(p.Changes) == 0 {
		fmt.Println("No changes required - configuration is in sync")
		fmt.Println("\n" + "================================================================")
		return
	}

	// Count changes by scope
	scopeCounts := make(map[string]int)
	actionCounts := make(map[string]int)

	for _, ch := range p.Changes {
		scopeCounts[ch.Scope]++
		actionCounts[ch.Action]++
	}

	fmt.Printf("Total changes: %d\n\n", len(p.Changes))

	// Print by scope
	fmt.Println("Changes by scope:")
	scopes := make([]string, 0, len(scopeCounts))
	for scope := range scopeCounts {
		scopes = append(scopes, scope)
	}
	sort.Strings(scopes)

	for _, scope := range scopes {
		count := scopeCounts[scope]
		fmt.Printf("  %-20s %d\n", scope+":", count)
	}

	fmt.Println("\nChanges by action:")
	actions := make([]string, 0, len(actionCounts))
	for action := range actionCounts {
		actions = append(actions, action)
	}
	sort.Strings(actions)

	for _, action := range actions {
		count := actionCounts[action]
		fmt.Printf("  %-20s %d\n", action+":", count)
	}

	if len(p.Warnings) > 0 {
		fmt.Printf("\nWarnings: %d\n", len(p.Warnings))
		for _, w := range p.Warnings {
			fmt.Printf("  - %s\n", w)
		}
	}

	fmt.Println("\n" + "================================================================")
}

// printStatePair prints a single state comparison line with delta
func printStatePair(label string, pair StatePair) {
	if pair.Current == 0 && pair.Desired == 0 {
		return
	}

	fmt.Printf("  %-20s %d → %d", label, pair.Current, pair.Desired)
	delta := pair.Desired - pair.Current
	if delta > 0 {
		fmt.Printf(" (+%d)\n", delta)
	} else if delta < 0 {
		fmt.Printf(" (%d)\n", delta)
	} else {
		fmt.Println(" (no change)")
	}
}
