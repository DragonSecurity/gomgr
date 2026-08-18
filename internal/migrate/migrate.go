// Package migrate converts a configuration directory in the layout used by the
// Python github-user-management tool into the layout gomgr reads.
//
// The two are close enough that most of the work is renaming keys. What makes
// this worth a command rather than a shell script is the handful of places they
// are not close, because each one is a chance to produce a configuration that
// loads cleanly and grants the wrong thing:
//
//   - The Python layout puts every team in one file, keyed by team name. gomgr
//     reads one team per file, named for the team. Turning a YAML mapping key
//     into a filename is the point where untrusted input becomes a path.
//   - A repository entry may carry no permission at all. There is no safe guess
//     for what that meant, so this refuses instead of picking one.
//   - The Python layout has org-level team defaults; gomgr has none. Defaults
//     are inlined onto every team that does not state its own.
//
// Conversion is planned in full before anything is written. A directory is
// either converted completely or not touched, so a refusal never leaves a
// half-translated configuration behind for someone to apply.
package migrate

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/DragonSecurity/gomgr/internal/config"
)

// MissingPermission is what to do about a repository entry that names no
// permission.
type MissingPermission string

const (
	// MissingPermissionRefuse stops the conversion and names the entry. The
	// default: the source says a team has some access to a repository without
	// saying how much, and every way of resolving that is a guess about
	// someone's privileges.
	MissingPermissionRefuse MissingPermission = "refuse"
	// MissingPermissionDrop omits the entry and reports it. Dropping grants
	// less access than the source, which is the safe direction to be wrong in,
	// but it is still a change and still gets reported.
	MissingPermissionDrop MissingPermission = "drop"
)

// Options configures a conversion.
type Options struct {
	From string
	To   string
	// Force allows writing into a directory that already has something in it.
	Force bool
	// OnMissingPermission handles a repository entry with no permission.
	OnMissingPermission MissingPermission
}

// OrgResult is what happened to one organization directory.
type OrgResult struct {
	Dir    string
	Org    string
	Teams  int
	Grants int
	// Dropped is everything in the source that is not in the output.
	Dropped []string
	// Unmapped names source keys gomgr has no equivalent for. Distinct from
	// Dropped only in that nobody asked for them to go.
	Unmapped []string
}

// Result is what happened to the whole conversion.
type Result struct {
	Orgs []OrgResult
}

// Lossy reports whether anything in the source did not survive. Callers use it
// to exit non-zero, so a pipeline cannot read a lossy conversion as a clean one.
func (r Result) Lossy() bool {
	for _, o := range r.Orgs {
		if len(o.Dropped) > 0 || len(o.Unmapped) > 0 {
			return true
		}
	}
	return false
}

// pyOrg is org.yaml in the Python layout.
type pyOrg struct {
	OrgName   string   `yaml:"org_name"`
	OrgOwners []string `yaml:"org_owners"`
	Defaults  struct {
		Team struct {
			Privacy             string `yaml:"privacy"`
			NotificationSetting string `yaml:"notification_setting"`
		} `yaml:"team"`
	} `yaml:"defaults"`
}

// pyApp is app.yaml in the Python layout. The flag names happen to be identical
// to gomgr's, so these carry across unchanged.
type pyApp struct {
	RemoveMembersWithoutTeam bool `yaml:"remove_members_without_team"`
	DeleteUnconfiguredTeams  bool `yaml:"delete_unconfigured_teams"`
	CreateRepo               bool `yaml:"create_repo"`
}

// pyTeam is one entry in the Python teams file. Repos is map[string]*string so
// that a null permission is distinguishable from an empty one; both are refused,
// but only because they are seen at all.
type pyTeam struct {
	Description         string             `yaml:"description"`
	Privacy             string             `yaml:"privacy"`
	NotificationSetting string             `yaml:"notification_setting"`
	Maintainer          []string           `yaml:"maintainer"`
	Member              []string           `yaml:"member"`
	Repos               map[string]*string `yaml:"repos"`
}

// knownTeamKeys is what pyTeam consumes. Anything else in a team entry is
// reported rather than silently discarded.
var knownTeamKeys = map[string]bool{
	"description": true, "privacy": true, "notification_setting": true,
	"maintainer": true, "member": true, "repos": true,
}

// outFiles is a converted org directory held in memory. Nothing reaches disk
// until every org has converted, so a refusal in the last one does not leave
// the first nine rewritten.
type outFiles struct {
	dir   string
	files map[string][]byte
}

