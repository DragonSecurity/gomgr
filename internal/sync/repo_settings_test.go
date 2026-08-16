package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-github/v90/github"
	"gopkg.in/yaml.v3"

	"github.com/DragonSecurity/gomgr/internal/config"
	"github.com/DragonSecurity/gomgr/internal/util"
)

func repoWith(name string, autoMerge, mergeCommit, deleteBranch bool) *github.Repository {
	return &github.Repository{
		Name:                github.Ptr(name),
		AllowAutoMerge:      github.Ptr(autoMerge),
		AllowMergeCommit:    github.Ptr(mergeCommit),
		DeleteBranchOnMerge: github.Ptr(deleteBranch),
		AllowSquashMerge:    github.Ptr(true),
		Visibility:          github.Ptr("private"),
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
	archived.Archived = github.Ptr(true)
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
