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
	grantOrg string
)

var grantCmd = &cobra.Command{
	Use:   "grant <target> <user> [permission]",
	Short: "Grant permanent access to a team or repository",
	Long: `Grant permanent access to a team or repository.

Unlike the 'access' command, this grants permanent access that is not automatically revoked.

Examples:
  # Grant permanent team membership
  gomgr grant developer allanice001 --org myorg
  gomgr grant developer allanice001 maintainer --org myorg

  # Grant permanent repo access
  gomgr grant myrepo allanice001 push --org myorg
  gomgr grant myrepo allanice001 admin --org myorg`,
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

		if grantOrg == "" {
			return fmt.Errorf("--org is required")
		}

		// Initialize GitHub client
		appCfg := config.AppConfig{Org: grantOrg}
		client, appInfo, err := gh.NewClientFromEnv(ctx, appCfg)
		if err != nil {
			return err
		}
		if appInfo != "" {
			log.Printf("auth: %s", appInfo)
		}

		// Determine if this is team or repo access
		grantType := "repo"
		if permission == "" || permission == "member" || permission == "maintainer" {
			grantType = "team"
			if permission == "" {
				permission = "member"
			}
		}

		// Grant access
		log.Printf("granting permanent %s access to %s for user %s...", grantType, target, user)
		switch grantType {
		case "team":
			err = jit.GrantTeamAccess(ctx, client, grantOrg, target, user, permission, 0)
		case "repo":
			err = jit.GrantRepoAccess(ctx, client, grantOrg, target, user, permission)
		}
		if err != nil {
			return err
		}

		log.Printf("✓ permanent access granted to %s", user)
		log.Printf("  type: %s", grantType)
		log.Printf("  target: %s", target)
		log.Printf("  permission: %s", permission)
		log.Println()
		log.Printf("Note: This is permanent access. To manage it via YAML config, add the user to the appropriate team or repo in your config files.")

		return nil
	},
}

func init() {
	grantCmd.Flags().StringVar(&grantOrg, "org", "", "GitHub organization (required)")
	_ = grantCmd.MarkFlagRequired("org")
	rootCmd.AddCommand(grantCmd)
}
