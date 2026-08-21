package sync

import (
	"reflect"
	"testing"

	"github.com/DragonSecurity/gomgr/internal/config"
)

func teamSet(pairs map[string]string) map[string]config.TeamConfig {
	out := map[string]config.TeamConfig{}
	for slug, parent := range pairs {
		tc := config.TeamConfig{Name: slug, Slug: slug}
		if parent != "" {
			tc.Parents = []string{parent}
		}
		out[slug] = tc
	}
	return out
}

func TestHierarchyOrderPutsParentsFirst(t *testing.T) {
	teams := teamSet(map[string]string{
		"grandchild": "child",
		"child":      "root",
		"root":       "",
		"loner":      "",
	})

	got := hierarchyOrder(teams)

	pos := map[string]int{}
	for i, slug := range got {
		pos[slug] = i
	}
	if len(got) != 4 {
		t.Fatalf("expected every team in the order, got %v", got)
	}
	if pos["root"] > pos["child"] || pos["child"] > pos["grandchild"] {
		t.Errorf("parents must precede children, got %v", got)
	}
}

func TestHierarchyOrderIsStable(t *testing.T) {
	teams := teamSet(map[string]string{
		"zebra": "", "alpha": "", "middle": "", "child-of-alpha": "alpha",
	})

	first := hierarchyOrder(teams)
	for i := 0; i < 20; i++ {
		if got := hierarchyOrder(teams); !reflect.DeepEqual(got, first) {
			t.Fatalf("order is not deterministic:\nfirst %v\nthen  %v", first, got)
		}
	}
	// Siblings alphabetical, and a child immediately follows its parent.
	want := []string{"alpha", "child-of-alpha", "middle", "zebra"}
	if !reflect.DeepEqual(first, want) {
		t.Errorf("want %v, got %v", want, first)
	}
}

// A cycle is refused by config validation, but an ordering function is the
// wrong place to lose a team, so it must still return all of them.
func TestHierarchyOrderKeepsEveryTeamInACycle(t *testing.T) {
	teams := teamSet(map[string]string{"a": "b", "b": "a", "free": ""})

	got := hierarchyOrder(teams)

	if len(got) != 3 {
		t.Fatalf("expected 3 teams, got %v", got)
	}
	seen := map[string]bool{}
	for _, slug := range got {
		if seen[slug] {
			t.Fatalf("team %q emitted twice: %v", slug, got)
		}
		seen[slug] = true
	}
}

func TestAncestors(t *testing.T) {
	teams := teamSet(map[string]string{
		"grandchild": "child", "child": "root", "root": "",
	})

	if got := ancestors("grandchild", teams); !reflect.DeepEqual(got, []string{"child", "root"}) {
		t.Errorf("want [child root], got %v", got)
	}
	if got := ancestors("root", teams); len(got) != 0 {
		t.Errorf("a root has no ancestors, got %v", got)
	}
}

func TestAncestorsStopsOnACycle(t *testing.T) {
	teams := teamSet(map[string]string{"a": "b", "b": "a"})

	// The assertion that matters is that this returns at all.
	if got := ancestors("a", teams); len(got) > 2 {
		t.Errorf("cycle was walked more than once: %v", got)
	}
}

func TestAncestorsStopsAtAnUndefinedParent(t *testing.T) {
	teams := teamSet(map[string]string{"child": "not-in-this-config"})

	if got := ancestors("child", teams); len(got) != 0 {
		t.Errorf("a parent outside the config is not an ancestor, got %v", got)
	}
}
