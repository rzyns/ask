package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yeasy/ask/internal/config"
	"github.com/yeasy/ask/internal/github"
	"github.com/yeasy/ask/internal/installer"
	"github.com/yeasy/ask/internal/repository"
)

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:               "install [url...]",
	Aliases:           []string{"add", "i"},
	ValidArgsFunction: completeSkillNames,
	Short:             "Install one or more skills from git repositories",
	Long: `Download and install skills into agent-specific directories. 
You can provide full git URLs or GitHub shorthands (owner/repo).
You can also specify versions: owner/repo@v1.0.0

Use --agent (-a) to specify target agents (claude, cursor, codex, opencode).
Multiple agents can be specified by repeating the flag.
If no agent is specified, skills are installed to .agent/skills/ by default.`,
	Example: `  # Install from GitHub shorthand
  ask skill install anthropics/pdf
  
  # Install to specific agents
  ask skill install pdf --agent claude --agent cursor
  ask skill install pdf -a claude -a cursor
  
  # Install globally for an agent
  ask skill install pdf --agent claude --global
  
  # Install multiple skills at once
  ask skill install pdf docx mcp-builder
  
  # Install specific version
  ask skill install browser-use/browser-use@v0.1.0
  
  # Install from subdirectory
  ask skill install anthropics/skills/skills/pdf
  
  # Install from GitHub browser URL
  ask skill install https://github.com/anthropics/skills/tree/main/skills/pdf
  
  # Install from full URL
  ask skill install https://github.com/browser-use/browser-use.git`,
	Args: cobra.MinimumNArgs(1),
	Run:  runInstall,
}

const maxInputLength = 255

func runInstall(cmd *cobra.Command, args []string) {
	// Check for offline mode
	if offline, _ := cmd.Flags().GetBool("offline"); offline || github.OfflineMode {
		fmt.Println("Error: Cannot install skills in offline mode.")
		os.Exit(1)
	}

	// Check for global flag
	global, _ := cmd.Flags().GetBool("global")

	// Get agent targets
	agents, _ := cmd.Flags().GetStringSlice("agent")

	// Validate agent names
	for _, agent := range agents {
		if !config.IsValidAgent(agent) {
			fmt.Printf("Error: Unknown agent '%s'. Supported agents: %s\n",
				agent, strings.Join(config.GetSupportedAgentNames(), ", "))
			os.Exit(1)
		}
	}

	// Ensure project is initialized for non-global, non-agent-specific operations
	if !global && len(agents) == 0 {
		if !ensureInitialized() {
			return
		}
	}

	// Track installation results
	var succeeded, failed []string

	// Pre-process args to separate skills and agents
	var skillArgs []string
	agentFlagChanged := cmd.Flags().Changed("agent")

	for _, arg := range args {
		if agentFlagChanged && config.IsValidAgent(arg) {
			agents = append(agents, arg)
		} else {
			skillArgs = append(skillArgs, arg)
		}
	}

	// Load config
	cfg, err := config.LoadConfig()
	if err != nil {
		def := config.DefaultConfig()
		cfg = &def
	}

	var expandedArgs []string
	for _, input := range skillArgs {
		if len(input) > maxInputLength {
			fmt.Printf("Error: Input '%s...' is too long (max %d chars)\n", input[:20], maxInputLength)
			failed = append(failed, input)
			continue
		}

		// Check if input matches a configured repository name
		var targetRepo *config.Repo
		for i := range cfg.Repos {
			r := &cfg.Repos[i]
			if r.Name == input {
				targetRepo = r
				break
			}
			if strings.Contains(r.URL, input) {
				if strings.HasPrefix(r.URL, input) || strings.Contains(r.URL, "/"+input) {
					targetRepo = r
					break
				}
			}
		}

		if targetRepo != nil {
			fmt.Printf("Fetching skills from repo '%s'...\n", input)

			var repos []github.Repository
			var err error

			if targetRepo.Type == "dir" {
				repos, err = repository.FetchSkillsViaGit(*targetRepo)
			}

			if err != nil || targetRepo.Type != "dir" {
				repos, err = repository.FetchSkills(*targetRepo)
				if err != nil {
					fmt.Printf("Failed to fetch skills from repo '%s': %v\n", input, err)
					failed = append(failed, input)
					continue
				}
			}

			if len(repos) == 0 {
				fmt.Printf("No skills found in repo '%s'\n", input)
				continue
			}

			fmt.Printf("Found %d skills in repo '%s'. Queueing for installation...\n", len(repos), input)
			for _, r := range repos {
				expandedArgs = append(expandedArgs, r.HTMLURL)
			}
		} else {
			expandedArgs = append(expandedArgs, input)
		}
	}

	opts := installer.InstallOptions{
		Global: global,
		Agents: agents,
		Config: cfg,
	}

	// Install each expanded skill
	for _, input := range expandedArgs {
		err := installer.Install(input, opts)
		if err != nil {
			failed = append(failed, input)
			fmt.Printf("Failed to install %s: %v\n", input, err)
		} else {
			succeeded = append(succeeded, input)
		}
	}

	// Print summary
	if len(args) > 1 {
		fmt.Println()
		fmt.Println("Installation Summary:")

		var targetDisplay string
		if len(agents) > 0 {
			targetDisplay = strings.Join(agents, ", ")
		} else if global {
			targetDisplay = "global"
		} else {
			wd, _ := os.Getwd()
			detected := config.DetectExistingToolDirs(wd)
			if len(detected) > 0 {
				var names []string
				for _, t := range detected {
					names = append(names, t.Name)
				}
				targetDisplay = strings.Join(names, ", ")
			} else {
				targetDisplay = ".agent/skills"
			}
		}

		if len(succeeded) > 0 {
			fmt.Printf("  ✓ Succeeded: %d (%s) -> to: %s\n", len(succeeded), strings.Join(succeeded, ", "), targetDisplay)
		}
		if len(failed) > 0 {
			fmt.Printf("  ✗ Failed: %d (%s)\n", len(failed), strings.Join(failed, ", "))
		}
	}

	if len(failed) > 0 {
		os.Exit(1)
	}
}

func registerInstallFlags(cmd *cobra.Command) {
	cmd.Flags().StringSliceP("agent", "a", []string{}, "Target agent(s) to install for (e.g. claude, cursor)")
}

func init() {
	skillCmd.AddCommand(installCmd)
	registerInstallFlags(installCmd)
}
