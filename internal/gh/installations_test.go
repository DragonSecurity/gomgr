package gh

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-github/v90/github"

	"github.com/DragonSecurity/gomgr/internal/config"
)

// installationsServer answers GET /app/installations, paginating when more than
// one page of fixtures is supplied.
func installationsServer(t *testing.T, pages [][]map[string]any) *httptest.Server {
	t.Helper()
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/app/installations" {
			http.NotFound(w, r)
			return
		}
		page := 1
		if p := r.URL.Query().Get("page"); p != "" {
			_, _ = fmt.Sscanf(p, "%d", &page)
		}
		if page < 1 || page > len(pages) {
			_ = json.NewEncoder(w).Encode([]map[string]any{})
			return
		}
		if page < len(pages) {
			w.Header().Set("Link", fmt.Sprintf(`<%s/app/installations?page=%d>; rel="next"`, srv.URL, page+1))
		}
		_ = json.NewEncoder(w).Encode(pages[page-1])
	}))
	t.Cleanup(srv.Close)
	return srv
}

func clientFor(t *testing.T, srv *httptest.Server) *github.Client {
	t.Helper()
	url := srv.URL + "/"
	c, err := github.NewClient(github.WithURLs(&url, &url))
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	return c
}

func TestListInstallations(t *testing.T) {
	srv := installationsServer(t, [][]map[string]any{{
		{"id": 22, "repository_selection": "selected", "account": map[string]any{"login": "Zebra"}},
		{"id": 11, "repository_selection": "all", "account": map[string]any{"login": "AlphaOrg"}},
	}})

	got, err := ListInstallations(context.Background(), clientFor(t, srv))
	if err != nil {
		t.Fatalf("ListInstallations: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 installations, got %+v", got)
	}
	// Sorted by org, and lowercased to match every other org comparison.
	if got[0].Org != "alphaorg" || got[1].Org != "zebra" {
		t.Errorf("expected sorted, lowercased orgs, got %q and %q", got[0].Org, got[1].Org)
	}
	if got[0].ID != 11 || got[0].RepositorySelection != "all" {
		t.Errorf("installation detail lost: %+v", got[0])
	}
	// "selected" is a different answer from "all": the app is on the org but
	// may still not see the repository a config names.
	if got[1].RepositorySelection != "selected" {
		t.Errorf("repository_selection lost: %+v", got[1])
	}
}

func TestListInstallationsPaginates(t *testing.T) {
	srv := installationsServer(t, [][]map[string]any{
		{{"id": 1, "account": map[string]any{"login": "one"}}},
		{{"id": 2, "account": map[string]any{"login": "two"}}},
	})

	got, err := ListInstallations(context.Background(), clientFor(t, srv))
	if err != nil {
		t.Fatalf("ListInstallations: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected both pages, got %+v", got)
	}
}

func TestListInstallationsEmpty(t *testing.T) {
	srv := installationsServer(t, [][]map[string]any{{}})

	got, err := ListInstallations(context.Background(), clientFor(t, srv))
	if err != nil {
		t.Fatalf("ListInstallations: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected none, got %+v", got)
	}
}

// A PAT authenticates a person, and a person has no app installations. Callers
// need to tell that apart from a real failure, so it is a sentinel.
func TestAppClientWithoutCredentials(t *testing.T) {
	t.Setenv("GITHUB_APP_ID", "")
	t.Setenv("GITHUB_APP_PRIVATE_KEY", "")

	_, err := AppClient(config.AppConfig{Org: "myorg"})
	if !errors.Is(err, ErrNoAppCredentials) {
		t.Fatalf("want ErrNoAppCredentials, got %v", err)
	}
}

// "find installation for org X" alone does not distinguish a typo from an app
// nobody installed, and the answer to both is a list this credential can fetch.
func TestNotInstalledErrorNamesTheInstalledOrgs(t *testing.T) {
	srv := installationsServer(t, [][]map[string]any{{
		{"id": 1, "account": map[string]any{"login": "dragondevcc"}},
		{"id": 2, "account": map[string]any{"login": "dragonsecurity"}},
	}})

	cause := errors.New("404 Not Found")
	err := notInstalledError(context.Background(), clientFor(t, srv), "kamuses", cause)

	msg := err.Error()
	if !strings.Contains(msg, "kamuses") {
		t.Errorf("the failing org must still be named: %s", msg)
	}
	if !strings.Contains(msg, "dragondevcc") || !strings.Contains(msg, "dragonsecurity") {
		t.Errorf("the installed orgs should be listed: %s", msg)
	}
	if !errors.Is(err, cause) {
		t.Error("the original cause must stay wrapped")
	}
}

// The list is a best effort. If fetching it also fails, the original error is
// what matters and must not be replaced by a complaint about the listing.
func TestNotInstalledErrorFallsBackWhenListingFails(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)

	cause := errors.New("404 Not Found")
	err := notInstalledError(context.Background(), clientFor(t, srv), "kamuses", cause)

	if !errors.Is(err, cause) {
		t.Fatalf("the original cause must survive, got %v", err)
	}
	if strings.Contains(err.Error(), "installed on:") {
		t.Errorf("no list should be claimed when it could not be fetched: %s", err)
	}
}
