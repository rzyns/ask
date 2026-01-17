package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/yeasy/ask/internal/config"
	"github.com/yeasy/ask/internal/skill"
)

// infoCmd represents the info command
var infoCmd = &cobra.Command{
	Use:   "info [skill-name]",
	Short: "Show detailed information about a skill",
	Long: `Display detailed metadata about an installed skill.
Reads from the SKILL.md file if available.
Use --global to check global skills.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		global, _ := cmd.Flags().GetBool("global")
		skillName := args[0]

		// Ensure project is initialized for non-global operations
		if !global {
			if !ensureInitialized() {
				return
			}
		}

		skillsDir := config.GetSkillsDirByScope(global)
		skillPath := filepath.Join(skillsDir, skillName)

		// Check if skill exists
		if _, err := os.Stat(skillPath); os.IsNotExist(err) {
			scopeLabel := "project"
			if global {
				scopeLabel = "global"
			}
			fmt.Printf("Skill '%s' is not installed (%s).\n", skillName, scopeLabel)
			os.Exit(1)
		}

		scopeLabel := "Project"
		if global {
			scopeLabel = "Global"
		}
		fmt.Printf("Skill: %s (%s)\n", skillName, scopeLabel)
		fmt.Printf("Path:  %s\n", skillPath)
		fmt.Println()

		// Try to parse SKILL.md
		if skill.FindSkillMD(skillPath) {
			meta, err := skill.ParseSkillMD(skillPath)
			if err != nil {
				fmt.Printf("Warning: Could not parse SKILL.md: %v\n", err)
			} else {
				if meta.Name != "" {
					fmt.Printf("  Name:        %s\n", meta.Name)
				}
				if meta.Description != "" {
					fmt.Printf("  Description: %s\n", meta.Description)
				}
				if meta.Version != "" {
					fmt.Printf("  Version:     %s\n", meta.Version)
				}
				if meta.Author != "" {
					fmt.Printf("  Author:      %s\n", meta.Author)
				}
				if len(meta.Dependencies) > 0 {
					fmt.Printf("  Dependencies:\n")
					for _, dep := range meta.Dependencies {
						fmt.Printf("    - %s\n", dep)
					}
				}
				if len(meta.Tags) > 0 {
					fmt.Printf("  Tags: %v\n", meta.Tags)
				}
			}
		} else {
			fmt.Println("  No SKILL.md found.")
		}

		// List files in skill directory
		fmt.Println()
		fmt.Println("Files:")
		entries, err := os.ReadDir(skillPath)
		if err == nil {
			for _, entry := range entries {
				if entry.IsDir() {
					fmt.Printf("  📁 %s/\n", entry.Name())
				} else {
					fmt.Printf("  📄 %s\n", entry.Name())
				}
			}
		}
	},
}

func init() {
	skillCmd.AddCommand(infoCmd)
}
