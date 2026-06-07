package sync

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/go-github/v88/github"

	"github.com/DragonSecurity/gomgr/internal/config"
	"github.com/DragonSecurity/gomgr/internal/gh"
	"github.com/DragonSecurity/gomgr/internal/util"
)

// newTestClient creates a gh.Client backed by the given httptest.Server.
func newTestClient(t *testing.T, server *httptest.Server) *gh.Client {
	t.Helper()
	url := server.URL + "/"
	client, err := github.NewClient(github.WithURLs(&url, &url))
	if err != nil {
		t.Fatalf("new github client: %v", err)
	}
	return &gh.Client{REST: client}
}

func TestApplyTeamCreate(t *testing.T) {
	var gotBody map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" && r.URL.Path == "/orgs/myorg/teams" {
			_ = json.NewDecoder(r.Body).Decode(&gotBody)
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]any{"id": 1, "slug": "backend"})
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	ch := util.Change{
		Scope:  "team",
		Target: "backend",
		Action: "create",
		Details: map[string]any{
			"org":         "myorg",
			"name":        "Backend",
			"privacy":     "closed",
			"description": "Backend team",
		},
	}

	err := applyTeamCreate(context.Background(), c, ch)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotBody["name"] != "Backend" {
		t.Errorf("expected name=Backend, got %v", gotBody["name"])
	}
}

func TestApplyTeamDelete(t *testing.T) {
	deleted := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "DELETE" && r.URL.Path == "/orgs/myorg/teams/old-team" {
			deleted = true
			w.WriteHeader(http.StatusNoContent)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	ch := util.Change{
		Scope:  "team",
		Target: "old-team",
		Action: "delete",
		Details: map[string]any{
			"org":  "myorg",
			"slug": "old-team",
		},
	}

	err := applyTeamDelete(context.Background(), c, ch)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !deleted {
		t.Error("expected DELETE request to be made")
	}
}

func TestApplyRepoEnsure(t *testing.T) {
	t.Run("regular repo", func(t *testing.T) {
		var gotBody map[string]any
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "POST" && r.URL.Path == "/orgs/myorg/repos" {
				_ = json.NewDecoder(r.Body).Decode(&gotBody)
				w.WriteHeader(http.StatusCreated)
				_ = json.NewEncoder(w).Encode(map[string]any{"id": 1, "name": "api"})
				return
			}
			http.NotFound(w, r)
		}))
		defer server.Close()

		c := newTestClient(t, server)
		ch := util.Change{
			Scope:  "repo",
			Target: "api",
			Action: "ensure",
			Details: map[string]any{
				"org":     "myorg",
				"name":    "api",
				"private": true,
			},
		}

		err := applyRepoEnsure(context.Background(), c, ch)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if gotBody["name"] != "api" {
			t.Errorf("expected name=api, got %v", gotBody["name"])
		}
	})

	t.Run("from template", func(t *testing.T) {
		created := false
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "POST" && r.URL.Path == "/repos/myorg/template-go/generate" {
				created = true
				w.WriteHeader(http.StatusCreated)
				_ = json.NewEncoder(w).Encode(map[string]any{"id": 2, "name": "new-api"})
				return
			}
			http.NotFound(w, r)
		}))
		defer server.Close()

		c := newTestClient(t, server)
		ch := util.Change{
			Scope:  "repo",
			Target: "new-api",
			Action: "ensure",
			Details: map[string]any{
				"org":     "myorg",
				"name":    "new-api",
				"private": true,
				"from":    "template-go",
			},
		}

		err := applyRepoEnsure(context.Background(), c, ch)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !created {
			t.Error("expected template creation request to be made")
		}
	})

	t.Run("422 name already exists is swallowed", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "POST" && r.URL.Path == "/orgs/myorg/repos" {
				w.WriteHeader(http.StatusUnprocessableEntity)
				_ = json.NewEncoder(w).Encode(map[string]any{
					"message": "Repository creation failed.",
					"errors": []map[string]any{
						{"resource": "Repository", "code": "custom", "field": "name", "message": "name already exists on this account"},
					},
				})
				return
			}
			http.NotFound(w, r)
		}))
		defer server.Close()

		c := newTestClient(t, server)
		ch := util.Change{
			Scope: "repo", Target: "api", Action: "ensure",
			Details: map[string]any{"org": "myorg", "name": "api", "private": true},
		}

		if err := applyRepoEnsure(context.Background(), c, ch); err != nil {
			t.Fatalf("expected 'already exists' 422 to be swallowed, got: %v", err)
		}
	})

	t.Run("422 other validation error surfaces", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "POST" && r.URL.Path == "/orgs/myorg/repos" {
				w.WriteHeader(http.StatusUnprocessableEntity)
				_ = json.NewEncoder(w).Encode(map[string]any{
					"message": "Repository creation failed.",
					"errors": []map[string]any{
						{"resource": "Repository", "code": "custom", "field": "name", "message": "name is reserved"},
					},
				})
				return
			}
			http.NotFound(w, r)
		}))
		defer server.Close()

		c := newTestClient(t, server)
		ch := util.Change{
			Scope: "repo", Target: "houston", Action: "ensure",
			Details: map[string]any{"org": "myorg", "name": "houston", "private": true},
		}

		err := applyRepoEnsure(context.Background(), c, ch)
		if err == nil {
			t.Fatal("expected non-'already exists' 422 to surface as an error, got nil")
		}
		if !containsSubstr(err.Error(), "create repo") {
			t.Errorf("expected 'create repo' in error, got: %v", err)
		}
	})
}

