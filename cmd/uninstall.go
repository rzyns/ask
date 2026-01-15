package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/yeasy/ask/internal/config"
)

// uninstallCmd represents the uninstall command
var uninstallCmd = &cobra.Command{
	Use:   "uninstall [skill-name]",
	Short: "Uninstall a skill",
	Long: `Remove a skill from the ./skills directory and update ask.yaml.
Example: ask uninstall browser-use`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		skillName := args[0]

		// 1. Remove from configuration
		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}

		targetName := filepath.Base(skillName)

		cfg.RemoveSkill(targetName)

		// 2. Remove directory
		skillPath := filepath.Join("skills", targetName)
		if _, err := os.Stat(skillPath); !os.IsNotExist(err) {
			fmt.Printf("Removing %s...\n", skillPath)
			err := os.RemoveAll(skillPath)
			if err != nil {
				fmt.Printf("Failed to remove skill directory: %v\n", err)
				os.Exit(1)
			}
		} else {
			fmt.Printf("Warning: Skill directory %s not found.\n", skillPath)
		}

		err = cfg.Save()
		if err != nil {
			fmt.Printf("Failed to update ask.yaml: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully uninstalled %s\n", targetName)
	},
}

func init() {
	rootCmd.AddCommand(uninstallCmd)
}
