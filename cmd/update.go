package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/yeasy/ask/internal/config"
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update [skill-name]",
	Short: "Update installed skills to latest version",
	Long: `Update one or all installed skills to their latest versions.
If no skill name is provided, updates all installed skills.`,
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

		if len(cfg.Skills) == 0 {
			fmt.Println("No skills installed.")
			return
		}

		// Determine which skills to update
		var skillsToUpdate []string
		if len(args) > 0 {
			// Update specific skill
			skillName := args[0]
			found := false
			for _, s := range cfg.Skills {
				if s == skillName {
					found = true
					break
				}
			}
			if !found {
				fmt.Printf("Skill '%s' is not installed.\n", skillName)
				os.Exit(1)
			}
			skillsToUpdate = []string{skillName}
		} else {
			// Update all skills
			skillsToUpdate = cfg.Skills
		}

		for _, skillName := range skillsToUpdate {
			skillPath := filepath.Join(config.DefaultSkillsDir, skillName)

			// Check if it's a git repository
			gitDir := filepath.Join(skillPath, ".git")
			if _, err := os.Stat(gitDir); os.IsNotExist(err) {
				fmt.Printf("Skipping %s (not a git repository)\n", skillName)
				continue
			}

			fmt.Printf("Updating %s...\n", skillName)

			// Run git pull
			gitCmd := exec.Command("git", "pull", "--rebase")
			gitCmd.Dir = skillPath
			gitCmd.Stdout = os.Stdout
			gitCmd.Stderr = os.Stderr

			if err := gitCmd.Run(); err != nil {
				fmt.Printf("  Failed to update %s: %v\n", skillName, err)
				continue
			}

			fmt.Printf("  Updated %s successfully!\n", skillName)
		}
	},
}

func init() {
	skillCmd.AddCommand(updateCmd)
}
