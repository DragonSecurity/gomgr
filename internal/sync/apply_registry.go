package sync

import (
	"context"
	"fmt"
	"math"

	"github.com/DragonSecurity/gomgr/internal/gh"
	"github.com/DragonSecurity/gomgr/internal/util"
)

// Handler applies a single planned change against GitHub.
//
// Implementations should be idempotent where possible; the planner already
// filters out no-op changes, but partial failures mid-apply are possible.
type Handler interface {
	Apply(ctx context.Context, c *gh.Client, ch util.Change) error
}

// HandlerFunc adapts an ordinary function to the Handler interface.
type HandlerFunc func(ctx context.Context, c *gh.Client, ch util.Change) error

// Apply implements Handler.
func (f HandlerFunc) Apply(ctx context.Context, c *gh.Client, ch util.Change) error {
	return f(ctx, c, ch)
}

// registration bundles a handler with the ordering precedence used during apply.
type registration struct {
	handler    Handler
	precedence int
}

// HandlerRegistry resolves change kinds (scope:action) to handlers and orders them.
type HandlerRegistry struct {
	entries map[string]registration
}

// NewHandlerRegistry returns an empty registry.
func NewHandlerRegistry() *HandlerRegistry {
	return &HandlerRegistry{entries: map[string]registration{}}
}

// Register binds a Handler to scope:action with the given precedence.
// Lower precedence runs first. Duplicate registrations panic so misconfigured
// registries fail fast at startup rather than during an apply.
func (r *HandlerRegistry) Register(scope, action string, precedence int, h Handler) {
	key := changeKey(scope, action)
	if _, exists := r.entries[key]; exists {
		panic(fmt.Sprintf("handler already registered for %s", key))
	}
	r.entries[key] = registration{handler: h, precedence: precedence}
}

// Lookup returns the Handler registered for scope:action, if any.
func (r *HandlerRegistry) Lookup(scope, action string) (Handler, bool) {
	e, ok := r.entries[changeKey(scope, action)]
	if !ok {
		return nil, false
	}
	return e.handler, true
}

// Precedence returns the ordering weight for scope:action. Unregistered keys
// sort last so unknown change kinds don't silently jump the queue.
func (r *HandlerRegistry) Precedence(scope, action string) int {
	if e, ok := r.entries[changeKey(scope, action)]; ok {
		return e.precedence
	}
	return math.MaxInt
}

func changeKey(scope, action string) string { return scope + ":" + action }

// defaultRegistry is the set of built-in handlers applyChanges consults.
var defaultRegistry = buildDefaultRegistry()

func buildDefaultRegistry() *HandlerRegistry {
	r := NewHandlerRegistry()

	// Creation / mutation phase (low precedence = runs first).
	r.Register("custom-role", "create", precedenceCustomRoleCreate, HandlerFunc(applyCustomRoleNoop))
	r.Register("custom-role", "update", precedenceCustomRoleUpdate, HandlerFunc(applyCustomRoleNoop))
	r.Register("team", "create", precedenceTeamCreate, HandlerFunc(applyTeamCreate))
	r.Register("repo", "ensure", precedenceRepoEnsure, HandlerFunc(applyRepoEnsure))
	r.Register("team", "update", precedenceTeamUpdate, HandlerFunc(applyTeamUpdate))
	r.Register("team-repo", "grant", precedenceTeamRepoGrant, HandlerFunc(applyTeamRepoGrant))
	r.Register("team-member", "ensure", precedenceTeamMemberEnsure, HandlerFunc(applyTeamMemberEnsure))
	r.Register("repo-file", "ensure", precedenceRepoFileEnsure, HandlerFunc(applyRepoFileEnsure))
	r.Register("repo-topics", "ensure", precedenceRepoTopicsEnsure, HandlerFunc(applyRepoTopicsEnsure))
	r.Register("repo-template", "ensure", precedenceRepoTemplateEnsure, HandlerFunc(applyRepoTemplateEnsure))
	r.Register("repo-pin", "ensure", precedenceRepoPinEnsure, HandlerFunc(applyRepoPinEnsure))

	// Cleanup phase (high precedence = runs last).
	r.Register("org-member", "remove", precedenceOrgMemberRemove, HandlerFunc(applyOrgMemberRemove))
	r.Register("team", "delete", precedenceTeamDelete, HandlerFunc(applyTeamDelete))
	r.Register("repo", "delete", precedenceRepoDelete, HandlerFunc(applyRepoDelete))
	r.Register("custom-role", "delete", precedenceCustomRoleDelete, HandlerFunc(applyCustomRoleNoop))

	return r
}

// applyCustomRoleNoop is a placeholder for custom-role changes; they are
// dispatched via applyCustomRoleChanges before the main loop runs, so this
// handler is only consulted for precedence ordering and should never execute.
func applyCustomRoleNoop(_ context.Context, _ *gh.Client, ch util.Change) error {
	return fmt.Errorf("custom-role change reached generic apply loop: %s:%s %s", ch.Scope, ch.Action, ch.Target)
}
