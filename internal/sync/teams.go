package sync

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"sort"
	"strings"

	"github.com/google/go-github/v84/github"

	"github.com/DragonSecurity/gomgr/internal/config"
	"github.com/DragonSecurity/gomgr/internal/gh"
	"github.com/DragonSecurity/gomgr/internal/templates"
	"github.com/DragonSecurity/gomgr/internal/util"
)

const defaultPerPage = 100

var validTopicRe = regexp.MustCompile(`^[a-z0-9-]+$`)

const (
	roleMaintainer = "maintainer"
	roleMember     = "member"
)

// Repository permission levels used across planning and apply.
const (
	permPull     = "pull"
	permTriage   = "triage"
	permPush     = "push"
	permMaintain = "maintain"
	permAdmin    = "admin"
)

const (
	precedenceCustomRoleCreate   = 5
	precedenceCustomRoleUpdate   = 5
	precedenceTeamCreate         = 10
	precedenceTeamUpdate         = 15
	precedenceRepoEnsure         = 10
	precedenceTeamRepoGrant      = 20
	precedenceTeamMemberEnsure   = 30
	precedenceRepoFileEnsure     = 40
	precedenceRepoTopicsEnsure   = 45
	precedenceRepoTemplateEnsure = 46
	precedenceRepoPinEnsure      = 47
	precedenceOrgMemberRemove    = 85
	precedenceTeamDelete         = 90
	precedenceRepoDelete         = 90
	precedenceCustomRoleDelete   = 95
)

const (
	errTermSHA            = "sha"
	errTermSHANotSupplied = "wasn't supplied"
	errTermRefExists      = "reference already exists"
)

type teamMemberChange struct {
	Org  string
	Slug string
	User string
	Role string // "member" or "maintainer"
}

type repoSettings struct {
	permission string
	topics     []string
	pinned     bool
	template   bool
	from       string
}

// validateTopic checks if a topic name meets GitHub requirements:
// - lowercase alphanumeric with hyphens
// - max 50 characters
// - cannot start with a hyphen
func validateTopic(topic string) error {
	if len(topic) == 0 {
		return fmt.Errorf("topic cannot be empty")
	}
	if len(topic) > 50 {
		return fmt.Errorf("topic exceeds 50 characters: %q", topic)
	}
	if topic[0] == '-' {
		return fmt.Errorf("topic cannot start with hyphen: %q", topic)
	}
	// Match lowercase alphanumeric and hyphens only
	if !validTopicRe.MatchString(topic) {
		return fmt.Errorf("topic contains invalid characters (must be lowercase alphanumeric with hyphens): %q", topic)
	}
	return nil
}

// normalizeYAMLMap converts both map[string]any and map[any]any (from YAML) to map[string]any.
func normalizeYAMLMap(v any) (map[string]any, bool) {
	switch m := v.(type) {
	case map[string]any:
		return m, true
	case map[any]any:
		result := make(map[string]any, len(m))
		for k, val := range m {
			result[fmt.Sprint(k)] = val
		}
		return result, true
	default:
		return nil, false
	}
}

// parseRepoConfig parses a repository value which can be either:
// - a simple string (permission only)
// - a map with permission, topics, pinned fields
func parseRepoConfig(val any) (repoSettings, error) {
	settings := repoSettings{}

	switch v := val.(type) {
	case string:
		// Simple case: just a permission string
		if v == "" {
			return settings, fmt.Errorf("permission cannot be empty string")
		}
		settings.permission = v
	default:
		m, ok := normalizeYAMLMap(val)
		if !ok {
			return settings, nil
		}
		if perm, ok := m["permission"].(string); ok {
			if perm == "" {
				return settings, fmt.Errorf("permission cannot be empty string")
			}
			settings.permission = perm
		} else if _, hasPermission := m["permission"]; hasPermission {
			return settings, fmt.Errorf("permission must be a string, got %T", m["permission"])
		}
		// Permission is optional if using advanced config for topics/pinning only

		if topics, ok := m["topics"].([]any); ok {
			for _, t := range topics {
				if tStr, ok := t.(string); ok {
					settings.topics = append(settings.topics, tStr)
				}
			}
		}
		if pinned, ok := m["pinned"].(bool); ok {
			settings.pinned = pinned
		}
		if template, ok := m["template"].(bool); ok {
			settings.template = template
		}
		if from, ok := m["from"].(string); ok {
			settings.from = from
		}
	}

	return settings, nil
}

