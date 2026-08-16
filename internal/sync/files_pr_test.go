package sync

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/DragonSecurity/gomgr/internal/gh"
	"github.com/DragonSecurity/gomgr/internal/util"
)

// prServer records the calls a pull-request-routed file change makes.
type prServer struct {
	mu       sync.Mutex
	calls    []string
	branches map[string]bool
	openPRs  int
	fileOn   map[string]string // branch:path -> content
	graphQL  []string
}

func newPRServer(t *testing.T) (*prServer, *gh.Client, func()) {
	t.Helper()
	rec := &prServer{branches: map[string]bool{"main": true}, fileOn: map[string]string{}}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		rec.mu.Lock()
		rec.calls = append(rec.calls, r.Method+" "+r.URL.Path)
		rec.mu.Unlock()
		path := r.URL.Path

		switch {
		case strings.Contains(path, "/git/ref/heads/"):
			branch := strings.SplitN(path, "/git/ref/heads/", 2)[1]
			if !rec.branches[branch] {
				http.NotFound(w, r)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"ref": "refs/heads/" + branch, "object": map[string]any{"sha": "basesha"},
			})

		case r.Method == http.MethodPost && strings.HasSuffix(path, "/git/refs"):
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			ref, _ := body["ref"].(string)
			rec.branches[strings.TrimPrefix(ref, "refs/heads/")] = true
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]any{"ref": ref})

		case strings.Contains(path, "/contents/"):
			file := strings.SplitN(path, "/contents/", 2)[1]
			if r.Method == http.MethodGet {
				branch := r.URL.Query().Get("ref")
				content, ok := rec.fileOn[branch+":"+file]
				if !ok {
					http.NotFound(w, r)
					return
				}
				_ = json.NewEncoder(w).Encode(map[string]any{
					"type": "file", "path": file, "sha": "filesha",
					"encoding": "base64", "content": base64.StdEncoding.EncodeToString([]byte(content)),
				})
				return
			}
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			branch, _ := body["branch"].(string)
			raw, _ := body["content"].(string)
			decoded, _ := base64.StdEncoding.DecodeString(raw)
			rec.fileOn[branch+":"+file] = string(decoded)
			_ = json.NewEncoder(w).Encode(map[string]any{"content": map[string]any{"path": file}})

		case strings.HasSuffix(path, "/pulls") && r.Method == http.MethodGet:
			if rec.openPRs == 0 {
				_ = json.NewEncoder(w).Encode([]map[string]any{})
				return
			}
			_ = json.NewEncoder(w).Encode([]map[string]any{
				{"number": 7, "node_id": "PR_kwexisting", "html_url": "https://github.com/myorg/infra/pull/7"},
			})

		case strings.HasSuffix(path, "/pulls") && r.Method == http.MethodPost:
			rec.openPRs++
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"number": 7, "node_id": "PR_kwnew", "html_url": "https://github.com/myorg/infra/pull/7",
			})

		case path == "/graphql":
			body, _ := io_ReadAll(r)
			rec.mu.Lock()
			rec.graphQL = append(rec.graphQL, body)
			rec.mu.Unlock()
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{}})

		default:
			_ = json.NewEncoder(w).Encode(map[string]any{})
		}
	})

	server := httptest.NewServer(mux)
	return rec, newTestClient(t, server), server.Close
}

func io_ReadAll(r *http.Request) (string, error) {
	buf := make([]byte, r.ContentLength)
	_, err := r.Body.Read(buf)
	return string(buf), err
}

func prChange(final bool) util.Change {
	return util.Change{
		Scope:  scopeRepoFilePR,
		Target: "infra:.github/renovate.json",
		Action: "ensure",
		Details: map[string]any{
			"org": "myorg", "repo": "infra", "path": ".github/renovate.json",
			"content": "{}\n", "message": "chore: sync Renovate config\n\nSigned-off-by: A <a@b.c>",
			"branch": "main", "head_branch": "gomgr/sync-files",
			"final": final,
		},
	}
}

func TestApplyRepoFilePullRequestOpensAndAutoMerges(t *testing.T) {
	rec, c, done := newPRServer(t)
	defer done()

	if err := applyRepoFilePullRequest(context.Background(), c, prChange(true)); err != nil {
		t.Fatalf("apply: %v", err)
	}

	if !rec.branches["gomgr/sync-files"] {
		t.Error("the head branch was not created")
	}
	if got := rec.fileOn["gomgr/sync-files:.github/renovate.json"]; got != "{}\n" {
		t.Errorf("file on head branch = %q, want the rendered content", got)
	}
	if rec.openPRs != 1 {
		t.Errorf("opened %d pull requests, want 1", rec.openPRs)
	}
	if len(rec.graphQL) != 1 || !strings.Contains(rec.graphQL[0], "enablePullRequestAutoMerge") {
		t.Errorf("auto-merge was not enabled: %v", rec.graphQL)
	}
	if !strings.Contains(rec.graphQL[0], "SQUASH") {
		t.Errorf("merge method not passed through: %v", rec.graphQL)
	}
	// Nothing may touch the base branch directly.
	for _, call := range rec.calls {
		if strings.Contains(call, "ref=main") {
			t.Errorf("unexpected write against the base branch: %s", call)
		}
	}
}

