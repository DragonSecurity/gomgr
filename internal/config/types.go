package config

import "strings"

type AppConfig struct {
	AppID      int64  `yaml:"app_id,omitempty"`
	PrivateKey string `yaml:"private_key,omitempty"`
	Org        string `yaml:"org"`

	DryWarnings struct {
		WarnUnmanagedTeams        bool `yaml:"warn_unmanaged_teams"`
		WarnMembersWithoutAnyTeam bool `yaml:"warn_members_without_any_team"`
		WarnUnmanagedRepos        bool `yaml:"warn_unmanaged_repos"`
		WarnUnmanagedCustomRoles  bool `yaml:"warn_unmanaged_custom_roles"`
	} `yaml:"dry_warnings"`
	RemoveMembersWithoutTeam   bool `yaml:"remove_members_without_team"`
	DeleteUnconfiguredTeams    bool `yaml:"delete_unconfigured_teams"`
	DeleteUnmanagedRepos       bool `yaml:"delete_unmanaged_repos"`
	DeleteUnmanagedCustomRoles bool `yaml:"delete_unmanaged_custom_roles"`
	CreateRepo                 bool `yaml:"create_repo"`

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
type FileSpec struct {
	Path    string   `yaml:"path"`
	Content string   `yaml:"content"`
	Message string   `yaml:"message,omitempty"`
	Branch  string   `yaml:"branch,omitempty"`
	Only    []string `yaml:"only,omitempty"`
}

type OrgConfig struct {
	Owners      []string           `yaml:"owners"`
	CustomRoles []CustomRoleConfig `yaml:"custom_roles,omitempty"`
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
	Permission string   `yaml:"permission,omitempty"` // pull|triage|push|maintain|admin
	Topics     []string `yaml:"topics,omitempty"`
	Pinned     bool     `yaml:"pinned,omitempty"`
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
}

// ResolvedSlug returns the team's slug, deriving it from the name if not explicitly set.
func (t TeamConfig) ResolvedSlug() string {
	if t.Slug != "" {
		return t.Slug
	}
	return strings.ToLower(strings.ReplaceAll(t.Name, " ", "-"))
}
