package migrate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/DragonSecurity/gomgr/internal/config"
)

// writeSrc lays out a Python-format org directory. Keys are paths relative to
// the org directory.
func writeSrc(t *testing.T, org string, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for rel, content := range files {
		p := filepath.Join(root, org, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

const orgYAML = `org_name: myorg
org_owners:
  - svc-account
defaults:
  team:
    privacy: closed
    notification_setting: notifications_disabled
`

func run(t *testing.T, from string, opts ...func(*Options)) (Result, string, error) {
	t.Helper()
	to := filepath.Join(t.TempDir(), "out")
	o := Options{From: from, To: to}
	for _, f := range opts {
		f(&o)
	}
	res, err := Run(o)
	return res, to, err
}

func readTeam(t *testing.T, to, org, slug string) config.TeamConfig {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(to, org, "teams", slug+".yaml"))
	if err != nil {
		t.Fatalf("read converted team: %v", err)
	}
	var tc config.TeamConfig
	if err := yaml.Unmarshal(b, &tc); err != nil {
		t.Fatal(err)
	}
	return tc
}

// Positive: a two-team org converts with members, maintainers and grants intact.
func TestRun_ConvertsTeamsAndGrants(t *testing.T) {
	from := writeSrc(t, "myorg", map[string]string{
		"org.yaml": orgYAML,
		"app.yaml": "create_repo: true\ndelete_unconfigured_teams: false\n",
		"teams/all-teams.yaml": `readonly:
  description: read access
  member:
    - alice
  repos:
    api: pull
writers:
  description: write access
  maintainer:
    - bob
  member:
    - bob
    - carol
  repos:
    api: push
    web: admin
`,
	})

	res, to, err := run(t, from)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(res.Orgs) != 1 || res.Orgs[0].Teams != 2 || res.Orgs[0].Grants != 3 {
		t.Fatalf("got %+v, want 1 org / 2 teams / 3 grants", res.Orgs)
	}
	if res.Lossy() {
		t.Fatalf("clean conversion reported lossy: %+v", res.Orgs[0])
	}

	writers := readTeam(t, to, "myorg", "writers")
	if writers.Name != "writers" || writers.Description != "write access" {
		t.Errorf("name/description not carried: %+v", writers)
	}
	if len(writers.Maintainers) != 1 || writers.Maintainers[0] != "bob" {
		t.Errorf("maintainer -> maintainers failed: %+v", writers.Maintainers)
	}
	if len(writers.Members) != 2 {
		t.Errorf("member -> members failed: %+v", writers.Members)
	}
	if writers.Repositories["api"] != "push" || writers.Repositories["web"] != "admin" {
		t.Errorf("repos -> repositories failed: %+v", writers.Repositories)
	}

	// The org's own file must name the org from org_name, not the directory.
	b, err := os.ReadFile(filepath.Join(to, "myorg", "app.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	var app struct {
		Org        string `yaml:"org"`
		CreateRepo bool   `yaml:"create_repo"`
	}
	if err := yaml.Unmarshal(b, &app); err != nil {
		t.Fatal(err)
	}
	if app.Org != "myorg" || !app.CreateRepo {
		t.Errorf("app.yaml = %+v, want org=myorg create_repo=true", app)
	}

	// And the result has to be loadable by gomgr, not merely written.
	if _, err := config.Load(filepath.Join(to, "myorg")); err != nil {
		t.Fatalf("converted config does not load: %v", err)
	}
}

// Positive: org defaults are inlined, but never over a team's own value.
func TestRun_InlinesTeamDefaults(t *testing.T) {
	from := writeSrc(t, "myorg", map[string]string{
		"org.yaml": orgYAML,
		"teams/all-teams.yaml": `inherits:
  description: takes the defaults
  member: [alice]
states-own:
  description: states its own
  privacy: secret
  notification_setting: notifications_enabled
  member: [bob]
`,
	})
	_, to, err := run(t, from)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	inherits := readTeam(t, to, "myorg", "inherits")
	if inherits.Privacy != "closed" || inherits.NotificationSetting != config.NotificationsDisabled {
		t.Errorf("defaults not inlined: privacy=%q notification=%q", inherits.Privacy, inherits.NotificationSetting)
	}
	own := readTeam(t, to, "myorg", "states-own")
	if own.Privacy != "secret" || own.NotificationSetting != config.NotificationsEnabled {
		t.Errorf("defaults overrode the team's own: privacy=%q notification=%q", own.Privacy, own.NotificationSetting)
	}
}

// A1: a team name that would escape the output directory is refused, and
// nothing is written.
func TestRun_RefusesPathTraversalInTeamName(t *testing.T) {
	from := writeSrc(t, "myorg", map[string]string{
		"org.yaml": orgYAML,
		"teams/all-teams.yaml": `../../../etc/cron.d/x:
  description: escape
  member: [mallory]
`,
	})
	res, to, err := run(t, from)
	if err == nil {
		t.Fatalf("expected refusal, got %+v", res)
	}
	if !strings.Contains(err.Error(), "path separator") && !strings.Contains(err.Error(), "not a usable file name") {
		t.Errorf("unhelpful error: %v", err)
	}
	if _, statErr := os.Stat(to); !os.IsNotExist(statErr) {
		t.Errorf("refusal still wrote to %s", to)
	}
}

// A2: two team names that converge on one file name are refused rather than
// silently collapsed into whichever was written last.
func TestRun_RefusesSlugCollision(t *testing.T) {
	from := writeSrc(t, "myorg", map[string]string{
		"org.yaml": orgYAML,
		"teams/all-teams.yaml": `web team:
  member: [alice]
  repos: {api: push}
Web-Team:
  member: [bob]
  repos: {api: admin}
`,
	})
	_, _, err := run(t, from)
	if err == nil {
		t.Fatal("expected a slug-collision refusal")
	}
	if !strings.Contains(err.Error(), "web-team.yaml") {
		t.Errorf("error should name the colliding file: %v", err)
	}
}

// A7: one team defined in two files is refused, rather than resolved by which
// file the directory listing happened to return second.
func TestRun_RefusesDuplicateTeamAcrossFiles(t *testing.T) {
	from := writeSrc(t, "myorg", map[string]string{
		"org.yaml":           orgYAML,
		"teams/a-teams.yaml": "writers:\n  repos: {api: push}\n",
		"teams/b-teams.yaml": "writers:\n  repos: {api: admin}\n",
	})
	_, _, err := run(t, from)
	if err == nil {
		t.Fatal("expected a duplicate-team refusal")
	}
	if !strings.Contains(err.Error(), "a-teams.yaml") || !strings.Contains(err.Error(), "b-teams.yaml") {
		t.Errorf("error should name both files: %v", err)
	}
}

// A3: a repository with no permission is refused by default, and dropped —
// reported, never guessed — only when asked.
func TestRun_MissingPermission(t *testing.T) {
	src := map[string]string{
		"org.yaml": orgYAML,
		"teams/all-teams.yaml": `writers:
  member: [alice]
  repos:
    api: push
    mystery:
`,
	}

	from := writeSrc(t, "myorg", src)
	_, _, err := run(t, from)
	if err == nil {
		t.Fatal("expected a refusal for the permissionless repo")
	}
	if !strings.Contains(err.Error(), "mystery") {
		t.Errorf("error should name the repo: %v", err)
	}

	from2 := writeSrc(t, "myorg", src)
	res, to, err := run(t, from2, func(o *Options) { o.OnMissingPermission = MissingPermissionDrop })
	if err != nil {
		t.Fatalf("drop mode should convert: %v", err)
	}
	if !res.Lossy() {
		t.Error("dropping a grant must report the conversion as lossy")
	}
	if len(res.Orgs[0].Dropped) != 1 || !strings.Contains(res.Orgs[0].Dropped[0], "mystery") {
		t.Errorf("drop not reported: %+v", res.Orgs[0].Dropped)
	}
	tc := readTeam(t, to, "myorg", "writers")
	if _, present := tc.Repositories["mystery"]; present {
		t.Error("dropped repo still in output")
	}
	if tc.Repositories["api"] != "push" {
		t.Error("dropping one grant lost another")
	}
}

// A4: a destination that already holds something is refused without --force.
func TestRun_RefusesNonEmptyDestination(t *testing.T) {
	from := writeSrc(t, "myorg", map[string]string{
		"org.yaml":             orgYAML,
		"teams/all-teams.yaml": "writers:\n  repos: {api: push}\n",
	})
	to := t.TempDir()
	if err := os.WriteFile(filepath.Join(to, "app.yaml"), []byte("org: live\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := Run(Options{From: from, To: to}); err == nil {
		t.Fatal("expected a refusal to write over a non-empty destination")
	}
	// The pre-existing file must still be intact.
	b, err := os.ReadFile(filepath.Join(to, "app.yaml"))
	if err != nil || !strings.Contains(string(b), "org: live") {
		t.Errorf("refusal damaged the destination: %q %v", b, err)
	}

	if _, err := Run(Options{From: from, To: to, Force: true}); err != nil {
		t.Fatalf("--force should allow it: %v", err)
	}
}

// A6: an org with no org_name is refused rather than named after its directory.
func TestRun_RefusesMissingOrgName(t *testing.T) {
	from := writeSrc(t, "some-directory", map[string]string{
		"org.yaml":             "org_owners:\n  - svc\n",
		"teams/all-teams.yaml": "writers:\n  repos: {api: push}\n",
	})
	_, _, err := run(t, from)
	if err == nil {
		t.Fatal("expected a refusal when org_name is absent")
	}
	if !strings.Contains(err.Error(), "org_name") {
		t.Errorf("error should name the missing key: %v", err)
	}
}

// A5: a key with no gomgr equivalent is reported, not silently dropped.
func TestRun_ReportsUnmappedKeys(t *testing.T) {
	from := writeSrc(t, "myorg", map[string]string{
		"org.yaml": orgYAML,
		"teams/all-teams.yaml": `writers:
  member: [alice]
  repos: {api: push}
  some_future_key: whatever
`,
	})
	res, _, err := run(t, from)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !res.Lossy() {
		t.Fatal("an unmapped key must make the conversion lossy")
	}
	if len(res.Orgs[0].Unmapped) != 1 || !strings.Contains(res.Orgs[0].Unmapped[0], "some_future_key") {
		t.Errorf("unmapped key not reported: %+v", res.Orgs[0].Unmapped)
	}
}

// A refusal in the second org must not leave the first one written, or a rerun
// would be merging into a partial conversion.
func TestRun_WritesNothingWhenAnyOrgFails(t *testing.T) {
	root := t.TempDir()
	for _, o := range []struct{ dir, teams string }{
		{"good", "writers:\n  repos: {api: push}\n"},
		{"bad", "writers:\n  repos:\n    mystery:\n"},
	} {
		p := filepath.Join(root, o.dir, "teams")
		if err := os.MkdirAll(p, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, o.dir, "org.yaml"), []byte(orgYAML), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(p, "all-teams.yaml"), []byte(o.teams), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	to := filepath.Join(t.TempDir(), "out")
	if _, err := Run(Options{From: root, To: to}); err == nil {
		t.Fatal("expected the bad org to fail the run")
	}
	if _, err := os.Stat(filepath.Join(to, "good")); !os.IsNotExist(err) {
		t.Error("the good org was written despite another org failing")
	}
}

func TestRun_RejectsUnknownMissingPermissionMode(t *testing.T) {
	from := writeSrc(t, "myorg", map[string]string{"org.yaml": orgYAML})
	if _, err := Run(Options{From: from, To: t.TempDir(), OnMissingPermission: "guess"}); err == nil {
		t.Fatal("expected an unknown-mode refusal")
	}
}

// R12: credentials in the source are not carried into the output. A migrated
// config should not be able to touch GitHub until someone deliberately gives it
// the means to.
func TestRun_DoesNotCarryCredentials(t *testing.T) {
	from := writeSrc(t, "myorg", map[string]string{
		"org.yaml": orgYAML,
		"app.yaml": `app_id: 12345
private_key: ./secret.pem
create_repo: true
`,
		"teams/all-teams.yaml": "writers:\n  repos: {api: push}\n",
	})
	_, to, err := run(t, from)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	b, err := os.ReadFile(filepath.Join(to, "myorg", "app.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	for _, leaked := range []string{"app_id", "12345", "private_key", "secret.pem"} {
		if strings.Contains(string(b), leaked) {
			t.Errorf("converted app.yaml carries %q:\n%s", leaked, b)
		}
	}
	if !strings.Contains(string(b), "create_repo") {
		t.Error("dropped a flag that should have carried across")
	}
}

// The Python layout says nesting as `parent: <Team Name>`; gomgr says it as a
// one-element `parents:` list. Before this was mapped, the key was reported as
// unmapped and the child silently lost the parent's repository access.
func TestRun_ConvertsParentToParents(t *testing.T) {
	from := writeSrc(t, "myorg", map[string]string{
		"org.yaml": orgYAML,
		"app.yaml": "create_repo: true\n",
		"teams/all-teams.yaml": `Platform:
  description: platform
  member:
    - alice
  repos:
    api: push
Platform Oncall:
  parent: Platform
  member:
    - bob
  repos:
    runbooks: admin
`,
	})

	res, to, err := run(t, from)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if res.Lossy() {
		t.Errorf("parent is mapped now, so nothing should be reported as lost: %+v", res.Orgs)
	}
	if got := res.Orgs[0].Nested; got != 1 {
		t.Errorf("expected 1 nested team reported, got %d", got)
	}

	child := readTeam(t, to, "myorg", "platform-oncall")
	if got := child.ParentSlug(); got != "platform" {
		t.Errorf("expected parent slug platform, got %q", got)
	}

	// The whole converted directory must load, which is what proves the parent
	// actually resolves against the other team file.
	root, err := config.Load(filepath.Join(to, "myorg"))
	if err != nil {
		t.Fatalf("converted config does not load: %v", err)
	}
	if err := root.Validate(); err != nil {
		t.Fatalf("converted config does not validate: %v", err)
	}
}

// A nested team cannot be secret at either end. The org-level default is the
// likely source, since it lands on teams that never mentioned privacy — so this
// is refused during conversion rather than left for the first `gomgr validate`.
func TestRun_RefusesSecretNestedTeam(t *testing.T) {
	from := writeSrc(t, "myorg", map[string]string{
		"org.yaml": `org_name: myorg
org_owners:
  - svc-account
defaults:
  team:
    privacy: secret
`,
		"app.yaml": "create_repo: true\n",
		"teams/all-teams.yaml": `Platform:
  member:
    - alice
Platform Oncall:
  parent: Platform
  member:
    - bob
`,
	})

	_, _, err := run(t, from)
	if err == nil {
		t.Fatal("a secret team in a hierarchy should be refused during conversion")
	}
	if !strings.Contains(err.Error(), "secret") {
		t.Errorf("the refusal should say why, got %v", err)
	}
}
