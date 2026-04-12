package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/DragonSecurity/gomgr/internal/config"
	"github.com/DragonSecurity/gomgr/internal/gh"
	insync "github.com/DragonSecurity/gomgr/internal/sync"
	"github.com/DragonSecurity/gomgr/internal/util"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Synchronize org state to match YAML configuration",
	Example: `  gomgr sync -c ./config
  gomgr sync -c ./config --dry
  gomgr sync -c ./config --timeout 5m --audit-log`,
	RunE: func(_ *cobra.Command, _ []string) error {
		if cfgDir == "" {
			return fmt.Errorf("--config/-c flag is required")
		}

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		if debug {
			util.EnableDebug()
		}
		util.AuditLog = auditLog

		cfg, err := config.Load(cfgDir)
		if err != nil {
			return err
		}

		client, appInfo, err := gh.NewClientFromEnv(ctx, cfg.App)
		if err != nil {
			return err
		}
		if appInfo != "" {
			util.Infof("auth: %s", appInfo)
		}

		plan, err := insync.BuildPlan(ctx, client, cfg)
		if err != nil {
			return err
		}

		if err := util.PrintPlan(plan); err != nil {
			return fmt.Errorf("print plan: %w", err)
		}

		if dryRun {
			util.PrintSummary(plan)
			util.Infof("dry-run: no changes applied")
			return nil
		}
		return insync.Apply(ctx, client, plan)
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
}