func TestApplyRepoFileEnsure_RaceCondition(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		errBody    map[string]any
		wantErr    bool
	}{
		{
			name:       "422 sha not supplied (race condition)",
			statusCode: 422,
			errBody: map[string]any{
				"message": `"sha" wasn't supplied`,
			},
			wantErr: false,
		},
		{
			name:       "409 reference already exists (race condition)",
			statusCode: 409,
			errBody: map[string]any{
				"message": "reference already exists",
			},
			wantErr: false,
		},
		{
			name:       "422 unrelated error",
			statusCode: 422,
			errBody: map[string]any{
				"message": "Validation Failed",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// GetContents returns 404 (file not found)
				if r.Method == "GET" {
					w.WriteHeader(http.StatusNotFound)
					_ = json.NewEncoder(w).Encode(map[string]any{"message": "Not Found"})
					return
				}
				// CreateFile returns the race condition error
				if r.Method == "PUT" {
					w.WriteHeader(tt.statusCode)
					_ = json.NewEncoder(w).Encode(tt.errBody)
					return
				}
				http.NotFound(w, r)
			}))
			defer server.Close()

			c := newTestClient(t, server)
			ch := util.Change{
				Scope:  "repo-file",
				Target: "api:README.md",
				Action: "ensure",
				Details: map[string]any{
					"org":     "myorg",
					"repo":    "api",
					"path":    "README.md",
					"content": "# API",
					"message": "add readme",
					"branch":  "main",
				},
			}

			err := applyRepoFileEnsure(context.Background(), c, ch)
			if (err != nil) != tt.wantErr {
				t.Errorf("applyRepoFileEnsure() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// codeownersFileResponse builds the JSON shape GetContents returns for a
// single existing file (base64-encoded content + sha).
func codeownersFileResponse(t *testing.T, path, sha, body string) []byte {
	t.Helper()
	resp := map[string]any{
		"name":     "CODEOWNERS",
		"path":     path,
		"sha":      sha,
		"size":     len(body),
		"type":     "file",
		"encoding": "base64",
		"content":  base64.StdEncoding.EncodeToString([]byte(body)),
	}
	b, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return b
}

func TestApplyRepoFileEnsure_ReconcileSkipsWhenContentMatches(t *testing.T) {
	const path = ".github/CODEOWNERS"
	const body = "* @octocat\n"
	var putCalled bool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" && r.URL.Path == "/repos/myorg/api/contents/.github/CODEOWNERS" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(codeownersFileResponse(t, path, "abc123", body))
			return
		}
		if r.Method == "PUT" {
			putCalled = true
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]any{})
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	ch := util.Change{
		Scope:  "repo-file",
		Target: "api:.github/CODEOWNERS",
		Action: "ensure",
		Details: map[string]any{
			"org":       "myorg",
			"repo":      "api",
			"path":      path,
			"content":   body,
			"message":   "chore: sync CODEOWNERS",
			"branch":    "main",
			"reconcile": true,
		},
	}

	if err := applyRepoFileEnsure(context.Background(), c, ch); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if putCalled {
		t.Error("expected no PUT (content matches) but UpdateFile was invoked")
	}
}

