package util

import (
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

// PrintPlan writes the plan's changes as a short, one-line-per-change list.
// File contents and other verbose detail fields are omitted; the summary
// block printed by PrintSummary carries the aggregate stats.
func PrintPlan(p Plan) error {
	if len(p.Changes) == 0 {
		fmt.Println("Plan: no changes")
		return nil
	}

	kinds := make([]string, len(p.Changes))
	longest := 0
	for i, ch := range p.Changes {
		kinds[i] = ch.Scope + ":" + ch.Action
		if len(kinds[i]) > longest {
			longest = len(kinds[i])
		}
	}

	fmt.Printf("Plan (%d changes):\n", len(p.Changes))
	for i, ch := range p.Changes {
		fmt.Printf("  %s %-*s  %s\n", changeSymbol(ch.Action), longest, kinds[i], formatTarget(ch))
	}
	return nil
}

// changeSymbol returns a one-character marker for the change's direction:
// `+` for creation, `-` for removal, `~` for in-place updates.
func changeSymbol(action string) string {
	switch action {
	case "create", "ensure", "grant":
		return "+"
	case "delete", "remove":
		return "-"
	case "update":
		return "~"
	}
	return "·"
}

// formatTarget prefixes ch.Target with the org when one is present in the
// details map. Typed detail structs (team-member, custom-role) fall through
// to the bare target.
func formatTarget(ch Change) string {
	if d, ok := ch.Details.(map[string]any); ok {
		if org, ok := d["org"].(string); ok && org != "" {
			return org + "/" + ch.Target
		}
	}
	return ch.Target
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
