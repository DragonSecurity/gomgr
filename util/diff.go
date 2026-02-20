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

type Plan struct {
	Changes  []Change `json:"changes"`
	Warnings []string `json:"warnings"`
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

	if len(p.Changes) == 0 {
		fmt.Println("No changes required - configuration is in sync")
		return
	}

	// Count changes by scope
	scopeCounts := make(map[string]int)
	actionCounts := make(map[string]int)

	for _, ch := range p.Changes {
		scopeCounts[ch.Scope]++
		actionCounts[ch.Action]++
	}

	fmt.Printf("\nTotal changes: %d\n\n", len(p.Changes))

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