func TestApplyRepoFileEnsure_ReconcileUpdatesWhenContentDrifts(t *testing.T) {
	const path = ".github/CODEOWNERS"
	const existing = "* @old-owner\n"
	const desired = "* @new-owner\n"

	var (
		putCalled bool
		putBody   map[string]any
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" && r.URL.Path == "/repos/myorg/api/contents/.github/CODEOWNERS" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(codeownersFileResponse(t, path, "sha-of-existing", existing))
			return
		}
		if r.Method == "PUT" && r.URL.Path == "/repos/myorg/api/contents/.github/CODEOWNERS" {
			putCalled = true
			_ = json.NewDecoder(r.Body).Decode(&putBody)
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]any{})
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	ch := util.Change{
		Scope:  "repo-file",
		Target: "api:.github/CODEOWNERS",
		Action: "ensure",
		Details: map[string]any{
			"org":       "myorg",
			"repo":      "api",
			"path":      path,
			"content":   desired,
			"message":   "chore: sync CODEOWNERS",
			"branch":    "main",
			"reconcile": true,
		},
	}

	if err := applyRepoFileEnsure(context.Background(), c, ch); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !putCalled {
		t.Fatal("expected UpdateFile to be invoked for drifted content")
	}
	if putBody["sha"] != "sha-of-existing" {
		t.Errorf("expected SHA of existing file in update payload, got %v", putBody["sha"])
	}
	decoded, err := base64.StdEncoding.DecodeString(putBody["content"].(string))
	if err != nil {
		t.Fatalf("decode update payload: %v", err)
	}
	if string(decoded) != desired {
		t.Errorf("expected update payload content %q, got %q", desired, string(decoded))
	}
}

func TestApplyRepoFileDelete_NoopWhenAbsent(t *testing.T) {
	var deleteCalled bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]any{"message": "Not Found"})
			return
		}
		if r.Method == "DELETE" {
			deleteCalled = true
			w.WriteHeader(http.StatusOK)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	ch := util.Change{
		Scope:  "repo-file",
		Target: "api:.github/CODEOWNERS",
		Action: "delete",
		Details: map[string]any{
			"org":     "myorg",
			"repo":    "api",
			"path":    ".github/CODEOWNERS",
			"message": "chore: remove stale CODEOWNERS",
			"branch":  "main",
		},
	}

	if err := applyRepoFileDelete(context.Background(), c, ch); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if deleteCalled {
		t.Error("expected no DELETE when file is absent, but DeleteFile was invoked")
	}
}

func TestApplyRepoFileDelete_DeletesWhenPresent(t *testing.T) {
	const path = ".github/CODEOWNERS"
	var (
		deleteCalled bool
		deleteBody   map[string]any
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" && r.URL.Path == "/repos/myorg/api/contents/.github/CODEOWNERS" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(codeownersFileResponse(t, path, "sha-of-stale", "* @old-owner\n"))
			return
		}
		if r.Method == "DELETE" && r.URL.Path == "/repos/myorg/api/contents/.github/CODEOWNERS" {
			deleteCalled = true
			_ = json.NewDecoder(r.Body).Decode(&deleteBody)
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]any{})
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	ch := util.Change{
		Scope:  "repo-file",
		Target: "api:.github/CODEOWNERS",
		Action: "delete",
		Details: map[string]any{
			"org":     "myorg",
			"repo":    "api",
			"path":    path,
			"message": "chore: remove stale CODEOWNERS",
			"branch":  "main",
		},
	}

	if err := applyRepoFileDelete(context.Background(), c, ch); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !deleteCalled {
		t.Fatal("expected DeleteFile to be invoked")
	}
	if deleteBody["sha"] != "sha-of-stale" {
		t.Errorf("expected SHA of existing file in delete payload, got %v", deleteBody["sha"])
	}
}

