package config

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// OrgDir is a configuration directory found on disk and the organization it
// names.
type OrgDir struct {
	// Dir is the path to the directory holding app.yaml.
	Dir string
	// Org is the `org:` key from that app.yaml, lowercased. Empty when the
	// file parses but names no organization.
	Org string
}

// DiscoverOrgDirs walks root and returns every directory containing an
// app.yaml, paired with the organization that app.yaml names.
//
// It reads only the `org:` key. A full config.Load would refuse a directory
// with an unrelated problem — a bad ruleset, a team file with a typo — and the
// question being asked here is which organizations are configured at all, which
// a broken team file does not change the answer to. Reporting those as missing
// would be worse than reporting them as present and broken.
//
// The walk does not descend into a directory that has an app.yaml. One config
// directory does not contain another, and teams/ is full of YAML that would
// otherwise be probed for no reason.
func DiscoverOrgDirs(root string) ([]OrgDir, error) {
	info, err := os.Stat(root)
	if err != nil {
		return nil, fmt.Errorf("read config root %s: %w", root, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("config root %s is not a directory", root)
	}

	var out []OrgDir
	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}
		// Hidden directories are .git, .github and friends — never a config
		// directory, and expensive to walk.
		if name := d.Name(); path != root && strings.HasPrefix(name, ".") {
			return filepath.SkipDir
		}
		appPath := filepath.Join(path, "app.yaml")
		if _, statErr := os.Stat(appPath); statErr != nil {
			return nil
		}
		org, readErr := readOrgKey(appPath)
		if readErr != nil {
			return readErr
		}
		out = append(out, OrgDir{Dir: path, Org: org})
		return filepath.SkipDir
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Dir < out[j].Dir })
	return out, nil
}

// readOrgKey pulls just the org name out of an app.yaml.
func readOrgKey(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read %s: %w", path, err)
	}
	var doc struct {
		Org string `yaml:"org"`
	}
	if err := yaml.Unmarshal(b, &doc); err != nil {
		return "", fmt.Errorf("parse %s: %w", path, err)
	}
	return strings.ToLower(strings.TrimSpace(doc.Org)), nil
}
