package cmd

import (
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/DragonSecurity/gomgr/internal/config"
)

var teamName string
var outFile string

var setupTeamCmd = &cobra.Command{
	Use:   "setup-team",
	Short: "Bootstrap a team YAML file for a given team name",
	Example: `  gomgr setup-team -c ./config -n "Backend"
  gomgr setup-team -n "Frontend" -f ./teams/frontend.yaml`,
	RunE: func(_ *cobra.Command, _ []string) error {
		slug := strings.ToLower(strings.ReplaceAll(teamName, " ", "-"))
		path := outFile
		if path == "" {
			path = filepath.Join(cfgDir, "teams", slug+".yaml")
		}
		return config.BootstrapTeamYAML(path, teamName)
	},
}

func init() {
	setupTeamCmd.Flags().StringVarP(&teamName, "name", "n", "", "Team display name (required)")
	_ = setupTeamCmd.MarkFlagRequired("name")
	setupTeamCmd.Flags().StringVarP(&outFile, "file", "f", "", "Force output file path")
	rootCmd.AddCommand(setupTeamCmd)
}
