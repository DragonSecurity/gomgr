package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/google/go-github/v91/github"
	"gopkg.in/yaml.v3"

	"github.com/DragonSecurity/gomgr/internal/config"
	"github.com/DragonSecurity/gomgr/internal/util"
)

func repoWith(name string, autoMerge, mergeCommit, deleteBranch bool) *github.Repository {
	return &github.Repository{
		Name:                new(name),
		AllowAutoMerge:      new(autoMerge),
		AllowMergeCommit:    new(mergeCommit),
		DeleteBranchOnMerge: new(deleteBranch),
		AllowSquashMerge:    new(true),
		Visibility:          new("private"),
	}
}

func cfgWithDefaults(d config.RepoSettingsConfig) *config.Root {
	cfg := &config.Root{App: config.AppConfig{Org: "myorg"}}
	cfg.Org.RepoDefaults = d
	return cfg
}

var houseStyle = config.RepoSettingsConfig{
	AllowAutoMerge:      ptrTo(true),
	AllowMergeCommit:    ptrTo(false),
	DeleteBranchOnMerge: ptrTo(true),
}

func TestPlanRepoSettingsOnlyPlansRealDifferences(t *testing.T) {
	cfg := cfgWithDefaults(houseStyle)
	bySettings := map[string]repoSettings{"conforming": {}, "legacy": {}}
	existing := map[string]*github.Repository{
		// Already matches the house style.
		"conforming": repoWith("conforming", true, false, true),
		// One gomgr did not create.
		"legacy": repoWith("legacy", false, true, false),
	}

	changes, warnings, _ := planRepoSettings(context.Background(), nil, cfg, bySettings, existing)
	if len(warnings) != 0 {
		t.Errorf("warnings = %v", warnings)
	}
	if len(changes) != 1 {
		t.Fatalf("got %d changes, want only the drifted repository: %+v", len(changes), changes)
	}
	if !strings.HasPrefix(changes[0].Target, "legacy") {
		t.Errorf("target = %q, want legacy", changes[0].Target)
	}
	// The plan line has to name the settings, since that is all a reviewer sees.
	for _, want := range []string{"allow_auto_merge", "allow_merge_commit", "delete_branch_on_merge"} {
		if !strings.Contains(changes[0].Target, want) {
			t.Errorf("target %q should name %s", changes[0].Target, want)
		}
	}
	d := changes[0].Details.(map[string]any)
	if d["allow_auto_merge"] != true || d["allow_merge_commit"] != false {
		t.Errorf("details = %+v", d)
	}
}

func TestPlanRepoSettingsRepoOverridesOrgDefault(t *testing.T) {
	cfg := cfgWithDefaults(houseStyle)
	bySettings := map[string]repoSettings{
		"special": {settings: config.RepoSettingsConfig{AllowMergeCommit: ptrTo(true)}},
	}
	// Matches the org default, but the override wants merge commits back on.
	existing := map[string]*github.Repository{"special": repoWith("special", true, false, true)}

	changes, _, _ := planRepoSettings(context.Background(), nil, cfg, bySettings, existing)
	if len(changes) != 1 {
		t.Fatalf("got %d changes, want the override applied: %+v", len(changes), changes)
	}
	d := changes[0].Details.(map[string]any)
	if d["allow_merge_commit"] != true {
		t.Errorf("the repository override must win over the org default: %+v", d)
	}
	if _, touched := d["allow_auto_merge"]; touched {
		t.Error("a setting the override does not mention and already matches must not be planned")
	}
}

func TestPlanRepoSettingsSkipsWhatItCannotTouch(t *testing.T) {
	cfg := cfgWithDefaults(houseStyle)
	archived := repoWith("frozen", false, true, false)
	archived.Archived = new(true)
	bySettings := map[string]repoSettings{"frozen": {}, "brand-new": {}}
	existing := map[string]*github.Repository{"frozen": archived}

	changes, warnings, _ := planRepoSettings(context.Background(), nil, cfg, bySettings, existing)
	if len(changes) != 0 {
		t.Errorf("changes = %+v; an archived repository cannot be edited and a new one is configured at creation", changes)
	}
	if len(warnings) != 1 || !strings.Contains(warnings[0], "archived") {
		t.Errorf("warnings = %v, want one about the archived repository", warnings)
	}
}

func TestPlanRepoSettingsIsInertWithoutConfig(t *testing.T) {
	cfg := &config.Root{App: config.AppConfig{Org: "myorg"}}
	bySettings := map[string]repoSettings{"anything": {}}
	existing := map[string]*github.Repository{"anything": repoWith("anything", false, true, false)}

	changes, warnings, _ := planRepoSettings(context.Background(), nil, cfg, bySettings, existing)
	if len(changes) != 0 || len(warnings) != 0 {
		t.Errorf("a configuration declaring nothing must plan nothing: %+v %v", changes, warnings)
	}
}

