package cmd

import (
	"github.com/spf13/cobra"
	"github.com/yeasy/ask/internal/config"
)

func loadConfigForCommand(cmd *cobra.Command) (*config.Config, error) {
	if cfgFile != "" {
		return config.LoadConfigFromPath(cfgFile)
	}
	global, _ := cmd.Flags().GetBool("global")
	return config.LoadConfigByScope(global)
}

func saveConfigForCommand(cmd *cobra.Command, cfg *config.Config) error {
	if cfgFile != "" {
		return cfg.SaveToPath(cfgFile)
	}
	global, _ := cmd.Flags().GetBool("global")
	return cfg.SaveByScope(global)
}
