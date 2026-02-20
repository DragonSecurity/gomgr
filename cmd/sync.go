package cmd

import (
	"context"
	"log"

	"github.com/DragonSecurity/gomgr/internal/config"
	"github.com/DragonSecurity/gomgr/internal/gh"
	"github.com/DragonSecurity/gomgr/internal/plan"
	insync "github.com/DragonSecurity/gomgr/internal/sync"
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Synchronize org state to match YAML configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		if debug {
			log.SetFlags(log.LstdFlags | log.Lshortfile)
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

		p, err := insync.BuildPlan(ctx, client, cfg)
		if err != nil {
			return err
		}

		plan.Print(p)

		if dryRun {
			log.Println("dry-run: no changes applied")
			return nil
		}
		return insync.Apply(ctx, client, p)
	},
}

func init() {
	syncCmd.PersistentFlags().StringVarP(&cfgDir, "config", "c", "", "Path to config directory (required)")
	_ = syncCmd.MarkPersistentFlagRequired("config")
	rootCmd.AddCommand(syncCmd)
}
