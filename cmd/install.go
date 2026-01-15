package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yeasy/ask/internal/config"
	"github.com/yeasy/ask/internal/git"
	"github.com/yeasy/ask/internal/skill"
)

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install [url]",
	Short: "Install a skill from a git repository",
	Long: `Download and install a skill into the ./skills directory. 
You can provide a full git URL or a GitHub shorthand (owner/repo).`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		input := args[0]

		// Check if it's a direct URL or shorthand
		isURL := strings.HasPrefix(input, "http") || strings.HasPrefix(input, "git@")

		var repoURL, subDir, skillName string

		if !isURL {
			parts := strings.Split(input, "/")
			if len(parts) > 2 {
				// It's a subdirectory install: owner/repo/path/to/skill
				owner := parts[0]
				repo := parts[1]
				subDir = strings.Join(parts[2:], "/")
				skillName = parts[len(parts)-1]
				repoURL = fmt.Sprintf("https://github.com/%s/%s.git", owner, repo)
			} else {
				// Standard install: owner/repo
				repoURL = "https://github.com/" + input
				parts := strings.Split(strings.TrimSuffix(repoURL, ".git"), "/")
				skillName = parts[len(parts)-1]
			}
		} else {
			// It's a URL
			repoURL = input
			parts := strings.Split(strings.TrimSuffix(repoURL, ".git"), "/")
			skillName = parts[len(parts)-1]
		}

		destPath := filepath.Join("skills", skillName)

		fmt.Printf("Installing %s...\n", skillName)

		if _, err := os.Stat(destPath); !os.IsNotExist(err) {
			fmt.Printf("Skill %s is already installed in %s\n", skillName, destPath)
			return
		}

		var err error
		if subDir != "" {
			err = git.InstallSubdir(repoURL, subDir, destPath)
		} else {
			err = git.Clone(repoURL, destPath)
		}

		if err != nil {
			fmt.Printf("Failed to install skill: %v\n", err)
			os.Exit(1)
		}

		// Update config
		cfg, err := config.LoadConfig()
		if err == nil {
			// Create skill info with metadata
			skillInfo := config.SkillInfo{
				Name:        skillName,
				Description: "Skill installed from " + input,
				URL:         repoURL,
			}
			if subDir != "" {
				skillInfo.URL = fmt.Sprintf("https://github.com/%s", input)
			}

			// Try to parse SKILL.md for better metadata
			if skill.FindSkillMD(destPath) {
				meta, err := skill.ParseSkillMD(destPath)
				if err == nil && meta != nil {
					if meta.Description != "" {
						skillInfo.Description = meta.Description
					}
				}
			}

			cfg.AddSkillInfo(skillInfo)
			err = cfg.Save()
			if err != nil {
				fmt.Printf("Warning: Failed to update ask.yaml: %v\n", err)
			}
		} else {
			// If config doesn't exist, we might be in a non-init project
			fmt.Println("Warning: ask.yaml not found. Run 'ask init' to track dependencies.")
		}

		fmt.Printf("Successfully installed %s!\n", skillName)
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
}
