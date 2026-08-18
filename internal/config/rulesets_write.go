package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// adoptedComment marks blocks gomgr wrote, so a reader can tell an adopted
// ruleset from a hand-written one.
const adoptedComment = "# Adopted from GitHub by `gomgr import rulesets`."

// InsertOrgRulesets adds ruleset entries to the top-level `rulesets:` block of
// an org.yaml, creating the block when the file has none.
//
// The file is edited as text rather than re-encoded from a parsed tree. yaml.v3
// can round-trip a document, but it normalizes as it goes — blank lines between
// entries disappear, quoting style shifts, long values re-wrap — and a config
// repository's diff is something people have to read. Parsing is used only to
// find the line to splice at, which is exactly what a parser is good for and a
// regex is not.
func InsertOrgRulesets(path string, rulesets []RulesetConfig) error {
	if len(rulesets) == 0 {
		return nil
	}
	doc, err := loadYAMLLines(path)
	if err != nil {
		return err
	}

	_, value := doc.field(doc.root, "rulesets")
	if value == nil {
		return doc.appendBlock(rulesets)
	}
	return doc.insertIntoSequence(value, 0, rulesets)
}

// InsertRepoRulesets adds ruleset entries to a repository's entry, handling the
// three shapes a repository entry can take: one that already has a `rulesets:`
// block, one that is a settings map without it, and one that is the bare
// permission string (`infra: push`), which has to grow into a map first.
//
// container is the key the entries sit under: "repos" in repos.yaml, or
// "repositories" in a team file.
func InsertRepoRulesets(path, container, repo string, rulesets []RulesetConfig) error {
	if len(rulesets) == 0 {
		return nil
	}
	doc, err := loadYAMLLines(path)
	if err != nil {
		return err
	}

	_, repos := doc.field(doc.root, container)
	if repos == nil {
		return fmt.Errorf("%s: no `%s:` block to add %q to", path, container, repo)
	}
	repoKey, repoValue := doc.field(repos, repo)
	if repoKey == nil {
		return fmt.Errorf("%s: repository %q is not declared here", path, repo)
	}

	keyIndent := repoKey.Column - 1

	// `infra: push` — rewrite the single line as a map before adding to it.
	if repoValue != nil && repoValue.Kind == yaml.ScalarNode && repoValue.Value != "" {
		body := []string{
			strings.Repeat(" ", keyIndent) + repo + ":",
			strings.Repeat(" ", keyIndent+2) + "permission: " + repoValue.Value,
		}
		block, err := renderRulesetBlock(rulesets, keyIndent+2)
		if err != nil {
			return err
		}
		doc.lines = replaceLine(doc.lines, repoKey.Line-1, append(body, block...))
		return doc.save()
	}

	if repoValue == nil || repoValue.Kind != yaml.MappingNode {
		return fmt.Errorf("%s: repository %q has an unexpected shape; add a `rulesets:` block by hand", path, repo)
	}

	if _, existing := doc.field(repoValue, "rulesets"); existing != nil {
		return doc.insertIntoSequence(existing, keyIndent+2, rulesets)
	}

	// A settings map with no rulesets yet: add the key at the end of the map.
	block, err := renderRulesetBlock(rulesets, keyIndent+2)
	if err != nil {
		return err
	}
	doc.lines = insertAfter(doc.lines, doc.blockEnd(repoValue, keyIndent+2), block)
	return doc.save()
}

// FindRepoDefinitionFile returns the file a repository's definition should be
// written to, and the key its entries sit under.
//
// repos.yaml is preferred when it declares the repository, so an imported
// ruleset lands with the rest of that repository's definition. A configuration
// still keeping its definitions in team files gets the team file, because
// writing to repos.yaml instead would define the repository in two places at
// once, which the loader refuses.
func FindRepoDefinitionFile(dir, repo string) (path, container string, err error) {
	reposPath := filepath.Join(dir, "repos.yaml")
	var rf ReposFile
	switch _, statErr := os.Stat(reposPath); {
	case statErr == nil:
		if err := readYAML(reposPath, &rf); err != nil {
			return "", "", err
		}
		for declared := range rf.Repos {
			if strings.EqualFold(declared, repo) {
				return reposPath, "repos", nil
			}
		}
	case !os.IsNotExist(statErr):
		return "", "", fmt.Errorf("read config file %s: %w", reposPath, statErr)
	}

	teamPath, err := FindTeamFileForRepo(dir, repo)
	if err != nil {
		return "", "", err
	}
	if teamPath == "" {
		return "", "", nil
	}
	return teamPath, "repositories", nil
}

// FindTeamFileForRepo returns the team file under <dir>/teams that declares
// repo, or "" when no file does. When several declare it the first by filename
// wins, which keeps the choice stable across runs.
func FindTeamFileForRepo(dir, repo string) (string, error) {
	teamDir := filepath.Join(dir, "teams")
	entries, err := os.ReadDir(teamDir)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("read teams directory %s: %w", teamDir, err)
	}

	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.HasSuffix(e.Name(), ".yaml") || strings.HasSuffix(e.Name(), ".yml") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)

	for _, name := range names {
		path := filepath.Join(teamDir, name)
		var team TeamConfig
		if err := readYAML(path, &team); err != nil {
			return "", err
		}
		for declared := range team.Repositories {
			if strings.EqualFold(declared, repo) {
				return path, nil
			}
		}
	}
	return "", nil
}

// yamlDoc is a file held as both text (what gets written back) and a parsed
// tree (what says where to write).
type yamlDoc struct {
	path    string
	lines   []string
	newline bool // the original file ended with a newline
	root    *yaml.Node
}

