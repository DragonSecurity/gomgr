package config

import (
	"errors"
	"os"
	"path/filepath"

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
	entries, _ := os.ReadDir(teamDir)
	for _, e := range entries {
		if e.IsDir() {
			continue
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
	return r, nil
}

func readYAML(path string, out any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(b, out)
}

func BootstrapTeamYAML(path string, name string) error {
	t := TeamConfig{
		Name:         name,
		Maintainers:  []string{},
		Members:      []string{},
		Repositories: map[string]string{},
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
