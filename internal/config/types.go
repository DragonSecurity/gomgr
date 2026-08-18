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
	} `yaml:"dry_warnings"`
	RemoveMembersWithoutTeam   bool `yaml:"remove_members_without_team"`
	DeleteUnconfiguredTeams    bool `yaml:"delete_unconfigured_teams"`
	DeleteUnmanagedRepos       bool `yaml:"delete_unmanaged_repos"`
	DeleteUnmanagedCustomRoles bool `yaml:"delete_unmanaged_custom_roles"`
	DeleteUnmanagedRulesets    bool `yaml:"delete_unmanaged_rulesets"`
	DeleteStaleCodeowners      bool `yaml:"delete_stale_codeowners"`
	CreateRepo                 bool `yaml:"create_repo"`

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
	Permission string             `yaml:"permission,omitempty"` // pull|triage|push|maintain|admin
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

type TeamConfig struct {
	Name        string   `yaml:"name"`
	Slug        string   `yaml:"slug,omitempty"`
	Description string   `yaml:"description,omitempty"`
	Privacy     string   `yaml:"privacy,omitempty"` // closed, secret
	Parents     []string `yaml:"parents,omitempty"`

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
	return strings.ToLower(strings.ReplaceAll(t.Name, " ", "-"))
}
