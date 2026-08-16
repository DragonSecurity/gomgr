package sync

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/DragonSecurity/gomgr/internal/config"
)

func TestMaterializeFileSpecs_LegacyFlags(t *testing.T) {
	app := config.AppConfig{
		AddDefaultReadme:  true,
		AddRenovateConfig: true,
		RenovateConfig:    `{"extends":["x"]}`,
	}
	files := materializeFileSpecs(app)
	if len(files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(files))
	}
	if files[0].Path != "README.md" {
		t.Errorf("expected README.md first, got %q", files[0].Path)
	}
	if files[1].Path != ".github/renovate.json" {
		t.Errorf("expected renovate.json second, got %q", files[1].Path)
	}
}

func TestMaterializeFileSpecs_SkipsRenovateWhenContentEmpty(t *testing.T) {
	app := config.AppConfig{AddRenovateConfig: true, RenovateConfig: "   "}
	files := materializeFileSpecs(app)
	if len(files) != 0 {
		t.Errorf("expected empty list when renovate content is blank, got %d entries", len(files))
	}
}

func TestMaterializeFileSpecs_UserOverridesLegacy(t *testing.T) {
	app := config.AppConfig{
		AddDefaultReadme: true,
		Files: []config.FileSpec{
			{Path: "README.md", Content: "# custom readme for {{.Repo}}", Message: "chore: README", Branch: "main"},
		},
	}
	files := materializeFileSpecs(app)
	if len(files) != 1 {
		t.Fatalf("expected 1 file after dedup, got %d", len(files))
	}
	if !strings.Contains(files[0].Content, "custom readme") {
		t.Errorf("expected user override to win, got %q", files[0].Content)
	}
}

const testSignOff = "dsec-gom[bot] <224359171+dsec-gom[bot]@users.noreply.github.com>"

func TestWithSignOff(t *testing.T) {
	cases := []struct {
		name    string
		message string
		signOff string
		want    string
	}{
		{
			name:    "empty signOff leaves the message untouched",
			message: "chore: add LICENSE",
			signOff: "",
			want:    "chore: add LICENSE",
		},
		{
			name:    "trailer is separated by a blank line",
			message: "chore: add LICENSE",
			signOff: testSignOff,
			want:    "chore: add LICENSE\n\nSigned-off-by: " + testSignOff,
		},
		{
			name:    "already signed off is left alone",
			message: "chore: add LICENSE\n\nSigned-off-by: Someone <someone@example.com>",
			signOff: testSignOff,
			want:    "chore: add LICENSE\n\nSigned-off-by: Someone <someone@example.com>",
		},
		{
			name:    "trailing newlines are collapsed before the trailer",
			message: "chore: add LICENSE\n\n",
			signOff: testSignOff,
			want:    "chore: add LICENSE\n\nSigned-off-by: " + testSignOff,
		},
		{
			name:    "multi-line body keeps its body",
			message: "chore: add LICENSE\n\nBecause the auditors asked.",
			signOff: testSignOff,
			want:    "chore: add LICENSE\n\nBecause the auditors asked.\n\nSigned-off-by: " + testSignOff,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := withSignOff(tc.message, tc.signOff); got != tc.want {
				t.Errorf("withSignOff(%q, %q) = %q, want %q", tc.message, tc.signOff, got, tc.want)
			}
		})
	}
}

