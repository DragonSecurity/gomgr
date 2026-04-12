package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/DragonSecurity/gomgr/internal/config"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate configuration without connecting to GitHub",
	Example: `  gomgr validate -c ./config
  gomgr validate --config /path/to/config`,
	RunE: func(_ *cobra.Command, _ []string) error {
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
