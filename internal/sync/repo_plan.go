package sync

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v90/github"

	"github.com/DragonSecurity/gomgr/internal/config"
	"github.com/DragonSecurity/gomgr/internal/gh"
	"github.com/DragonSecurity/gomgr/internal/util"
)

// repoPlanner holds the state that planning a repository needs, so the phases
// below can be read one at a time.
//
// Everything here was previously local to one function that had grown past
// three hundred lines and a cyclomatic complexity of fifty-one. That length is
// not a style problem: it hid a `continue` intended to skip a permission grant
// which in fact skipped the repository's topics, pinning, codeowners and file
// writes as well.
type repoPlanner struct {
	ctx context.Context
	c   *gh.Client
	cfg *config.Root
	st  *State
	org string

	// exists is updated as repositories are planned for creation, so it answers
	// "will this repository be there by apply time".
	exists map[string]bool
	// preexisting is fixed at the start and answers "is there anything there to
	// read right now" — which is what deciding to probe a file depends on.
	preexisting map[string]bool
	// repos are the repositories GitHub reported, by lowercase name.
	repos map[string]*github.Repository

	settings     map[string]repoSettings
	managedRepos map[string]bool

	fileSpecs     []config.FileSpec
	userFilePaths map[string]bool
	emittedFiles  map[string]bool
	probeFor      func(string) fileProbe
	router        fileRouter

	// Accumulated across teams, because several teams may name one repository.
	topics    map[string][]string
	pinned    map[string]bool
	templates map[string]bool
	owners    map[string][]string
	ownerSeen map[string]map[string]bool
	repoNames map[string]string
}

func newRepoPlanner(ctx context.Context, c *gh.Client, cfg *config.Root, st *State) (*repoPlanner, error) {
	p := &repoPlanner{
		ctx: ctx, c: c, cfg: cfg, st: st, org: st.Org,
		exists:        map[string]bool{},
		preexisting:   map[string]bool{},
		repos:         map[string]*github.Repository{},
		userFilePaths: map[string]bool{},
		emittedFiles:  map[string]bool{},
		topics:        map[string][]string{},
		pinned:        map[string]bool{},
		templates:     map[string]bool{},
		owners:        map[string][]string{},
		ownerSeen:     map[string]map[string]bool{},
		repoNames:     map[string]string{},
	}

	for _, r := range st.ActualRepos {
		name := strings.ToLower(r.GetName())
		p.exists[name] = true
		p.preexisting[name] = true
		p.repos[name] = r
	}

	all, managed, err := collectRepoSettings(cfg, p.org)
	if err != nil {
		return nil, err
	}
	p.managedRepos = managed
	if p.settings, err = resolveAllTemplates(all, p.org); err != nil {
		return nil, err
	}

	p.fileSpecs = materializeFileSpecs(cfg.App)
	for _, fs := range p.fileSpecs {
		p.userFilePaths[fs.Path] = true
	}

	p.router = newRouteDecider(cfg, st, p.settings).route
	probe := newFileProbe(c)
	p.probeFor = func(repo string) fileProbe {
		// A repository this run is creating has nothing to read yet.
		if !p.preexisting[repo] {
			return nil
		}
		return probe
	}
	return p, nil
}

// planRepoPerms plans everything that follows from a team naming a repository:
// the repository itself, the grant, its topics, pins, templates, codeowners,
// files, settings and visibility.
func planRepoPerms(ctx context.Context, c *gh.Client, cfg *config.Root, st *State) ([]util.Change, error) {
	p, err := newRepoPlanner(ctx, c, cfg, st)
	if err != nil {
		return nil, err
	}

	out, err := p.planPerTeamRepo()
	if err != nil {
		return nil, err
	}
	out = append(out, p.planCodeownerFiles()...)
	out = append(out, p.planAccumulated()...)

	p.recordStats()

	out, err = p.dropSatisfiedGrants(out)
	if err != nil {
		return nil, err
	}

	extra, err := p.planRepoLevelSettings()
	if err != nil {
		return nil, err
	}
	out = append(out, extra...)

	markFinalPullRequestFiles(out)
	return out, nil
}

