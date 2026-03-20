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

	"github.com/DragonSecurity/gomgr/internal/config"
	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v83/github"
	"golang.org/x/oauth2"
)

type Client struct {
	REST       *github.Client
	httpClient *http.Client
}

func NewClientFromEnv(ctx context.Context, app config.AppConfig) (*Client, string, error) {
	// PAT
	if tok := os.Getenv("GITHUB_TOKEN"); tok != "" {
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: tok})
		tc := oauth2.NewClient(ctx, ts)
		return &Client{REST: github.NewClient(tc), httpClient: tc}, "PAT", nil
	}
	// App
	appID := app.AppID
	if v := os.Getenv("GITHUB_APP_ID"); v != "" && appID == 0 {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			appID = id
		}
	}
	key := firstNonEmpty(app.PrivateKey, os.Getenv("GITHUB_APP_PRIVATE_KEY"))
	if appID == 0 || key == "" {
		return nil, "", errors.New("no auth found: set GITHUB_TOKEN or app_id+private_key")
	}
	pemBytes, err := maybeReadPEM(key)
	if err != nil {
		return nil, "", err
	}
	atr, err := ghinstallation.NewAppsTransport(http.DefaultTransport, appID, pemBytes)
	if err != nil {
		return nil, "", fmt.Errorf("app transport: %w", err)
	}
	tmp := github.NewClient(&http.Client{Transport: atr})
	inst, _, err := tmp.Apps.FindOrganizationInstallation(ctx, app.Org)
	if err != nil {
		return nil, "", fmt.Errorf("find installation for org %q: %w", app.Org, err)
	}
	itr := ghinstallation.NewFromAppsTransport(atr, inst.GetID())
	httpClient := &http.Client{Transport: itr, Timeout: 30 * time.Second}
	return &Client{REST: github.NewClient(httpClient), httpClient: httpClient}, "Github App", nil
}

func maybeReadPEM(s string) ([]byte, error) {
	if strings.Contains(s, "BEGIN") {
		return []byte(s), nil
	}
	b, err := os.ReadFile(s)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(b)
	if block == nil {
		return nil, fmt.Errorf("invalid PEM at %s", s)
	}
	return b, nil
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

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.github.com/graphql", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create graphql request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("execute graphql request: %w", err)
	}
	defer resp.Body.Close()

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
