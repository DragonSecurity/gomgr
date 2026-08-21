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

// permRank orders repository permissions so two can be compared. An unknown or
// absent permission ranks below every real one, so a value gomgr does not
// recognize can never be read as covering a grant it does not cover.
func permRank(p string) int {
	switch normalizePermission(p) {
	case permPull:
		return 1
	case permTriage:
		return 2
	case permPush:
		return 3
	case permMaintain:
		return 4
	case permAdmin:
		return 5
	default:
		return 0
	}
}

// higherPerm returns whichever of the two grants more.
func higherPerm(a, b string) string {
	if permRank(b) > permRank(a) {
		return b
	}
	return a
}

// planCollaborators finds direct repository grants that give someone more than
// their team membership and the organization's default repository permission
// already do, and — when RemoveExcessCollaborators is set — revokes them.
//
// This is the only part of gomgr that looks at access granted outside a team.
// The GitHub UI's "add collaborator" button writes exactly such a grant, and
// nothing in a YAML file mentions it, so without this a hand-made admin grant
// outlives every sync that claims the configuration is authoritative.
//
// Two properties keep it from being destructive by accident:
//
//   - It only ever removes. A direct grant is never created to match config,
//     because the way to give someone access here is to put them in a team.
//   - It compares against the union of what GitHub reports today and what this
//     configuration asks for. A team gomgr does not manage still counts toward
//     what someone is entitled to, so enabling this in a half-adopted org
//     revokes the grants that are genuinely redundant and leaves the rest.
//
// Outside collaborators — accounts in no team at all — are entitled to nothing
// under this model, so every direct grant they hold is excess. That is the
// intended reading of "all access flows through teams", and it is why this
// takes an explicit flag rather than riding along with another one.
func planCollaborators(ctx context.Context, c *gh.Client, cfg *config.Root, st *State, desired map[string]config.TeamConfig) ([]util.Change, []string, error) {
	if !cfg.App.RemoveExcessCollaborators && !cfg.App.DryWarnings.WarnExcessCollaborators {
		return nil, nil, nil
	}
	org := st.Org

	entitled, err := entitlements(ctx, c, st, desired)
	if err != nil {
		return nil, nil, err
	}

	baseline, err := st.orgDefaultRepoPermission(ctx, c)
	if err != nil {
		return nil, nil, err
	}

	// Owners hold admin on every repository in the organization by virtue of
	// being owners; a direct grant on top of that is redundant rather than
	// excessive, and revoking it changes nothing about what they can do.
	owners, err := fetchOrgOwners(ctx, c, org)
	if err != nil {
		return nil, nil, err
	}
	for _, o := range cfg.Org.Owners {
		owners[strings.ToLower(o)] = true
	}

	members, err := st.orgMembers(ctx, c)
	if err != nil {
		return nil, nil, err
	}

	var changes []util.Change
	var excess []string
	for _, repo := range sortedSet(st.ManagedRepos) {
		if ctx.Err() != nil {
			return nil, nil, ctx.Err()
		}
		direct, err := fetchDirectCollaborators(ctx, c, org, repo)
		if err != nil {
			return nil, nil, err
		}
		for _, login := range sortedKeysOf(direct) {
			if owners[login] {
				continue
			}
			held := direct[login]
			covered := entitled[repo][login]
			if members[login] {
				covered = higherPerm(covered, baseline)
			}
			if permRank(held) <= permRank(covered) {
				continue
			}
			excess = append(excess, fmt.Sprintf("%s on %s/%s (%s, entitled to %s)",
				login, org, repo, held, describePerm(covered)))
			if !cfg.App.RemoveExcessCollaborators {
				continue
			}
			changes = append(changes, util.Change{
				Scope:  scopeRepoCollaborator,
				Target: repo + "/" + login,
				Action: util.ActionRemove,
				Details: map[string]any{
					"org":       org,
					"repo":      repo,
					"user":      login,
					"held":      held,
					"entitled":  describePerm(covered),
					"team_only": cfg.App.RemoveExcessCollaborators,
				},
			})
		}
	}

	var warnings []string
	if len(excess) > 0 && cfg.App.DryWarnings.WarnExcessCollaborators {
		verb := "found"
		if cfg.App.RemoveExcessCollaborators {
			verb = "revoking"
		}
		warnings = append(warnings, fmt.Sprintf(
			"%s %d direct repository grant(s) beyond team access: %v", verb, len(excess), excess))
	}
	return changes, warnings, nil
}

// describePerm names a permission for a human, including the absence of one.
func describePerm(p string) string {
	if permRank(p) == 0 {
		return "no team access"
	}
	return p
}