// Run converts every organization directory under opts.From into opts.To.
func Run(opts Options) (Result, error) {
	if opts.From == "" || opts.To == "" {
		return Result{}, errors.New("both --from and --to are required")
	}
	switch opts.OnMissingPermission {
	case "":
		opts.OnMissingPermission = MissingPermissionRefuse
	case MissingPermissionRefuse, MissingPermissionDrop:
	default:
		return Result{}, fmt.Errorf("unknown --on-missing-permission %q (must be %s or %s)",
			opts.OnMissingPermission, MissingPermissionRefuse, MissingPermissionDrop)
	}
	if err := checkDestination(opts.To, opts.Force); err != nil {
		return Result{}, err
	}

	entries, err := os.ReadDir(opts.From)
	if err != nil {
		return Result{}, fmt.Errorf("read source directory %s: %w", opts.From, err)
	}

	var res Result
	var pending []outFiles
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		src := filepath.Join(opts.From, e.Name())
		if _, err := os.Stat(filepath.Join(src, "org.yaml")); err != nil {
			continue // not an org directory
		}
		out, orgRes, err := convertOrg(src, e.Name(), opts)
		if err != nil {
			return Result{}, err
		}
		pending = append(pending, out)
		res.Orgs = append(res.Orgs, orgRes)
	}
	if len(pending) == 0 {
		return Result{}, fmt.Errorf("no organization directories found under %s (an org directory is one containing org.yaml)", opts.From)
	}

	for _, p := range pending {
		if err := writeOrg(opts.To, p); err != nil {
			return Result{}, err
		}
	}
	return res, nil
}

// checkDestination refuses to write into a directory that already holds
// something, because the obvious mistake is pointing --to at a live config.
func checkDestination(to string, force bool) error {
	entries, err := os.ReadDir(to)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read destination directory %s: %w", to, err)
	}
	if len(entries) > 0 && !force {
		return fmt.Errorf("destination %s is not empty — refusing to write over it (pass --force to overwrite)", to)
	}
	return nil
}

func convertOrg(src, dirName string, opts Options) (outFiles, OrgResult, error) {
	res := OrgResult{Dir: dirName}
	out := outFiles{dir: dirName, files: map[string][]byte{}}

	var org pyOrg
	if err := readYAML(filepath.Join(src, "org.yaml"), &org); err != nil {
		return out, res, err
	}
	// The org name decides which GitHub organization the result manages.
	// Guessing it from the directory name would let a rename apply one org's
	// configuration to another.
	if org.OrgName == "" {
		return out, res, fmt.Errorf("%s/org.yaml: org_name is required — it decides which organization the converted config manages, and the directory name is not a safe substitute", dirName)
	}
	res.Org = org.OrgName

	var app pyApp
	appPath := filepath.Join(src, "app.yaml")
	if _, err := os.Stat(appPath); err == nil {
		if err := readYAML(appPath, &app); err != nil {
			return out, res, err
		}
	}

	appOut := map[string]any{"org": org.OrgName}
	if app.RemoveMembersWithoutTeam {
		appOut["remove_members_without_team"] = true
	}
	if app.DeleteUnconfiguredTeams {
		appOut["delete_unconfigured_teams"] = true
	}
	if app.CreateRepo {
		appOut["create_repo"] = true
	}
	appBytes, err := marshal(appOut)
	if err != nil {
		return out, res, err
	}
	out.files["app.yaml"] = appBytes

	orgOut := map[string]any{"owners": org.OrgOwners}
	if org.OrgOwners == nil {
		orgOut["owners"] = []string{}
	}
	orgBytes, err := marshal(orgOut)
	if err != nil {
		return out, res, err
	}
	out.files["org.yaml"] = orgBytes

	teams, err := readTeams(src, dirName)
	if err != nil {
		return out, res, err
	}

	seenSlug := map[string]string{} // slug -> the team name that claimed it
	names := make([]string, 0, len(teams))
	for name := range teams {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		node := teams[name]
		var pt pyTeam
		if err := node.Decode(&pt); err != nil {
			// PyYAML keeps the last of a duplicated mapping key without
			// comment, so a repository listed twice under one team resolves to
			// whichever line came second and the first is simply gone. This
			// parser rejects it, which is the useful behavior — but the raw
			// error does not say why anyone should care.
			if strings.Contains(err.Error(), "already defined") {
				return out, res, fmt.Errorf("%s: team %q: %w\n"+
					"    A key listed twice takes its second value silently in the tool that wrote this. "+
					"Delete the duplicate in the source so the surviving value is the one someone chose", dirName, name, err)
			}
			return out, res, fmt.Errorf("%s: team %q: %w", dirName, name, err)
		}
		var raw map[string]any
		if err := node.Decode(&raw); err != nil {
			return out, res, fmt.Errorf("%s: team %q: %w", dirName, name, err)
		}
		for k := range raw {
			if !knownTeamKeys[k] {
				res.Unmapped = append(res.Unmapped, fmt.Sprintf("team %q: key %q has no gomgr equivalent", name, k))
			}
		}

		tc := config.TeamConfig{
			Name:        name,
			Description: pt.Description,
			Maintainers: pt.Maintainer,
			Members:     pt.Member,
		}
		// gomgr has no defaults block, so an org-level default becomes an
		// explicit value on every team that did not state its own.
		tc.Privacy = firstNonEmpty(pt.Privacy, org.Defaults.Team.Privacy)
		tc.NotificationSetting = firstNonEmpty(pt.NotificationSetting, org.Defaults.Team.NotificationSetting)

		repos := map[string]any{}
		repoNames := make([]string, 0, len(pt.Repos))
		for r := range pt.Repos {
			repoNames = append(repoNames, r)
		}
		sort.Strings(repoNames)
		for _, r := range repoNames {
			perm := pt.Repos[r]
			if perm == nil || strings.TrimSpace(*perm) == "" {
				if opts.OnMissingPermission == MissingPermissionDrop {
					res.Dropped = append(res.Dropped, fmt.Sprintf("team %q: repo %q had no permission — dropped", name, r))
					continue
				}
				return out, res, fmt.Errorf(
					"%s: team %q, repo %q has no permission. There is no safe default for this — "+
						"decide what it should be and set it in the source, or pass "+
						"--on-missing-permission=drop to leave the grant out", dirName, name, r)
			}
			repos[r] = strings.TrimSpace(*perm)
			res.Grants++
		}
		if len(repos) > 0 {
			tc.Repositories = repos
		}

		slug := tc.ResolvedSlug()
		if err := safeSegment(slug); err != nil {
			return out, res, fmt.Errorf("%s: team %q: %w", dirName, name, err)
		}
		if prev, dup := seenSlug[slug]; dup {
			return out, res, fmt.Errorf(
				"%s: teams %q and %q both convert to %s.yaml — rename one in the source, "+
					"because writing both would silently keep only the second", dirName, prev, name, slug)
		}
		seenSlug[slug] = name

		b, err := marshal(tc)
		if err != nil {
			return out, res, err
		}
		out.files[filepath.Join("teams", slug+".yaml")] = b
		res.Teams++
	}

	if d := org.Defaults.Team.NotificationSetting; d != "" && res.Teams == 0 {
		res.Unmapped = append(res.Unmapped, fmt.Sprintf("defaults.team.notification_setting %q had no teams to inline onto", d))
	}
	return out, res, nil
}