func TestPlanRepoVisibilityNeedsBothKeys(t *testing.T) {
	existing := map[string]*github.Repository{"gomgr": repoWith("gomgr", true, false, true)} // private
	bySettings := map[string]repoSettings{"gomgr": {visibility: "public"}}

	t.Run("declared but not enabled only warns", func(t *testing.T) {
		cfg := &config.Root{App: config.AppConfig{Org: "myorg"}}
		changes, warnings := planRepoVisibility(cfg, bySettings, existing)
		if len(changes) != 0 {
			t.Errorf("changing visibility needs reconcile_visibility, got %+v", changes)
		}
		if len(warnings) != 1 || !strings.Contains(warnings[0], "reconcile_visibility") {
			t.Errorf("warnings = %v, want one naming the flag", warnings)
		}
	})

	t.Run("enabled and declared plans the change", func(t *testing.T) {
		cfg := &config.Root{App: config.AppConfig{Org: "myorg", ReconcileVisibility: true}}
		changes, _ := planRepoVisibility(cfg, bySettings, existing)
		if len(changes) != 1 {
			t.Fatalf("got %+v", changes)
		}
		if !strings.Contains(changes[0].Target, "private -> public") {
			t.Errorf("the plan line must show the direction of travel, got %q", changes[0].Target)
		}
	})

	t.Run("enabled but undeclared does nothing", func(t *testing.T) {
		cfg := &config.Root{App: config.AppConfig{Org: "myorg", ReconcileVisibility: true}}
		changes, warnings := planRepoVisibility(cfg, map[string]repoSettings{"gomgr": {}}, existing)
		if len(changes) != 0 || len(warnings) != 0 {
			t.Errorf("a repository that declares no visibility is never touched: %+v %v", changes, warnings)
		}
	})
}

// TestOrgDefaultsCannotSetVisibility is the guard that matters: no
// organization-wide key exists that could move repositories between public and
// private, so one edit can never expose thirty-four repositories.
func TestOrgDefaultsCannotSetVisibility(t *testing.T) {
	var block map[string]any
	if err := yamlUnmarshalStrict(t, "allow_auto_merge: true\nvisibility: public\n", &block); err != nil {
		t.Fatalf("fixture: %v", err)
	}
	parsed, err := config.ParseRepoSettings(block)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if parsed.AllowAutoMerge == nil || !*parsed.AllowAutoMerge {
		t.Error("the settings block should still parse the settings it does own")
	}
	// RepoSettingsConfig has no visibility field at all — there is nowhere for
	// an org default to put one.
	if strings.Contains(structFields(parsed), "Visibility") {
		t.Error("RepoSettingsConfig must not carry visibility")
	}
}

func TestApplyRepoSettingsVerifiesTheResult(t *testing.T) {
	t.Run("succeeds when GitHub applies it", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var sent map[string]any
			_ = json.NewDecoder(r.Body).Decode(&sent)
			sent["name"] = "legacy"
			_ = json.NewEncoder(w).Encode(sent)
		}))
		defer server.Close()

		ch := util.Change{Scope: scopeRepoSettings, Action: "ensure", Details: map[string]any{
			"org": "myorg", "repo": "legacy", "allow_auto_merge": true, "allow_merge_commit": false,
		}}
		if err := applyRepoSettingsEnsure(context.Background(), newTestClient(t, server), ch); err != nil {
			t.Fatalf("apply: %v", err)
		}
	})

	t.Run("fails when GitHub ignores it", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			// Accepts the request, reports the old value back.
			_ = json.NewEncoder(w).Encode(map[string]any{"name": "legacy", "allow_auto_merge": false})
		}))
		defer server.Close()

		ch := util.Change{Scope: scopeRepoSettings, Action: "ensure", Details: map[string]any{
			"org": "myorg", "repo": "legacy", "allow_auto_merge": true,
		}}
		err := applyRepoSettingsEnsure(context.Background(), newTestClient(t, server), ch)
		if err == nil || !strings.Contains(err.Error(), "did not take effect") {
			t.Fatalf("err = %v, want a report that the change did not stick", err)
		}
	})
}

func TestRepoSettingsValidateRejectsNoMergeMethod(t *testing.T) {
	all := config.RepoSettingsConfig{
		AllowSquashMerge: ptrTo(false),
		AllowMergeCommit: ptrTo(false),
		AllowRebaseMerge: ptrTo(false),
	}
	if err := all.Validate("org.yaml"); err == nil {
		t.Fatal("GitHub requires one merge method to stay on; this should be caught before the API says so")
	}
}

