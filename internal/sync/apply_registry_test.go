package sync

import (
	"context"
	"errors"
	"math"
	"testing"

	"github.com/DragonSecurity/gomgr/internal/gh"
	"github.com/DragonSecurity/gomgr/internal/util"
)

func TestHandlerRegistry_RegisterAndLookup(t *testing.T) {
	r := NewHandlerRegistry()
	want := errors.New("boom")
	r.Register("team", "create", 10, HandlerFunc(func(context.Context, *gh.Client, util.Change) error {
		return want
	}))

	h, ok := r.Lookup("team", "create")
	if !ok {
		t.Fatal("expected Lookup to find team:create")
	}
	if err := h.Apply(context.Background(), nil, util.Change{}); !errors.Is(err, want) {
		t.Errorf("expected registered handler to run, got %v", err)
	}

	if _, ok := r.Lookup("repo", "delete"); ok {
		t.Error("expected Lookup to miss unregistered keys")
	}
}

func TestHandlerRegistry_Precedence(t *testing.T) {
	r := NewHandlerRegistry()
	r.Register("team", "create", 10, HandlerFunc(noopHandler))
	r.Register("repo", "delete", 90, HandlerFunc(noopHandler))

	if got := r.Precedence("team", "create"); got != 10 {
		t.Errorf("team:create precedence = %d, want 10", got)
	}
	if got := r.Precedence("repo", "delete"); got != 90 {
		t.Errorf("repo:delete precedence = %d, want 90", got)
	}
	if got := r.Precedence("unknown", "action"); got != math.MaxInt {
		t.Errorf("unknown precedence = %d, want MaxInt (sort last)", got)
	}
}

func TestHandlerRegistry_DuplicatePanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on duplicate registration")
		}
	}()
	r := NewHandlerRegistry()
	r.Register("team", "create", 10, HandlerFunc(noopHandler))
	r.Register("team", "create", 20, HandlerFunc(noopHandler))
}

func TestDefaultRegistry_HasKnownKinds(t *testing.T) {
	kinds := []struct{ scope, action string }{
		{"team", "create"},
		{"team", "update"},
		{"team", "delete"},
		{"team-member", "ensure"},
		{"repo", "ensure"},
		{"repo", "delete"},
		{"team-repo", "grant"},
		{"repo-file", "ensure"},
		{"repo-topics", "ensure"},
		{"repo-template", "ensure"},
		{"repo-pin", "ensure"},
		{"org-member", "remove"},
		{"custom-role", "create"},
		{"custom-role", "update"},
		{"custom-role", "delete"},
	}
	for _, k := range kinds {
		if _, ok := defaultRegistry.Lookup(k.scope, k.action); !ok {
			t.Errorf("defaultRegistry missing handler for %s:%s", k.scope, k.action)
		}
	}
}

func TestDefaultRegistry_PrecedenceOrdering(t *testing.T) {
	// Create must run before member additions, which must run before deletions.
	if defaultRegistry.Precedence("team", "create") >= defaultRegistry.Precedence("team-member", "ensure") {
		t.Error("team:create should precede team-member:ensure")
	}
	if defaultRegistry.Precedence("team-member", "ensure") >= defaultRegistry.Precedence("team", "delete") {
		t.Error("team-member:ensure should precede team:delete")
	}
	if defaultRegistry.Precedence("repo", "ensure") >= defaultRegistry.Precedence("repo", "delete") {
		t.Error("repo:ensure should precede repo:delete")
	}
}

func noopHandler(context.Context, *gh.Client, util.Change) error { return nil }
