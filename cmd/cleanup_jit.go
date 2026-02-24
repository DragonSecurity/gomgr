package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/DragonSecurity/gomgr/internal/config"
	"github.com/DragonSecurity/gomgr/internal/gh"
	"github.com/DragonSecurity/gomgr/internal/jit"
	"github.com/DragonSecurity/gomgr/internal/util"
	"github.com/spf13/cobra"
)

var (
	cleanupOrg      string
	cleanupStateDir string
)

var cleanupJitCmd = &cobra.Command{
	Use:   "cleanup-jit",
	Short: "Clean up expired JIT access grants",
	Long: `Manually clean up expired Just-In-Time (JIT) access grants.

This command revokes access for all expired grants and removes them from the state.
It's useful for manual cleanup or can be run periodically in a cron job.

Example:
  gomgr cleanup-jit --org myorg`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		if debug {
			util.EnableDebug()
		}

		if cleanupOrg == "" {
			return fmt.Errorf("--org is required")
		}

		// Initialize GitHub client
		appCfg := config.AppConfig{Org: cleanupOrg}
		client, appInfo, err := gh.NewClientFromEnv(ctx, appCfg)
		if err != nil {
			return err
		}
		if appInfo != "" {
			log.Printf("auth: %s", appInfo)
		}

		// Initialize JIT state
		state, err := jit.NewState(cleanupStateDir)
		if err != nil {
			return fmt.Errorf("init JIT state: %w", err)
		}

		// Get expired grants before cleanup
		expired := state.GetExpired()
		if len(expired) == 0 {
			log.Println("no expired grants to clean up")
			return nil
		}

		log.Printf("found %d expired grant(s), revoking access...", len(expired))
		for _, g := range expired {
			log.Printf("  - %s: %s/%s for user %s (expired at %s)",
				g.Type, g.Org, g.Target, g.User, g.ExpiresAt.Format("2006-01-02 15:04:05"))
		}

		// Clean up expired grants
		if err := jit.CleanupExpiredGrants(ctx, client, state); err != nil {
			return fmt.Errorf("cleanup failed: %w", err)
		}

		log.Printf("✓ successfully cleaned up %d expired grant(s)", len(expired))
		return nil
	},
}

func init() {
	cleanupJitCmd.Flags().StringVar(&cleanupOrg, "org", "", "GitHub organization (required)")
	_ = cleanupJitCmd.MarkFlagRequired("org")
	cleanupJitCmd.Flags().StringVar(&cleanupStateDir, "state-dir", "", "Directory to store JIT state (default: ~/.gomgr)")
	rootCmd.AddCommand(cleanupJitCmd)
}