// planPerTeamRepo walks every repository each team names.
func (p *repoPlanner) planPerTeamRepo() ([]util.Change, error) {
	var out []util.Change
	for _, t := range p.cfg.Team {
		if p.ctx.Err() != nil {
			return nil, p.ctx.Err()
		}
		slug := t.ResolvedSlug()
		for repo := range t.Repositories {
			key := strings.ToLower(repo)
			settings := p.settings[key]

			if change, ok := p.planRepoCreation(repo, key, settings); ok {
				out = append(out, change)
			}
			if grant, ok := p.planGrant(slug, repo, key, settings); ok {
				out = append(out, grant)
			}
			if err := p.accumulate(repo, key, settings); err != nil {
				return nil, err
			}

			files, err := planRepoFiles(p.ctx, p.probeFor(key), p.router,
				p.org, repo, key, p.fileSpecs, p.cfg.App.SignOff, p.emittedFiles)
			if err != nil {
				return nil, err
			}
			out = append(out, files...)
		}
	}
	return out, nil
}

// planRepoCreation emits a repo:ensure for a repository that is not there yet.
func (p *repoPlanner) planRepoCreation(repo, key string, settings repoSettings) (util.Change, bool) {
	if p.exists[key] || !p.cfg.App.CreateRepo {
		return util.Change{}, false
	}
	details := map[string]any{"org": p.org, "name": repo}
	if settings.visibility != "" {
		details["visibility"] = settings.visibility
	} else {
		details["private"] = true
	}
	if settings.from != "" {
		details["from"] = settings.from
	}
	if settings.template {
		details["template"] = true
	}
	p.exists[key] = true
	return util.Change{Scope: "repo", Target: key, Action: "ensure", Details: details}, true
}

// planGrant emits a team-repo:grant, unless the entry declares no permission.
//
// Granting the empty string is not "leave it alone": GitHub reads it as its own
// default of read, so the grant is planned, applied, and planned again next run
// while the team keeps whatever access it had. Saying nothing is the only way to
// say nothing.
//
// Only the grant is skipped. An entry may declare topics or rulesets and no
// permission at all, and everything else it declares still applies.
func (p *repoPlanner) planGrant(slug, repo, key string, settings repoSettings) (util.Change, bool) {
	if settings.permission == "" {
		return util.Change{}, false
	}
	return util.Change{
		Scope:  "team-repo",
		Target: slug + "/" + key,
		Action: "grant",
		Details: map[string]any{
			"org": p.org, "slug": slug, "repo": repo, "permission": settings.permission,
		},
	}, true
}

// accumulate records what this team asks of the repository. Topics and
// codeowners are unions across teams, so they are gathered before being planned.
func (p *repoPlanner) accumulate(repo, key string, settings repoSettings) error {
	if settings.template {
		p.templates[key] = true
	}
	if settings.pinned {
		p.pinned[key] = true
	}

	if len(settings.topics) > 0 {
		seen := map[string]bool{}
		merged := p.topics[key]
		for _, topic := range merged {
			seen[topic] = true
		}
		for _, topic := range settings.topics {
			if err := validateTopic(topic); err != nil {
				return fmt.Errorf("invalid topic for repo %s: %w", repo, err)
			}
			if !seen[topic] {
				merged = append(merged, topic)
				seen[topic] = true
			}
		}
		p.topics[key] = merged
	}

	if len(settings.codeowners) > 0 {
		if _, ok := p.repoNames[key]; !ok {
			p.repoNames[key] = repo
		}
		if p.ownerSeen[key] == nil {
			p.ownerSeen[key] = map[string]bool{}
		}
		for _, co := range settings.codeowners {
			if p.ownerSeen[key][co] {
				continue
			}
			p.ownerSeen[key][co] = true
			p.owners[key] = append(p.owners[key], co)
		}
	}
	return nil
}

