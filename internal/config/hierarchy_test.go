package config

import "testing"

func teamsRoot(teams ...TeamConfig) *Root {
	return &Root{App: AppConfig{Org: "myorg"}, Team: teams}
}

func TestValidateAcceptsAHierarchy(t *testing.T) {
	r := teamsRoot(
		TeamConfig{Name: "Platform", Privacy: "closed"},
		TeamConfig{Name: "Platform Oncall", Privacy: "closed", Parents: []string{"Platform"}},
	)
	if err := r.Validate(); err != nil {
		t.Fatalf("expected a valid hierarchy, got %v", err)
	}
}

// The parent may be written as either the team's name or its slug; both are
// slugified before the lookup.
func TestValidateAcceptsAParentNamedBySlug(t *testing.T) {
	r := teamsRoot(
		TeamConfig{Name: "Platform Team", Privacy: "closed"},
		TeamConfig{Name: "Oncall", Privacy: "closed", Parents: []string{"platform-team"}},
	)
	if err := r.Validate(); err != nil {
		t.Fatalf("expected slug form to resolve, got %v", err)
	}
}

func TestValidateRejectsMoreThanOneParent(t *testing.T) {
	r := teamsRoot(
		TeamConfig{Name: "A", Privacy: "closed"},
		TeamConfig{Name: "B", Privacy: "closed"},
		TeamConfig{Name: "C", Privacy: "closed", Parents: []string{"A", "B"}},
	)
	if err := r.Validate(); err == nil {
		t.Fatal("GitHub allows one parent; two should be refused")
	}
}

func TestValidateRejectsAnUnknownParent(t *testing.T) {
	r := teamsRoot(TeamConfig{Name: "Orphan", Privacy: "closed", Parents: []string{"Nobody"}})
	if err := r.Validate(); err == nil {
		t.Fatal("a parent no team file defines should be refused")
	}
}

func TestValidateRejectsSelfParenting(t *testing.T) {
	r := teamsRoot(TeamConfig{Name: "Loop", Privacy: "closed", Parents: []string{"Loop"}})
	if err := r.Validate(); err == nil {
		t.Fatal("a team that is its own parent should be refused")
	}
}

func TestValidateRejectsACycle(t *testing.T) {
	r := teamsRoot(
		TeamConfig{Name: "A", Privacy: "closed", Parents: []string{"C"}},
		TeamConfig{Name: "B", Privacy: "closed", Parents: []string{"A"}},
		TeamConfig{Name: "C", Privacy: "closed", Parents: []string{"B"}},
	)
	err := r.Validate()
	if err == nil {
		t.Fatal("a cycle should be refused rather than deadlocking the apply order")
	}
}

// GitHub refuses a secret team at either end of a nesting, so both ends are
// checked here rather than discovered halfway through an apply.
func TestValidateRejectsSecretInAHierarchy(t *testing.T) {
	t.Run("secret child", func(t *testing.T) {
		r := teamsRoot(
			TeamConfig{Name: "Parent", Privacy: "closed"},
			TeamConfig{Name: "Child", Privacy: "secret", Parents: []string{"Parent"}},
		)
		if err := r.Validate(); err == nil {
			t.Fatal("a secret child should be refused")
		}
	})
	t.Run("secret parent", func(t *testing.T) {
		r := teamsRoot(
			TeamConfig{Name: "Parent", Privacy: "secret"},
			TeamConfig{Name: "Child", Privacy: "closed", Parents: []string{"Parent"}},
		)
		if err := r.Validate(); err == nil {
			t.Fatal("a secret parent should be refused")
		}
	})
	t.Run("secret and unnested is fine", func(t *testing.T) {
		r := teamsRoot(TeamConfig{Name: "Security", Privacy: "secret"})
		if err := r.Validate(); err != nil {
			t.Fatalf("a secret team with no hierarchy is legal, got %v", err)
		}
	})
}

func TestParentSlug(t *testing.T) {
	cases := []struct {
		name string
		in   []string
		want string
	}{
		{"none", nil, ""},
		{"empty list", []string{}, ""},
		{"blank entry", []string{"   "}, ""},
		{"name is slugified", []string{"Platform Team"}, "platform-team"},
		{"slug passes through", []string{"platform-team"}, "platform-team"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := (TeamConfig{Parents: tc.in}).ParentSlug(); got != tc.want {
				t.Errorf("want %q, got %q", tc.want, got)
			}
		})
	}
}
