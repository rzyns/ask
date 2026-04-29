package hermes

import (
	"fmt"
	"os"

	"github.com/yeasy/ask/internal/config"
)

// UninstallOptions controls Hermes ownership-aware uninstall behavior.
type UninstallOptions struct {
	Forget      bool
	DeleteFiles bool
}

// UninstallAction describes the removal operation performed.
type UninstallAction struct {
	Forgot        bool
	RemovedFiles  bool
	RemovedSource bool
}

// UninstallSkill applies Hermes ownership rules to a single lock entry.
func UninstallSkill(lockFile *config.LockFile, name string, opts UninstallOptions) (UninstallAction, error) {
	if lockFile == nil {
		return UninstallAction{}, fmt.Errorf("%s is not tracked by ASK for Hermes", name)
	}
	entry := lockFile.GetEntryForAgent(name, string(config.AgentHermes))
	if entry == nil {
		return UninstallAction{}, fmt.Errorf("%s is not tracked by ASK for Hermes; remove it manually or import it first", name)
	}

	switch entry.Ownership {
	case string(HermesSkillOwnershipASK), "":
		var action UninstallAction
		if entry.TargetPath != "" {
			if err := os.RemoveAll(entry.TargetPath); err != nil {
				return action, fmt.Errorf("remove Hermes target: %w", err)
			}
			action.RemovedFiles = true
		}
		if entry.SourcePath != "" && entry.SourcePath != entry.TargetPath {
			if err := os.RemoveAll(entry.SourcePath); err != nil {
				return action, fmt.Errorf("remove ASK source: %w", err)
			}
			action.RemovedSource = true
		}
		lockFile.RemoveEntryForAgent(name, string(config.AgentHermes))
		action.Forgot = true
		return action, nil
	case string(HermesSkillOwnershipImported):
		if !opts.Forget && !opts.DeleteFiles {
			return UninstallAction{}, fmt.Errorf("%s was imported in-place and was not installed by ASK; use --forget to remove tracking or --delete-files to remove the Hermes directory", name)
		}
		var action UninstallAction
		if opts.DeleteFiles && entry.TargetPath != "" {
			if err := os.RemoveAll(entry.TargetPath); err != nil {
				return action, fmt.Errorf("remove imported Hermes directory: %w", err)
			}
			action.RemovedFiles = true
		}
		lockFile.RemoveEntryForAgent(name, string(config.AgentHermes))
		action.Forgot = true
		return action, nil
	case string(HermesSkillOwnershipBundled):
		return UninstallAction{}, fmt.Errorf("%s is a bundled Hermes skill and cannot be uninstalled by ASK", name)
	default:
		return UninstallAction{}, fmt.Errorf("%s has unsupported Hermes ownership %q", name, entry.Ownership)
	}
}
