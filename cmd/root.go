package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var (
	cfgDir   string
	debug    bool
	dryRun   bool
	timeout  time.Duration
	auditLog bool
)

var rootCmd = &cobra.Command{
	Use:   "gomgr",
	Short: "GitHub Organization Manager (Go)",
	Long:  "Sync GitHub org owners, teams, members, and repo permissions from YAML.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgDir, "config", "c", "", "Path to config directory (required)")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Enable verbose debug logs")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry", false, "Show a plan without applying changes")
	rootCmd.PersistentFlags().DurationVar(&timeout, "timeout", 10*time.Minute, "Overall context timeout for the sync operation")
	rootCmd.PersistentFlags().BoolVar(&auditLog, "audit-log", false, "Emit structured JSON audit log entries to stderr")
}
