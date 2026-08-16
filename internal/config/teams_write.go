package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// adoptedTeamComment marks team files gomgr generated.
const adoptedTeamComment = "# Adopted from GitHub by `gomgr import teams`."

// WriteTeamFile writes an imported team to <dir>/teams/<slug>.yaml and returns
// the path it wrote.
//
// Unlike the ruleset import, which has to splice into files somebody else
// wrote, every team file here is new — so this is a plain encode, and the only
// thing worth guarding is not landing on top of a file that already exists.
func WriteTeamFile(dir string, team TeamConfig) (string, error) {
	teamDir := filepath.Join(dir, "teams")
	if err := os.MkdirAll(teamDir, 0o755); err != nil {
		return "", fmt.Errorf("create teams directory: %w", err)
	}

	path := filepath.Join(teamDir, team.ResolvedSlug()+".yaml")
	if _, err := os.Stat(path); err == nil {
		return "", fmt.Errorf("%s already exists; move it aside to re-import this team", path)
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("check %s: %w", path, err)
	}

	body, err := encodeTeam(team)
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(path, body, 0o644); err != nil {
		return "", fmt.Errorf("write %s: %w", path, err)
	}
	return path, nil
}

func encodeTeam(team TeamConfig) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString(adoptedTeamComment + "\n")

	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(team); err != nil {
		return nil, fmt.Errorf("encode team %q: %w", team.Name, err)
	}
	if err := enc.Close(); err != nil {
		return nil, fmt.Errorf("encode team %q: %w", team.Name, err)
	}
	return buf.Bytes(), nil
}
