package cmd

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/DragonSecurity/gomgr/internal/config"
	"github.com/DragonSecurity/gomgr/internal/gh"
	"github.com/DragonSecurity/gomgr/internal/jit"
	"github.com/DragonSecurity/gomgr/internal/util"
	"github.com/spf13/cobra"
)

var (
	accessOrg      string
	accessDuration string
	accessStateDir string
)

var accessCmd = &cobra.Command{
	Use:   "access <target> <user> [permission]",
	Short: "Grant temporary (JIT) access to a team or repository",
	Long: `Grant temporary Just-In-Time (JIT) access to a team or repository.

Access is automatically revoked after the specified duration (default: 1 hour).

Examples:
  # Grant team access for 1 hour (default)
  gomgr access developer allanice001 --org myorg
  gomgr access developer allanice001 maintainer --org myorg

  # Grant repo access for 1 hour (default)
  gomgr access myrepo allanice001 push --org myorg

  # Grant access for a custom duration
  gomgr access myrepo allanice001 push --org myorg --duration 2h
  gomgr access developer allanice001 --org myorg --duration 30m`,
	Args: cobra.RangeArgs(2, 3),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		if debug {
			util.EnableDebug()
		}

		target := args[0]
		user := args[1]
		permission := ""
		if len(args) == 3 {
			permission = args[2]
		}

		if accessOrg == "" {
			return fmt.Errorf("--org is required")
		}

		// Parse duration
		duration, err := time.ParseDuration(accessDuration)
		if err != nil {
			return fmt.Errorf("invalid duration %q: %w", accessDuration, err)
		}
		if duration <= 0 {
			return fmt.Errorf("duration must be positive")
		}

		// Initialize GitHub client (we need minimal app.yaml or env vars)
		appCfg := config.AppConfig{Org: accessOrg}
		client, appInfo, err := gh.NewClientFromEnv(ctx, appCfg)
		if err != nil {
			return err
		}
		if appInfo != "" {
			log.Printf("auth: %s", appInfo)
		}

		// Initialize JIT state
		state, err := jit.NewState(accessStateDir)
		if err != nil {
			return fmt.Errorf("init JIT state: %w", err)
		}

		// Clean up expired grants first
		log.Println("checking for expired grants...")
		if err := jit.CleanupExpiredGrants(ctx, client, state); err != nil {
			log.Printf("warning: cleanup failed: %v", err)
		}

		// Determine if this is team or repo access
		// For simplicity: if permission looks like a team role (member/maintainer), treat as team
		// Otherwise, treat as repo
		grantType := "repo"
		if permission == "" || permission == "member" || permission == "maintainer" {
			grantType = "team"
			if permission == "" {
				permission = "member"
			}
		}

		// Grant access
		log.Printf("granting %s access to %s for user %s (duration: %s)...", grantType, target, user, duration)
		switch grantType {
		case "team":
			err = jit.GrantTeamAccess(ctx, client, accessOrg, target, user, permission, duration)
		case "repo":
			err = jit.GrantRepoAccess(ctx, client, accessOrg, target, user, permission)
		}
		if err != nil {
			return err
		}

		// Record the grant
		grant := &jit.Grant{
			Type:       grantType,
			User:       user,
			Org:        accessOrg,
			Target:     target,
			Permission: permission,
			GrantedAt:  time.Now(),
			ExpiresAt:  time.Now().Add(duration),
		}
		if err := state.AddGrant(grant); err != nil {
			return fmt.Errorf("failed to record grant: %w", err)
		}

		log.Printf("✓ access granted to %s", user)
		log.Printf("  type: %s", grantType)
		log.Printf("  target: %s", target)
		log.Printf("  permission: %s", permission)
		log.Printf("  expires: %s", grant.ExpiresAt.Format(time.RFC3339))
		log.Println()
		log.Printf("Access will be automatically revoked after %s", duration)
		log.Printf("Run 'gomgr cleanup-jit --org %s' to manually clean up expired grants", accessOrg)

		return nil
	},
}

func init() {
	accessCmd.Flags().StringVar(&accessOrg, "org", "", "GitHub organization (required)")
	_ = accessCmd.MarkFlagRequired("org")
	accessCmd.Flags().StringVar(&accessDuration, "duration", "1h", "Access duration (e.g., 30m, 1h, 2h)")
	accessCmd.Flags().StringVar(&accessStateDir, "state-dir", "", "Directory to store JIT state (default: ~/.gomgr)")
	rootCmd.AddCommand(accessCmd)
}
