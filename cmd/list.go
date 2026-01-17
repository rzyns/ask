package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yeasy/ask/internal/config"
	"github.com/yeasy/ask/internal/skill"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed skills",
	Long: `List all skills currently installed.
Use --global to show global skills, --all to show both project and global skills.
Use --agent (-a) to list skills for specific agents (checks agent directories).`,
	Run: func(cmd *cobra.Command, args []string) {
		global, _ := cmd.Flags().GetBool("global")
		all, _ := cmd.Flags().GetBool("all")
		agents, _ := cmd.Flags().GetStringSlice("agent")

		// Validate agent names
		for _, agent := range agents {
			if !config.IsValidAgent(agent) {
				fmt.Printf("Error: Unknown agent '%s'. Supported agents: %s\n",
					agent, strings.Join(config.GetSupportedAgentNames(), ", "))
				os.Exit(1)
			}
		}

		if len(agents) > 0 {
			// Show skills for specific agents by checking directories
			for _, agentName := range agents {
				agentType, _ := config.ResolveAgentType(agentName)

				if all || (!global) {
					// Project level
					dir, _ := config.GetAgentSkillsDir(agentType, false)
					showAgentSkills(agentName, dir, "Project")
				}

				if all || global {
					// Global level
					dir, _ := config.GetAgentSkillsDir(agentType, true)
					showAgentSkills(agentName, dir, "Global")
				}
			}
			return
		}

		if all {
			// Show both project and global skills from config
			showSkills("Project", false)
			fmt.Println()
			showSkills("Global", true)
		} else {
			scope := "Project"
			if global {
				scope = "Global"
			}
			showSkills(scope, global)
		}
	},
}

func showAgentSkills(agentName, dir, scope string) {
	fmt.Printf("%s Skills for %s (%s):\n", scope, agentName, dir)

	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("  (directory not created)")
		} else {
			fmt.Printf("  Error reading directory: %v\n", err)
		}
		fmt.Println()
		return
	}

	if len(entries) == 0 {
		fmt.Println("  (none)")
		fmt.Println()
		return
	}

	count := 0
	for _, entry := range entries {
		if entry.IsDir() {
			fmt.Printf("  %s\n", entry.Name())

			// Try to get description from SKILL.md
			skillPath := filepath.Join(dir, entry.Name())
			if skill.FindSkillMD(skillPath) {
				if meta, err := skill.ParseSkillMD(skillPath); err == nil && meta != nil && meta.Description != "" {
					fmt.Printf("    Description: %s\n", meta.Description)
				}
			}
			count++
		}
	}

	if count == 0 {
		fmt.Println("  (none)")
	}
	fmt.Println()
}

func showSkills(scope string, global bool) {
	cfg, err := config.LoadConfigByScope(global)
	if err != nil {
		if os.IsNotExist(err) && !global {
			fmt.Printf("%s Skills: No ask.yaml found. Run 'ask init' first.\n", scope)
			return
		}
		if !global {
			fmt.Printf("Error loading config: %v\n", err)
		}
		return
	}

	if len(cfg.Skills) == 0 && len(cfg.SkillsInfo) == 0 {
		fmt.Printf("%s Skills: (none)\n", scope)
		return
	}

	fmt.Printf("%s Skills:\n", scope)
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
}

func init() {
	skillCmd.AddCommand(listCmd)
	listCmd.Flags().Bool("all", false, "show both project and global skills")
	listCmd.Flags().StringSliceP("agent", "a", []string{}, "list skills for specific agent(s)")
}
