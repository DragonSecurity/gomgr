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
	RemoveMembersWithoutTeam   bool   `yaml:"remove_members_without_team"`
	DeleteUnconfiguredTeams    bool   `yaml:"delete_unconfigured_teams"`
	DeleteUnmanagedRepos       bool   `yaml:"delete_unmanaged_repos"`
	DeleteUnmanagedCustomRoles bool   `yaml:"delete_unmanaged_custom_roles"`
	CreateRepo                 bool   `yaml:"create_repo"`
	AddRenovateConfig          bool   `yaml:"add_renovate_config"`
	RenovateConfig             string `yaml:"renovate_config"`
	AddDefaultReadme           bool   `yaml:"add_default_readme"`
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
