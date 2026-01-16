package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/yeasy/ask/internal/config"
	"github.com/yeasy/ask/internal/github"
)

// outdatedCmd represents the outdated command
var outdatedCmd = &cobra.Command{
	Use:   "outdated",
	Short: "Check for outdated skills",
	Long:  `Check which installed skills have updates available.`,
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

		lockFile, _ := config.LoadLockFile()

		fmt.Println("Checking for updates...")
		fmt.Println()

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "SKILL\tCURRENT\tLATEST\tSTATUS")

		outdatedCount := 0

		for _, skillName := range cfg.Skills {
			skillPath := filepath.Join(config.DefaultSkillsDir, skillName)

			// Check if it's a git repository
			gitDir := filepath.Join(skillPath, ".git")
			if _, err := os.Stat(gitDir); os.IsNotExist(err) {
				fmt.Fprintf(w, "%s\t-\t-\tNot a git repo\n", skillName)
				continue
			}

			// Get current commit
			currentCommit := getShortCommit(skillPath)

			// Fetch latest from remote (skip if offline)
			remoteCommit := ""
			status := "✓ Up to date"

			if !github.OfflineMode {
				fetchCmd := exec.Command("git", "fetch", "--quiet")
				fetchCmd.Dir = skillPath
				fetchCmd.Run()

				// Get remote HEAD commit
				remoteCommit = getRemoteHeadCommit(skillPath)
			} else {
				remoteCommit = "?"
				status = "? Unknown (Offline)"
			}

			// Get lock file info
			lockEntry := lockFile.GetEntry(skillName)
			lockedVersion := ""
			if lockEntry != nil && lockEntry.Version != "" {
				lockedVersion = lockEntry.Version
			}

			// Compare
			// Compare
			if !github.OfflineMode {
				status = "✓ Up to date"
				if currentCommit != remoteCommit && remoteCommit != "" {
					status = "⬆ Update available"
					outdatedCount++
				}
			}

			currentDisplay := currentCommit
			if lockedVersion != "" {
				currentDisplay = lockedVersion + " (" + currentCommit + ")"
			}

			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", skillName, currentDisplay, remoteCommit, status)
		}

		w.Flush()

		fmt.Println()
		if outdatedCount > 0 {
			fmt.Printf("%d skill(s) can be updated. Run 'ask update' to update all.\n", outdatedCount)
		} else {
			fmt.Println("All skills are up to date.")
		}
	},
}

// getShortCommit returns the short commit hash
func getShortCommit(repoPath string) string {
	cmd := exec.Command("git", "rev-parse", "--short", "HEAD")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return "-"
	}
	return strings.TrimSpace(string(output))
}

// getRemoteHeadCommit returns the short commit hash of remote HEAD
func getRemoteHeadCommit(repoPath string) string {
	cmd := exec.Command("git", "rev-parse", "--short", "origin/HEAD")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		// Try origin/main or origin/master
		cmd = exec.Command("git", "rev-parse", "--short", "origin/main")
		cmd.Dir = repoPath
		output, err = cmd.Output()
		if err != nil {
			cmd = exec.Command("git", "rev-parse", "--short", "origin/master")
			cmd.Dir = repoPath
			output, err = cmd.Output()
			if err != nil {
				return "-"
			}
		}
	}
	return strings.TrimSpace(string(output))
}

func init() {
	skillCmd.AddCommand(outdatedCmd)
}
