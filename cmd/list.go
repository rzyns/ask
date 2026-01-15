package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/yeasy/ask/internal/config"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed skills",
	Long:  `List all skills currently tracked in ask.yaml.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadConfig()
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Println("No ask.yaml found. Run 'ask init' first.")
				return
			}
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}

		if len(cfg.Skills) == 0 && len(cfg.SkillsInfo) == 0 {
			fmt.Println("No skills installed.")
			return
		}

		fmt.Println("Installed Skills:")
		fmt.Println()

		// Show skills with metadata first
		shown := make(map[string]bool)
		for _, skill := range cfg.SkillsInfo {
			shown[skill.Name] = true
			fmt.Printf("  %s\n", skill.Name)
			if skill.Description != "" {
				fmt.Printf("    Description: %s\n", skill.Description)
			}
			if skill.URL != "" {
				fmt.Printf("    URL: %s\n", skill.URL)
			}
			fmt.Println()
		}

		// Show legacy skills without metadata
		for _, skill := range cfg.Skills {
			if !shown[skill] {
				fmt.Printf("  %s\n", skill)
				fmt.Println()
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
