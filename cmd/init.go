package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/yeasy/ask/internal/config"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new ASK project",
	Long: `Initialize a new Agent Skills Kit project. 
This will create ask.yaml and the skills directory (default: .agent/skills/).`,
	Run: func(_ *cobra.Command, _ []string) {
		if _, err := os.Stat("ask.yaml"); err == nil {
			fmt.Println("ask.yaml already exists in this directory.")
			return
		}

		// Create skills directory using default path
		skillsDir := config.DefaultSkillsDir
		if err := os.MkdirAll(skillsDir, 0755); err != nil {
			fmt.Printf("Error creating skills directory: %v\n", err)
			os.Exit(1)
		}

		err := config.CreateDefaultConfig()
		if err != nil {
			fmt.Printf("Error creating ask.yaml: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("✓ Initialized ASK project")
		fmt.Println("  Created: ask.yaml")
		fmt.Printf("  Created: %s/\n", skillsDir)
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
