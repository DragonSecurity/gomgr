package gh

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v90/github"
	"golang.org/x/oauth2"

	"github.com/DragonSecurity/gomgr/internal/config"
)

type Client struct {
	REST       *github.Client
	httpClient *http.Client
	// GraphQLURL is the endpoint used by DoGraphQL. Empty means GitHub's public
	// GraphQL API. Tests may override it to point at a local server.
	GraphQLURL string
}

const defaultMaxRetries = 3
const defaultGraphQLURL = "https://api.github.com/graphql"

func NewClientFromEnv(ctx context.Context, app config.AppConfig) (*Client, string, error) {
	// PAT
	if tok := os.Getenv("GITHUB_TOKEN"); tok != "" {
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: tok})
		tc := oauth2.NewClient(ctx, ts)
		tc.Transport = newRetryTransport(tc.Transport, defaultMaxRetries)
		rest, err := github.NewClient(github.WithHTTPClient(tc))
		if err != nil {
			return nil, "", fmt.Errorf("new github client: %w", err)
		}
		return &Client{REST: rest, httpClient: tc}, "PAT", nil
	}
	// App
	appID := app.AppID
	if v := os.Getenv("GITHUB_APP_ID"); v != "" && appID == 0 {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			// Silently ignoring this used to send a typo'd env var all the way
			// to "no auth found", which says nothing about the actual mistake.
			return nil, "", fmt.Errorf("GITHUB_APP_ID is not a number: %q", v)
		}
		appID = id
	}
	key := firstNonEmpty(app.PrivateKey, os.Getenv("GITHUB_APP_PRIVATE_KEY"))
	if appID == 0 || key == "" {
		return nil, "", missingAuthError(appID, key)
	}
	pemBytes, err := maybeReadPEM(key)
	if err != nil {
		return nil, "", err
	}
	atr, err := ghinstallation.NewAppsTransport(http.DefaultTransport, appID, pemBytes)
	if err != nil {
		return nil, "", fmt.Errorf("app transport: %w", err)
	}
	tmp, err := github.NewClient(github.WithHTTPClient(&http.Client{Transport: atr}))
	if err != nil {
		return nil, "", fmt.Errorf("new github client: %w", err)
	}
	inst, _, err := tmp.Apps.GetOrganizationInstallation(ctx, app.Org)
	if err != nil {
		return nil, "", fmt.Errorf("find installation for org %q: %w", app.Org, err)
	}
	itr := ghinstallation.NewFromAppsTransport(atr, inst.GetID())
	httpClient := &http.Client{Transport: newRetryTransport(itr, defaultMaxRetries), Timeout: 30 * time.Second}
	rest, err := github.NewClient(github.WithHTTPClient(httpClient))
	if err != nil {
		return nil, "", fmt.Errorf("new github client: %w", err)
	}
	return &Client{REST: rest, httpClient: httpClient}, "Github App", nil
}

// missingAuthError explains which half of the App credentials is missing, and
// every place the missing half can come from.
//
// The two halves are not equally sensitive, and the message says so: an App ID
// identifies the app but grants nothing, so it belongs in a committed app.yaml,
// while the private key is the actual credential and should arrive by flag or
// environment. Saying only "no auth found" leaves someone who supplied one of
// the two with no idea which one failed.
func missingAuthError(appID int64, key string) error {
	switch {
	case appID == 0 && key == "":
		return errors.New("no GitHub credentials found. Either export GITHUB_TOKEN for a personal " +
			"access token, or supply a GitHub App: --app-id and --private-key, app_id and private_key " +
			"in app.yaml, or GITHUB_APP_ID and GITHUB_APP_PRIVATE_KEY")
	case appID == 0:
		return errors.New("a GitHub App private key was supplied but no App ID. Pass --app-id, " +
			"or set app_id in app.yaml — an App ID is not a secret, so it is safe to commit alongside " +
			"the org name — or export GITHUB_APP_ID")
	default:
		return fmt.Errorf("GitHub App ID %d was supplied but no private key. Pass --private-key <path>, "+
			"set private_key in app.yaml, or export GITHUB_APP_PRIVATE_KEY", appID)
	}
}

func maybeReadPEM(s string) ([]byte, error) {
	var (
		data   []byte
		source string
	)
	if strings.Contains(s, "BEGIN") {
		data = []byte(s)
		source = "inline key"
	} else {
		b, err := os.ReadFile(s)
		if err != nil {
			return nil, err
		}
		data = b
		source = s
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("invalid PEM at %s", source)
	}
	if !isPrivateKeyBlockType(block.Type) {
		return nil, fmt.Errorf("invalid PEM at %s: expected a private key block, got %q", source, block.Type)
	}
	return data, nil
}

func isPrivateKeyBlockType(t string) bool {
	switch t {
	case "RSA PRIVATE KEY", "PRIVATE KEY", "EC PRIVATE KEY":
		return true
	}
	return false
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

// DoGraphQL executes a GraphQL query or mutation
func (c *Client) DoGraphQL(ctx context.Context, query string, variables map[string]any, result any) error {
	if c == nil || c.httpClient == nil {
		return fmt.Errorf("graphql client httpClient is nil")
	}
	if strings.TrimSpace(query) == "" {
		return fmt.Errorf("graphql query must not be empty")
	}
	if ctx == nil {
		return fmt.Errorf("context must not be nil")
	}

	reqBody := map[string]any{
		"query": query,
	}
	if len(variables) > 0 {
		reqBody["variables"] = variables
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshal graphql request: %w", err)
	}

	url := c.GraphQLURL
	if url == "" {
		url = defaultGraphQLURL
	}
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create graphql request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("execute graphql request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("graphql request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response to check for GraphQL errors
	var gqlResp struct {
		Data   json.RawMessage `json:"data"`
		Errors []struct {
			Message string `json:"message"`
			Path    []any  `json:"path,omitempty"`
		} `json:"errors"`
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read graphql response: %w", err)
	}

	if err := json.Unmarshal(respBody, &gqlResp); err != nil {
		return fmt.Errorf("decode graphql response: %w", err)
	}

	// Check for GraphQL errors
	if len(gqlResp.Errors) > 0 {
		msgs := make([]string, len(gqlResp.Errors))
		for i, e := range gqlResp.Errors {
			msgs[i] = e.Message
		}
		return fmt.Errorf("graphql error: %s", strings.Join(msgs, "; "))
	}

	if result != nil && len(gqlResp.Data) > 0 {
		if err := json.Unmarshal(gqlResp.Data, result); err != nil {
			return fmt.Errorf("decode graphql data: %w", err)
		}
	}

	return nil
}

// NewClientWithHTTP builds a Client around an HTTP client the caller already
// has.
//
// DoGraphQL needs that client and the field is unexported, so without this
// there is no way to construct a GraphQL-capable Client outside this package —
// which left the GraphQL paths untestable.
func NewClientWithHTTP(rest *github.Client, httpClient *http.Client, graphQLURL string) *Client {
	return &Client{REST: rest, httpClient: httpClient, GraphQLURL: graphQLURL}
}