// parseTemplateRef splits a template reference into org and repo parts.
// Supports "repo-name" (uses defaultOrg) or "org/repo-name".
func parseTemplateRef(ref, defaultOrg string) (org, repo string) {
	if strings.Contains(ref, "/") {
		parts := strings.SplitN(ref, "/", 2)
		return parts[0], parts[1]
	}
	return defaultOrg, ref
}

// resolveTemplate resolves template inheritance for a repository configuration.
// If the repo has a "from" field, it looks up the template repository and merges settings.
// Topics are combined (union), template flag is not inherited, and permission can be overridden.
func resolveTemplate(_ string, settings repoSettings, allRepos map[string]repoSettings, defaultOrg string) (repoSettings, error) {
	if settings.from == "" {
		return settings, nil
	}

	// Parse template reference (supports "repo-name" or "org/repo-name")
	templateOrg, templateRepo := parseTemplateRef(settings.from, defaultOrg)

	// Only support same-org templates for now
	if templateOrg != defaultOrg {
		return settings, fmt.Errorf("cross-organization template references not yet supported: %q", settings.from)
	}

	// Look up template repository in the current configuration
	templateKey := strings.ToLower(templateRepo)
	templateSettings, exists := allRepos[templateKey]
	if !exists {
		return settings, fmt.Errorf("template repository %q not found in configuration", templateRepo)
	}

	if !templateSettings.template {
		return settings, fmt.Errorf("repository %q is referenced as template but not marked with template: true", templateRepo)
	}

	// Merge settings: inherit from template, override with repo-specific
	result := settings

	// Inherit permission if not specified
	if result.permission == "" && templateSettings.permission != "" {
		result.permission = templateSettings.permission
	}

	// Merge topics (union): template topics + repo-specific topics
	// Clear existing topics first since we'll rebuild the list
	result.topics = nil
	topicSet := make(map[string]bool)

	// Add template topics first
	for _, topic := range templateSettings.topics {
		topicSet[topic] = true
		result.topics = append(result.topics, topic)
	}

	// Add repo-specific topics that aren't already in the set
	for _, topic := range settings.topics {
		if !topicSet[topic] {
			topicSet[topic] = true
			result.topics = append(result.topics, topic)
		}
	}

	// Don't inherit template or pinned flags
	// result.template is already false (or explicitly set)
	// result.pinned is already set from repo config

	return result, nil
}

