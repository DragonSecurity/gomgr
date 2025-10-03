package cmd

import (
	"fmt"

	"github.com/earthboundkid/versioninfo/v2"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Version:", versioninfo.Version)
		fmt.Println("Revision:", versioninfo.Revision)
		fmt.Println("DirtyBuild:", versioninfo.DirtyBuild)
		fmt.Println("LastCommit:", versioninfo.LastCommit)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
