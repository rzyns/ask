package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/yeasy/ask/internal/config"
	"github.com/yeasy/ask/internal/hermes"
	"github.com/yeasy/ask/internal/installer"
	"github.com/yeasy/ask/internal/ui"
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update [skill-name]",
	Short: "Update installed skills to latest version",
	Long: `Update one or all installed skills to their latest versions.
If no skill name is provided, updates all installed skills.
Use --global to update global skills.`,
	Example: `  # Update all installed skills
  ask skill update
  
  # Update a specific skill
  ask skill update pdf
  
  # Update global skills
  ask skill update --global`,
	RunE: runUpdate,
}

var installHermesUpdate = installer.Install

func init() {
	skillCmd.AddCommand(updateCmd)
	registerUpdateFlags(updateCmd)
}

func registerUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("agent", "a", "", "target agent for update (use 'hermes' for Hermes-aware updates)")
	cmd.Flags().Bool("force", false, "force update despite local modifications when supported")
}

func runUpdate(cmd *cobra.Command, args []string) error {
	agent, _ := cmd.Flags().GetString("agent")
	if strings.EqualFold(strings.TrimSpace(agent), string(config.AgentHermes)) {
		return runHermesUpdate(cmd, args)
	}
	return runGenericUpdate(cmd, args)
}

func runHermesUpdate(cmd *cobra.Command, args []string) error {
	global, _ := cmd.Flags().GetBool("global")
	force, _ := cmd.Flags().GetBool("force")
	if !global {
		if !ensureInitialized() {
			return nil
		}
	}
	cfg, err := config.LoadConfigByScope(global)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	lockFile, err := config.LoadLockFileByScope(global)
	if err != nil {
		return fmt.Errorf("load lock file: %w", err)
	}
	plan, err := hermes.PlanUpdate(hermes.UpdateOptions{LockFile: lockFile, Names: args, Force: force})
	if err != nil {
		return err
	}
	for _, skipped := range plan.Skipped {
		ui.Warn(fmt.Sprintf("Skipping %s: %s", skipped.Entry.Name, skipped.Reason))
	}
	for _, blocked := range plan.Blocked {
		ui.Warn(fmt.Sprintf("Refusing %s: %s", blocked.Entry.Name, blocked.Reason))
	}
	if len(plan.Updateable) == 0 {
		fmt.Println("No Hermes skills to update.")
		return nil
	}
	for _, candidate := range plan.Updateable {
		fmt.Printf("Updating Hermes skill %s...\n", candidate.Entry.Name)
		if err := installHermesUpdate(candidate.Input, installer.InstallOptions{
			Global:                   global,
			Agents:                   []string{string(config.AgentHermes)},
			Config:                   cfg,
			SkipScore:                true,
			ReplaceExisting:          true,
			ReplaceExistingName:      candidate.Entry.Name,
			ReplaceExistingSource:    candidate.Entry.SourcePath,
			ReplaceExistingTarget:    candidate.Entry.TargetPath,
			SuppressGenericLockEntry: true,
			SourceMetadata: &installer.InstallSourceMetadata{
				Source:           candidate.SourceMetadata.Source,
				SourceIdentifier: candidate.SourceMetadata.SourceIdentifier,
				UpdateStrategy:   candidate.SourceMetadata.UpdateStrategy,
			},
		}); err != nil {
			return fmt.Errorf("update Hermes skill %s: %w", candidate.Entry.Name, err)
		}
	}
	return nil
}

func runGenericUpdate(cmd *cobra.Command, args []string) error {
	global, _ := cmd.Flags().GetBool("global")

	// Ensure project is initialized for non-global operations
	if !global {
		if !ensureInitialized() {
			return nil
		}
	}

	cfg, err := config.LoadConfigByScope(global)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	if len(cfg.Skills) == 0 && len(cfg.SkillsInfo) == 0 {
		scopeLabel := "project"
		if global {
			scopeLabel = "global"
		}
		fmt.Printf("No %s skills installed.\n", scopeLabel)
		return nil
	}

	allSkills := cfg.GetAllSkillNames()

	var skillsToUpdate []string
	if len(args) > 0 {
		skillName := args[0]
		found := false
		for _, s := range cfg.Skills {
			if s == skillName {
				found = true
				break
			}
		}
		if !found {
			for _, si := range cfg.SkillsInfo {
				if si.Name == skillName {
					found = true
					break
				}
			}
		}
		if !found {
			return fmt.Errorf("skill %q is not installed", skillName)
		}
		skillsToUpdate = []string{skillName}
	} else {
		skillsToUpdate = allSkills
	}

	skillsDir, err := config.GetSkillsDirByScope(global)
	if err != nil {
		return err
	}
	for _, skillName := range skillsToUpdate {
		skillPath := filepath.Join(skillsDir, skillName)

		gitDir := filepath.Join(skillPath, ".git")
		if _, err := os.Stat(gitDir); os.IsNotExist(err) {
			ui.Debug(fmt.Sprintf("Skipping %s (not a git repository)", skillName))
			continue
		}

		ui.Debug(fmt.Sprintf("Updating %s...", skillName))

		pullCtx, pullCancel := context.WithTimeout(context.Background(), 60*time.Second)
		gitCmd := exec.CommandContext(pullCtx, "git", "pull", "--ff-only")
		gitCmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
		gitCmd.Dir = skillPath
		gitCmd.Stdout = os.Stdout
		gitCmd.Stderr = os.Stderr

		runErr := gitCmd.Run()
		pullCancel()
		if runErr != nil {
			ui.Warn(fmt.Sprintf("  Failed to update %s: %v", skillName, runErr))
			continue
		}

		ui.Debug(fmt.Sprintf("  Updated %s successfully!", skillName))
	}
	return nil
}
