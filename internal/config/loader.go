package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

func Load(dir string) (*Root, error) {
	r := &Root{}
	// app.yaml
	if err := readYAML(filepath.Join(dir, "app.yaml"), &r.App); err != nil {
		return nil, err
	}
	// org.yaml
	if err := readYAML(filepath.Join(dir, "org.yaml"), &r.Org); err != nil {
		return nil, err
	}
	// teams/*.yaml
	teamDir := filepath.Join(dir, "teams")
	entries, err := os.ReadDir(teamDir)
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("read teams directory %s: %w", teamDir, err)
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".yaml") && !strings.HasSuffix(name, ".yml") {
			continue // ignore non-YAML files like .DS_Store, README, etc.
		}
		var t TeamConfig
		if err := readYAML(filepath.Join(teamDir, e.Name()), &t); err != nil {
			return nil, err
		}
		r.Team = append(r.Team, t)
	}
	if r.App.Org == "" {
		return nil, errors.New("app.org is required")
	}
	if err := r.Validate(); err != nil {
		return nil, err
	}
	return r, nil
}

func readYAML(path string, out any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read config file %s: %w", path, err)
	}
	if err := yaml.Unmarshal(b, out); err != nil {
		return fmt.Errorf("parse YAML in %s: %w", path, err)
	}
	return nil
}

// Validate checks that the loaded configuration is semantically correct.
func (r *Root) Validate() error {
	validPrivacy := map[string]bool{"": true, "closed": true, "secret": true}
	validBaseRole := map[string]bool{"read": true, "triage": true, "write": true, "maintain": true, "admin": true}

	for _, t := range r.Team {
		if t.Name == "" {
			return fmt.Errorf("team name must not be empty")
		}
		if !validPrivacy[t.Privacy] {
			return fmt.Errorf("team %q has invalid privacy %q (must be closed or secret)", t.Name, t.Privacy)
		}
		for repo := range t.Repositories {
			if err := validateRepoName(repo); err != nil {
				return fmt.Errorf("team %q: %w", t.Name, err)
			}
		}
		for _, u := range t.Maintainers {
			if err := validateUsername(u); err != nil {
				return fmt.Errorf("team %q maintainer: %w", t.Name, err)
			}
		}
		for _, u := range t.Members {
			if err := validateUsername(u); err != nil {
				return fmt.Errorf("team %q member: %w", t.Name, err)
			}
		}
	}
	for _, cr := range r.Org.CustomRoles {
		if cr.Name == "" {
			return fmt.Errorf("custom role name must not be empty")
		}
		if !validBaseRole[cr.BaseRole] {
			return fmt.Errorf("custom role %q has invalid base_role %q (must be read|triage|write|maintain|admin)", cr.Name, cr.BaseRole)
		}
	}
	for _, u := range r.Org.Owners {
		if err := validateUsername(u); err != nil {
			return fmt.Errorf("org owner: %w", err)
		}
	}
	if err := validateFileSpecs(r.App.Files); err != nil {
		return err
	}
	return nil
}

// validateFileSpecs ensures every templated file has a path and content and
// that no two specs target the same path (which would make apply order
// ambiguous).
func validateFileSpecs(files []FileSpec) error {
	seen := map[string]bool{}
	for i, f := range files {
		if strings.TrimSpace(f.Path) == "" {
			return fmt.Errorf("app.files[%d]: path must not be empty", i)
		}
		if strings.TrimSpace(f.Content) == "" {
			return fmt.Errorf("app.files[%d] (%s): content must not be empty", i, f.Path)
		}
		if seen[f.Path] {
			return fmt.Errorf("app.files[%d]: duplicate path %q", i, f.Path)
		}
		seen[f.Path] = true
	}
	return nil
}

var validRepoName = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

func validateRepoName(name string) error {
	if len(name) == 0 || len(name) > 100 {
		return fmt.Errorf("repo name must be 1-100 characters: %q", name)
	}
	if !validRepoName.MatchString(name) {
		return fmt.Errorf("repo name contains invalid characters: %q", name)
	}
	if name == "." || name == ".." {
		return fmt.Errorf("repo name cannot be %q", name)
	}
	return nil
}

var validUsername = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?$`)

func validateUsername(name string) error {
	if len(name) == 0 || len(name) > 39 {
		return fmt.Errorf("username must be 1-39 characters: %q", name)
	}
	if !validUsername.MatchString(name) {
		return fmt.Errorf("username contains invalid characters: %q", name)
	}
	if strings.Contains(name, "--") {
		return fmt.Errorf("username cannot contain consecutive hyphens: %q", name)
	}
	return nil
}

var validTeamSlug = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

// ValidateCodeOwner checks that a CODEOWNERS entry is well-formed. Accepts
// bare usernames (octocat), explicit user refs (@octocat), and team refs
// (@org/team-slug). Emails are not yet supported.
func ValidateCodeOwner(co string) error {
	if co == "" {
		return fmt.Errorf("codeowner must not be empty")
	}
	if strings.ContainsAny(co, " \t\n\r") {
		return fmt.Errorf("codeowner contains whitespace: %q", co)
	}
	name := strings.TrimPrefix(co, "@")
	if strings.Contains(name, "/") {
		parts := strings.SplitN(name, "/", 2)
		if err := validateUsername(parts[0]); err != nil {
			return fmt.Errorf("codeowner %q: org segment: %w", co, err)
		}
		if parts[1] == "" || len(parts[1]) > 100 {
			return fmt.Errorf("codeowner %q: team slug must be 1-100 characters", co)
		}
		if !validTeamSlug.MatchString(parts[1]) {
			return fmt.Errorf("codeowner %q: team slug contains invalid characters", co)
		}
		return nil
	}
	if err := validateUsername(name); err != nil {
		return fmt.Errorf("codeowner %q: %w", co, err)
	}
	return nil
}

func BootstrapTeamYAML(path string, name string) error {
	t := TeamConfig{
		Name:         name,
		Maintainers:  []string{},
		Members:      []string{},
		Repositories: map[string]any{},
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	b, err := yaml.Marshal(t)
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}
