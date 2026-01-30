package cmd

import (
	"encoding/json"
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
	Run: runList,
}

func runList(cmd *cobra.Command, _ []string) {
	global, _ := cmd.Flags().GetBool("global")
	all, _ := cmd.Flags().GetBool("all")
	agents, _ := cmd.Flags().GetStringSlice("agent")
	jsonOutput, _ := cmd.Flags().GetBool("json")

	// Ensure project is initialized for non-global operations
	if !global && !all && len(agents) == 0 {
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

	var collectedItems []SkillListItem

	if len(agents) > 0 {
		// Show skills for specific agents by checking directories
		for _, agentName := range agents {
			agentType, _ := config.ResolveAgentType(agentName)

			if all || (!global) {
				// Project level
				dir, _ := config.GetAgentSkillsDir(agentType, false)
				items := showAgentSkills(agentName, dir, "Project", jsonOutput)
				if jsonOutput {
					collectedItems = append(collectedItems, items...)
				}
			}

			if all || global {
				// Global level
				dir, _ := config.GetAgentSkillsDir(agentType, true)
				items := showAgentSkills(agentName, dir, "Global", jsonOutput)
				if jsonOutput {
					collectedItems = append(collectedItems, items...)
				}
			}
		}
	} else {
		if all {
			// Show both project and global skills from config
			items := showSkills("Project", false, jsonOutput)
			if jsonOutput {
				collectedItems = append(collectedItems, items...)
			}
			if !jsonOutput {
				fmt.Println()
			}
			itemsGlobal := showSkills("Global", true, jsonOutput)
			if jsonOutput {
				collectedItems = append(collectedItems, itemsGlobal...)
			}
		} else {
			scope := "Project"
			if global {
				scope = "Global"
			}
			items := showSkills(scope, global, jsonOutput)
			if jsonOutput {
				collectedItems = append(collectedItems, items...)
			}
		}
	}

	if jsonOutput {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(collectedItems); err != nil {
			fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", err)
		}
	}
}

// SkillListItem represents a skill in the list output
type SkillListItem struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	URL         string `json:"url,omitempty"`
	Scope       string `json:"scope"`
	Agent       string `json:"agent,omitempty"`
	Path        string `json:"path,omitempty"`
}

func showAgentSkills(agentName, dir, scope string, jsonOutput bool) []SkillListItem {
	var items []SkillListItem

	if !jsonOutput {
		fmt.Printf("%s Skills for %s (%s):\n", scope, agentName, dir)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		if !jsonOutput {
			if os.IsNotExist(err) {
				fmt.Println("  (directory not created)")
			} else {
				fmt.Printf("  Error reading directory: %v\n", err)
			}
			fmt.Println()
		}
		return nil
	}

	if len(entries) == 0 {
		if !jsonOutput {
			fmt.Println("  (none)")
			fmt.Println()
		}
		return nil
	}

	count := 0
	for _, entry := range entries {
		if entry.IsDir() {
			item := SkillListItem{
				Name:  entry.Name(),
				Scope: scope,
				Agent: agentName,
				Path:  filepath.Join(dir, entry.Name()),
			}

			// Try to get description from SKILL.md
			skillPath := filepath.Join(dir, entry.Name())
			if skill.FindSkillMD(skillPath) {
				if meta, err := skill.ParseSkillMD(skillPath); err == nil && meta != nil && meta.Description != "" {
					item.Description = meta.Description
				}
			}

			if jsonOutput {
				items = append(items, item)
			} else {
				fmt.Printf("  %s\n", entry.Name())
				if item.Description != "" {
					fmt.Printf("    Description: %s\n", item.Description)
				}
			}
			count++
		}
	}

	if !jsonOutput {
		if count == 0 {
			fmt.Println("  (none)")
		}
		fmt.Println()
	}
	return items
}

func showSkills(scope string, global bool, jsonOutput bool) []SkillListItem {
	var items []SkillListItem
	cfg, err := config.LoadConfigByScope(global)
	if err != nil {
		if !jsonOutput {
			if os.IsNotExist(err) && !global {
				fmt.Printf("%s Skills: No ask.yaml found. Run 'ask init' first.\n", scope)
				return nil
			}
			if !global {
				fmt.Printf("Error loading config: %v\n", err)
			}
		}
		return nil
	}

	if len(cfg.Skills) == 0 && len(cfg.SkillsInfo) == 0 {
		if !jsonOutput {
			fmt.Printf("%s Skills: (none)\n", scope)
		}
		return nil
	}

	if !jsonOutput {
		fmt.Printf("%s Skills:\n", scope)
		fmt.Println()
	}

	// Show skills with metadata first
	shown := make(map[string]bool)
	for _, skill := range cfg.SkillsInfo {
		shown[skill.Name] = true
		item := SkillListItem{
			Name:        skill.Name,
			Description: skill.Description,
			URL:         skill.URL,
			Scope:       scope,
		}
		if jsonOutput {
			items = append(items, item)
		} else {
			fmt.Printf("  %s\n", skill.Name)
			if skill.Description != "" {
				fmt.Printf("    Description: %s\n", skill.Description)
			}
			if skill.URL != "" {
				fmt.Printf("    URL: %s\n", skill.URL)
			}
			fmt.Println()
		}
	}

	// Show legacy skills without metadata
	for _, skill := range cfg.Skills {
		if !shown[skill] {
			item := SkillListItem{
				Name:  skill,
				Scope: scope,
			}
			if jsonOutput {
				items = append(items, item)
			} else {
				fmt.Printf("  %s\n", skill)
				fmt.Println()
			}
		}
	}
	return items
}

func registerListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("all", false, "show both project and global skills")
	cmd.Flags().StringSliceP("agent", "a", []string{}, "list skills for specific agent(s)")
	cmd.Flags().Bool("json", false, "output results in JSON format")
}

func init() {
	skillCmd.AddCommand(listCmd)
	registerListFlags(listCmd)

	// Register agent flag completion
	_ = listCmd.RegisterFlagCompletionFunc("agent", completeAgentNames)
}
