package cmd

import (
	"context"

	"github.com/DragonSecurity/gomgr/internal/config"
	"github.com/DragonSecurity/gomgr/internal/gh"
	"github.com/DragonSecurity/gomgr/internal/util"
)

// newClient authenticates against GitHub for the loaded configuration,
// applying any credentials given on the command line first.
//
// Credentials are resolved in the order --app-id/--private-key, then app.yaml,
// then the environment. That order exists because a config directory shared
// through a repository has nowhere safe to keep a private key: the committed
// app.yaml names the org and the behavior flags, CI supplies the credentials
// as secrets, and a developer running the same directory locally supplies them
// on the command line.
//
// There is deliberately no --token flag. An App ID is not a secret and a
// private key is passed by path, but a personal access token on argv is visible
// in `ps` and lands in shell history, so GITHUB_TOKEN stays the only way in.
func newClient(ctx context.Context, cfg *config.Root) (*gh.Client, error) {
	if appIDFlag != 0 {
		cfg.App.AppID = appIDFlag
	}
	if privateKeyFlag != "" {
		cfg.App.PrivateKey = privateKeyFlag
	}

	client, authKind, err := gh.NewClientFromEnv(ctx, cfg.App)
	if err != nil {
		return nil, err
	}
	if authKind != "" {
		util.Infof("auth: %s", authKind)
	}
	return client, nil
}
