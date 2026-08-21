package config

import (
	"os"
	"strconv"
	"strings"
)

type AppConfig struct {
	AppID      int64  `yaml:"app_id,omitempty"`
	PrivateKey string `yaml:"private_key,omitempty"`
	Org        string `yaml:"org"`

	DryWarnings struct {
		WarnUnmanagedTeams        bool `yaml:"warn_unmanaged_teams"`
		WarnMembersWithoutAnyTeam bool `yaml:"warn_members_without_any_team"`
		WarnUnmanagedRepos        bool `yaml:"warn_unmanaged_repos"`
		WarnUnmanagedCustomRoles  bool `yaml:"warn_unmanaged_custom_roles"`
		WarnUnmanagedRulesets     bool `yaml:"warn_unmanaged_rulesets"`
		WarnUnmanagedOwners       bool `yaml:"warn_unmanaged_owners"`
		WarnExcessCollaborators   bool `yaml:"warn_excess_collaborators"`
	} `yaml:"dry_warnings"`
	RemoveMembersWithoutTeam   bool `yaml:"remove_members_without_team"`
	DeleteUnconfiguredTeams    bool `yaml:"delete_unconfigured_teams"`
	DeleteUnmanagedRepos       bool `yaml:"delete_unmanaged_repos"`
	DeleteUnmanagedCustomRoles bool `yaml:"delete_unmanaged_custom_roles"`
	DeleteUnmanagedRulesets    bool `yaml:"delete_unmanaged_rulesets"`
	DeleteStaleCodeowners      bool `yaml:"delete_stale_codeowners"`
	CreateRepo                 bool `yaml:"create_repo"`

	// DemoteUnconfiguredOwners drops an organization owner that org.yaml does
	// not list back to plain member. Off by default, and inert unless
	// org.owners names at least one login: an empty list means "gomgr does not
	// manage owners here", never "this organization should have none". The
	// authenticated user is never demoted by their own run, because an
	// organization whose last owner demotes themselves cannot be repaired
	// through the API that did it.
	//
	// A demoted owner becomes an ordinary member, which puts them in reach of
	// RemoveMembersWithoutTeam on the same run if they belong to no team.
	DemoteUnconfiguredOwners bool `yaml:"demote_unconfigured_owners"`

	// RemoveExcessCollaborators revokes a direct repository grant that gives
	// someone more access than their team membership and the organization's
	// default repository permission already do.
	//
	// This is the one thing that closes the door the GitHub UI opens: a
	// repository's "add collaborator" button grants access that no YAML file
	// mentions, and no other part of gomgr looks at it. Revoking a direct
	// grant leaves team-derived access untouched, so someone who should have
	// push through a team keeps push; only the grant stapled on beside it goes.
	//
	// Off by default. With DryWarnings.WarnExcessCollaborators set instead,
	// the same grants are reported and left alone.
	RemoveExcessCollaborators bool `yaml:"remove_excess_collaborators"`

	// ArchiveUnmanagedRepos archives a repository no team names, instead of
	// leaving it alone.
	//
	// This is the reversible sibling of DeleteUnmanagedRepos, and the one to
	// reach for first. An archived repository can be un-archived by anyone with
	// admin on it; a deleted one is ninety days of GitHub support at best and
	// gone at worst. A configuration that has drifted — a repository created
	// last week that nobody has declared yet — costs an un-archive rather than
	// a restore request.
	//
	// If both this and DeleteUnmanagedRepos are set, this one wins and the
	// deletion is reported as skipped. Being wrong in the recoverable direction
	// is the point of having it.
	ArchiveUnmanagedRepos bool `yaml:"archive_unmanaged_repos"`

	// IgnoreArchived skips archived repositories when planning team permission
	// grants. GitHub refuses to change permissions on an archived repository,
	// so a configuration that still names one plans a grant that fails on every
	// run. Off by default, because silently skipping a repository is also a way
	// to not notice that it was archived out from under the config.
	IgnoreArchived bool `yaml:"ignore_archived"`

	// ReconcileVisibility allows gomgr to change the visibility of a repository
	// that already exists, and only one that explicitly declares a visibility.
	//
	// Off by default, and deliberately a second key rather than something an
	// org-wide default can trigger. Making a private repository public is the
	// most damaging thing this tool can do — a deleted repository can be
	// restored for ninety days, a disclosed one cannot be undisclosed — so it
	// takes a deliberate edit in two files, and can never happen to thirty-four
	// repositories because one line changed.
	ReconcileVisibility bool `yaml:"reconcile_visibility"`

	// SignOff is the identity used for the Signed-off-by trailer appended to
	// every commit gomgr writes, in "Name <email>" form. Set it when the org
	// enforces DCO — a ruleset with a commit_message_pattern rule requiring
	// "Signed-off-by:" rejects gomgr's file-sync pushes otherwise, because
	// those commits go straight to the default branch and never pass through
	// a pull request where a DCO check could run.
	//
	// Empty (the default) appends nothing, preserving prior behavior for orgs
	// that do not require sign-off.
	SignOff string `yaml:"sign_off,omitempty"`

	// FileChanges controls how gomgr lands the files it writes.
	FileChanges FileChangeConfig `yaml:"file_changes,omitempty"`

	// Files declares templated files that should exist in every managed
	// repository. Each entry's Content is rendered through text/template with
	// {Org, Repo} context. Only (optional) limits which repos an entry applies
	// to via path.Match-style globs.
	Files []FileSpec `yaml:"files,omitempty"`

	// Legacy convenience flags. These are still honored but are materialized
	// into Files entries at load time. Prefer Files for new configurations.
	AddRenovateConfig bool   `yaml:"add_renovate_config,omitempty"`
	RenovateConfig    string `yaml:"renovate_config,omitempty"`
	AddDefaultReadme  bool   `yaml:"add_default_readme,omitempty"`
}