func TestApplyRepoFileEnsure_NoReconcileLeavesExistingAlone(t *testing.T) {
	const path = "README.md"
	var putCalled bool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" && r.URL.Path == "/repos/myorg/api/contents/README.md" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(codeownersFileResponse(t, path, "sha-readme", "# hand-edited readme\n"))
			return
		}
		if r.Method == "PUT" {
			putCalled = true
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]any{})
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	ch := util.Change{
		Scope:  "repo-file",
		Target: "api:README.md",
		Action: "ensure",
		Details: map[string]any{
			"org":     "myorg",
			"repo":    "api",
			"path":    path,
			"content": "# template default\n",
			"message": "chore: add readme",
			"branch":  "main",
			// reconcile not set → defaults to false
		},
	}

	if err := applyRepoFileEnsure(context.Background(), c, ch); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if putCalled {
		t.Error("expected reconcile=false to leave the hand-edited file alone, but UpdateFile was invoked")
	}
}

func TestApplyRepoTopicsEnsure(t *testing.T) {
	tests := []struct {
		name   string
		topics any
	}{
		{
			name:   "string slice",
			topics: []string{"backend", "api"},
		},
		{
			name:   "any slice",
			topics: []any{"backend", "api"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method == "PUT" && r.URL.Path == "/repos/myorg/api/topics" {
					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(map[string]any{"names": []string{"backend", "api"}})
					return
				}
				http.NotFound(w, r)
			}))
			defer server.Close()

			c := newTestClient(t, server)
			ch := util.Change{
				Scope:  "repo-topics",
				Target: "api",
				Action: "ensure",
				Details: map[string]any{
					"org":    "myorg",
					"repo":   "api",
					"topics": tt.topics,
				},
			}

			err := applyRepoTopicsEnsure(context.Background(), c, ch)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestApplyHandlers_InvalidDetails(t *testing.T) {
	handlers := map[string]func(context.Context, *gh.Client, util.Change) error{
		"team:create":          applyTeamCreate,
		"team:update":          applyTeamUpdate,
		"team:delete":          applyTeamDelete,
		"repo:ensure":          applyRepoEnsure,
		"team-repo:grant":      applyTeamRepoGrant,
		"repo-file:ensure":     applyRepoFileEnsure,
		"repo-topics:ensure":   applyRepoTopicsEnsure,
		"repo-template:ensure": applyRepoTemplateEnsure,
		"repo-pin:ensure":      applyRepoPinEnsure,
		"repo:delete":          applyRepoDelete,
	}

	for key, handler := range handlers {
		t.Run(key, func(t *testing.T) {
			ch := util.Change{
				Scope:   key[:4], // doesn't matter much
				Target:  "test",
				Action:  "test",
				Details: "not-a-map", // wrong type
			}
			err := handler(context.Background(), nil, ch)
			if err == nil {
				t.Errorf("expected error for invalid details type, got nil")
			}
			if err != nil && !containsSubstr(err.Error(), "invalid details") {
				t.Errorf("expected 'invalid details' in error, got: %v", err)
			}
		})
	}

	// Test team-member:ensure separately (expects teamMemberChange, not map)
	t.Run("team-member:ensure", func(t *testing.T) {
		ch := util.Change{
			Scope:   "team-member",
			Target:  "test",
			Action:  "ensure",
			Details: "not-a-struct",
		}
		err := applyTeamMemberEnsure(context.Background(), nil, ch)
		if err == nil {
			t.Error("expected error for invalid details type, got nil")
		}
	})
}

func containsSubstr(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestApplyTeamUpdate(t *testing.T) {
	var gotBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "PATCH" && r.URL.Path == "/orgs/myorg/teams/backend" {
			_ = json.NewDecoder(r.Body).Decode(&gotBody)
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]any{"id": 1, "slug": "backend"})
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	ch := util.Change{
		Scope:  "team",
		Target: "backend",
		Action: "update",
		Details: map[string]any{
			"org":         "myorg",
			"slug":        "backend",
			"name":        "Backend",
			"description": "Updated description",
			"privacy":     "secret",
		},
	}

	err := applyTeamUpdate(context.Background(), c, ch)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotBody["name"] != "Backend" {
		t.Errorf("expected name=Backend, got %v", gotBody["name"])
	}
	if gotBody["description"] != "Updated description" {
		t.Errorf("expected description='Updated description', got %v", gotBody["description"])
	}
	if gotBody["privacy"] != "secret" {
		t.Errorf("expected privacy=secret, got %v", gotBody["privacy"])
	}
}

func TestApplyTeamMemberEnsure(t *testing.T) {
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "PUT" && r.URL.Path == "/orgs/myorg/teams/backend/memberships/alice" {
			gotPath = r.URL.Path
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]any{"state": "active", "role": "member"})
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	ch := util.Change{
		Scope:   "team-member",
		Target:  "backend",
		Action:  "ensure",
		Details: teamMemberChange{Org: "myorg", Slug: "backend", User: "alice", Role: "member"},
	}

	err := applyTeamMemberEnsure(context.Background(), c, ch)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/orgs/myorg/teams/backend/memberships/alice" {
		t.Errorf("expected PUT to memberships path, got %s", gotPath)
	}
}

