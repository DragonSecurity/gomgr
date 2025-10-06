package cmd

import (
	"fmt"
	rdebug "runtime/debug"
	"strings"

	"github.com/DragonSecurity/gomgr/internal/version"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Version:", version.Version)

		// Optional: print VCS info if the build has it
		if info, ok := rdebug.ReadBuildInfo(); ok {
			var rev, t, dirty string
			for _, s := range info.Settings {
				switch s.Key {
				case "vcs.revision":
					rev = s.Value
					if len(rev) > 12 {
						rev = rev[:12]
					}
				case "vcs.time":
					t = s.Value
				case "vcs.modified":
					if s.Value == "true" {
						dirty = "dirty"
					} else {
						dirty = "clean"
					}
				}
			}
			if rev != "" || t != "" {
				if rev == "" {
					rev = "unknown"
				}
				if t == "" {
					t = "unknown"
				}
				if dirty == "" {
					dirty = "unknown"
				}
				fmt.Printf("Revision: %s\n", rev)
				fmt.Printf("DirtyBuild: %s\n", strings.ToLower(dirty))
				fmt.Printf("LastCommit: %s\n", t)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
