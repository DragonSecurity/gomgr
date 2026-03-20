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
