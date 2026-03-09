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

		// Detect existing agent directories
		cwd, _ := os.Getwd()
		detected := config.DetectExistingToolDirs(cwd)

		fmt.Println("✓ Initialized ASK project")
		fmt.Println("  Created: ask.yaml")
		fmt.Printf("  Created: %s/\n", skillsDir)

		if len(detected) > 0 {
			fmt.Println()
			fmt.Println("  Detected agents:")
			for _, t := range detected {
				fmt.Printf("    • %s (%s)\n", t.Name, t.SkillsDir)
			}
			fmt.Println()
			fmt.Println("  Skills will be synced to all detected agents automatically.")
		}

		fmt.Println()
		fmt.Println("Next steps:")
		fmt.Println("  ask search          Browse available skills")
		fmt.Println("  ask install <name>  Install a skill")
		fmt.Println("  ask doctor          Check your setup")
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