func yamlUnmarshalStrict(t *testing.T, src string, out any) error {
	t.Helper()
	return yaml.Unmarshal([]byte(src), out)
}

func structFields(v any) string { return fmt.Sprintf("%#v", v) }

// securityRepo builds a repository reporting the security_and_analysis
// statuses given, in the shape GitHub returns them.
func securityRepo(name string, statuses map[string]string) *github.Repository {
	sa := &github.SecurityAndAnalysis{}
	if v, ok := statuses["secret_scanning"]; ok {
		sa.SecretScanning = &github.SecretScanning{Status: new(v)}
	}
	if v, ok := statuses["secret_scanning_push_protection"]; ok {
		sa.SecretScanningPushProtection = &github.SecretScanningPushProtection{Status: new(v)}
	}
	if v, ok := statuses["dependabot_security_updates"]; ok {
		sa.DependabotSecurityUpdates = &github.DependabotSecurityUpdates{Status: new(v)}
	}
	return &github.Repository{Name: new(name), Visibility: new("public"), SecurityAndAnalysis: sa}
}

func TestPlanRepoSettingsReconcilesSecurityAnalyses(t *testing.T) {
	cfg := cfgWithDefaults(config.RepoSettingsConfig{
		SecretScanning:               ptrTo(true),
		SecretScanningPushProtection: ptrTo(true),
	})
	bySettings := map[string]repoSettings{"ward": {}, "already-on": {}}
	existing := map[string]*github.Repository{
		"ward": securityRepo("ward", map[string]string{
			"secret_scanning": "disabled", "secret_scanning_push_protection": "disabled",
		}),
		"already-on": securityRepo("already-on", map[string]string{
			"secret_scanning": "enabled", "secret_scanning_push_protection": "enabled",
		}),
	}

	changes, warnings, err := planRepoSettings(context.Background(), nil, cfg, bySettings, existing)
	if err != nil {
		t.Fatalf("plan: %v", err)
	}
	if len(warnings) != 0 {
		t.Errorf("warnings = %v, want none; secret scanning is being turned on in the same edit", warnings)
	}
	if len(changes) != 1 {
		t.Fatalf("got %d changes, want only the repository that is off: %+v", len(changes), changes)
	}
	if !strings.HasPrefix(changes[0].Target, "ward") {
		t.Errorf("target = %q, want ward", changes[0].Target)
	}
	details, err := extractDetails(changes[0])
	if err != nil {
		t.Fatalf("details: %v", err)
	}
	for _, name := range []string{"secret_scanning", "secret_scanning_push_protection"} {
		if details[name] != true {
			t.Errorf("details[%q] = %v, want true", name, details[name])
		}
	}
}

// A repository GitHub declines to describe is not a repository that has the
// feature switched off, and planning a change for it would put every such
// repository in every plan forever.
func TestPlanRepoSettingsLeavesUnreportedSecurityAlone(t *testing.T) {
	cfg := cfgWithDefaults(config.RepoSettingsConfig{SecretScanning: ptrTo(true)})
	bySettings := map[string]repoSettings{"quiet": {}}
	existing := map[string]*github.Repository{
		"quiet": {Name: new("quiet"), Visibility: new("public")},
	}

	changes, warnings, err := planRepoSettings(context.Background(), nil, cfg, bySettings, existing)
	if err != nil {
		t.Fatalf("plan: %v", err)
	}
	if len(changes) != 0 {
		t.Errorf("changes = %+v, want none", changes)
	}
	if len(warnings) != 1 || !strings.Contains(warnings[0], "did not report secret_scanning") {
		t.Errorf("warnings = %v, want one saying GitHub did not report it", warnings)
	}
}