// paginate calls fn repeatedly, advancing through pages until there are no more.
func paginate(fn func(opts *github.ListOptions) (*github.Response, error)) error {
	opts := &github.ListOptions{PerPage: defaultPerPage}
	for {
		resp, err := fn(opts)
		if err != nil {
			return err
		}
		if resp == nil || resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return nil
}

// ---- planning ----

func planTeams(_ context.Context, _ *gh.Client, cfg *config.Root, st *State) ([]util.Change, map[string]config.TeamConfig, error) {
	var out []util.Change
	desired := map[string]config.TeamConfig{}

	// build desired map
	for _, t := range cfg.Team {
		slug := t.ResolvedSlug()
		if slug == "" {
			continue
		}
		t.Slug = slug
		desired[slug] = t
	}

	// use prefetched teams
	actualBySlug := map[string]*github.Team{}
	for _, t := range st.ActualTeams {
		actualBySlug[t.GetSlug()] = t
	}

	// Track state
	st.CurrentTeams = len(st.ActualTeams)
	st.DesiredTeams = len(desired)

	for slug, want := range desired {
		if _, ok := actualBySlug[slug]; !ok {
			out = append(out, util.Change{
				Scope:  "team",
				Target: slug,
				Action: "create",
				Details: map[string]any{
					"org":         st.Org,
					"name":        want.Name,
					"privacy":     want.Privacy,
					"description": want.Description,
				},
			})
			continue
		}
		// Compare & update description/privacy
		existing := actualBySlug[slug]
		needsUpdate := false
		updateDetails := map[string]any{
			"org":  st.Org,
			"slug": slug,
			"name": want.Name,
		}
		if want.Description != existing.GetDescription() {
			needsUpdate = true
			updateDetails["description"] = want.Description
		}
		if want.Privacy != "" && want.Privacy != existing.GetPrivacy() {
			needsUpdate = true
			updateDetails["privacy"] = want.Privacy
		}
		if needsUpdate {
			out = append(out, util.Change{
				Scope:   "team",
				Target:  slug,
				Action:  "update",
				Details: updateDetails,
			})
		}
	}
	return out, desired, nil
}

func planTeamMembership(ctx context.Context, c *gh.Client, st *State, desiredBySlug map[string]config.TeamConfig) ([]util.Change, error) {
	var out []util.Change
	org := st.Org

	totalCurrentMembers := 0
	totalDesiredMembers := 0

	validatedUsers := map[string]bool{}

	for slug, want := range desiredBySlug {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		// actual role map
		got := map[string]string{}
		// maintainers
		mopts := &github.TeamListTeamMembersOptions{Role: roleMaintainer, ListOptions: github.ListOptions{PerPage: defaultPerPage}}
		if err := paginate(func(opts *github.ListOptions) (*github.Response, error) {
			mopts.ListOptions = *opts
			users, resp, err := c.REST.Teams.ListTeamMembersBySlug(ctx, org, slug, mopts)
			if err != nil {
				var ghErr *github.ErrorResponse
				if errors.As(err, &ghErr) && ghErr.Response != nil && ghErr.Response.StatusCode == http.StatusNotFound {
					return &github.Response{}, nil
				}
				return nil, err
			}
			for _, u := range users {
				got[strings.ToLower(u.GetLogin())] = roleMaintainer
			}
			return resp, nil
		}); err != nil {
			return nil, err
		}
		// members
		memOpts := &github.TeamListTeamMembersOptions{Role: roleMember, ListOptions: github.ListOptions{PerPage: defaultPerPage}}
		if err := paginate(func(opts *github.ListOptions) (*github.Response, error) {
			memOpts.ListOptions = *opts
			users, resp, err := c.REST.Teams.ListTeamMembersBySlug(ctx, org, slug, memOpts)
			if err != nil {
				var ghErr *github.ErrorResponse
				if errors.As(err, &ghErr) && ghErr.Response != nil && ghErr.Response.StatusCode == http.StatusNotFound {
					return &github.Response{}, nil
				}
				return nil, err
			}
			for _, u := range users {
				if _, ok := got[strings.ToLower(u.GetLogin())]; !ok {
					got[strings.ToLower(u.GetLogin())] = roleMember
				}
			}
			return resp, nil
		}); err != nil {
			return nil, err
		}

		// desired role map
		wantRole := map[string]string{}
		for _, u := range want.Maintainers {
			wantRole[strings.ToLower(u)] = roleMaintainer
		}
		for _, u := range want.Members {
			if _, ok := wantRole[strings.ToLower(u)]; !ok {
				wantRole[strings.ToLower(u)] = roleMember
			}
		}

		// Validate that all desired users exist on GitHub
		for user := range wantRole {
			if validatedUsers[user] {
				continue
			}
			_, _, err := c.REST.Users.Get(ctx, user)
			if err != nil {
				return nil, fmt.Errorf("user %q in team %q not found on GitHub: %w", user, slug, err)
			}
			validatedUsers[user] = true
		}

		// Track member counts
		totalCurrentMembers += len(got)
		totalDesiredMembers += len(wantRole)

		for user, role := range wantRole {
			if got[user] == role {
				continue
			}
			out = append(out, util.Change{
				Scope:   "team-member",
				Target:  slug,
				Action:  "ensure",
				Details: teamMemberChange{Org: org, Slug: slug, User: user, Role: role},
			})
		}
		// (optional) removals left for later
	}

	// Update state
	st.CurrentTeamMembers = totalCurrentMembers
	st.DesiredTeamMembers = totalDesiredMembers

	return out, nil
}

// collectRepoSettings gathers and validates all repository settings from config.
func collectRepoSettings(cfg *config.Root, _ string) (allSettings map[string]repoSettings, managedRepos map[string]bool, err error) {
	allSettings = map[string]repoSettings{}
	managedRepos = map[string]bool{}

	for _, t := range cfg.Team {
		slug := t.ResolvedSlug()
		for repo, val := range t.Repositories {
			r := strings.ToLower(repo)
			managedRepos[r] = true

			settings, err := parseRepoConfig(val)
			if err != nil {
				return nil, nil, fmt.Errorf("invalid config for repo %s in team %s: %w", repo, slug, err)
			}
			allSettings[r] = settings
		}
	}
	return allSettings, managedRepos, nil
}

// resolveAllTemplates resolves template inheritance for all repository settings.
func resolveAllTemplates(allSettings map[string]repoSettings, org string) (map[string]repoSettings, error) {
	resolved := make(map[string]repoSettings, len(allSettings))
	for repo, settings := range allSettings {
		r, err := resolveTemplate(repo, settings, allSettings, org)
		if err != nil {
			return nil, fmt.Errorf("error resolving template for repo %s: %w", repo, err)
		}
		resolved[repo] = r
	}
	return resolved, nil
}

// teamRepoPermKey is "team-slug/repo-name" (lowercase).
type teamRepoPermKey = string

// fetchCurrentPermissions fetches the current team-repo permission grants from GitHub.
// Returns the total count and a map of "team/repo" -> permission string.
func fetchCurrentPermissions(ctx context.Context, c *gh.Client, cfg *config.Root, org string) (int, map[teamRepoPermKey]string, error) {
	count := 0
	permMap := map[teamRepoPermKey]string{}
	for _, t := range cfg.Team {
		if ctx.Err() != nil {
			return 0, nil, ctx.Err()
		}
		teamSlug := t.ResolvedSlug()
		if err := paginate(func(opts *github.ListOptions) (*github.Response, error) {
			teamRepos, resp, err := c.REST.Teams.ListTeamReposBySlug(ctx, org, teamSlug, opts)
			if err != nil {
				var ghErr *github.ErrorResponse
				if errors.As(err, &ghErr) && ghErr.Response != nil && ghErr.Response.StatusCode == http.StatusNotFound {
					return &github.Response{}, nil
				}
				return nil, err
			}
			count += len(teamRepos)
			for _, repo := range teamRepos {
				repoName := strings.ToLower(repo.GetName())
				perm := extractRepoPerm(repo)
				permMap[teamSlug+"/"+repoName] = perm
			}
			return resp, nil
		}); err != nil {
			return 0, nil, fmt.Errorf("fetch permissions for team %s: %w", teamSlug, err)
		}
	}
	return count, permMap, nil
}

// extractRepoPerm returns the highest permission level granted to a team for a repo.
func extractRepoPerm(repo *github.Repository) string {
	p := repo.Permissions
	if p == nil {
		return ""
	}
	switch {
	case p.Admin != nil && *p.Admin:
		return permAdmin
	case p.Maintain != nil && *p.Maintain:
		return permMaintain
	case p.Push != nil && *p.Push:
		return permPush
	case p.Triage != nil && *p.Triage:
		return permTriage
	case p.Pull != nil && *p.Pull:
		return permPull
	default:
		return ""
	}
}

func planRepoPerms(ctx context.Context, c *gh.Client, cfg *config.Root, st *State) ([]util.Change, error) {
	var out []util.Change
	org := st.Org

	// use prefetched repos
	existing := map[string]bool{}
	existingRepos := map[string]*github.Repository{}
	for _, r := range st.ActualRepos {
		repoName := strings.ToLower(r.GetName())
		existing[repoName] = true
		existingRepos[repoName] = r
	}

	allRepoSettings, managedRepos, err := collectRepoSettings(cfg, org)
	if err != nil {
		return nil, err
	}

	resolvedSettings, err := resolveAllTemplates(allRepoSettings, org)
	if err != nil {
		return nil, err
	}

	desiredTopics := map[string][]string{}
	desiredPinned := map[string]bool{}
	desiredTemplates := map[string]bool{}
	emittedFiles := map[string]bool{} // tracks repo-level file changes to avoid duplicates

	for _, t := range cfg.Team {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		slug := t.ResolvedSlug()
		for repo := range t.Repositories {
			r := strings.ToLower(repo)
			settings := resolvedSettings[r]

			if !existing[r] && cfg.App.CreateRepo {
				details := map[string]any{
					"org":     org,
					"name":    repo,
					"private": true,
				}
				if settings.from != "" {
					details["from"] = settings.from
				}
				if settings.template {
					details["template"] = true
				}
				out = append(out, util.Change{
					Scope:   "repo",
					Target:  r,
					Action:  "ensure",
					Details: details,
				})
				existing[r] = true
			}

			if settings.template {
				desiredTemplates[r] = true
			}

			out = append(out, util.Change{
				Scope:  "team-repo",
				Target: slug + "/" + r,
				Action: "grant",
				Details: map[string]any{
					"org":        org,
					"slug":       slug,
					"repo":       repo,
					"permission": settings.permission,
				},
			})

			if len(settings.topics) > 0 {
				existingTopics := desiredTopics[r]
				topicSet := map[string]bool{}
				for _, topic := range existingTopics {
					topicSet[topic] = true
				}
				for _, topic := range settings.topics {
					if err := validateTopic(topic); err != nil {
						return nil, fmt.Errorf("invalid topic for repo %s: %w", repo, err)
					}
					if !topicSet[topic] {
						existingTopics = append(existingTopics, topic)
						topicSet[topic] = true
					}
				}
				desiredTopics[r] = existingTopics
			}

			if settings.pinned {
				desiredPinned[r] = true
			}

			// Emit file changes only once per repo (skip if already emitted from another team)
			if cfg.App.AddDefaultReadme && !emittedFiles[r+":README.md"] {
				readmeContent, err := templates.GenerateReadme(org, repo)
				if err != nil {
					return nil, fmt.Errorf("failed to generate README for %s: %w", repo, err)
				}
				out = append(out, util.Change{
					Scope:  "repo-file",
					Target: r + ":README.md",
					Action: "ensure",
					Details: map[string]any{
						"org":     org,
						"repo":    repo,
						"path":    "README.md",
						"content": readmeContent,
						"message": "chore: add default README",
						"branch":  "main",
					},
				})
				emittedFiles[r+":README.md"] = true
			}
			if cfg.App.AddRenovateConfig && cfg.App.RenovateConfig != "" && !emittedFiles[r+":renovate"] {
				out = append(out, util.Change{
					Scope:  "repo-file",
					Target: r + ":.github/renovate.json",
					Action: "ensure",
					Details: map[string]any{
						"org":     org,
						"repo":    repo,
						"path":    ".github/renovate.json",
						"content": cfg.App.RenovateConfig,
						"message": "chore: add Renovate config",
						"branch":  "main",
					},
				})
				emittedFiles[r+":renovate"] = true
			}
		}
	}

	// Plan topic updates
	for repo, topics := range desiredTopics {
		if len(topics) > 20 {
			return nil, fmt.Errorf("repo %s has %d topics (max 20 allowed)", repo, len(topics))
		}
		needsUpdate := false
		if existingRepo, ok := existingRepos[repo]; ok {
			currentTopics := existingRepo.Topics
			if len(currentTopics) != len(topics) {
				needsUpdate = true
			} else {
				currentSet := make(map[string]bool)
				for _, t := range currentTopics {
					currentSet[t] = true
				}
				for _, t := range topics {
					if !currentSet[t] {
						needsUpdate = true
						break
					}
				}
			}
		} else {
			needsUpdate = true
		}
		if needsUpdate {
			out = append(out, util.Change{
				Scope:  "repo-topics",
				Target: repo,
				Action: "ensure",
				Details: map[string]any{
					"org":    org,
					"repo":   repo,
					"topics": topics,
				},
			})
		}
	}

	// Plan pinning changes
	for repo, shouldPin := range desiredPinned {
		if shouldPin {
			out = append(out, util.Change{
				Scope:  "repo-pin",
				Target: repo,
				Action: "ensure",
				Details: map[string]any{
					"org":    org,
					"repo":   repo,
					"pinned": true,
				},
			})
		}
	}

	// Plan template marking changes
	for repo, shouldBeTemplate := range desiredTemplates {
		if shouldBeTemplate {
			needsUpdate := false
			if existingRepo, ok := existingRepos[repo]; ok {
				if !existingRepo.GetIsTemplate() {
					needsUpdate = true
				}
			} else {
				needsUpdate = true
			}
			if needsUpdate {
				out = append(out, util.Change{
					Scope:  "repo-template",
					Target: repo,
					Action: "ensure",
					Details: map[string]any{
						"org":      org,
						"repo":     repo,
						"template": true,
					},
				})
			}
		}
	}

	st.ManagedRepos = managedRepos
	st.CurrentRepos = len(existing)
	st.DesiredRepos = len(managedRepos)

	currentPerms, currentPermMap, err := fetchCurrentPermissions(ctx, c, cfg, org)
	if err != nil {
		return nil, fmt.Errorf("fetch current permissions: %w", err)
	}
	st.CurrentRepoPerms = currentPerms
	desiredPermsCount := 0
	for _, t := range cfg.Team {
		desiredPermsCount += len(t.Repositories)
	}
	st.DesiredRepoPerms = desiredPermsCount

	// Filter out no-op grants where the permission already matches
	filtered := out[:0]
	for _, ch := range out {
		if ch.Scope == "team-repo" && ch.Action == "grant" {
			d := ch.Details.(map[string]any)
			slug := d["slug"].(string)
			repo := strings.ToLower(d["repo"].(string))
			desired := normalizePermission(d["permission"].(string))
			current := currentPermMap[slug+"/"+repo]
			if current == desired {
				continue // skip, already has correct permission
			}
		}
		filtered = append(filtered, ch)
	}

	return filtered, nil
}

// planTeamCleanups generates delete changes for teams not in the desired set.
func planTeamCleanups(st *State, org string, desired map[string]config.TeamConfig) ([]util.Change, error) {
	var out []util.Change
	for _, at := range st.ActualTeams {
		if _, ok := desired[at.GetSlug()]; !ok {
			out = append(out, util.Change{Scope: "team", Target: at.GetSlug(), Action: "delete", Details: map[string]any{"org": org, "slug": at.GetSlug()}})
		}
	}
	return out, nil
}

// planMemberCleanups generates remove changes for org members not in any team.
func planMemberCleanups(ctx context.Context, c *gh.Client, org string) ([]util.Change, error) {
	var out []util.Change
	memOpt := &github.ListMembersOptions{
		Role:        roleMember,
		ListOptions: github.ListOptions{PerPage: defaultPerPage},
	}
	var members []*github.User
	if err := paginate(func(opts *github.ListOptions) (*github.Response, error) {
		memOpt.ListOptions = *opts
		us, resp, err := c.REST.Organizations.ListMembers(ctx, org, memOpt)
		if err != nil {
			return nil, err
		}
		members = append(members, us...)
		return resp, nil
	}); err != nil {
		return nil, err
	}
	inAnyTeam := map[string]bool{}
	var allTeams []*github.Team
	if err := paginate(func(opts *github.ListOptions) (*github.Response, error) {
		ts, resp, err := c.REST.Teams.ListTeams(ctx, org, opts)
		if err != nil {
			return nil, err
		}
		allTeams = append(allTeams, ts...)
		return resp, nil
	}); err != nil {
		return nil, err
	}
	for _, t := range allTeams {
		tmOpt := &github.TeamListTeamMembersOptions{Role: "all", ListOptions: github.ListOptions{PerPage: defaultPerPage}}
		if err := paginate(func(opts *github.ListOptions) (*github.Response, error) {
			tmOpt.ListOptions = *opts
			us, resp, err := c.REST.Teams.ListTeamMembersBySlug(ctx, org, t.GetSlug(), tmOpt)
			if err != nil {
				return nil, err
			}
			for _, u := range us {
				inAnyTeam[strings.ToLower(u.GetLogin())] = true
			}
			return resp, nil
		}); err != nil {
			return nil, err
		}
	}
	for _, u := range members {
		login := strings.ToLower(u.GetLogin())
		if !inAnyTeam[login] {
			out = append(out, util.Change{Scope: "org-member", Target: login, Action: "remove", Details: map[string]any{"org": org, "user": login}})
		}
	}
	return out, nil
}

// planRepoCleanups generates delete/warning changes for unmanaged repositories.
func planRepoCleanups(cfg *config.Root, st *State) ([]util.Change, []string, error) {
	var out []util.Change
	var warnings []string
	org := st.Org
	var unmanagedRepos []string
	for _, repo := range st.ActualRepos {
		repoName := strings.ToLower(repo.GetName())
		if !st.ManagedRepos[repoName] {
			unmanagedRepos = append(unmanagedRepos, repo.GetName())
			if cfg.App.DeleteUnmanagedRepos {
				out = append(out, util.Change{
					Scope:  "repo",
					Target: repoName,
					Action: "delete",
					Details: map[string]any{
						"org":  org,
						"repo": repo.GetName(),
					},
				})
			}
		}
	}
	if cfg.App.DryWarnings.WarnUnmanagedRepos && len(unmanagedRepos) > 0 {
		warnings = append(warnings, fmt.Sprintf("Found %d unmanaged repositories: %v", len(unmanagedRepos), unmanagedRepos))
	}
	return out, warnings, nil
}

func planCleanups(ctx context.Context, c *gh.Client, cfg *config.Root, st *State, desired map[string]config.TeamConfig) ([]util.Change, []string, error) {
	var out []util.Change
	var warnings []string
	org := st.Org

	if cfg.App.DeleteUnconfiguredTeams {
		if ctx.Err() != nil {
			return nil, nil, ctx.Err()
		}
		changes, err := planTeamCleanups(st, org, desired)
		if err != nil {
			return nil, nil, err
		}
		out = append(out, changes...)
	}

	if cfg.App.RemoveMembersWithoutTeam {
		if ctx.Err() != nil {
			return nil, nil, ctx.Err()
		}
		changes, err := planMemberCleanups(ctx, c, org)
		if err != nil {
			return nil, nil, err
		}
		out = append(out, changes...)
	}

	if cfg.App.DeleteUnmanagedRepos || cfg.App.DryWarnings.WarnUnmanagedRepos {
		if ctx.Err() != nil {
			return nil, nil, ctx.Err()
		}
		changes, w, err := planRepoCleanups(cfg, st)
		if err != nil {
			return nil, nil, err
		}
		out = append(out, changes...)
		warnings = append(warnings, w...)
	}

	return out, warnings, nil
}

// containsErrorMessage checks if a GitHub ErrorResponse contains a specific error message
// in either the main Message field or in any of the individual Error messages in the Errors array.
func containsErrorMessage(ghErr *github.ErrorResponse, searchTerms ...string) bool {
	// Check main message (only if not empty)
	if ghErr.Message != "" {
		allFound := true
		for _, term := range searchTerms {
			if !strings.Contains(ghErr.Message, term) {
				allFound = false
				break
			}
		}
		if allFound {
			return true
		}
	}

	// Check individual errors in the Errors array
	for _, e := range ghErr.Errors {
		allFound := true
		for _, term := range searchTerms {
			if !strings.Contains(e.Message, term) {
				allFound = false
				break
			}
		}
		if allFound {
			return true
		}
	}

	return false
}

// ---- apply ----

// applyHandlers maps change keys (scope:action) to handler functions.
var applyHandlers = map[string]func(context.Context, *gh.Client, util.Change) error{
	"team:create":          applyTeamCreate,
	"team:update":          applyTeamUpdate,
	"team:delete":          applyTeamDelete,
	"team-member:ensure":   applyTeamMemberEnsure,
	"repo:ensure":          applyRepoEnsure,
	"team-repo:grant":      applyTeamRepoGrant,
	"repo-file:ensure":     applyRepoFileEnsure,
	"repo-topics:ensure":   applyRepoTopicsEnsure,
	"repo-template:ensure": applyRepoTemplateEnsure,
	"repo-pin:ensure":      applyRepoPinEnsure,
	"repo:delete":          applyRepoDelete,
	"org-member:remove":    applyOrgMemberRemove,
}

func applyChanges(ctx context.Context, c *gh.Client, changes []util.Change) error {
	precedence := map[string]int{
		"custom-role:create":   precedenceCustomRoleCreate,
		"custom-role:update":   precedenceCustomRoleUpdate,
		"team:create":          precedenceTeamCreate,
		"team:update":          precedenceTeamUpdate,
		"repo:ensure":          precedenceRepoEnsure,
		"team-repo:grant":      precedenceTeamRepoGrant,
		"team-member:ensure":   precedenceTeamMemberEnsure,
		"repo-file:ensure":     precedenceRepoFileEnsure,
		"repo-topics:ensure":   precedenceRepoTopicsEnsure,
		"repo-template:ensure": precedenceRepoTemplateEnsure,
		"repo-pin:ensure":      precedenceRepoPinEnsure,
		"org-member:remove":    precedenceOrgMemberRemove,
		"team:delete":          precedenceTeamDelete,
		"repo:delete":          precedenceRepoDelete,
		"custom-role:delete":   precedenceCustomRoleDelete,
	}

	sort.Slice(changes, func(i, j int) bool {
		ai := changes[i].Scope + ":" + changes[i].Action
		aj := changes[j].Scope + ":" + changes[j].Action
		return precedence[ai] < precedence[aj]
	})

	// Apply custom role changes first
	if err := applyCustomRoleChanges(ctx, c, changes); err != nil {
		return err
	}

	// Count non-custom-role changes for progress display
	total := 0
	for _, ch := range changes {
		if !strings.HasPrefix(ch.Scope, "custom-role") {
			total++
		}
	}

	applied := 0
	for _, ch := range changes {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		// Skip custom role changes - already handled above
		if strings.HasPrefix(ch.Scope, "custom-role") {
			continue
		}

		applied++
		log.Printf("[%d/%d] %s:%s %s", applied, total, ch.Scope, ch.Action, ch.Target)

		if err := gh.RespectRate(ctx, c.REST); err != nil {
			util.Warnf("rate limit check failed: %v", err)
		}

		key := ch.Scope + ":" + ch.Action
		handler, ok := applyHandlers[key]
		if !ok {
			util.Warnf("no handler for change %s:%s on %s", ch.Scope, ch.Action, ch.Target)
			continue
		}
		if err := handler(ctx, c, ch); err != nil {
			util.Audit(ch.Scope, ch.Target, ch.Action, "error")
			return err
		}
		util.Audit(ch.Scope, ch.Target, ch.Action, "ok")
	}
	return nil
}