func TestApplyRepoTemplateEnsure(t *testing.T) {
	var gotBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "PATCH" && r.URL.Path == "/repos/myorg/api" {
			_ = json.NewDecoder(r.Body).Decode(&gotBody)
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]any{"id": 1, "name": "api", "is_template": true})
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	ch := util.Change{
		Scope:  "repo-template",
		Target: "api",
		Action: "ensure",
		Details: map[string]any{
			"org":      "myorg",
			"repo":     "api",
			"template": true,
		},
	}

	err := applyRepoTemplateEnsure(context.Background(), c, ch)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotBody["is_template"] != true {
		t.Errorf("expected is_template=true, got %v", gotBody["is_template"])
	}
}

func TestApplyRepoPinEnsure(t *testing.T) {
	ch := util.Change{
		Scope:  "repo-pin",
		Target: "api",
		Action: "ensure",
		Details: map[string]any{
			"org":    "myorg",
			"repo":   "api",
			"pinned": true,
		},
	}

	err := applyRepoPinEnsure(context.Background(), nil, ch)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestApplyRepoDelete(t *testing.T) {
	deleted := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "DELETE" && r.URL.Path == "/repos/myorg/old-repo" {
			deleted = true
			w.WriteHeader(http.StatusNoContent)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	ch := util.Change{
		Scope:  "repo",
		Target: "old-repo",
		Action: "delete",
		Details: map[string]any{
			"org":  "myorg",
			"repo": "old-repo",
		},
	}

	err := applyRepoDelete(context.Background(), c, ch)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !deleted {
		t.Error("expected DELETE request to be made")
	}
}

func TestApplyTeamRepoGrant(t *testing.T) {
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "PUT" && r.URL.Path == "/orgs/myorg/teams/backend/repos/myorg/api" {
			gotPath = r.URL.Path
			w.WriteHeader(http.StatusNoContent)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	ch := util.Change{
		Scope:  "team-repo",
		Target: "backend/api",
		Action: "grant",
		Details: map[string]any{
			"org":        "myorg",
			"slug":       "backend",
			"repo":       "api",
			"permission": "push",
		},
	}

	err := applyTeamRepoGrant(context.Background(), c, ch)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/orgs/myorg/teams/backend/repos/myorg/api" {
		t.Errorf("expected PUT to team repos path, got %s", gotPath)
	}
}

// shortenTeamRepoGrantRetries swaps the grant backoff for near-zero delays so
// retry tests stay fast, restoring the original schedule afterward.
func shortenTeamRepoGrantRetries(t *testing.T) {
	t.Helper()
	orig := teamRepoGrantRetryDelays
	teamRepoGrantRetryDelays = []time.Duration{time.Millisecond, time.Millisecond, time.Millisecond, time.Millisecond}
	t.Cleanup(func() { teamRepoGrantRetryDelays = orig })
}

func TestApplyTeamRepoGrant_RetriesOn404(t *testing.T) {
	shortenTeamRepoGrantRetries(t)

	var calls int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "PUT" && r.URL.Path == "/orgs/myorg/teams/admins/repos/myorg/houston" {
			// Simulate eventual consistency: the freshly created repo is not
			// visible for the first two grant attempts, then succeeds.
			if atomic.AddInt32(&calls, 1) <= 2 {
				w.WriteHeader(http.StatusNotFound)
				_ = json.NewEncoder(w).Encode(map[string]any{"message": "Not Found"})
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	ch := util.Change{
		Scope:  "team-repo",
		Target: "admins/houston",
		Action: "grant",
		Details: map[string]any{
			"org":        "myorg",
			"slug":       "admins",
			"repo":       "houston",
			"permission": "admin",
		},
	}

	if err := applyTeamRepoGrant(context.Background(), c, ch); err != nil {
		t.Fatalf("expected success after retries, got: %v", err)
	}
	if got := atomic.LoadInt32(&calls); got != 3 {
		t.Errorf("expected 3 grant attempts (2 x 404 + success), got %d", got)
	}
}

func TestApplyTeamRepoGrant_ExhaustsRetries(t *testing.T) {
	shortenTeamRepoGrantRetries(t)

	var calls int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{"message": "Not Found"})
	}))
	defer server.Close()

	c := newTestClient(t, server)
	ch := util.Change{
		Scope:  "team-repo",
		Target: "admins/houston",
		Action: "grant",
		Details: map[string]any{
			"org":        "myorg",
			"slug":       "admins",
			"repo":       "houston",
			"permission": "admin",
		},
	}

	err := applyTeamRepoGrant(context.Background(), c, ch)
	if err == nil {
		t.Fatal("expected error after exhausting retries")
	}
	if want := len(teamRepoGrantRetryDelays) + 1; int(atomic.LoadInt32(&calls)) != want {
		t.Errorf("expected %d grant attempts, got %d", want, atomic.LoadInt32(&calls))
	}
	if !containsSubstr(err.Error(), "grant") {
		t.Errorf("expected 'grant' in error, got: %v", err)
	}
}

