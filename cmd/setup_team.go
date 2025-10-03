package cmd

import (
	"path/filepath"
	"strings"

	"github.com/DragonSecurity/github-org-manager-go/internal/config"
	"github.com/spf13/cobra"
)

var teamName string
var outFile string

var setupTeamCmd = &cobra.Command{
	Use:   "setup-team",
	Short: "Bootstrap a team YAML file for a given team name",
	RunE: func(cmd *cobra.Command, args []string) error {
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
