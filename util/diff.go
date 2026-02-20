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

func PrintPlan(p Plan) {
	b, _ := json.MarshalIndent(p, "", "  ")
	fmt.Println(string(b))
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

		if p.Stats.Teams.Current > 0 || p.Stats.Teams.Desired > 0 {
			fmt.Printf("  Teams:              %d → %d", p.Stats.Teams.Current, p.Stats.Teams.Desired)
			delta := p.Stats.Teams.Desired - p.Stats.Teams.Current
			if delta > 0 {
				fmt.Printf(" (+%d)\n", delta)
			} else if delta < 0 {
				fmt.Printf(" (%d)\n", delta)
			} else {
				fmt.Println(" (no change)")
			}
		}

		if p.Stats.TeamMembers.Current > 0 || p.Stats.TeamMembers.Desired > 0 {
			fmt.Printf("  Team Members:       %d → %d", p.Stats.TeamMembers.Current, p.Stats.TeamMembers.Desired)
			delta := p.Stats.TeamMembers.Desired - p.Stats.TeamMembers.Current
			if delta > 0 {
				fmt.Printf(" (+%d)\n", delta)
			} else if delta < 0 {
				fmt.Printf(" (%d)\n", delta)
			} else {
				fmt.Println(" (no change)")
			}
		}

		if p.Stats.Repositories.Current > 0 || p.Stats.Repositories.Desired > 0 {
			fmt.Printf("  Repositories:       %d → %d", p.Stats.Repositories.Current, p.Stats.Repositories.Desired)
			delta := p.Stats.Repositories.Desired - p.Stats.Repositories.Current
			if delta > 0 {
				fmt.Printf(" (+%d)\n", delta)
			} else if delta < 0 {
				fmt.Printf(" (%d)\n", delta)
			} else {
				fmt.Println(" (no change)")
			}
		}

		if p.Stats.RepoPermissions.Current > 0 || p.Stats.RepoPermissions.Desired > 0 {
			fmt.Printf("  Repo Permissions:   %d → %d", p.Stats.RepoPermissions.Current, p.Stats.RepoPermissions.Desired)
			delta := p.Stats.RepoPermissions.Desired - p.Stats.RepoPermissions.Current
			if delta > 0 {
				fmt.Printf(" (+%d)\n", delta)
			} else if delta < 0 {
				fmt.Printf(" (%d)\n", delta)
			} else {
				fmt.Println(" (no change)")
			}
		}
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