func (p *repoPlanner) planCodeownerFiles() []util.Change {
	out := planCodeowners(p.org, p.owners, p.repoNames, p.userFilePaths, p.cfg.App.SignOff, p.emittedFiles)
	if p.cfg.App.DeleteStaleCodeowners {
		out = append(out, planCodeownersDeletions(p.org, p.managedRepos, p.repoNames,
			p.owners, p.userFilePaths, p.cfg.App.SignOff, p.emittedFiles)...)
	}
	return out
}

// planAccumulated emits the changes gathered across every team: topics, pins
// and template flags, each compared against what GitHub already has.
func (p *repoPlanner) planAccumulated() []util.Change {
	var out []util.Change
	for _, repo := range sortedKeys(p.topics) {
		if change, ok := p.planTopics(repo, p.topics[repo]); ok {
			out = append(out, change)
		}
	}
	for _, repo := range sortedKeys(p.pinned) {
		if p.pinned[repo] {
			out = append(out, util.Change{
				Scope: "repo-pin", Target: repo, Action: "ensure",
				Details: map[string]any{"org": p.org, "repo": repo, "pinned": true},
			})
		}
	}
	for _, repo := range sortedKeys(p.templates) {
		if p.templates[repo] && !p.repos[repo].GetIsTemplate() {
			out = append(out, util.Change{
				Scope: "repo-template", Target: repo, Action: "ensure",
				Details: map[string]any{"org": p.org, "repo": repo, "template": true},
			})
		}
	}
	return out
}

// planTopics emits a topic change when the wanted set differs from the current
// one. A repository that is not there yet always needs its topics written.
func (p *repoPlanner) planTopics(repo string, want []string) (util.Change, bool) {
	current, known := p.repos[repo]
	if known && topicsMatch(current.Topics, want) {
		return util.Change{}, false
	}
	return util.Change{
		Scope: "repo-topics", Target: repo, Action: "ensure",
		Details: map[string]any{"org": p.org, "repo": repo, "topics": want},
	}, true
}

func topicsMatch(current, want []string) bool {
	if len(current) != len(want) {
		return false
	}
	have := make(map[string]bool, len(current))
	for _, t := range current {
		have[t] = true
	}
	for _, t := range want {
		if !have[t] {
			return false
		}
	}
	return true
}

func (p *repoPlanner) recordStats() {
	p.st.ManagedRepos = p.managedRepos
	p.st.CurrentRepos = len(p.exists)
	p.st.DesiredRepos = len(p.managedRepos)

	desired := 0
	for _, t := range p.cfg.Team {
		desired += len(t.Repositories)
	}
	p.st.DesiredRepoPerms = desired
}

// dropSatisfiedGrants removes grants the team already holds, which needs the
// current permissions and so cannot be decided while walking the teams.
func (p *repoPlanner) dropSatisfiedGrants(in []util.Change) ([]util.Change, error) {
	count, current, err := fetchCurrentPermissions(p.ctx, p.c, p.cfg, p.org)
	if err != nil {
		return nil, fmt.Errorf("fetch current permissions: %w", err)
	}
	p.st.CurrentRepoPerms = count

	out := in[:0]
	for _, ch := range in {
		if ch.Scope == "team-repo" && ch.Action == "grant" {
			d, ok := ch.Details.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("invalid details for team-repo:grant: %T", ch.Details)
			}
			slug, _ := d["slug"].(string)
			repo, _ := d["repo"].(string)
			perm, _ := d["permission"].(string)
			if current[slug+"/"+strings.ToLower(repo)] == normalizePermission(perm) {
				continue
			}
		}
		out = append(out, ch)
	}
	return out, nil
}

// planRepoLevelSettings covers the settings that belong to the repository
// rather than to any team naming it.
func (p *repoPlanner) planRepoLevelSettings() ([]util.Change, error) {
	settings, warnings, err := planRepoSettings(p.ctx, newRepoDetailFetcher(p.c), p.cfg, p.settings, p.repos)
	if err != nil {
		return nil, err
	}
	visibility, visWarnings := planRepoVisibility(p.cfg, p.settings, p.repos)
	p.st.RepoWarnings = append(p.st.RepoWarnings, warnings...)
	p.st.RepoWarnings = append(p.st.RepoWarnings, visWarnings...)
	return append(settings, visibility...), nil
}