// FileSpec declares a file that gomgr should ensure exists in managed repos.
// Content is a Go text/template; Path, Message and Branch are literal strings.
// Only restricts which repositories the file applies to (path.Match globs
// against the repo name). An empty Only matches every managed repo.
//
// Reconcile controls drift handling for files that already exist. When false
// (the default) gomgr only writes the file when it is missing, leaving any
// hand-edited content alone. When true gomgr compares the rendered content to
// what is on the default branch and pushes an update commit if they differ —
// useful for config-derived files like CODEOWNERS that should track the YAML.
type FileSpec struct {
	Path      string   `yaml:"path"`
	Content   string   `yaml:"content"`
	Message   string   `yaml:"message,omitempty"`
	Branch    string   `yaml:"branch,omitempty"`
	Only      []string `yaml:"only,omitempty"`
	Reconcile bool     `yaml:"reconcile,omitempty"`
}

type OrgConfig struct {
	Owners      []string           `yaml:"owners"`
	CustomRoles []CustomRoleConfig `yaml:"custom_roles,omitempty"`

	// RepoDefaults are the repository settings every managed repository gets
	// unless its own entry in teams/*.yaml overrides them. Visibility is not
	// among them on purpose: see AppConfig.ReconcileVisibility.
	RepoDefaults RepoSettingsConfig `yaml:"repo_defaults,omitempty"`

	// Rulesets declares organization-wide rulesets — the guard rails that
	// apply across repositories, narrowed by their repository_name conditions.
	// Repository-specific rulesets live on the repository entry in teams/*.yaml
	// and stack on top of these; GitHub evaluates every ruleset that matches and
	// enforces the strictest outcome.
	Rulesets []RulesetConfig `yaml:"rulesets,omitempty"`
}

// CustomRoleConfig defines a custom repository role for the organization
// Requires GitHub Enterprise Cloud
type CustomRoleConfig struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description,omitempty"`
	BaseRole    string   `yaml:"base_role"` // read, triage, write, maintain, admin
	Permissions []string `yaml:"permissions,omitempty"`
}

type RepoConfig struct {
	Permission string `yaml:"permission,omitempty"` // pull|triage|push|maintain|admin

	// Archived declares whether this repository should be archived. It is a
	// pointer because omitting it and setting it to false are different
	// instructions: an absent key leaves GitHub's state alone, so a repository
	// somebody archived by hand stays archived until a configuration says
	// otherwise in as many words. Un-archiving never happens by omission.
	Archived   *bool              `yaml:"archived,omitempty"`
	Topics     []string           `yaml:"topics,omitempty"`
	Pinned     bool               `yaml:"pinned,omitempty"`
	Rulesets   []RulesetConfig    `yaml:"rulesets,omitempty"`
	Settings   RepoSettingsConfig `yaml:"settings,omitempty"`
	Visibility string             `yaml:"visibility,omitempty"` // public|private|internal

	// Files are this repository's own templated files. An entry here replaces
	// the app.files entry with the same path for this repository alone, so an
	// exception is stated by naming the repository rather than by relying on
	// where it sits in a list.
	Files []FileSpec `yaml:"files,omitempty"`
}

// The two values GitHub accepts for a team's privacy. A secret team cannot
// take part in a hierarchy at either end, which is why the name is needed
// somewhere both the loader and the team validation can see it.
const (
	privacyClosed = "closed"
	privacySecret = "secret"
)

