package cmd

import (
	"context"
	"log"

	"github.com/DragonSecurity/github-org-manager-go/internal/config"
	"github.com/DragonSecurity/github-org-manager-go/internal/gh"
	insync "github.com/DragonSecurity/github-org-manager-go/internal/sync"
	"github.com/DragonSecurity/github-org-manager-go/util"
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Synchronize org state to match YAML configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		if debug {
			util.EnableDebug()
		}

		cfg, err := config.Load(cfgDir)
		if err != nil {
			return err
		}

		client, appInfo, err := gh.NewClientFromEnv(ctx, cfg.App)
		if err != nil {
			return err
		}
		if appInfo != "" {
			log.Printf("auth: %s", appInfo)
		}

		plan, err := insync.BuildPlan(ctx, client, cfg)
		if err != nil {
			return err
		}

		util.PrintPlan(plan)

		if dryRun {
			log.Println("dry-run: no changes applied")
			return nil
		}
		return insync.Apply(ctx, client, plan)
	},
}

func init() { rootCmd.AddCommand(syncCmd) }
