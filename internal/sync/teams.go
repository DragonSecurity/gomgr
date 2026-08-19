package sync

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"sort"
	"strings"

	"github.com/google/go-github/v90/github"

	"github.com/DragonSecurity/gomgr/internal/config"
	"github.com/DragonSecurity/gomgr/internal/gh"
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
	// permRead is GitHub's name for permPull in some API responses.
	permRead = "read"
)

// Repository visibilities. GitHub rejects anything else.
const (
	visPublic   = "public"
	visPrivate  = "private"
	visInternal = "internal"
)

// Change scopes, joining the ones already declared next to the code that plans
// them (scopeOrgRuleset, scopeRepoSettings, scopeRepoFilePR and friends).
const (
	scopeRepo         = "repo"
	scopeRepoFile     = "repo-file"
	scopeRepoPin      = "repo-pin"
	scopeRepoTemplate = "repo-template"
	scopeRepoTopics   = "repo-topics"
	scopeTeam         = "team"
	scopeTeamMember   = "team-member"
	scopeTeamRepo     = "team-repo"
	scopeOrgMember    = "org-member"
	scopeCustomRole   = "custom-role"
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
	precedenceRepoSettingsEnsure = 44
	precedenceRepoTopicsEnsure   = 45
	precedenceRepoTemplateEnsure = 46
	precedenceRepoPinEnsure      = 47
	precedenceRepoVisibility     = 48
	// Rulesets go on last of the mutating changes. A guard rail that requires a
	// pull request would otherwise reject the file writes above it, which push
	// straight to the default branch in the same run.
	precedenceOrgRulesetCreate  = 60
	precedenceOrgRulesetUpdate  = 60
	precedenceRepoRulesetCreate = 62
	precedenceRepoRulesetUpdate = 62
	precedenceRepoFileDelete    = 80
	precedenceOrgMemberRemove   = 85
	precedenceOrgRulesetDelete  = 86
	precedenceRepoRulesetDelete = 86
	precedenceTeamDelete        = 90
	precedenceRepoDelete        = 90
	precedenceCustomRoleDelete  = 95
)

const (
	errTermSHA            = "sha"
	errTermSHANotSupplied = "wasn't supplied"
	errTermRefExists      = "reference already exists"
	errTermNameExists     = "already exists"
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
	visibility string // "", "public", "private", or "internal"
	codeowners []string
	rulesets   []config.RulesetConfig
	settings   config.RepoSettingsConfig
	files      []config.FileSpec
}

