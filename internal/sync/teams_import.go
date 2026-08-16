package sync

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/google/go-github/v90/github"

	"github.com/DragonSecurity/gomgr/internal/config"
	"github.com/DragonSecurity/gomgr/internal/gh"
	"github.com/DragonSecurity/gomgr/internal/util"
)

// ImportedTeam is a team that exists on GitHub, rendered as configuration.
type ImportedTeam struct {
	Config config.TeamConfig
	// Parent is the slug of the team this one is nested under, if any. gomgr
	// does not manage team hierarchy, so this is reported rather than written.
	Parent string
}

// SkippedTeam is a team that could not be turned into configuration.
type SkippedTeam struct {
	Slug   string
	Reason string
}

// TeamImportResult is what a team scan found.
type TeamImportResult struct {
	// Teams holds teams the configuration does not declare, sorted by slug.
	Teams []ImportedTeam
	// Skipped lists teams that exist but could not be represented.
	Skipped []SkippedTeam
	// AlreadyDeclared counts teams skipped because a team file declares them.
	AlreadyDeclared int
	// Ungranted lists repositories in the organization that no team — declared
	// or imported — grants access to. They have no home in the configuration.
	Ungranted []string
	// DeletionRisk is true when Ungranted is non-empty and the configuration
	// has delete_unmanaged_repos set, which means the next sync would delete
	// exactly those repositories.
	DeletionRisk bool
}

// Total returns how many teams are available to adopt.
func (r *TeamImportResult) Total() int { return len(r.Teams) }

// ImportTeams discovers the teams that exist on GitHub and renders them as
// configuration: name, privacy, description, maintainers, members, and the
// repositories each team can reach with the permission it holds.
//
// This is the command for bringing an organization under management that was
// never under management before. Teams the configuration already declares are
// left alone, on the same reasoning as the ruleset import: a declared team is
// gomgr's to define, and overwriting it from the live state is backwards.
func ImportTeams(ctx context.Context, c *gh.Client, cfg *config.Root) (*TeamImportResult, error) {
	org := cfg.App.Org
	result := &TeamImportResult{}

	declared := map[string]bool{}
	for _, t := range cfg.Team {
		declared[strings.ToLower(t.ResolvedSlug())] = true
	}

	var actual []*github.Team
	if err := paginate(func(opts *github.ListOptions) (*github.Response, error) {
		teams, resp, err := c.REST.Teams.ListTeams(ctx, org, opts)
		if err != nil {
			return nil, err
		}
		actual = append(actual, teams...)
		return resp, nil
	}); err != nil {
		return nil, fmt.Errorf("list teams for %s: %w", org, err)
	}

	granted := map[string]bool{}
	for _, team := range actual {
		slug := team.GetSlug()
		repos, err := teamRepoPermissions(ctx, c, org, slug)
		if err != nil {
			return nil, err
		}
		for repo := range repos {
			granted[repo] = true
		}

		if declared[strings.ToLower(slug)] {
			result.AlreadyDeclared++
			continue
		}

		imported, err := teamToConfig(ctx, c, org, team, repos)
		if err != nil {
			return nil, err
		}
		if err := validateTeamConfig(imported.Config); err != nil {
			result.Skipped = append(result.Skipped, SkippedTeam{Slug: slug, Reason: err.Error()})
			continue
		}
		result.Teams = append(result.Teams, imported)
	}

	sort.Slice(result.Teams, func(i, j int) bool {
		return result.Teams[i].Config.ResolvedSlug() < result.Teams[j].Config.ResolvedSlug()
	})

	// Repositories no team reaches have nowhere to live in the configuration,
	// and under delete_unmanaged_repos they are what the next sync deletes.
	ungranted, err := ungrantedRepos(ctx, c, org, granted)
	if err != nil {
		return nil, err
	}
	result.Ungranted = ungranted
	result.DeletionRisk = len(ungranted) > 0 && cfg.App.DeleteUnmanagedRepos

	return result, nil
}