func TestPlanRepoFiles_SignsOffCustomAndDefaultMessages(t *testing.T) {
	specs := []config.FileSpec{
		{Path: "README.md", Content: "hi", Message: "chore: readme"},
		{Path: "LICENSE", Content: "MIT\n"},
	}

	changes, err := planRepoFiles(context.Background(), nil, "Acme", "widgets", "widgets", specs, testSignOff, map[string]bool{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(changes) != 2 {
		t.Fatalf("expected 2 changes, got %d", len(changes))
	}

	want := "chore: readme\n\nSigned-off-by: " + testSignOff
	if got := changes[0].Details.(map[string]any)["message"]; got != want {
		t.Errorf("custom message: got %q, want %q", got, want)
	}
	want = "chore: add LICENSE\n\nSigned-off-by: " + testSignOff
	if got := changes[1].Details.(map[string]any)["message"]; got != want {
		t.Errorf("default message: got %q, want %q", got, want)
	}
}

func TestPlanCodeowners_SignsOffSyncAndDeleteMessages(t *testing.T) {
	owners := map[string][]string{"widgets": {"@acme/platform"}}
	names := map[string]string{"widgets": "widgets"}

	changes := planCodeowners("acme", owners, names, map[string]bool{}, testSignOff, map[string]bool{})
	if len(changes) != 1 {
		t.Fatalf("expected 1 codeowners change, got %d", len(changes))
	}
	want := "chore: sync CODEOWNERS\n\nSigned-off-by: " + testSignOff
	if got := changes[0].Details.(map[string]any)["message"]; got != want {
		t.Errorf("sync message: got %q, want %q", got, want)
	}

	managed := map[string]bool{"widgets": true}
	deletions := planCodeownersDeletions("acme", managed, names, map[string][]string{}, map[string]bool{}, testSignOff, map[string]bool{})
	if len(deletions) != 1 {
		t.Fatalf("expected 1 deletion change, got %d", len(deletions))
	}
	want = "chore: remove stale CODEOWNERS\n\nSigned-off-by: " + testSignOff
	if got := deletions[0].Details.(map[string]any)["message"]; got != want {
		t.Errorf("delete message: got %q, want %q", got, want)
	}
}

func TestPlanRepoFiles_RendersAndDedupes(t *testing.T) {
	specs := []config.FileSpec{
		{Path: "README.md", Content: "# {{.Repo}} in {{.Org}}", Message: "chore: readme", Branch: "main"},
		{Path: "LICENSE", Content: "MIT\n"},
	}
	emitted := map[string]bool{}

	changes, err := planRepoFiles(context.Background(), nil, "Acme", "widgets", "widgets", specs, "", emitted)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(changes) != 2 {
		t.Fatalf("expected 2 changes, got %d", len(changes))
	}

	readmeDetails := changes[0].Details.(map[string]any)
	if readmeDetails["content"] != "# widgets in Acme" {
		t.Errorf("expected rendered README content, got %q", readmeDetails["content"])
	}
	if readmeDetails["message"] != "chore: readme" {
		t.Errorf("expected custom commit message, got %q", readmeDetails["message"])
	}
	if readmeDetails["branch"] != "main" {
		t.Errorf("expected branch main, got %q", readmeDetails["branch"])
	}

	licenseDetails := changes[1].Details.(map[string]any)
	if licenseDetails["message"] != "chore: add LICENSE" {
		t.Errorf("expected default commit message, got %q", licenseDetails["message"])
	}
	if licenseDetails["branch"] != "main" {
		t.Errorf("expected default branch main, got %q", licenseDetails["branch"])
	}

	// Calling again should be a no-op because emitted tracks both paths now.
	more, err := planRepoFiles(context.Background(), nil, "Acme", "widgets", "widgets", specs, "", emitted)
	if err != nil {
		t.Fatalf("unexpected error on second call: %v", err)
	}
	if len(more) != 0 {
		t.Errorf("expected no new changes on second call, got %d", len(more))
	}
}

func TestPlanRepoFiles_OnlyGlobSkipsNonMatch(t *testing.T) {
	specs := []config.FileSpec{
		{Path: "LICENSE", Content: "MIT", Only: []string{"public-*"}},
	}
	changes, err := planRepoFiles(context.Background(), nil, "Acme", "internal-api", "internal-api", specs, "", map[string]bool{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(changes) != 0 {
		t.Errorf("expected no changes for non-matching repo, got %d", len(changes))
	}

	changes, err = planRepoFiles(context.Background(), nil, "Acme", "public-docs", "public-docs", specs, "", map[string]bool{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(changes) != 1 {
		t.Errorf("expected 1 change for matching repo, got %d", len(changes))
	}
}

func TestPlanRepoFiles_BadTemplatePropagates(t *testing.T) {
	specs := []config.FileSpec{{Path: "bad.md", Content: "{{.Missing}}"}}
	_, err := planRepoFiles(context.Background(), nil, "Acme", "widgets", "widgets", specs, "", map[string]bool{})
	if err == nil {
		t.Fatal("expected template error")
	}
}

func TestMaterializeFileSpecs_PreservesUserOrder(t *testing.T) {
	app := config.AppConfig{
		Files: []config.FileSpec{
			{Path: "LICENSE", Content: "MIT"},
			{Path: "CODEOWNERS", Content: "* @team"},
		},
	}
	files := materializeFileSpecs(app)
	if len(files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(files))
	}
	if files[0].Path != "LICENSE" || files[1].Path != "CODEOWNERS" {
		t.Errorf("expected LICENSE, CODEOWNERS order, got %q, %q", files[0].Path, files[1].Path)
	}
}

func TestRenderCodeowners(t *testing.T) {
	tests := []struct {
		name   string
		owners []string
		want   string
	}{
		{"empty", nil, ""},
		{"bare username", []string{"octocat"}, "* @octocat\n"},
		{"already prefixed", []string{"@octocat"}, "* @octocat\n"},
		{"team ref", []string{"@my-org/team"}, "* @my-org/team\n"},
		{"dedup", []string{"octocat", "@octocat"}, "* @octocat\n"},
		{"multiple", []string{"a", "@b", "@org/t"}, "* @a @b @org/t\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := renderCodeowners(tt.owners)
			if got != tt.want {
				t.Errorf("renderCodeowners(%v) = %q, want %q", tt.owners, got, tt.want)
			}
		})
	}
}

func TestPlanCodeowners_EmitsPerRepo(t *testing.T) {
	owners := map[string][]string{
		"api": {"allanice001"},
		"web": {"@org/frontend"},
	}
	names := map[string]string{"api": "api", "web": "web"}
	emitted := map[string]bool{}

	changes := planCodeowners("acme", owners, names, map[string]bool{}, "", emitted)
	if len(changes) != 2 {
		t.Fatalf("expected 2 changes, got %d", len(changes))
	}

	// Deterministic order (sorted): api, web
	if changes[0].Target != "api:.github/CODEOWNERS" {
		t.Errorf("expected api first, got %q", changes[0].Target)
	}
	d := changes[0].Details.(map[string]any)
	if d["path"] != ".github/CODEOWNERS" {
		t.Errorf("expected .github/CODEOWNERS path, got %q", d["path"])
	}
	if d["content"] != "* @allanice001\n" {
		t.Errorf("unexpected api content: %q", d["content"])
	}
	if d["branch"] != "main" {
		t.Errorf("expected branch main, got %q", d["branch"])
	}
	if d["reconcile"] != true {
		t.Errorf("expected synthesized CODEOWNERS to be reconciled, got %v", d["reconcile"])
	}
	if !emitted["api:.github/CODEOWNERS"] {
		t.Error("expected emitted set to be updated")
	}
}

func TestPlanCodeowners_SkipsWhenUserDeclared(t *testing.T) {
	owners := map[string][]string{"api": {"octocat"}}
	names := map[string]string{"api": "api"}
	userFiles := map[string]bool{".github/CODEOWNERS": true}

	changes := planCodeowners("acme", owners, names, userFiles, "", map[string]bool{})
	if len(changes) != 0 {
		t.Errorf("expected user-declared CODEOWNERS to win, got %d synthesized changes", len(changes))
	}
}

func TestPlanCodeowners_SkipsRepoWithoutOwners(t *testing.T) {
	owners := map[string][]string{
		"api":   {"octocat"},
		"empty": nil,
	}
	names := map[string]string{"api": "api", "empty": "empty"}
	changes := planCodeowners("acme", owners, names, map[string]bool{}, "", map[string]bool{})
	if len(changes) != 1 {
		t.Fatalf("expected 1 change (api only), got %d", len(changes))
	}
	if changes[0].Target != "api:.github/CODEOWNERS" {
		t.Errorf("expected api change, got %q", changes[0].Target)
	}
}

func TestPlanCodeowners_RespectsEmittedSet(t *testing.T) {
	owners := map[string][]string{"api": {"octocat"}}
	names := map[string]string{"api": "api"}
	emitted := map[string]bool{"api:.github/CODEOWNERS": true}

	changes := planCodeowners("acme", owners, names, map[string]bool{}, "", emitted)
	if len(changes) != 0 {
		t.Errorf("expected no changes when already emitted, got %d", len(changes))
	}
}

func TestPlanCodeownersDeletions_OnlyForReposWithoutOwners(t *testing.T) {
	managed := map[string]bool{"api": true, "web": true, "infra": true}
	names := map[string]string{"api": "api", "web": "web", "infra": "infra"}
	owners := map[string][]string{"api": {"octocat"}}

	changes := planCodeownersDeletions("acme", managed, names, owners, map[string]bool{}, "", map[string]bool{})
	if len(changes) != 2 {
		t.Fatalf("expected 2 delete changes (web, infra), got %d", len(changes))
	}
	// Sorted: infra, web
	if changes[0].Target != "infra:.github/CODEOWNERS" {
		t.Errorf("expected infra first, got %q", changes[0].Target)
	}
	if changes[0].Action != "delete" {
		t.Errorf("expected delete action, got %q", changes[0].Action)
	}
	d := changes[0].Details.(map[string]any)
	if d["path"] != ".github/CODEOWNERS" {
		t.Errorf("expected .github/CODEOWNERS path, got %q", d["path"])
	}
	if d["message"] == "" {
		t.Error("expected non-empty commit message")
	}
}

func TestPlanCodeownersDeletions_SkipsWhenUserDeclared(t *testing.T) {
	managed := map[string]bool{"api": true}
	names := map[string]string{"api": "api"}
	owners := map[string][]string{} // no owners -> would normally delete
	userFiles := map[string]bool{".github/CODEOWNERS": true}

	changes := planCodeownersDeletions("acme", managed, names, owners, userFiles, "", map[string]bool{})
	if len(changes) != 0 {
		t.Errorf("expected user-declared CODEOWNERS to suppress deletion, got %d", len(changes))
	}
}

func TestPlanCodeownersDeletions_RespectsEmittedSet(t *testing.T) {
	managed := map[string]bool{"api": true}
	names := map[string]string{"api": "api"}
	emitted := map[string]bool{"api:.github/CODEOWNERS": true}

	changes := planCodeownersDeletions("acme", managed, names, map[string][]string{}, map[string]bool{}, "", emitted)
	if len(changes) != 0 {
		t.Errorf("expected no changes when already emitted (write wins), got %d", len(changes))
	}
}

// staticProbe answers from a fixed map of "repo:path" -> content. A key that is
// absent means the file does not exist on the branch.
func staticProbe(files map[string]string) fileProbe {
	return func(_ context.Context, _, repo, path, _ string) (remoteFile, error) {
		if content, ok := files[repo+":"+path]; ok {
			return remoteFile{exists: true, content: content}, nil
		}
		return remoteFile{}, nil
	}
}

func TestFileNeedsWrite(t *testing.T) {
	tests := []struct {
		name      string
		current   remoteFile
		desired   string
		reconcile bool
		want      bool
	}{
		{name: "missing file is written", current: remoteFile{}, desired: "x", want: true},
		{
			name:    "existing file is left alone without reconcile",
			current: remoteFile{exists: true, content: "old"}, desired: "new", want: false,
		},
		{
			name:    "identical content is a no-op even with reconcile",
			current: remoteFile{exists: true, content: "same"}, desired: "same", reconcile: true, want: false,
		},
		{
			name:    "drifted content is written with reconcile",
			current: remoteFile{exists: true, content: "old"}, desired: "new", reconcile: true, want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fileNeedsWrite(tt.current, tt.desired, tt.reconcile); got != tt.want {
				t.Errorf("fileNeedsWrite = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestPlanRepoFilesSkipsFilesThatAlreadyMatch is the regression this whole
// change exists for: a reconciled file identical to what is on the branch used
// to appear in every plan forever, claiming a commit that would never happen.
func TestPlanRepoFilesSkipsFilesThatAlreadyMatch(t *testing.T) {
	specs := []config.FileSpec{{
		Path:      ".github/renovate.json",
		Content:   "{\n  \"extends\": [\"config:base\"]\n}\n",
		Reconcile: true,
	}}
	rendered := specs[0].Content

	t.Run("identical content plans nothing", func(t *testing.T) {
		probe := staticProbe(map[string]string{"infra:.github/renovate.json": rendered})
		got, err := planRepoFiles(context.Background(), probe, "myorg", "infra", "infra", specs, "", map[string]bool{})
		if err != nil {
			t.Fatalf("plan: %v", err)
		}
		if len(got) != 0 {
			t.Errorf("got %d changes, want none for a file that already matches: %+v", len(got), got)
		}
	})

	t.Run("drifted content plans a write", func(t *testing.T) {
		probe := staticProbe(map[string]string{"infra:.github/renovate.json": "{}\n"})
		got, err := planRepoFiles(context.Background(), probe, "myorg", "infra", "infra", specs, "", map[string]bool{})
		if err != nil {
			t.Fatalf("plan: %v", err)
		}
		if len(got) != 1 {
			t.Fatalf("got %d changes, want 1", len(got))
		}
	})

	t.Run("missing file plans a write", func(t *testing.T) {
		got, err := planRepoFiles(context.Background(), staticProbe(nil), "myorg", "infra", "infra", specs, "", map[string]bool{})
		if err != nil {
			t.Fatalf("plan: %v", err)
		}
		if len(got) != 1 {
			t.Fatalf("got %d changes, want 1", len(got))
		}
	})

	t.Run("existing file without reconcile plans nothing", func(t *testing.T) {
		noReconcile := []config.FileSpec{{Path: "README.md", Content: "new"}}
		probe := staticProbe(map[string]string{"infra:README.md": "hand-edited"})
		got, err := planRepoFiles(context.Background(), probe, "myorg", "infra", "infra", noReconcile, "", map[string]bool{})
		if err != nil {
			t.Fatalf("plan: %v", err)
		}
		if len(got) != 0 {
			t.Errorf("got %+v, want nothing: the file exists and drift is not tracked", got)
		}
	})

	t.Run("a nil probe plans everything", func(t *testing.T) {
		// What a repository being created in this same run gets.
		got, err := planRepoFiles(context.Background(), nil, "myorg", "infra", "infra", specs, "", map[string]bool{})
		if err != nil {
			t.Fatalf("plan: %v", err)
		}
		if len(got) != 1 {
			t.Fatalf("got %d changes, want 1", len(got))
		}
	})
}

func TestNewFileProbeReadsAndCaches(t *testing.T) {
	var calls int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/contents/present.txt") {
			atomic.AddInt32(&calls, 1)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"type": "file", "name": "present.txt", "path": "present.txt",
				"encoding": "base64",
				"content":  base64.StdEncoding.EncodeToString([]byte("hello\n")),
			})
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	probe := newFileProbe(newTestClient(t, server))

	got, err := probe(context.Background(), "myorg", "infra", "present.txt", "main")
	if err != nil {
		t.Fatalf("probe: %v", err)
	}
	if !got.exists || got.content != "hello\n" {
		t.Errorf("got %+v, want the decoded file", got)
	}

	// A second read of the same file must not hit the API again: one repository
	// can be reached through several teams.
	if _, err := probe(context.Background(), "myorg", "infra", "present.txt", "main"); err != nil {
		t.Fatalf("probe: %v", err)
	}
	if n := atomic.LoadInt32(&calls); n != 1 {
		t.Errorf("API called %d times, want 1 — the probe should memoize", n)
	}

	missing, err := probe(context.Background(), "myorg", "infra", "absent.txt", "main")
	if err != nil {
		t.Fatalf("a 404 is not an error, it means the file is missing: %v", err)
	}
	if missing.exists {
		t.Errorf("got %+v, want a missing file", missing)
	}
}
