package config

type AppConfig struct {
	AppID      int64  `yaml:"app_id,omitempty"`
	PrivateKey string `yaml:"private_key,omitempty"`
	Org        string `yaml:"org"`

	DryWarnings struct {
		WarnUnmanagedTeams        bool `yaml:"warn_unmanaged_teams"`
		WarnMembersWithoutAnyTeam bool `yaml:"warn_members_without_any_team"`
	} `yaml:"dry_warnings"`
	RemoveMembersWithoutTeam bool   `yaml:"remove_members_without_team"`
	DeleteUnconfiguredTeams  bool   `yaml:"delete_unconfigured_teams"`
	CreateRepo               bool   `yaml:"create_repo"`
	AddRenovateConfig        bool   `yaml:"add_renovate_config"`
	RenovateConfig           string `yaml:"renovate_config"`
	AddDefaultReadme         bool   `yaml:"add_default_readme"`
}

type OrgConfig struct {
	Owners []string `yaml:"owners"`
}

type TeamConfig struct {
	Name        string   `yaml:"name"`
	Slug        string   `yaml:"slug,omitempty"`
	Description string   `yaml:"description,omitempty"`
	Privacy     string   `yaml:"privacy,omitempty"` // closed, secret
	Parents     []string `yaml:"parents,omitempty"`

	Maintainers []string `yaml:"maintainers,omitempty"`
	Members     []string `yaml:"members,omitempty"`

	// repo => permission (pull|triage|push|maintain|admin)
	Repositories map[string]string `yaml:"repositories,omitempty"`
}

type Root struct {
	App  AppConfig    `yaml:"app"`
	Org  OrgConfig    `yaml:"org"`
	Team []TeamConfig `yaml:"teams"`
}
