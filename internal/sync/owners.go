package sync

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/google/go-github/v90/github"

	"github.com/DragonSecurity/gomgr/internal/config"
	"github.com/DragonSecurity/gomgr/internal/gh"
	"github.com/DragonSecurity/gomgr/internal/util"
)

// The two roles an organization membership can carry.
const (
	orgRoleAdmin  = "admin"
	orgRoleMember = "member"
)

// planOrgOwners reconciles org.yaml's `owners:` against the organization's
// actual admins.
//
// An empty owner list means gomgr does not manage owners for this org. It never
// means the org should have none: a configuration that omits the key, or that
// loses it to a bad edit, must not be able to strip every administrator from an
// organization — including the account that would be needed to put one back.
//
// Promotion is unconditional, because adding an owner cannot lock anyone out.
// Demotion is not: it needs DemoteUnconfiguredOwners, and it never touches the
// authenticated user, for the same reason.
func planOrgOwners(ctx context.Context, c *gh.Client, cfg *config.Root, st *State) ([]util.Change, []string, error) {
	if len(cfg.Org.Owners) == 0 {
		return nil, nil, nil
	}
	org := st.Org

	want := make(map[string]bool, len(cfg.Org.Owners))
	for _, o := range cfg.Org.Owners {
		if o = strings.TrimSpace(o); o != "" {
			want[strings.ToLower(o)] = true
		}
	}
	if len(want) == 0 {
		return nil, nil, nil
	}

	current, err := fetchOrgOwners(ctx, c, org)
	if err != nil {
		return nil, nil, err
	}
	st.CurrentOwners = len(current)
	st.DesiredOwners = len(want)

	var changes []util.Change
	for _, login := range sortedSet(want) {
		if !current[login] {
			changes = append(changes, util.Change{
				Scope:  scopeOrgOwner,
				Target: login,
				Action: util.ActionEnsure,
				Details: map[string]any{
					"org":  org,
					"user": login,
				},
			})
		}
	}

	var extra []string
	for _, login := range sortedSet(current) {
		if !want[login] {
			extra = append(extra, login)
		}
	}
	if len(extra) == 0 {
		return changes, nil, nil
	}

	var warnings []string
	if !cfg.App.DemoteUnconfiguredOwners {
		if cfg.App.DryWarnings.WarnUnmanagedOwners {
			warnings = append(warnings, fmt.Sprintf(
				"Found %d organization owner(s) not listed in org.yaml: %v (set demote_unconfigured_owners to drop them to member)",
				len(extra), extra))
		}
		return changes, warnings, nil
	}

	// Only looked up when a demotion is actually on the table, and only useful
	// under PAT auth — an installation token authenticates an app, which cannot
	// be an owner and so cannot demote itself.
	self := authenticatedLogin(ctx, c)
	for _, login := range extra {
		if self != "" && login == self {
			warnings = append(warnings, fmt.Sprintf(
				"Not demoting %q: it is the account this run authenticated as, and an owner that demotes itself cannot undo it", login))
			continue
		}
		changes = append(changes, util.Change{
			Scope:  scopeOrgOwner,
			Target: login,
			Action: util.ActionRemove,
			Details: map[string]any{
				"org":  org,
				"user": login,
			},
		})
	}
	return changes, warnings, nil
}

// fetchOrgOwners returns the organization's current admins, lowercased.
func fetchOrgOwners(ctx context.Context, c *gh.Client, org string) (map[string]bool, error) {
	opt := &github.ListMembersOptions{
		Role:        orgRoleAdmin,
		ListOptions: github.ListOptions{PerPage: defaultPerPage},
	}
	owners := map[string]bool{}
	if err := paginate(func(opts *github.ListOptions) (*github.Response, error) {
		opt.ListOptions = *opts
		users, resp, err := c.REST.Organizations.ListMembers(ctx, org, opt)
		if err != nil {
			return nil, err
		}
		for _, u := range users {
			owners[strings.ToLower(u.GetLogin())] = true
		}
		return resp, nil
	}); err != nil {
		return nil, fmt.Errorf("list owners of org %q: %w", org, err)
	}
	return owners, nil
}

// authenticatedLogin returns the login this run is acting as, lowercased, or ""
// when there is not one. A GitHub App installation has no user behind it, and
// the endpoint answers with an error rather than a login — which is the answer,
// not a failure, so it is not reported as one.
func authenticatedLogin(ctx context.Context, c *gh.Client) string {
	user, _, err := c.REST.Users.Get(ctx, "")
	if err != nil {
		return ""
	}
	return strings.ToLower(user.GetLogin())
}

// sortedSet returns a set's members in a stable order.
func sortedSet(set map[string]bool) []string {
	out := make([]string, 0, len(set))
	for k, v := range set {
		if v {
			out = append(out, k)
		}
	}
	sort.Strings(out)
	return out
}
