package gh

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/DragonSecurity/gomgr/internal/config"
	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v75/github"
	"golang.org/x/oauth2"
)

type Client struct{ REST *github.Client }

func NewClientFromEnv(ctx context.Context, app config.AppConfig) (*Client, string, error) {
	// PAT
	if tok := os.Getenv("GITHUB_TOKEN"); tok != "" {
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: tok})
		tc := oauth2.NewClient(ctx, ts)
		return &Client{REST: github.NewClient(tc)}, "PAT", nil
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
	return &Client{REST: github.NewClient(httpClient)}, "Github App", nil
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
	if _, err := x509.ParsePKCS1PrivateKey(block.Bytes); err != nil {
		_, _ = x509.ParsePKCS8PrivateKey(block.Bytes)
	}
	return b, nil
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