// teamToConfig assembles one team's configuration from the API.
func teamToConfig(ctx context.Context, c *gh.Client, org string, team *github.Team, repos map[string]string) (ImportedTeam, error) {
	slug := team.GetSlug()

	maintainers, err := teamMembers(ctx, c, org, slug, roleMaintainer)
	if err != nil {
		return ImportedTeam{}, err
	}
	members, err := teamMembers(ctx, c, org, slug, roleMember)
	if err != nil {
		return ImportedTeam{}, err
	}

	cfgTeam := config.TeamConfig{
		Name:        team.GetName(),
		Description: team.GetDescription(),
		Privacy:     team.GetPrivacy(),
		Maintainers: maintainers,
		Members:     members,
	}
	// Only record the slug when it is not what the name would derive to, so an
	// imported file carries the same amount of detail a hand-written one would.
	if slug != cfgTeam.ResolvedSlug() {
		cfgTeam.Slug = slug
	}
	if len(repos) > 0 {
		cfgTeam.Repositories = make(map[string]any, len(repos))
		for repo, perm := range repos {
			cfgTeam.Repositories[repo] = perm
		}
	}

	return ImportedTeam{Config: cfgTeam, Parent: team.GetParent().GetSlug()}, nil
}

// teamMembers lists the logins holding a role on a team, sorted so that
// re-importing produces the same file.
func teamMembers(ctx context.Context, c *gh.Client, org, slug, role string) ([]string, error) {
	var out []string
	listOpts := &github.TeamListTeamMembersOptions{
		Role:        role,
		ListOptions: github.ListOptions{PerPage: defaultPerPage},
	}
	if err := paginate(func(opts *github.ListOptions) (*github.Response, error) {
		listOpts.ListOptions = *opts
		users, resp, err := c.REST.Teams.ListTeamMembersBySlug(ctx, org, slug, listOpts)
		if err != nil {
			if isNotFound(err) {
				return &github.Response{}, nil
			}
			return nil, err
		}
		for _, u := range users {
			out = append(out, u.GetLogin())
		}
		return resp, nil
	}); err != nil {
		return nil, fmt.Errorf("list %ss of team %s: %w", role, slug, err)
	}
	sort.Strings(out)
	return out, nil
}

// teamRepoPermissions maps repository name to the permission a team holds on it.
func teamRepoPermissions(ctx context.Context, c *gh.Client, org, slug string) (map[string]string, error) {
	out := map[string]string{}
	if err := paginate(func(opts *github.ListOptions) (*github.Response, error) {
		repos, resp, err := c.REST.Teams.ListTeamReposBySlug(ctx, org, slug, opts)
		if err != nil {
			var ghErr *github.ErrorResponse
			if errors.As(err, &ghErr) && ghErr.Response != nil && ghErr.Response.StatusCode == http.StatusNotFound {
				return &github.Response{}, nil
			}
			return nil, err
		}
		for _, repo := range repos {
			if perm := extractRepoPerm(repo); perm != "" {
				out[repo.GetName()] = perm
			}
		}
		return resp, nil
	}); err != nil {
		return nil, fmt.Errorf("list repositories for team %s: %w", slug, err)
	}
	return out, nil
}

// ungrantedRepos returns the organization's repositories that no team reaches.
func ungrantedRepos(ctx context.Context, c *gh.Client, org string, granted map[string]bool) ([]string, error) {
	lowered := make(map[string]bool, len(granted))
	for repo := range granted {
		lowered[strings.ToLower(repo)] = true
	}

	var out []string
	repoOpt := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{PerPage: defaultPerPage},
		Type:        "all",
	}
	if err := paginate(func(opts *github.ListOptions) (*github.Response, error) {
		repoOpt.ListOptions = *opts
		repos, resp, err := c.REST.Repositories.ListByOrg(ctx, org, repoOpt)
		if err != nil {
			return nil, err
		}
		for _, r := range repos {
			if !lowered[strings.ToLower(r.GetName())] {
				out = append(out, r.GetName())
			}
		}
		return resp, nil
	}); err != nil {
		return nil, fmt.Errorf("list repos for %s: %w", org, err)
	}
	sort.Strings(out)
	return out, nil
}

// validateTeamConfig runs one imported team through the checks a hand-written
// team file gets, so anything gomgr's schema cannot express is reported rather
// than written out and rediscovered on the next load.
func validateTeamConfig(team config.TeamConfig) error {
	root := &config.Root{
		App:  config.AppConfig{Org: "placeholder"},
		Team: []config.TeamConfig{team},
	}
	return root.Validate()
}

// LogTeamImportWarnings emits the caveats a team import cannot encode in the
// files it writes.
func LogTeamImportWarnings(result *TeamImportResult) {
	for _, t := range result.Teams {
		if t.Parent != "" {
			util.Warnf("team %q is nested under %q; gomgr does not manage team hierarchy, so the imported file does not record it",
				t.Config.ResolvedSlug(), t.Parent)
		}
	}
}