// entitlements maps repository -> login -> the highest permission that
// membership of some team grants.
//
// Both halves are a union of observed and configured state. A team that exists
// on GitHub but not in the config still grants what it grants; a team this run
// is about to create grants what the config says it will. Erring toward the
// larger entitlement means the only grants revoked are ones nothing else
// justifies.
func entitlements(ctx context.Context, c *gh.Client, st *State, desired map[string]config.TeamConfig) (map[string]map[string]string, error) {
	actualMembers, err := st.allTeamMembers(ctx, c)
	if err != nil {
		return nil, err
	}
	actualRepos, err := st.allTeamRepos(ctx, c)
	if err != nil {
		return nil, err
	}

	// Team membership, by slug.
	membersBySlug := map[string]map[string]bool{}
	for slug, logins := range actualMembers {
		membersBySlug[slug] = map[string]bool{}
		for login := range logins {
			membersBySlug[slug][login] = true
		}
	}
	for slug, team := range desired {
		if membersBySlug[slug] == nil {
			membersBySlug[slug] = map[string]bool{}
		}
		for _, u := range team.Maintainers {
			membersBySlug[slug][strings.ToLower(u)] = true
		}
		for _, u := range team.Members {
			membersBySlug[slug][strings.ToLower(u)] = true
		}
	}

	// Repository permissions each team holds directly, by slug.
	permsBySlug := map[string]map[string]string{}
	for slug, repos := range actualRepos {
		permsBySlug[slug] = map[string]string{}
		for repo, perm := range repos {
			permsBySlug[slug][repo] = perm
		}
	}
	for slug, team := range desired {
		if permsBySlug[slug] == nil {
			permsBySlug[slug] = map[string]string{}
		}
		for repo, val := range team.Repositories {
			settings, err := parseRepoConfig(val)
			if err != nil {
				// The loader has already rejected anything malformed; a value
				// this cannot read here is not a reason to stop, only a reason
				// not to widen someone's entitlement on the strength of it.
				continue
			}
			key := strings.ToLower(repo)
			permsBySlug[slug][key] = higherPerm(permsBySlug[slug][key], settings.permission)
		}
	}

	// Nesting grants the parent's repository access to the child, so fold each
	// ancestor's permissions down before reading them off.
	effective := map[string]map[string]string{}
	for slug := range membersBySlug {
		merged := map[string]string{}
		for repo, perm := range permsBySlug[slug] {
			merged[repo] = higherPerm(merged[repo], perm)
		}
		for _, parent := range ancestors(slug, desired) {
			for repo, perm := range permsBySlug[parent] {
				merged[repo] = higherPerm(merged[repo], perm)
			}
		}
		effective[slug] = merged
	}

	out := map[string]map[string]string{}
	for slug, repos := range effective {
		for repo, perm := range repos {
			if out[repo] == nil {
				out[repo] = map[string]string{}
			}
			for login := range membersBySlug[slug] {
				out[repo][login] = higherPerm(out[repo][login], perm)
			}
		}
	}
	return out, nil
}

// affiliationDirect asks GitHub for only the grants made against a repository
// itself, excluding access that arrives through a team or through organization
// membership. That exclusion is the whole distinction this file draws.
const affiliationDirect = "direct"

// fetchDirectCollaborators returns the logins holding a grant made against this
// repository itself, and the permission each holds.
func fetchDirectCollaborators(ctx context.Context, c *gh.Client, org, repo string) (map[string]string, error) {
	opt := &github.ListCollaboratorsOptions{
		Affiliation: affiliationDirect,
		ListOptions: github.ListOptions{PerPage: defaultPerPage},
	}
	out := map[string]string{}
	if err := paginate(func(opts *github.ListOptions) (*github.Response, error) {
		opt.ListOptions = *opts
		users, resp, err := c.REST.Repositories.ListCollaborators(ctx, org, repo, opt)
		if err != nil {
			return nil, err
		}
		for _, u := range users {
			out[strings.ToLower(u.GetLogin())] = collaboratorPermission(u)
		}
		return resp, nil
	}); err != nil {
		return nil, fmt.Errorf("list direct collaborators of %s/%s: %w", org, repo, err)
	}
	return out, nil
}

// collaboratorPermission reads the permission off a collaborator entry.
// role_name is the precise answer and names custom roles too; the boolean
// permissions block is the fallback for responses that omit it.
func collaboratorPermission(u *github.User) string {
	if role := u.GetRoleName(); role != "" {
		return normalizePermission(role)
	}
	switch p := u.Permissions; {
	case p == nil:
		return ""
	case p.GetAdmin():
		return permAdmin
	case p.GetMaintain():
		return permMaintain
	case p.GetPush():
		return permPush
	case p.GetTriage():
		return permTriage
	case p.GetPull():
		return permPull
	}
	return ""
}

// sortedKeysOf returns a string-keyed map's keys in a stable order.
func sortedKeysOf[V any](m map[string]V) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