// TestApplyRepoFilePullRequestDefersAutoMerge is the ordering guarantee: a
// non-final change must not enable auto-merge, or GitHub could merge the pull
// request while gomgr is still adding commits and the rest would never land.
func TestApplyRepoFilePullRequestDefersAutoMerge(t *testing.T) {
	rec, c, done := newPRServer(t)
	defer done()

	if err := applyRepoFilePullRequest(context.Background(), c, prChange(false)); err != nil {
		t.Fatalf("apply: %v", err)
	}
	if len(rec.graphQL) != 0 {
		t.Errorf("auto-merge enabled on a non-final change: %v", rec.graphQL)
	}
}

func TestApplyRepoFilePullRequestReusesOpenPR(t *testing.T) {
	rec, c, done := newPRServer(t)
	defer done()
	rec.openPRs = 1 // a pull request from an earlier run is already open

	if err := applyRepoFilePullRequest(context.Background(), c, prChange(true)); err != nil {
		t.Fatalf("apply: %v", err)
	}
	for _, call := range rec.calls {
		if call == "POST /repos/myorg/infra/pulls" {
			t.Error("opened a second pull request instead of adding to the open one")
		}
	}
}

func TestApplyRepoFilePullRequestSkipsIdenticalContent(t *testing.T) {
	rec, c, done := newPRServer(t)
	defer done()
	rec.branches["gomgr/sync-files"] = true
	rec.fileOn["gomgr/sync-files:.github/renovate.json"] = "{}\n"

	if err := applyRepoFilePullRequest(context.Background(), c, prChange(true)); err != nil {
		t.Fatalf("apply: %v", err)
	}
	for _, call := range rec.calls {
		if strings.HasPrefix(call, "PUT /repos/myorg/infra/contents/") {
			t.Error("wrote an empty commit for content already on the branch")
		}
	}
}

func TestAutoMergeFailureDoesNotFailTheSync(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "/git/ref/heads/gomgr/sync-files"):
			http.NotFound(w, r)
		case strings.Contains(r.URL.Path, "/git/ref/heads/main"):
			_ = json.NewEncoder(w).Encode(map[string]any{"object": map[string]any{"sha": "basesha"}})
		case strings.Contains(r.URL.Path, "/contents/"):
			if r.Method == http.MethodGet {
				http.NotFound(w, r)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{})
		case strings.HasSuffix(r.URL.Path, "/pulls") && r.Method == http.MethodGet:
			_ = json.NewEncoder(w).Encode([]map[string]any{})
		case strings.HasSuffix(r.URL.Path, "/pulls"):
			_ = json.NewEncoder(w).Encode(map[string]any{"number": 7, "node_id": "PR_kw"})
		case r.URL.Path == "/graphql":
			// What GitHub says when auto-merge is off for the repository.
			_ = json.NewEncoder(w).Encode(map[string]any{
				"errors": []map[string]any{{"message": "Pull request Auto merge is not allowed for this repository"}},
			})
		default:
			_ = json.NewEncoder(w).Encode(map[string]any{})
		}
	}))
	defer server.Close()

	c := newTestClient(t, server)

	if err := applyRepoFilePullRequest(context.Background(), c, prChange(true)); err != nil {
		t.Fatalf("a repository without auto-merge must still get its pull request: %v", err)
	}
}

func TestSplitCommitMessage(t *testing.T) {
	title, body := splitCommitMessage("chore: sync Renovate config\n\nSigned-off-by: A <a@b.c>")
	if title != "chore: sync Renovate config" {
		t.Errorf("title = %q", title)
	}
	if !strings.Contains(body, "Signed-off-by: A <a@b.c>") {
		t.Errorf("body must carry the trailer so the merge commit keeps it: %q", body)
	}
}

func TestMarkFinalPullRequestFiles(t *testing.T) {
	mk := func(repo, path string) util.Change {
		return util.Change{Scope: scopeRepoFilePR, Details: map[string]any{"org": "myorg", "repo": repo, "path": path}}
	}
	changes := []util.Change{
		mk("infra", "a"), mk("infra", "b"),
		{Scope: "repo-file", Details: map[string]any{"org": "myorg", "repo": "docs"}},
		mk("other", "a"),
	}
	markFinalPullRequestFiles(changes)

	final := func(i int) bool {
		d, _ := changes[i].Details.(map[string]any)
		v, _ := d["final"].(bool)
		return v
	}
	if final(0) {
		t.Error("only the last change per repository is final")
	}
	if !final(1) {
		t.Error("the last infra change should be final")
	}
	if !final(3) {
		t.Error("a repository with one change has that change as final")
	}
}
