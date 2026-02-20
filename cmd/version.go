package cmd

import (
	"fmt"

	"github.com/DragonSecurity/gomgr/internal/version"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Run: func(cmd *cobra.Command, args []string) {
		info := version.GetBuildInfo()
		fmt.Println("Version:", info.Version)

		if info.Revision != "" {
			fmt.Printf("Revision: %s\n", info.Revision)
			fmt.Printf("Modified: %v\n", info.Modified)
		}

		if info.CommitTime != "" {
			fmt.Printf("LastCommit: %s\n", info.CommitTime)
		}
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