func TestApplyTeamRepoGrant_NoRetryOnNon404(t *testing.T) {
	shortenTeamRepoGrantRetries(t)

	var calls int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusUnprocessableEntity)
		_ = json.NewEncoder(w).Encode(map[string]any{"message": "Validation Failed"})
	}))
	defer server.Close()

	c := newTestClient(t, server)
	ch := util.Change{
		Scope:  "team-repo",
		Target: "admins/houston",
		Action: "grant",
		Details: map[string]any{
			"org":        "myorg",
			"slug":       "admins",
			"repo":       "houston",
			"permission": "admin",
		},
	}

	if err := applyTeamRepoGrant(context.Background(), c, ch); err == nil {
		t.Fatal("expected error for 422 response")
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Errorf("expected exactly 1 attempt (no retry on non-404), got %d", got)
	}
}

func TestApplyOrgMemberRemove(t *testing.T) {
	removed := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "DELETE" && r.URL.Path == "/orgs/myorg/memberships/bob" {
			removed = true
			w.WriteHeader(http.StatusNoContent)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	ch := util.Change{
		Scope:  "org-member",
		Target: "bob",
		Action: "remove",
		Details: map[string]any{
			"org":  "myorg",
			"user": "bob",
		},
	}

	err := applyOrgMemberRemove(context.Background(), c, ch)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !removed {
		t.Error("expected DELETE request to be made")
	}
}

func TestApplyOrgMemberRemove_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{"message": "Not Found"})
	}))
	defer server.Close()

	c := newTestClient(t, server)
	ch := util.Change{
		Scope:  "org-member",
		Target: "ghost",
		Action: "remove",
		Details: map[string]any{
			"org":  "myorg",
			"user": "ghost",
		},
	}

	err := applyOrgMemberRemove(context.Background(), c, ch)
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
	if !containsSubstr(err.Error(), "remove member") {
		t.Errorf("expected 'remove member' in error, got: %v", err)
	}
}

func TestApplyOrgMemberRemove_InvalidDetails(t *testing.T) {
	ch := util.Change{
		Scope:   "org-member",
		Target:  "test",
		Action:  "remove",
		Details: "not-a-map",
	}
	err := applyOrgMemberRemove(context.Background(), nil, ch)
	if err == nil {
		t.Error("expected error for invalid details type")
	}
}

