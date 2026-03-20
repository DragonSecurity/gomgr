package cmd

import (
	"fmt"

	"github.com/DragonSecurity/gomgr/internal/config"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate configuration without connecting to GitHub",
	Example: `  gomgr validate -c ./config
  gomgr validate --config /path/to/config`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if cfgDir == "" {
			return fmt.Errorf("--config/-c flag is required")
		}
		cfg, err := config.Load(cfgDir)
		if err != nil {
			return err
		}
		_ = cfg
		fmt.Println("Configuration is valid.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)
}