// The two values GitHub accepts for a team's notification setting. They are
// compared against what the API reports and sent back to it, so they are
// constants rather than literals repeated at each end.
const (
	NotificationsEnabled  = "notifications_enabled"
	NotificationsDisabled = "notifications_disabled"
)

type TeamConfig struct {
	Name        string `yaml:"name"`
	Slug        string `yaml:"slug,omitempty"`
	Description string `yaml:"description,omitempty"`
	Privacy     string `yaml:"privacy,omitempty"` // closed, secret

	// Parents nests this team under another. GitHub allows exactly one parent,
	// so at most one entry is accepted; the field stays a list because it
	// always was one, and a config that wrote `parents: []` should keep
	// loading. The entry may be a team name or a slug — it is resolved to a
	// slug against the other teams in the config.
	//
	// Nesting grants access rather than membership: a child team inherits the
	// repository permissions of every team above it, while the members of the
	// two teams stay separate rosters. gomgr sets the relationship and lets
	// GitHub apply the inheritance; it does not write the parent's grants onto
	// the child.
	//
	// Neither a parent nor a child may be `secret` — GitHub rejects the
	// combination — which loader validation reports rather than discovering at
	// apply time.
	Parents []string `yaml:"parents,omitempty"`

	// NotificationSetting is "notifications_enabled" or
	// "notifications_disabled". Empty leaves GitHub's setting alone, which for
	// a team gomgr creates means enabled — so an org that wants its teams quiet
	// has to say so on every team, and a config that drops this field turns
	// notifications back on for everyone in it.
	NotificationSetting string `yaml:"notification_setting,omitempty"`

	Maintainers []string `yaml:"maintainers,omitempty"`
	Members     []string `yaml:"members,omitempty"`

	// repo => permission (pull|triage|push|maintain|admin) or RepoConfig for advanced settings
	// For backward compatibility, supports both:
	//   repositories:
	//     infra: maintain               # simple string permission
	//     api:                          # or advanced RepoConfig
	//       permission: push
	//       topics: [backend, api]
	//       pinned: true
	Repositories map[string]any `yaml:"repositories,omitempty"`
}

type Root struct {
	App  AppConfig    `yaml:"app"`
	Org  OrgConfig    `yaml:"org"`
	Team []TeamConfig `yaml:"teams"`

	// Repos are repository definitions from repos.yaml: everything about a
	// repository that is not some team's permission on it. Values take the same
	// shape as an advanced entry under a team's `repositories:`, minus
	// `permission`, which belongs to the (team, repo) pair rather than to the
	// repository.
	//
	// A repository may be defined here or in a team file, but not in both — see
	// ReposFile.
	Repos map[string]any
}

// ReposFile is repos.yaml: repository definitions keyed by repository name.
//
// Splitting these out of teams/*.yaml separates the two things a team entry was
// saying at once. A team file says which teams hold what access; repos.yaml says
// what the repository is. Keeping both in one place meant a repository named by
// two teams had two definitions and no rule for which won, and the answer turned
// out to be "whichever file sorted last", including for the permission granted.
//
// Definitions in team files are still honored so a configuration can move across
// one repository at a time, but a repository defined in both places is refused
// rather than resolved by precedence, because a precedence rule is what this is
// getting rid of.
type ReposFile struct {
	Repos map[string]any `yaml:"repos"`
}

// ResolvedAppID returns the GitHub App ID in effect, falling back to
// GITHUB_APP_ID when app.yaml does not set one. It is 0 under PAT auth, where
// there is no app.
func (a AppConfig) ResolvedAppID() int64 {
	if a.AppID != 0 {
		return a.AppID
	}
	if v := os.Getenv("GITHUB_APP_ID"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			return id
		}
	}
	return 0
}

// ResolvedSlug returns the team's slug, deriving it from the name if not explicitly set.
func (t TeamConfig) ResolvedSlug() string {
	if t.Slug != "" {
		return t.Slug
	}
	return Slugify(t.Name)
}

// Slugify turns a team name into the slug GitHub would derive from it.
func Slugify(name string) string {
	return strings.ToLower(strings.ReplaceAll(name, " ", "-"))
}

// ParentSlug returns the slug of the team this one is nested under, or "" when
// it is not nested. Validation has already rejected more than one entry, so
// only the first is consulted.
func (t TeamConfig) ParentSlug() string {
	for _, p := range t.Parents {
		if p = strings.TrimSpace(p); p != "" {
			return Slugify(p)
		}
	}
	return ""
}
