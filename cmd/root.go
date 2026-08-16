package cmd

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var (
	cfgDir           string
	debug            bool
	dryRun           bool
	timeout          time.Duration
	auditLog         bool
	continueOnError  bool
	appIDFlag        int64
	privateKeyFlag   string
	detailedExitCode bool
)

// ErrPendingChanges is returned by a --dry run under --detailed-exitcode when
// the plan is not empty. It carries no message: the plan has already been
// printed, and the exit status is the whole point.
var ErrPendingChanges = errors.New("pending changes")

// exitCodePendingChanges is what --detailed-exitcode reports for a non-empty
// plan. Distinct from 1 so a pipeline can tell "there is work to do" from
// "the run failed", which grepping the output cannot do reliably.
const exitCodePendingChanges = 2

var rootCmd = &cobra.Command{
	Use:   "gomgr",
	Short: "GitHub Organization Manager (Go)",
	Long:  "Sync GitHub org owners, teams, members, and repo permissions from YAML.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		if errors.Is(err, ErrPendingChanges) {
			os.Exit(exitCodePendingChanges)
		}
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
	rootCmd.PersistentFlags().BoolVar(&continueOnError, "continue-on-error", false, "Keep applying remaining changes after a failure, then report all errors at the end")
	rootCmd.PersistentFlags().Int64Var(&appIDFlag, "app-id", 0, "GitHub App ID, overriding app.yaml and GITHUB_APP_ID")
	rootCmd.PersistentFlags().StringVar(&privateKeyFlag, "private-key", "", "Path to the GitHub App private key PEM, overriding app.yaml and GITHUB_APP_PRIVATE_KEY")
	rootCmd.PersistentFlags().BoolVar(&detailedExitCode, "detailed-exitcode", false,
		"With --dry, exit 2 when the plan has changes and 0 when it does not, so a pipeline can branch on it without parsing the output")
}