var validVisibilities = map[string]bool{
	"":          true,
	visPublic:   true,
	visPrivate:  true,
	visInternal: true,
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
		if err := config.RejectUnknownRepoKeys(m); err != nil {
			return settings, err
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
		if visibility, ok := m["visibility"].(string); ok {
			if !validVisibilities[visibility] {
				return settings, fmt.Errorf("invalid visibility %q (must be public, private, or internal)", visibility)
			}
			settings.visibility = visibility
		} else if _, has := m["visibility"]; has {
			return settings, fmt.Errorf("visibility must be a string, got %T", m["visibility"])
		}

		if raw, has := m["settings"]; has {
			parsed, err := config.ParseRepoSettings(raw)
			if err != nil {
				return settings, err
			}
			settings.settings = parsed
		}

		if raw, has := m["rulesets"]; has {
			rulesets, err := config.ParseRulesets(raw)
			if err != nil {
				return settings, err
			}
			settings.rulesets = rulesets
		}

		if raw, has := m["codeowners"]; has {
			items, ok := raw.([]any)
			if !ok {
				return settings, fmt.Errorf("codeowners must be a list, got %T", raw)
			}
			for _, item := range items {
				coStr, ok := item.(string)
				if !ok {
					return settings, fmt.Errorf("codeowners entries must be strings, got %T", item)
				}
				coStr = strings.TrimSpace(coStr)
				if err := config.ValidateCodeOwner(coStr); err != nil {
					return settings, err
				}
				settings.codeowners = append(settings.codeowners, coStr)
			}
		}

		if _, has := m["files"]; has {
			specs, err := config.RepoFileSpecs(m)
			if err != nil {
				return settings, err
			}
			settings.files = specs
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

	// Merge codeowners (union): template codeowners + repo-specific codeowners
	result.codeowners = nil
	ownerSet := make(map[string]bool)
	for _, co := range templateSettings.codeowners {
		ownerSet[co] = true
		result.codeowners = append(result.codeowners, co)
	}
	for _, co := range settings.codeowners {
		if !ownerSet[co] {
			ownerSet[co] = true
			result.codeowners = append(result.codeowners, co)
		}
	}

	// Merge rulesets by name: the template supplies the guard rails, and a
	// repo-specific ruleset of the same name replaces the template's version
	// rather than being added alongside it (GitHub would enforce both).
	result.rulesets = nil
	rulesetSeen := map[string]bool{}
	for _, rs := range settings.rulesets {
		rulesetSeen[strings.ToLower(rs.Name)] = true
		result.rulesets = append(result.rulesets, rs)
	}
	for _, rs := range templateSettings.rulesets {
		if !rulesetSeen[strings.ToLower(rs.Name)] {
			result.rulesets = append(result.rulesets, rs)
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
				Scope:  scopeTeam,
				Target: slug,
				Action: util.ActionCreate,
				Details: map[string]any{
					"org":                  st.Org,
					"name":                 want.Name,
					"privacy":              want.Privacy,
					"description":          want.Description,
					"notification_setting": want.NotificationSetting,
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
		if want.NotificationSetting != "" && want.NotificationSetting != existing.GetNotificationSetting() {
			needsUpdate = true
			updateDetails["notification_setting"] = want.NotificationSetting
		}
		if needsUpdate {
			out = append(out, util.Change{
				Scope:   scopeTeam,
				Target:  slug,
				Action:  util.ActionUpdate,
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
				Scope:   scopeTeamMember,
				Target:  slug,
				Action:  util.ActionEnsure,
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
//
// Two things are collected, because a repository entry says two different kinds
// of thing. Permission belongs to the (team, repo) pair — several teams may hold
// different access to one repository, and that is the point of teams. Everything
// else — topics, visibility, settings, rulesets, codeowners, pinning — describes
// the repository itself, so it is folded into one definition per repository and
// returned keyed by repo alone.
//
// perTeam keeps each team's own declaration so a grant can be planned from what
// that team asked for. Reading the permission out of the repo-keyed map instead
// was a privilege escalation: the map was overwritten unconditionally, so the
// last team file to name a repository decided what *every* team naming it was
// granted, and a team declaring pull could be handed admin because an unrelated
// file sorted later.
func collectRepoSettings(cfg *config.Root, _ string) (allSettings map[string]repoSettings, managedRepos map[string]bool, perTeam map[teamRepoPermKey]repoSettings, err error) {
	allSettings = map[string]repoSettings{}
	managedRepos = map[string]bool{}
	perTeam = map[teamRepoPermKey]repoSettings{}
	// declaredBy names the team that first stated a repo-level field, so a
	// conflict can say which two files disagree.
	declaredBy := map[string]string{}
	// fromReposFile marks repositories repos.yaml defines, so a team file
	// defining one too can be refused rather than merged.
	fromReposFile := map[string]bool{}

	for repo, val := range cfg.Repos {
		r := strings.ToLower(repo)
		settings, err := parseRepoConfig(val)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("invalid config for repo %s in repos.yaml: %w", repo, err)
		}
		allSettings[r] = settings
		fromReposFile[r] = true
	}

	for _, t := range cfg.Team {
		slug := t.ResolvedSlug()
		for repo, val := range t.Repositories {
			r := strings.ToLower(repo)
			managedRepos[r] = true

			settings, err := parseRepoConfig(val)
			if err != nil {
				return nil, nil, nil, fmt.Errorf("invalid config for repo %s in team %s: %w", repo, slug, err)
			}
			perTeam[slug+"/"+r] = settings

			if fromReposFile[r] && statesRepoDefinition(settings) {
				return nil, nil, nil, fmt.Errorf(
					"repo %s is defined in repos.yaml and again in team %s — "+
						"leave the permission in the team file and move the rest to repos.yaml",
					repo, slug)
			}

			existing, seen := allSettings[r]
			if !seen {
				allSettings[r] = settings
				if statesRepoDefinition(settings) {
					declaredBy[r] = slug
				}
				continue
			}
			merged, err := mergeRepoDefinition(existing, settings, repo, declaredBy[r], slug)
			if err != nil {
				return nil, nil, nil, err
			}
			allSettings[r] = merged
			if declaredBy[r] == "" && statesRepoDefinition(settings) {
				declaredBy[r] = slug
			}
		}
	}
	return allSettings, managedRepos, perTeam, nil
}

// statesRepoDefinition reports whether an entry says anything about the
// repository beyond which permission its team should hold.
func statesRepoDefinition(s repoSettings) bool {
	return len(s.topics) > 0 || s.pinned || s.template || s.from != "" ||
		s.visibility != "" || len(s.codeowners) > 0 || len(s.rulesets) > 0 ||
		!s.settings.IsEmpty()
}

// mergeRepoDefinition folds one team's declaration of a repository into what an
// earlier team declared. Set-like fields (topics, codeowners) union, flags OR,
// and scalar fields must agree — two teams may each state part of a repository's
// definition, but they may not contradict each other, because only one of the
// two can win and neither file says which.
//
// Permission is deliberately not merged: it is per (team, repo) and lives in
// collectRepoSettings' perTeam map. The copy kept here exists only so template
// inheritance can read a template repository's permission; the first team to
// state one defines it.
func mergeRepoDefinition(into, from repoSettings, repo, prevTeam, team string) (repoSettings, error) {
	conflict := func(field string, a, b any) error {
		where := prevTeam
		if where == "" {
			where = "an earlier team"
		}
		return fmt.Errorf("repo %s is defined by more than one team and they disagree on %s: "+
			"team %s says %v, team %s says %v — state it in one team only",
			repo, field, where, a, team, b)
	}

	out := into
	if out.permission == "" {
		out.permission = from.permission
	}

	if from.from != "" {
		if out.from != "" && out.from != from.from {
			return out, conflict("from", out.from, from.from)
		}
		out.from = from.from
	}
	if from.visibility != "" {
		if out.visibility != "" && out.visibility != from.visibility {
			return out, conflict("visibility", out.visibility, from.visibility)
		}
		out.visibility = from.visibility
	}
	out.pinned = out.pinned || from.pinned
	out.template = out.template || from.template
	out.topics = unionStrings(out.topics, from.topics)
	out.codeowners = unionStrings(out.codeowners, from.codeowners)

	if len(from.rulesets) > 0 {
		if len(out.rulesets) > 0 && !reflect.DeepEqual(out.rulesets, from.rulesets) {
			return out, conflict("rulesets", "one set", "another")
		}
		out.rulesets = from.rulesets
	}

	merged, err := mergeRepoSettingsConfig(out.settings, from.settings, conflict)
	if err != nil {
		return out, err
	}
	out.settings = merged
	return out, nil
}

// mergeRepoSettingsConfig merges two settings blocks field by field. A field one
// side leaves unset takes the other's value; a field both set must agree.
func mergeRepoSettingsConfig(into, from config.RepoSettingsConfig, conflict func(string, any, any) error) (config.RepoSettingsConfig, error) {
	fields := []struct {
		name string
		into **bool
		from *bool
	}{
		{"allow_auto_merge", &into.AllowAutoMerge, from.AllowAutoMerge},
		{"allow_squash_merge", &into.AllowSquashMerge, from.AllowSquashMerge},
		{"allow_merge_commit", &into.AllowMergeCommit, from.AllowMergeCommit},
		{"allow_rebase_merge", &into.AllowRebaseMerge, from.AllowRebaseMerge},
		{"delete_branch_on_merge", &into.DeleteBranchOnMerge, from.DeleteBranchOnMerge},
		{"allow_update_branch", &into.AllowUpdateBranch, from.AllowUpdateBranch},
	}
	for _, f := range fields {
		if f.from == nil {
			continue
		}
		if *f.into != nil && **f.into != *f.from {
			return into, conflict("settings."+f.name, **f.into, *f.from)
		}
		v := *f.from
		*f.into = &v
	}
	return into, nil
}

// unionStrings appends the members of b that a does not already have, keeping
// a's order. Topics and codeowners are unions across teams by design.
func unionStrings(a, b []string) []string {
	if len(b) == 0 {
		return a
	}
	seen := make(map[string]bool, len(a))
	for _, s := range a {
		seen[s] = true
	}
	out := a
	for _, s := range b {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}

// resolveTeamPerms reduces each team's own repository entry to the permission it
// asks for, applying the same template inheritance the repository-level settings
// get: an entry naming a `from:` template and stating no permission of its own
// inherits the template's, exactly as it did when the permission was read from
// the repository-keyed map.
func resolveTeamPerms(perTeam map[teamRepoPermKey]repoSettings, all map[string]repoSettings, org string) (map[teamRepoPermKey]string, error) {
	out := make(map[teamRepoPermKey]string, len(perTeam))
	for key, s := range perTeam {
		resolved, err := resolveTemplate("", s, all, org)
		if err != nil {
			return nil, fmt.Errorf("error resolving template for %s: %w", key, err)
		}
		out[key] = resolved.permission
	}
	return out, nil
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

// planTeamCleanups generates delete changes for teams not in the desired set.
func planTeamCleanups(st *State, org string, desired map[string]config.TeamConfig) ([]util.Change, error) {
	var out []util.Change
	for _, at := range st.ActualTeams {
		if _, ok := desired[at.GetSlug()]; !ok {
			out = append(out, util.Change{Scope: scopeTeam, Target: at.GetSlug(), Action: util.ActionDelete, Details: map[string]any{"org": org, "slug": at.GetSlug()}})
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
			out = append(out, util.Change{Scope: scopeOrgMember, Target: login, Action: util.ActionRemove, Details: map[string]any{"org": org, "user": login}})
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
					Scope:  scopeRepo,
					Target: repoName,
					Action: util.ActionDelete,
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

func applyChanges(ctx context.Context, c *gh.Client, changes []util.Change) error {
	return applyChangesWith(ctx, c, changes, defaultRegistry, ApplyOptions{})
}

func applyChangesWith(ctx context.Context, c *gh.Client, changes []util.Change, reg *HandlerRegistry, opts ApplyOptions) error {
	sort.SliceStable(changes, func(i, j int) bool {
		return reg.Precedence(changes[i].Scope, changes[i].Action) <
			reg.Precedence(changes[j].Scope, changes[j].Action)
	})

	// Apply custom role changes first — they have their own dispatcher. These
	// are a prerequisite for dependent changes, so a failure here always aborts
	// regardless of ContinueOnError.
	if err := applyCustomRoleChanges(ctx, c, changes); err != nil {
		return err
	}

	// Count non-custom-role changes for progress display.
	total := 0
	for _, ch := range changes {
		if !strings.HasPrefix(ch.Scope, "custom-role") {
			total++
		}
	}

	var failed []error
	applied := 0
	for _, ch := range changes {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if strings.HasPrefix(ch.Scope, "custom-role") {
			continue
		}

		applied++
		util.Infof("[%d/%d] %s:%s %s", applied, total, ch.Scope, ch.Action, ch.Target)

		if err := gh.RespectRate(ctx, c.REST); err != nil {
			util.Warnf("rate limit check failed: %v", err)
		}

		handler, ok := reg.Lookup(ch.Scope, ch.Action)
		if !ok {
			util.Warnf("no handler for change %s:%s on %s", ch.Scope, ch.Action, ch.Target)
			continue
		}
		if err := handler.Apply(ctx, c, ch); err != nil {
			util.Audit(ch.Scope, ch.Target, ch.Action, "error")
			if !opts.ContinueOnError {
				return err
			}
			wrapped := fmt.Errorf("%s:%s %s: %w", ch.Scope, ch.Action, ch.Target, err)
			util.Warnf("continuing after error: %v", wrapped)
			failed = append(failed, wrapped)
			continue
		}
		util.Audit(ch.Scope, ch.Target, ch.Action, "ok")
	}

	if len(failed) > 0 {
		util.Warnf("%d of %d changes failed", len(failed), total)
		return fmt.Errorf("apply completed with %d error(s): %w", len(failed), errors.Join(failed...))
	}
	return nil
}