func TestApplyCustomRoleCreate(t *testing.T) {
	var gotBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" && r.URL.Path == "/orgs/myorg/custom-repository-roles" {
			_ = json.NewDecoder(r.Body).Decode(&gotBody)
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]any{"id": 1, "name": "deployer"})
			return
		}
		// Rate limit endpoint
		if r.URL.Path == "/rate_limit" {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"resources": map[string]any{
					"core": map[string]any{"limit": 5000, "remaining": 4999, "reset": 9999999999},
				},
			})
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	changes := []util.Change{
		{
			Scope:  "custom-role",
			Target: "deployer",
			Action: "create",
			Details: customRoleChange{
				Org:         "myorg",
				Name:        "deployer",
				Description: "Deploy role",
				BaseRole:    "read",
				Permissions: []string{"manage_actions"},
			},
		},
	}

	err := applyCustomRoleChanges(context.Background(), c, changes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotBody["name"] != "deployer" {
		t.Errorf("expected name=deployer, got %v", gotBody["name"])
	}
}

func TestApplyCustomRoleUpdate(t *testing.T) {
	var gotBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "PATCH" && r.URL.Path == "/orgs/myorg/custom-repository-roles/42" {
			_ = json.NewDecoder(r.Body).Decode(&gotBody)
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]any{"id": 42, "name": "deployer"})
			return
		}
		if r.URL.Path == "/rate_limit" {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"resources": map[string]any{
					"core": map[string]any{"limit": 5000, "remaining": 4999, "reset": 9999999999},
				},
			})
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	changes := []util.Change{
		{
			Scope:  "custom-role",
			Target: "deployer",
			Action: "update",
			Details: customRoleChange{
				Org:         "myorg",
				ID:          42,
				Name:        "deployer",
				Description: "Updated desc",
				BaseRole:    "write",
				Permissions: []string{"manage_actions", "create_releases"},
			},
		},
	}

	err := applyCustomRoleChanges(context.Background(), c, changes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotBody["name"] != "deployer" {
		t.Errorf("expected name=deployer, got %v", gotBody["name"])
	}
}

func TestApplyCustomRoleDelete(t *testing.T) {
	deleted := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "DELETE" && r.URL.Path == "/orgs/myorg/custom-repository-roles/42" {
			deleted = true
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if r.URL.Path == "/rate_limit" {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"resources": map[string]any{
					"core": map[string]any{"limit": 5000, "remaining": 4999, "reset": 9999999999},
				},
			})
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	changes := []util.Change{
		{
			Scope:  "custom-role",
			Target: "deployer",
			Action: "delete",
			Details: customRoleChange{
				Org:  "myorg",
				ID:   42,
				Name: "deployer",
			},
		},
	}

	err := applyCustomRoleChanges(context.Background(), c, changes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !deleted {
		t.Error("expected DELETE request to be made")
	}
}

func TestApplyCustomRoleChanges_SkipsNonCustomRole(t *testing.T) {
	changes := []util.Change{
		{Scope: "team", Target: "backend", Action: "create", Details: map[string]any{"org": "myorg"}},
	}
	// Should not error - just skip non-custom-role changes
	err := applyCustomRoleChanges(context.Background(), nil, changes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPlanCustomRoleCleanups(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/orgs/myorg/custom-repository-roles" && r.Method == "GET" {
			resp := map[string]any{
				"total_count": 2,
				"custom_roles": []map[string]any{
					{"id": 1, "name": "deployer"},
					{"id": 2, "name": "stale-role"},
				},
			}
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	cfg := &config.Root{
		App: config.AppConfig{
			Org:                        "myorg",
			DeleteUnmanagedCustomRoles: true,
		},
		Org: config.OrgConfig{
			CustomRoles: []config.CustomRoleConfig{
				{Name: "deployer", BaseRole: "read"},
			},
		},
	}
	st := &State{Org: "myorg"}

	changes, warnings, err := planCustomRoleCleanups(context.Background(), c, cfg, st)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(warnings) != 0 {
		t.Errorf("expected 0 warnings, got %d", len(warnings))
	}
	// Should have a delete change for "stale-role"
	found := false
	for _, ch := range changes {
		if ch.Scope == "custom-role" && ch.Action == "delete" && ch.Target == "stale-role" {
			found = true
		}
	}
	if !found {
		t.Error("expected custom-role:delete change for stale-role")
	}
}
