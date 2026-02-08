package cmd

import (
	"github.com/spf13/cobra"
)

// skillCmd represents the skill parent command
var skillCmd = &cobra.Command{
	Use:     "skill",
	Aliases: []string{"skills"},
	Short:   "Manage agent skills",
	Long: `Manage agent skills - search, install, update, and remove skills.

Examples:
  ask skill search browser       # Search for browser-related skills
  ask skill install browser-use  # Install a skill
  ask skill list                 # List installed skills
  ask skill update               # Update all skills`,
}

func init() {
	rootCmd.AddCommand(skillCmd)
	skillCmd.AddCommand(checkCmd)
	skillCmd.AddCommand(promptCmd)
}