func loadYAMLLines(path string) (*yamlDoc, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	text := string(raw)
	doc := &yamlDoc{path: path, newline: strings.HasSuffix(text, "\n")}
	doc.lines = strings.Split(strings.TrimSuffix(text, "\n"), "\n")

	var node yaml.Node
	if err := yaml.Unmarshal(raw, &node); err != nil {
		return nil, fmt.Errorf("parse YAML in %s: %w", path, err)
	}
	if len(node.Content) == 0 || node.Content[0].Kind != yaml.MappingNode {
		return nil, fmt.Errorf("%s: expected a YAML mapping at the top level", path)
	}
	doc.root = node.Content[0]
	return doc, nil
}

func (d *yamlDoc) save() error {
	out := strings.Join(d.lines, "\n")
	if d.newline {
		out += "\n"
	}
	info, err := os.Stat(d.path)
	mode := os.FileMode(0o644)
	if err == nil {
		mode = info.Mode().Perm()
	}
	return os.WriteFile(d.path, []byte(out), mode)
}

// field looks up a key in a mapping node, matching case-insensitively so a
// repository written as `Infra` is still found by `infra`.
func (d *yamlDoc) field(mapping *yaml.Node, key string) (keyNode, valueNode *yaml.Node) {
	if mapping == nil || mapping.Kind != yaml.MappingNode {
		return nil, nil
	}
	for i := 0; i+1 < len(mapping.Content); i += 2 {
		if strings.EqualFold(mapping.Content[i].Value, key) {
			return mapping.Content[i], mapping.Content[i+1]
		}
	}
	return nil, nil
}

// insertIntoSequence appends entries to an existing `rulesets:` sequence,
// matching the indentation the file already uses for its items.
func (d *yamlDoc) insertIntoSequence(seq *yaml.Node, keyIndent int, rulesets []RulesetConfig) error {
	itemIndent := keyIndent + 2
	if seq.Kind == yaml.SequenceNode && len(seq.Content) > 0 {
		itemIndent = seq.Content[0].Column - 1
		// A sequence item's Column points at the first key after the dash, so
		// step back over "- " to find where the dash itself sits.
		if itemIndent >= 2 {
			itemIndent -= 2
		}
	}

	items, err := renderRulesets(rulesets, itemIndent)
	if err != nil {
		return err
	}
	d.lines = insertAfter(d.lines, d.blockEnd(seq, itemIndent), items)
	return d.save()
}

// appendBlock adds a whole `rulesets:` key at the end of the document.
func (d *yamlDoc) appendBlock(rulesets []RulesetConfig) error {
	block, err := renderRulesetBlock(rulesets, 0)
	if err != nil {
		return err
	}
	if len(d.lines) > 0 && strings.TrimSpace(d.lines[len(d.lines)-1]) != "" {
		d.lines = append(d.lines, "")
	}
	d.lines = append(d.lines, block...)
	return d.save()
}

// blockEnd returns the 0-based index of the last line belonging to node.
//
// yaml.Node records where a node starts but not where it ends, so the end is
// the deepest line any descendant reports, extended over any following lines
// that are indented past the block — the continuation of a multi-line value.
func (d *yamlDoc) blockEnd(node *yaml.Node, indent int) int {
	end := deepestLine(node) - 1
	if end < 0 {
		end = 0
	}
	for end+1 < len(d.lines) {
		next := d.lines[end+1]
		if strings.TrimSpace(next) == "" {
			break
		}
		if lineIndent(next) <= indent {
			break
		}
		end++
	}
	if end >= len(d.lines) {
		end = len(d.lines) - 1
	}
	return end
}

func deepestLine(node *yaml.Node) int {
	if node == nil {
		return 0
	}
	deepest := node.Line
	for _, child := range node.Content {
		if l := deepestLine(child); l > deepest {
			deepest = l
		}
	}
	return deepest
}

func lineIndent(line string) int {
	return len(line) - len(strings.TrimLeft(line, " "))
}

func insertAfter(lines []string, index int, block []string) []string {
	if index < -1 {
		index = -1
	}
	if index >= len(lines) {
		return append(lines, block...)
	}
	out := make([]string, 0, len(lines)+len(block))
	out = append(out, lines[:index+1]...)
	out = append(out, block...)
	out = append(out, lines[index+1:]...)
	return out
}

func replaceLine(lines []string, index int, block []string) []string {
	if index < 0 || index >= len(lines) {
		return append(lines, block...)
	}
	out := make([]string, 0, len(lines)+len(block))
	out = append(out, lines[:index]...)
	out = append(out, block...)
	out = append(out, lines[index+1:]...)
	return out
}

// renderRulesetBlock renders a `rulesets:` key and its items at the given
// indentation, preceded by the marker comment.
func renderRulesetBlock(rulesets []RulesetConfig, indent int) ([]string, error) {
	items, err := renderRulesets(rulesets, indent+2)
	if err != nil {
		return nil, err
	}
	pad := strings.Repeat(" ", indent)
	out := []string{pad + adoptedComment, pad + "rulesets:"}
	return append(out, items...), nil
}

// renderRulesets encodes ruleset entries as YAML sequence items indented by
// indent spaces.
func renderRulesets(rulesets []RulesetConfig, indent int) ([]string, error) {
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(rulesets); err != nil {
		return nil, fmt.Errorf("encode rulesets: %w", err)
	}
	if err := enc.Close(); err != nil {
		return nil, fmt.Errorf("encode rulesets: %w", err)
	}

	pad := strings.Repeat(" ", indent)
	var out []string
	for _, line := range strings.Split(strings.TrimSuffix(buf.String(), "\n"), "\n") {
		if strings.TrimSpace(line) == "" {
			out = append(out, "")
			continue
		}
		out = append(out, pad+line)
	}
	return out, nil
}
