package sync

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strings"

	"github.com/DragonSecurity/gomgr/internal/config"
	"github.com/DragonSecurity/gomgr/internal/gh"
	"github.com/DragonSecurity/gomgr/internal/templates"
	"github.com/DragonSecurity/gomgr/internal/util"
	"github.com/google/go-github/v83/github"
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
	validTopic := regexp.MustCompile(`^[a-z0-9-]+$`)
	if !validTopic.MatchString(topic) {
		return fmt.Errorf("topic contains invalid characters (must be lowercase alphanumeric with hyphens): %q", topic)
	}
	return nil
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
	case map[string]any:
		// Advanced case: RepoConfig structure
		if perm, ok := v["permission"].(string); ok {
			if perm == "" {
				return settings, fmt.Errorf("permission cannot be empty string")
			}
			settings.permission = perm
		} else if _, hasPermission := v["permission"]; hasPermission {
			return settings, fmt.Errorf("permission must be a string, got %T", v["permission"])
		}
		// Permission is optional if using advanced config for topics/pinning only

		if topics, ok := v["topics"].([]any); ok {
			for _, t := range topics {
				if tStr, ok := t.(string); ok {
					settings.topics = append(settings.topics, tStr)
				}
			}
		}
		if pinned, ok := v["pinned"].(bool); ok {
			settings.pinned = pinned
		}
		if template, ok := v["template"].(bool); ok {
			settings.template = template
		}
		if from, ok := v["from"].(string); ok {
			settings.from = from
		}
	case map[any]any:
		// YAML might unmarshal as map[any]any
		if perm, ok := v["permission"].(string); ok {
			if perm == "" {
				return settings, fmt.Errorf("permission cannot be empty string")
			}
			settings.permission = perm
		} else if _, hasPermission := v["permission"]; hasPermission {
			return settings, fmt.Errorf("permission must be a string, got %T", v["permission"])
		}

		if topics, ok := v["topics"].([]any); ok {
			for _, t := range topics {
				if tStr, ok := t.(string); ok {
					settings.topics = append(settings.topics, tStr)
				}
			}
		}
		if pinned, ok := v["pinned"].(bool); ok {
			settings.pinned = pinned
		}
		if template, ok := v["template"].(bool); ok {
			settings.template = template
		}
		if from, ok := v["from"].(string); ok {
			settings.from = from
		}
	}

	return settings, nil
}

