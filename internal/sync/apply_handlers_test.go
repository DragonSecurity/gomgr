package sync

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-github/v84/github"

	"github.com/DragonSecurity/gomgr/internal/config"
	"github.com/DragonSecurity/gomgr/internal/gh"
	"github.com/DragonSecurity/gomgr/internal/util"
)

// newTestClient creates a gh.Client backed by the given httptest.Server.
func newTestClient(t *testing.T, server *httptest.Server) *gh.Client {
	t.Helper()
	client := github.NewClient(nil)
	url := server.URL + "/"
	client.BaseURL, _ = client.BaseURL.Parse(url)
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