func TestWarnSecretScanningDependency(t *testing.T) {
	off := securityRepo("ward", map[string]string{"secret_scanning": "disabled"})

	t.Run("warns when secret scanning stays off", func(t *testing.T) {
		got := warnSecretScanningDependency(
			config.RepoSettingsConfig{SecretScanningPushProtection: ptrTo(true)}, off, "ward")
		if len(got) != 1 || !strings.Contains(got[0], "secret_scanning: true") {
			t.Errorf("warnings = %v, want one naming the setting to add", got)
		}
	})

	t.Run("stays quiet when the same run turns it on", func(t *testing.T) {
		got := warnSecretScanningDependency(config.RepoSettingsConfig{
			SecretScanning:               ptrTo(true),
			SecretScanningPushProtection: ptrTo(true),
		}, off, "ward")
		if len(got) != 0 {
			t.Errorf("warnings = %v, want none", got)
		}
	})

	t.Run("stays quiet when it is already on", func(t *testing.T) {
		on := securityRepo("ward", map[string]string{"secret_scanning": "enabled"})
		got := warnSecretScanningDependency(
			config.RepoSettingsConfig{SecretScanningPushProtection: ptrTo(true)}, on, "ward")
		if len(got) != 0 {
			t.Errorf("warnings = %v, want none", got)
		}
	})

	t.Run("stays quiet when GitHub did not say", func(t *testing.T) {
		got := warnSecretScanningDependency(
			config.RepoSettingsConfig{SecretScanningPushProtection: ptrTo(true)},
			&github.Repository{Name: new("ward")}, "ward")
		if len(got) != 0 {
			t.Errorf("warnings = %v, want none; a guess here is noise", got)
		}
	})
}

// Several security features are separate settingFields but share one nested
// object in the edit, so each must write into it rather than replace it.
func TestApplyRepoSettingsSendsOneSecurityObject(t *testing.T) {
	var sent map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&sent)
		_ = json.NewEncoder(w).Encode(map[string]any{"name": "ward"})
	}))
	defer server.Close()

	ch := util.Change{Scope: scopeRepoSettings, Action: "ensure", Details: map[string]any{
		"org": "myorg", "repo": "ward",
		"secret_scanning": true, "secret_scanning_push_protection": true,
		"dependabot_security_updates": false,
	}}
	if err := applyRepoSettingsEnsure(context.Background(), newTestClient(t, server), ch); err != nil {
		t.Fatalf("apply: %v", err)
	}

	sa, ok := sent["security_and_analysis"].(map[string]any)
	if !ok {
		t.Fatalf("request = %v, want a security_and_analysis object", sent)
	}
	want := map[string]string{
		"secret_scanning": "enabled", "secret_scanning_push_protection": "enabled",
		"dependabot_security_updates": "disabled",
	}
	for feature, status := range want {
		got, ok := sa[feature].(map[string]any)
		if !ok {
			t.Errorf("%s missing from the edit: %v", feature, sa)
			continue
		}
		if got["status"] != status {
			t.Errorf("%s status = %v, want %q", feature, got["status"], status)
		}
	}
}

func TestRepoSettingsValidateRejectsPushProtectionWithoutScanning(t *testing.T) {
	c := config.RepoSettingsConfig{
		SecretScanning:               ptrTo(false),
		SecretScanningPushProtection: ptrTo(true),
	}
	err := c.Validate("org.yaml")
	if err == nil || !strings.Contains(err.Error(), "secret_scanning_push_protection") {
		t.Fatalf("err = %v, want the contradiction named", err)
	}
}

// A setting added to RepoSettingsConfig and nowhere else loads, validates and
// is then silently ignored — the configuration says one thing and gomgr does
// another, with nothing anywhere reporting it. This walks the struct so that
// the next field cannot be half-wired.
func TestEveryRepoSettingIsWiredUp(t *testing.T) {
	typ := reflect.TypeOf(config.RepoSettingsConfig{})
	for i := range typ.NumField() {
		field := typ.Field(i)
		t.Run(field.Name, func(t *testing.T) {
			var only config.RepoSettingsConfig
			reflect.ValueOf(&only).Elem().Field(i).Set(reflect.ValueOf(ptrTo(true)))

			if only.IsEmpty() {
				t.Error("IsEmpty does not look at this field, so a config stating only it is skipped entirely")
			}
			if merged := only.MergedWith(config.RepoSettingsConfig{}); reflect.ValueOf(merged).Field(i).IsNil() {
				t.Error("MergedWith drops this field, so a repository stating it is ignored")
			}
			if merged := (config.RepoSettingsConfig{}).MergedWith(only); reflect.ValueOf(merged).Field(i).IsNil() {
				t.Error("MergedWith does not fall through to the org default for this field")
			}

			noConflict := func(string, any, any) error { return nil }
			merged, err := mergeRepoSettingsConfig(config.RepoSettingsConfig{}, only, noConflict)
			if err != nil {
				t.Fatalf("merge: %v", err)
			}
			if reflect.ValueOf(merged).Field(i).IsNil() {
				t.Error("mergeRepoSettingsConfig drops this field, so a repository two teams declare loses it")
			}

			wired := false
			for _, f := range settingFields {
				if f.want(only) != nil {
					wired = true
					break
				}
			}
			if !wired {
				t.Error("no settingField reconciles this; it would be accepted in config and never applied")
			}
		})
	}
}
