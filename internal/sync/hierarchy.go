package sync

import (
	"sort"

	"github.com/DragonSecurity/gomgr/internal/config"
)

// hierarchyOrder returns every team slug in an order where a parent always
// precedes its children, and siblings are alphabetical.
//
// Two things depend on it. Creating a child before its parent exists fails, so
// team:create changes have to come out in this order — apply sorts by
// precedence with a stable sort, which preserves the order the planner emitted
// within one precedence class. And a plan that lists the same teams in a
// different order on every run is a plan nobody can diff.
//
// Cycles are rejected by config validation before this runs. If one reaches
// here anyway, the teams in it are appended in slug order rather than dropped:
// an ordering function is the wrong place to lose a team.
func hierarchyOrder(teams map[string]config.TeamConfig) []string {
	slugs := make([]string, 0, len(teams))
	for slug := range teams {
		slugs = append(slugs, slug)
	}
	sort.Strings(slugs)

	children := map[string][]string{}
	var roots []string
	for _, slug := range slugs {
		parent := teams[slug].ParentSlug()
		// A parent outside the config is a root here. Validation already
		// refused that case, so this only matters to callers that build a
		// team map directly.
		if _, ok := teams[parent]; parent == "" || !ok {
			roots = append(roots, slug)
			continue
		}
		children[parent] = append(children[parent], slug)
	}

	out := make([]string, 0, len(slugs))
	placed := make(map[string]bool, len(slugs))
	var walk func(slug string)
	walk = func(slug string) {
		if placed[slug] {
			return
		}
		placed[slug] = true
		out = append(out, slug)
		for _, child := range children[slug] {
			walk(child)
		}
	}
	for _, slug := range roots {
		walk(slug)
	}
	for _, slug := range slugs {
		if !placed[slug] {
			out = append(out, slug)
			placed[slug] = true
		}
	}
	return out
}

// ancestors returns the slugs above a team, nearest first. It stops at a team
// the config does not define and at a repeat, so a malformed map cannot spin.
func ancestors(slug string, teams map[string]config.TeamConfig) []string {
	var out []string
	seen := map[string]bool{slug: true}
	for cur := teams[slug].ParentSlug(); cur != ""; cur = teams[cur].ParentSlug() {
		if seen[cur] {
			return out
		}
		if _, ok := teams[cur]; !ok {
			return out
		}
		seen[cur] = true
		out = append(out, cur)
	}
	return out
}
