package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/DragonSecurity/gomgr/internal/config"
	"github.com/DragonSecurity/gomgr/internal/gh"
	insync "github.com/DragonSecurity/gomgr/internal/sync"
	"github.com/DragonSecurity/gomgr/internal/util"
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Synchronize org state to match YAML configuration",
	Example: `  gomgr sync -c ./config
  gomgr sync -c ./config --dry
  gomgr sync -c ./config --timeout 5m --audit-log`,
	RunE: func(cmd *cobra.Command, args []string) error {
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
			log.Printf("auth: %s", appInfo)
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
			log.Println("dry-run: no changes applied")
			return nil
		}
		return insync.Apply(ctx, client, plan)
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
}