// readTeams reads every YAML file under teams/ as a mapping of team name to
// team body, refusing a name that appears in more than one file. Keeping the
// last one read would reintroduce exactly the file-order precedence this
// codebase removed from team repository definitions.
func readTeams(src, dirName string) (map[string]yaml.Node, error) {
	teamDir := filepath.Join(src, "teams")
	entries, err := os.ReadDir(teamDir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", teamDir, err)
	}
	all := map[string]yaml.Node{}
	from := map[string]string{}
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || (!strings.HasSuffix(name, ".yaml") && !strings.HasSuffix(name, ".yml")) {
			continue
		}
		var doc map[string]yaml.Node
		if err := readYAML(filepath.Join(teamDir, name), &doc); err != nil {
			return nil, err
		}
		for team, node := range doc {
			if prev, dup := all[team]; dup {
				_ = prev
				return nil, fmt.Errorf("%s: team %q is defined in both %s and %s — "+
					"remove one, rather than leaving which wins to the order the files are read",
					dirName, team, from[team], name)
			}
			all[team] = node
			from[team] = name
		}
	}
	return all, nil
}

// safeSegment refuses a slug that is anything other than one ordinary path
// component. This is the boundary where a YAML mapping key becomes a filename,
// so it rejects rather than sanitizes: quietly rewriting "../x" into "-x" would
// produce a team nobody named.
func safeSegment(s string) error {
	if s == "" {
		return errors.New("converts to an empty file name")
	}
	if s == "." || s == ".." || strings.Contains(s, "..") {
		return fmt.Errorf("converts to %q, which is not a usable file name", s)
	}
	if strings.ContainsAny(s, `/\`) || strings.ContainsRune(s, 0) {
		return fmt.Errorf("converts to %q, which contains a path separator", s)
	}
	if s != filepath.Base(s) {
		return fmt.Errorf("converts to %q, which is not a single path component", s)
	}
	return nil
}

func writeOrg(to string, out outFiles) error {
	if err := safeSegment(out.dir); err != nil {
		return fmt.Errorf("organization directory %q: %w", out.dir, err)
	}
	base := filepath.Join(to, out.dir)
	for rel, content := range out.files {
		dst := filepath.Join(base, rel)
		// Belt and braces: every segment was checked above, so this can only
		// fire if that check is ever weakened.
		if !strings.HasPrefix(filepath.Clean(dst), filepath.Clean(base)+string(os.PathSeparator)) {
			return fmt.Errorf("refusing to write %s: outside the destination directory", dst)
		}
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return fmt.Errorf("create %s: %w", filepath.Dir(dst), err)
		}
		if err := os.WriteFile(dst, content, 0o644); err != nil {
			return fmt.Errorf("write %s: %w", dst, err)
		}
	}
	return nil
}

func readYAML(path string, into any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	if err := yaml.Unmarshal(b, into); err != nil {
		return fmt.Errorf("parse %s: %w", path, err)
	}
	return nil
}

func marshal(v any) ([]byte, error) {
	b, err := yaml.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("marshal: %w", err)
	}
	return b, nil
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
