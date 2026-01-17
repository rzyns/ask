package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yeasy/ask/internal/config"
)

// uninstallCmd represents the uninstall command
var uninstallCmd = &cobra.Command{
	Use:   "uninstall [skill-name]",
	Short: "Uninstall a skill",
	Long: `Remove a skill from the skills directory and update configuration.
Use --global to uninstall from global installation (~/.ask/skills).

Use --agent (-a) to specify target agents (claude, cursor, codex, opencode).
If no agent is specified, uninstalls from .agent/skills/ by default.`,
	Example: `  ask skill uninstall browser-use
  ask skill uninstall --global browser-use
  ask skill uninstall browser-use --agent claude --agent cursor`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		skillName := args[0]
		global, _ := cmd.Flags().GetBool("global")
		agents, _ := cmd.Flags().GetStringSlice("agent")

		// Ensure project is initialized for non-global operations
		if !global && len(agents) == 0 {
			if !ensureInitialized() {
				return
			}
		}

		// Validate agent names
		for _, agent := range agents {
			if !config.IsValidAgent(agent) {
				fmt.Printf("Error: Unknown agent '%s'. Supported agents: %s\n",
					agent, strings.Join(config.GetSupportedAgentNames(), ", "))
				os.Exit(1)
			}
		}

		// 1. Remove from configuration (always update project/global config regardless of agents)
		// This keeps track of what was installed, even if we remove it from agent dirs
		cfg, err := config.LoadConfigByScope(global)
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}

		targetName := filepath.Base(skillName)

		cfg.RemoveSkill(targetName)
		cfg.RemoveSkillInfo(targetName)

		// Determine target directories
		var targetDirs []string
		var scopeLabel string

		if len(agents) > 0 {
			// Uninstall from specific agent directories
			for _, agentName := range agents {
				agentType, _ := config.ResolveAgentType(agentName)
				dir, err := config.GetAgentSkillsDir(agentType, global)
				if err != nil {
					fmt.Printf("Error getting skills dir for agent %s: %v\n", agentName, err)
					continue
				}
				targetDirs = append(targetDirs, dir)
			}
			scopeLabel = strings.Join(agents, ", ")
			if global {
				scopeLabel += " (global)"
			}
		} else {
			if global {
				targetDirs = []string{config.GetSkillsDirByScope(true)}
				scopeLabel = "global"
			} else {
				// Use active/detected directories
				wd, _ := os.Getwd()
				targetDirs = cfg.GetActiveSkillsDirs(wd)
				scopeLabel = "detected targets"
			}
		}

		// 2. Remove directories
		for _, dir := range targetDirs {
			skillPath := filepath.Join(dir, targetName)
			if _, err := os.Stat(skillPath); !os.IsNotExist(err) {
				fmt.Printf("Removing %s...\n", skillPath)
				err := os.RemoveAll(skillPath)
				if err != nil {
					fmt.Printf("Failed to remove skill directory %s: %v\n", skillPath, err)
				}
			} else {
				fmt.Printf("Warning: Skill directory %s not found.\n", skillPath)
			}
		}

		err = cfg.SaveByScope(global)
		if err != nil {
			configFile := "ask.yaml"
			if global {
				configFile = "~/.ask/config.yaml"
			}
			fmt.Printf("Failed to update %s: %v\n", configFile, err)
			os.Exit(1)
		}

		// 3. Remove from lock file
		lockFile, _ := config.LoadLockFileByScope(global)
		lockFile.RemoveEntry(targetName)
		if err := lockFile.SaveByScope(global); err != nil {
			lockFileName := "ask.lock"
			if global {
				lockFileName = "~/.ask/ask.lock"
			}
			fmt.Printf("Warning: Failed to update %s: %v\n", lockFileName, err)
		}

		fmt.Printf("Successfully uninstalled %s (%s)\n", targetName, scopeLabel)
	},
}

func init() {
	skillCmd.AddCommand(uninstallCmd)
	uninstallCmd.Flags().StringSliceP("agent", "a", []string{}, "target agent(s) for uninstallation")
	uninstallCmd.Flags().BoolP("global", "g", false, "uninstall globally")
}