// resolveTemplate resolves template inheritance for a repository configuration.
// If the repo has a "from" field, it looks up the template repository and merges settings.
// Topics are combined (union), template flag is not inherited, and permission can be overridden.
func resolveTemplate(repoName string, settings repoSettings, allRepos map[string]repoSettings, defaultOrg string) (repoSettings, error) {
	if settings.from == "" {
		return settings, nil
	}

	// Parse template reference (supports "repo-name" or "org/repo-name")
	templateOrg := defaultOrg
	templateRepo := settings.from
	if strings.Contains(settings.from, "/") {
		parts := strings.SplitN(settings.from, "/", 2)
		if len(parts) != 2 {
			return settings, fmt.Errorf("invalid template reference format: %q (expected 'repo' or 'org/repo')", settings.from)
		}
		templateOrg = parts[0]
		templateRepo = parts[1]
	}

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

// ---- planning ----

func planTeams(ctx context.Context, c *gh.Client, cfg *config.Root, st *State) ([]util.Change, map[string]config.TeamConfig, error) {
	var out []util.Change
	desired := map[string]config.TeamConfig{}

	// build desired map
	for _, t := range cfg.Team {
		slug := t.Slug
		if slug == "" && t.Name != "" {
			slug = strings.ToLower(strings.ReplaceAll(t.Name, " ", "-"))
		}
		if slug == "" {
			continue
		}
		t.Slug = slug
		desired[slug] = t
	}

	// list actual teams
	var actual []*github.Team
	opt := &github.ListOptions{PerPage: 100}
	for {
		ts, resp, err := c.REST.Teams.ListTeams(ctx, st.Org, opt)
		if err != nil {
			return nil, nil, err
		}
		actual = append(actual, ts...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	actualBySlug := map[string]*github.Team{}
	for _, t := range actual {
		actualBySlug[t.GetSlug()] = t
	}

	// Track state
	st.CurrentTeams = len(actual)
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
		// TODO: compare & update description/privacy/parents as needed
	}
	return out, desired, nil
}

func planTeamMembership(ctx context.Context, c *gh.Client, cfg *config.Root, st *State, desiredBySlug map[string]config.TeamConfig) ([]util.Change, error) {
	var out []util.Change
	org := st.Org

	totalCurrentMembers := 0
	totalDesiredMembers := 0

	for slug, want := range desiredBySlug {
		// actual role map
		got := map[string]string{}
		// maintainers
		mopts := &github.TeamListTeamMembersOptions{Role: "maintainer", ListOptions: github.ListOptions{PerPage: 100}}
		for {
			users, resp, err := c.REST.Teams.ListTeamMembersBySlug(ctx, org, slug, mopts)
			if err != nil {
				var ghErr *github.ErrorResponse
				if errors.As(err, &ghErr) && ghErr.Response != nil && ghErr.Response.StatusCode == http.StatusNotFound {
					break
				}
				return nil, err
			}
			for _, u := range users {
				got[strings.ToLower(u.GetLogin())] = "maintainer"
			}
			if resp.NextPage == 0 {
				break
			}
			mopts.Page = resp.NextPage
		}
		// members
		opts := &github.TeamListTeamMembersOptions{Role: "member", ListOptions: github.ListOptions{PerPage: 100}}
		for {
			users, resp, err := c.REST.Teams.ListTeamMembersBySlug(ctx, org, slug, opts)
			if err != nil {
				var ghErr *github.ErrorResponse
				if errors.As(err, &ghErr) && ghErr.Response != nil && ghErr.Response.StatusCode == http.StatusNotFound {
					break
				}
				return nil, err
			}
			for _, u := range users {
				if _, ok := got[strings.ToLower(u.GetLogin())]; !ok {
					got[strings.ToLower(u.GetLogin())] = "member"
				}
			}
			if resp.NextPage == 0 {
				break
			}
			opts.Page = resp.NextPage
		}

		// desired role map
		wantRole := map[string]string{}
		for _, u := range want.Maintainers {
			wantRole[strings.ToLower(u)] = "maintainer"
		}
		for _, u := range want.Members {
			if _, ok := wantRole[strings.ToLower(u)]; !ok {
				wantRole[strings.ToLower(u)] = "member"
			}
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

func planRepoPerms(ctx context.Context, c *gh.Client, cfg *config.Root, st *State) ([]util.Change, error) {
	var out []util.Change
	org := st.Org

	existing := map[string]bool{}
	existingRepos := map[string]*github.Repository{}
	opt := &github.RepositoryListByOrgOptions{ListOptions: github.ListOptions{PerPage: 100}, Type: "all"}
	for {
		repos, resp, err := c.REST.Repositories.ListByOrg(ctx, org, opt)
		if err != nil {
			return nil, err
		}
		for _, r := range repos {
			repoName := strings.ToLower(r.GetName())
			existing[repoName] = true
			existingRepos[repoName] = r
		}
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	// Track which repos are managed
	managedRepos := map[string]bool{}

	// Map to track desired topics and pinned state per repo
	desiredTopics := map[string][]string{}
	desiredPinned := map[string]bool{}
	desiredTemplates := map[string]bool{}

	// First pass: collect all repository settings
	allRepoSettings := map[string]repoSettings{}
	repoToTeams := map[string][]string{} // track which teams reference each repo

	for _, t := range cfg.Team {
		slug := t.Slug
		if slug == "" {
			slug = strings.ToLower(strings.ReplaceAll(t.Name, " ", "-"))
		}
		for repo, val := range t.Repositories {
			r := strings.ToLower(repo)
			managedRepos[r] = true

			settings, err := parseRepoConfig(val)
			if err != nil {
				return nil, fmt.Errorf("invalid config for repo %s in team %s: %w", repo, slug, err)
			}

			// Store settings for later template resolution
			allRepoSettings[r] = settings
			repoToTeams[r] = append(repoToTeams[r], slug)
		}
	}

	// Second pass: resolve templates
	resolvedSettings := make(map[string]repoSettings)
	for repo, settings := range allRepoSettings {
		resolved, err := resolveTemplate(repo, settings, allRepoSettings, org)
		if err != nil {
			return nil, fmt.Errorf("error resolving template for repo %s: %w", repo, err)
		}
		resolvedSettings[repo] = resolved
	}

	// Third pass: process repositories with resolved settings
	for _, t := range cfg.Team {
		slug := t.Slug
		if slug == "" {
			slug = strings.ToLower(strings.ReplaceAll(t.Name, " ", "-"))
		}
		for repo := range t.Repositories {
			r := strings.ToLower(repo)
			settings := resolvedSettings[r]

			if !existing[r] && cfg.App.CreateRepo {
				details := map[string]any{
					"org":     org,
					"name":    repo,
					"private": true,
				}
				// Include template information if present
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

			// Mark repository as template if configured
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

			// Aggregate topics from all teams (union)
			if len(settings.topics) > 0 {
				existingTopics := desiredTopics[r]
				topicSet := map[string]bool{}
				for _, topic := range existingTopics {
					topicSet[topic] = true
				}
				for _, topic := range settings.topics {
					// Validate topic before adding
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

			// Pinned state: if any team wants it pinned, pin it
			if settings.pinned {
				desiredPinned[r] = true
			}

			if cfg.App.AddDefaultReadme {
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
			}
			if cfg.App.AddRenovateConfig && cfg.App.RenovateConfig != "" {
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
			}
		}
	}

	// Plan topic updates for managed repos - only if different from current state
	for repo, topics := range desiredTopics {
		if len(topics) > 0 {
			// GitHub allows max 20 topics per repo
			if len(topics) > 20 {
				return nil, fmt.Errorf("repo %s has %d topics (max 20 allowed)", repo, len(topics))
			}

			// Check if topics differ from current state
			needsUpdate := false
			if existingRepo, ok := existingRepos[repo]; ok {
				currentTopics := existingRepo.Topics
				if len(currentTopics) != len(topics) {
					needsUpdate = true
				} else {
					// Compare topics (order-independent)
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
				// Repo doesn't exist yet, will be created
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
	}

	// Plan pinning changes - check current pinning state
	// Note: GitHub REST API doesn't provide pinning status directly
	// We'll generate changes for all repos marked as pinned
	// A future enhancement could use GraphQL to query current pinning state
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
			// Check if repo needs to be marked as template
			needsUpdate := false
			if existingRepo, ok := existingRepos[repo]; ok {
				if !existingRepo.GetIsTemplate() {
					needsUpdate = true
				}
			} else {
				// Repo will be created, needs template marking
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

	// Store managed repos in state for cleanup phase
	st.ManagedRepos = managedRepos

	// Track repository counts
	st.CurrentRepos = len(existing)
	st.DesiredRepos = len(managedRepos)

	// Count permissions (team-repo grants)
	// Note: This requires additional API calls to get accurate current state.
	// These calls are intentional for precise state tracking and run only during
	// dry-run planning. The overhead is acceptable for the visibility benefit.
	currentPermsCount := 0
	desiredPermsCount := 0

	// Count current permissions from GitHub
	for _, t := range cfg.Team {
		teamSlug := t.Slug
		if teamSlug == "" {
			teamSlug = strings.ToLower(strings.ReplaceAll(t.Name, " ", "-"))
		}
		// List team repos to count current permissions
		repoOpts := &github.ListOptions{PerPage: 100}
		for {
			teamRepos, resp, err := c.REST.Teams.ListTeamReposBySlug(ctx, org, teamSlug, repoOpts)
			if err != nil {
				// If team doesn't exist yet, skip counting
				var ghErr *github.ErrorResponse
				if errors.As(err, &ghErr) && ghErr.Response != nil && ghErr.Response.StatusCode == http.StatusNotFound {
					break
				}
				// Ignore other errors for counting purposes
				break
			}
			currentPermsCount += len(teamRepos)
			if resp.NextPage == 0 {
				break
			}
			repoOpts.Page = resp.NextPage
		}
	}

	// Count desired permissions
	for _, t := range cfg.Team {
		desiredPermsCount += len(t.Repositories)
	}

	st.CurrentRepoPerms = currentPermsCount
	st.DesiredRepoPerms = desiredPermsCount

	return out, nil
}

func planCleanups(ctx context.Context, c *gh.Client, cfg *config.Root, st *State, desired map[string]config.TeamConfig) ([]util.Change, []string, error) {
	var out []util.Change
	var warnings []string
	org := st.Org
	if cfg.App.DeleteUnconfiguredTeams {
		var actual []*github.Team
		opt := &github.ListOptions{PerPage: 100}
		for {
			ts, resp, err := c.REST.Teams.ListTeams(ctx, org, opt)
			if err != nil {
				return nil, nil, err
			}
			actual = append(actual, ts...)
			if resp.NextPage == 0 {
				break
			}
			opt.Page = resp.NextPage
		}
		for _, at := range actual {
			if _, ok := desired[at.GetSlug()]; !ok {
				out = append(out, util.Change{Scope: "team", Target: at.GetSlug(), Action: "delete", Details: map[string]any{"org": org, "slug": at.GetSlug()}})
			}
		}
	}

	if cfg.App.RemoveMembersWithoutTeam {
		// list all org members
		memOpt := &github.ListMembersOptions{
			Role: "member",
			ListOptions: github.ListOptions{
				PerPage: 100,
			},
		}
		var members []*github.User
		for {
			us, resp, err := c.REST.Organizations.ListMembers(ctx, org, memOpt)
			if err != nil {
				return nil, nil, err
			}
			members = append(members, us...)
			if resp.NextPage == 0 {
				break
			}
			memOpt.Page = resp.NextPage
		}
		// compute members who are in any team
		inAnyTeam := map[string]bool{}
		teamOpt := &github.ListOptions{PerPage: 100}
		for {
			ts, resp, err := c.REST.Teams.ListTeams(ctx, org, teamOpt)
			if err != nil {
				return nil, nil, err
			}
			for _, t := range ts {
				page := &github.TeamListTeamMembersOptions{Role: "all", ListOptions: github.ListOptions{PerPage: 100}}
				for {
					us, r2, err := c.REST.Teams.ListTeamMembersBySlug(ctx, org, t.GetSlug(), page)
					if err != nil {
						return nil, nil, err
					}
					for _, u := range us {
						inAnyTeam[strings.ToLower(u.GetLogin())] = true
					}
					if r2.NextPage == 0 {
						break
					}
					page.Page = r2.NextPage
				}
			}
			if resp.NextPage == 0 {
				break
			}
			teamOpt.Page = resp.NextPage
		}
		for _, u := range members {
			login := strings.ToLower(u.GetLogin())
			if !inAnyTeam[login] {
				out = append(out, util.Change{Scope: "org-member", Target: login, Action: "remove", Details: map[string]any{"org": org, "user": login}})
			}
		}
	}

	// Warn about or delete unmanaged repositories
	if cfg.App.DeleteUnmanagedRepos || cfg.App.DryWarnings.WarnUnmanagedRepos {
		var actualRepos []*github.Repository
		repoOpt := &github.RepositoryListByOrgOptions{ListOptions: github.ListOptions{PerPage: 100}, Type: "all"}
		for {
			repos, resp, err := c.REST.Repositories.ListByOrg(ctx, org, repoOpt)
			if err != nil {
				return nil, nil, err
			}
			actualRepos = append(actualRepos, repos...)
			if resp.NextPage == 0 {
				break
			}
			repoOpt.Page = resp.NextPage
		}
		var unmanagedRepos []string
		for _, repo := range actualRepos {
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
		// Add warning if configured and there are unmanaged repos
		if cfg.App.DryWarnings.WarnUnmanagedRepos && len(unmanagedRepos) > 0 {
			warnings = append(warnings, fmt.Sprintf("Found %d unmanaged repositories: %v", len(unmanagedRepos), unmanagedRepos))
		}
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

func applyChanges(ctx context.Context, c *gh.Client, changes []util.Change) error {
	precedence := map[string]int{
		"custom-role:create":   5, // Create custom roles first, before teams/repos
		"custom-role:update":   5,
		"team:create":          10,
		"repo:ensure":          10,
		"team-repo:grant":      20,
		"team-member:ensure":   30,
		"repo-file:ensure":     40,
		"repo-topics:ensure":   45,
		"repo-template:ensure": 46,
		"repo-pin:ensure":      47,
		"team:delete":          90,
		"repo:delete":          90,
		"custom-role:delete":   95, // Delete custom roles last
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

	for _, ch := range changes {
		// Skip custom role changes - already handled above
		if strings.HasPrefix(ch.Scope, "custom-role") {
			continue
		}

		switch ch.Scope + ":" + ch.Action {
		case "team:create":
			d, _ := ch.Details.(map[string]any)
			org := fmt.Sprint(d["org"])
			name := fmt.Sprint(d["name"])
			var privacyPtr, descPtr *string
			if v, ok := d["privacy"]; ok && fmt.Sprint(v) != "" {
				pv := fmt.Sprint(v)
				privacyPtr = github.Ptr(pv)
			}
			if v, ok := d["description"]; ok && fmt.Sprint(v) != "" {
				dv := fmt.Sprint(v)
				descPtr = github.Ptr(dv)
			}
			newTeam := github.NewTeam{Name: name, Privacy: privacyPtr, Description: descPtr}
			_, _, err := c.REST.Teams.CreateTeam(ctx, org, newTeam)
			if err != nil {
				return fmt.Errorf("create team %q: %w", name, err)
			}

		case "team:delete":
			d, _ := ch.Details.(map[string]any)
			org := fmt.Sprint(d["org"])
			slug := fmt.Sprint(d["slug"])
			_, err := c.REST.Teams.DeleteTeamBySlug(ctx, org, slug)
			if err != nil {
				return fmt.Errorf("delete team %s: %w", slug, err)
			}

		case "team-member:ensure":
			d, ok := ch.Details.(teamMemberChange)
			if !ok {
				return fmt.Errorf("invalid details for team-member:ensure")
			}
			_, _, err := c.REST.Teams.AddTeamMembershipBySlug(ctx, d.Org, d.Slug, d.User, &github.TeamAddTeamMembershipOptions{Role: d.Role})
			if err != nil {
				return fmt.Errorf("add %s as %s to %s: %w", d.User, d.Role, d.Slug, err)
			}

		case "repo:ensure":
			d, _ := ch.Details.(map[string]any)
			org := fmt.Sprint(d["org"])
			name := fmt.Sprint(d["name"])
			private := true
			if v, ok := d["private"]; ok {
				private = fmt.Sprint(v) != "false"
			}
			isTemplate := false
			if v, ok := d["template"]; ok {
				isTemplate = fmt.Sprint(v) == "true"
			}

			// Check if this repo should be created from a template
			if fromTemplate, ok := d["from"]; ok && fromTemplate != "" {
				templateRef := fmt.Sprint(fromTemplate)
				// Parse template reference (supports "repo-name" or "org/repo-name")
				templateOrg := org
				templateRepo := templateRef
				if strings.Contains(templateRef, "/") {
					parts := strings.SplitN(templateRef, "/", 2)
					if len(parts) == 2 {
						templateOrg = parts[0]
						templateRepo = parts[1]
					}
				}

				// Create repository from template
				_, _, err := c.REST.Repositories.CreateFromTemplate(ctx, templateOrg, templateRepo, &github.TemplateRepoRequest{
					Name:    github.Ptr(name),
					Owner:   github.Ptr(org),
					Private: github.Ptr(private),
				})
				if err != nil {
					var ghErr *github.ErrorResponse
					if errors.As(err, &ghErr) && ghErr.Response != nil && ghErr.Response.StatusCode == 422 {
						// already exists race
					} else {
						return fmt.Errorf("create repo %s/%s from template %s/%s: %w", org, name, templateOrg, templateRepo, err)
					}
				}
			} else {
				// Create regular repository
				_, _, err := c.REST.Repositories.Create(ctx, org, &github.Repository{
					Name:                github.Ptr(name),
					Private:             github.Ptr(private),
					IsTemplate:          github.Ptr(isTemplate),
					AllowAutoMerge:      github.Ptr(true),
					AllowMergeCommit:    github.Ptr(false),
					DeleteBranchOnMerge: github.Ptr(true),
					HasIssues:           github.Ptr(true),
				})
				if err != nil {
					var ghErr *github.ErrorResponse
					if errors.As(err, &ghErr) && ghErr.Response != nil && ghErr.Response.StatusCode == 422 {
						// already exists race
					} else {
						return fmt.Errorf("create repo %s/%s: %w", org, name, err)
					}
				}
			}

		case "team-repo:grant":
			d, _ := ch.Details.(map[string]any)
			org := fmt.Sprint(d["org"])
			slug := fmt.Sprint(d["slug"])
			repo := fmt.Sprint(d["repo"])
			perm := normalizePermission(fmt.Sprint(d["permission"]))
			_, err := c.REST.Teams.AddTeamRepoBySlug(ctx, org, slug, org, repo, &github.TeamAddTeamRepoOptions{Permission: perm})
			if err != nil {
				return fmt.Errorf("grant %s on %s/%s to %s: %w", perm, org, repo, slug, err)
			}

		case "repo-file:ensure":
			d, _ := ch.Details.(map[string]any)
			org := fmt.Sprint(d["org"])
			repo := fmt.Sprint(d["repo"])
			path := fmt.Sprint(d["path"])
			content := []byte(fmt.Sprint(d["content"]))
			message := fmt.Sprint(d["message"])
			branch := fmt.Sprint(d["branch"])
			file, _, resp, err := c.REST.Repositories.GetContents(ctx, org, repo, path, &github.RepositoryContentGetOptions{Ref: branch})
			if err != nil && (resp == nil || resp.StatusCode != http.StatusNotFound) {
				return fmt.Errorf("check %s/%s:%s: %w", org, repo, path, err)
			}
			if file == nil {
				_, _, err := c.REST.Repositories.CreateFile(ctx, org, repo, path, &github.RepositoryContentFileOptions{
					Message: github.Ptr(message),
					Content: content,
					Branch:  github.Ptr(branch),
				})
				if err != nil {
					// Handle race condition: If repository was created from template,
					// files may exist even though GetContents returned nil.
					// This can happen due to timing - template files are copied asynchronously.
					// GitHub returns 422 with "sha wasn't supplied" or 409 with "reference already exists"
					// when trying to create a file that already exists.
					var ghErr *github.ErrorResponse
					if errors.As(err, &ghErr) && ghErr.Response != nil {
						// Check if this is a race condition error
						isRaceCondition := (ghErr.Response.StatusCode == 422 && containsErrorMessage(ghErr, "sha", "wasn't supplied")) ||
							(ghErr.Response.StatusCode == 409 && containsErrorMessage(ghErr, "reference already exists"))

						if !isRaceCondition {
							return fmt.Errorf("create file %s in %s/%s: %w", path, org, repo, err)
						}
						// File already exists (likely from template), which is what we want - skip error
					} else {
						return fmt.Errorf("create file %s in %s/%s: %w", path, org, repo, err)
					}
				}
			} else {
				// optional: update if differs (skipped for now)
			}

		case "repo-topics:ensure":
			d, _ := ch.Details.(map[string]any)
			org := fmt.Sprint(d["org"])
			repo := fmt.Sprint(d["repo"])

			// Handle topics - may come as []string or []any from planning
			var topicsRaw []string
			if v, ok := d["topics"]; ok {
				switch topics := v.(type) {
				case []string:
					topicsRaw = topics
				case []any:
					for _, t := range topics {
						if tStr, ok := t.(string); ok {
							topicsRaw = append(topicsRaw, tStr)
						}
					}
				default:
					return fmt.Errorf("invalid type for topics for %s/%s: %T", org, repo, v)
				}
			}

			_, _, err := c.REST.Repositories.ReplaceAllTopics(ctx, org, repo, topicsRaw)
			if err != nil {
				return fmt.Errorf("set topics on %s/%s: %w", org, repo, err)
			}

		case "repo-template:ensure":
			d, _ := ch.Details.(map[string]any)
			org := fmt.Sprint(d["org"])
			repo := fmt.Sprint(d["repo"])

			// Mark repository as a template
			_, _, err := c.REST.Repositories.Edit(ctx, org, repo, &github.Repository{
				IsTemplate: github.Ptr(true),
			})
			if err != nil {
				return fmt.Errorf("mark repo %s/%s as template: %w", org, repo, err)
			}

		case "repo-pin:ensure":
			d, _ := ch.Details.(map[string]any)
			org := fmt.Sprint(d["org"])
			repo := fmt.Sprint(d["repo"])

			// Note: GitHub's GraphQL API does not support pinning repositories to organization profiles.
			// The pinRepository mutation only works for user profiles, not organizations.
			// This is a known limitation of the GitHub API.
			// See: https://github.com/orgs/community/discussions/184845
			util.Warnf("Skipping pin for %s/%s: GitHub API does not support pinning to organization profiles", org, repo)

		case "repo:delete":
			d, _ := ch.Details.(map[string]any)
			org := fmt.Sprint(d["org"])
			repo := fmt.Sprint(d["repo"])
			_, err := c.REST.Repositories.Delete(ctx, org, repo)
			if err != nil {
				return fmt.Errorf("delete repo %s/%s: %w", org, repo, err)
			}

		default:
			// no-op for unhandled changes
		}
	}
	return nil
}

func normalizePermission(p string) string {
	// Use lowercase comparison to match built-in roles case-insensitively
	switch strings.ToLower(p) {
	case "read", "pull":
		return "pull"
	case "triage":
		return "triage"
	case "write", "push":
		return "push"
	case "maintain":
		return "maintain"
	case "admin":
		return "admin"
	default:
		// For custom repository roles (GitHub Enterprise Cloud), pass through the role name as-is
		// preserving the original case since custom role names may be case-sensitive
		// Custom roles must be created in the GitHub organization before use
		// Examples: "actions-manager", "release-manager", "runner-admin"
		return p
	}
}
