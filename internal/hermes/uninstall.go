package hermes

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/yeasy/ask/internal/config"
)

// UninstallOptions controls Hermes ownership-aware uninstall behavior.
type UninstallOptions struct {
	Global      bool
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

	safeName, err := validateHermesSkillName(name)
	if err != nil {
		return UninstallAction{}, err
	}

	switch entry.Ownership {
	case string(HermesSkillOwnershipASK), "":
		var action UninstallAction
		if entry.TargetPath != "" {
			if err := validateASKOwnedTargetPath(*entry, safeName, opts.Global); err != nil {
				return action, err
			}
			if err := os.RemoveAll(entry.TargetPath); err != nil {
				return action, fmt.Errorf("remove Hermes target: %w", err)
			}
			action.RemovedFiles = true
		}
		if entry.SourcePath != "" && entry.SourcePath != entry.TargetPath {
			if err := validateASKOwnedSourcePath(*entry, safeName, opts.Global); err != nil {
				return action, err
			}
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
			if err := validateImportedTargetPath(*entry, safeName, opts.Global); err != nil {
				return action, err
			}
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

func validateHermesSkillName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" || name == "." || name == ".." || filepath.Base(name) != name || strings.ContainsAny(name, `/\\`) {
		return "", fmt.Errorf("invalid Hermes skill name %q", name)
	}
	return name, nil
}

func validateASKOwnedTargetPath(entry config.LockEntry, skillName string, global bool) error {
	return validateHermesLockedPath(entry.TargetPath, skillName, scopedHermesSkillRoots(global), "Hermes target")
}

func validateASKOwnedSourcePath(entry config.LockEntry, skillName string, global bool) error {
	return validateHermesLockedPath(entry.SourcePath, skillName, scopedASKSkillRoots(global), "ASK source")
}

func validateImportedTargetPath(entry config.LockEntry, skillName string, global bool) error {
	return validateHermesLockedPath(entry.TargetPath, skillName, scopedHermesSkillRoots(global), "imported Hermes directory")
}

type expectedSkillRootFunc func() (string, bool)

func scopedHermesSkillRoots(global bool) []expectedSkillRootFunc {
	if global {
		return []expectedSkillRootFunc{globalHermesSkillsRoot}
	}
	return []expectedSkillRootFunc{projectHermesSkillsRoot}
}

func scopedASKSkillRoots(global bool) []expectedSkillRootFunc {
	if global {
		return []expectedSkillRootFunc{globalASKSkillsRoot}
	}
	return []expectedSkillRootFunc{projectASKSkillsRoot}
}

func validateHermesLockedPath(path, skillName string, rootFuncs []expectedSkillRootFunc, label string) error {
	cleanPath, err := cleanLockedSkillPath(path, skillName, label)
	if err != nil {
		return err
	}
	for _, rootFunc := range rootFuncs {
		root, ok := rootFunc()
		if !ok {
			continue
		}
		if pathIsDirectChildOfRoot(cleanPath, root, skillName) {
			return nil
		}
	}
	return fmt.Errorf("refusing to remove %s %q: path is outside expected Hermes/ASK skill roots for %q", label, path, skillName)
}

func cleanLockedSkillPath(path, skillName, label string) (string, error) {
	if strings.TrimSpace(path) == "" {
		return "", fmt.Errorf("refusing to remove empty %s path for %q", label, skillName)
	}
	cleanPath := filepath.Clean(path)
	if cleanPath != path {
		return "", fmt.Errorf("refusing to remove %s %q: path is not clean", label, path)
	}
	if !filepath.IsAbs(cleanPath) {
		absPath, err := filepath.Abs(cleanPath)
		if err != nil {
			return "", fmt.Errorf("refusing to remove %s %q: resolve absolute path: %w", label, path, err)
		}
		cleanPath = absPath
	}
	volume := filepath.VolumeName(cleanPath)
	if cleanPath == volume+string(os.PathSeparator) {
		return "", fmt.Errorf("refusing to remove %s %q: path is filesystem root", label, path)
	}
	if filepath.Base(cleanPath) != skillName {
		return "", fmt.Errorf("refusing to remove %s %q: basename does not match skill %q", label, path, skillName)
	}
	return cleanPath, nil
}

func pathIsDirectChildOfRoot(path, root, skillName string) bool {
	root = filepath.Clean(root)
	if !filepath.IsAbs(root) {
		absRoot, err := filepath.Abs(root)
		if err != nil {
			return false
		}
		root = absRoot
	}
	return path == filepath.Join(root, skillName)
}

func projectHermesSkillsRoot() (string, bool) {
	root, err := filepath.Abs(filepath.Join(".hermes", "skills"))
	return root, err == nil
}

func globalHermesSkillsRoot() (string, bool) {
	root, err := config.GetAgentSkillsDir(config.AgentHermes, true)
	return root, err == nil
}

func projectASKSkillsRoot() (string, bool) {
	root, err := filepath.Abs(config.DefaultSkillsDir)
	return root, err == nil
}

func globalASKSkillsRoot() (string, bool) {
	root, err := config.GetGlobalSkillsDir()
	return root, err == nil
}
