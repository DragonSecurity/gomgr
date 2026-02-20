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
			if info.Modified {
				fmt.Println("DirtyBuild: true")
			} else {
				fmt.Println("DirtyBuild: false")
			}
		}

		if info.CommitTime != "" {
			fmt.Printf("LastCommit: %s\n", info.CommitTime)
		}
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
