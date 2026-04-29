package hermes

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/yeasy/ask/internal/config"
)

// UninstallOptions describes a conservative Hermes uninstall request.
type UninstallOptions struct {
	LockFile    *config.LockFile
	Name        string
	SkillsDir   string
	SourceDir   string
	Forget      bool
	DeleteFiles bool
}

// UninstallDecision is the file/tracking action set approved by PlanUninstall.
type UninstallDecision struct {
	Entry          config.LockEntry
	TargetPath     string
	SourcePath     string
	RemoveTarget   bool
	RemoveSource   bool
	RemoveTracking bool
}

// PlanUninstall chooses safe Hermes uninstall actions from lock provenance.
func PlanUninstall(opts UninstallOptions) (UninstallDecision, error) {
	name := strings.TrimSpace(opts.Name)
	if name == "" || filepath.Base(name) != name || name == "." || name == ".." {
		return UninstallDecision{}, fmt.Errorf("invalid Hermes skill name %q", opts.Name)
	}
	if opts.Forget && opts.DeleteFiles {
		return UninstallDecision{}, fmt.Errorf("--forget and --delete-files are mutually exclusive")
	}
	skillsDir := filepath.Clean(strings.TrimSpace(opts.SkillsDir))
	if skillsDir == "." || skillsDir == string(filepath.Separator) || skillsDir == "" {
		return UninstallDecision{}, fmt.Errorf("invalid Hermes skills directory")
	}
	sourceDir := filepath.Clean(strings.TrimSpace(opts.SourceDir))
	if sourceDir == "." || sourceDir == "" {
		sourceDir = ""
	}
	expectedTarget := filepath.Join(skillsDir, name)
	entry, err := findHermesUninstallEntry(opts.LockFile, name, expectedTarget)
	if err != nil {
		return UninstallDecision{}, err
	}
	if entry == nil {
		if _, statErr := os.Lstat(expectedTarget); statErr == nil {
			return UninstallDecision{}, fmt.Errorf("Hermes skill %q is unmanaged by ASK; import it for tracking or remove it manually", name)
		} else if os.IsNotExist(statErr) {
			return UninstallDecision{}, fmt.Errorf("Hermes skill %q is not installed for Hermes", name)
		} else {
			return UninstallDecision{}, fmt.Errorf("cannot inspect Hermes skill %q: %w", name, statErr)
		}
	}

	targetPath := strings.TrimSpace(entry.TargetPath)
	if targetPath == "" {
		targetPath = expectedTarget
	}
	targetPath = filepath.Clean(targetPath)
	if !pathWithinDir(targetPath, skillsDir) || filepath.Base(targetPath) != name {
		return UninstallDecision{}, fmt.Errorf("refusing to uninstall Hermes skill %q outside Hermes skills dir: %s", name, targetPath)
	}
	if isBundledHermesSkillPath(targetPath) || strings.EqualFold(strings.TrimSpace(entry.Ownership), string(HermesSkillOwnershipBundled)) {
		return UninstallDecision{}, fmt.Errorf("Hermes skill %q is bundled and cannot be uninstalled by ASK", name)
	}

	decision := UninstallDecision{Entry: *entry, TargetPath: targetPath, SourcePath: strings.TrimSpace(entry.SourcePath), RemoveTracking: true}
	switch strings.TrimSpace(entry.Ownership) {
	case string(HermesSkillOwnershipASK):
		if opts.Forget {
			return decision, nil
		}
		decision.RemoveTarget = true
		decision.RemoveSource = decision.SourcePath != ""
		if decision.RemoveSource {
			decision.SourcePath = filepath.Clean(decision.SourcePath)
			if sourceDir != "" && !pathWithinDir(decision.SourcePath, sourceDir) {
				return UninstallDecision{}, fmt.Errorf("refusing unsafe ASK source removal for Hermes skill %q outside ASK skills dir: %s", name, decision.SourcePath)
			}
			if samePath(decision.SourcePath, decision.TargetPath) || filepath.Base(decision.SourcePath) != name {
				return UninstallDecision{}, fmt.Errorf("refusing unsafe ASK source removal for Hermes skill %q: %s", name, decision.SourcePath)
			}
			if err := verifyUninstallChecksum(*entry, decision.SourcePath, decision.TargetPath); err != nil {
				return UninstallDecision{}, err
			}
		}
		return decision, nil
	case string(HermesSkillOwnershipImported):
		if opts.Forget {
			return decision, nil
		}
		if opts.DeleteFiles {
			decision.RemoveTarget = true
			return decision, nil
		}
		return UninstallDecision{}, fmt.Errorf("Hermes skill %q was imported in-place; use --forget to remove tracking only or --delete-files to delete files", name)
	default:
		return UninstallDecision{}, fmt.Errorf("Hermes skill %q is not ASK-owned; refusing destructive uninstall", name)
	}
}

func findHermesUninstallEntry(lockFile *config.LockFile, name, expectedTarget string) (*config.LockEntry, error) {
	if lockFile == nil {
		return nil, nil
	}
	var candidates []*config.LockEntry
	for i := range lockFile.Skills {
		entry := &lockFile.Skills[i]
		if entry.Name != name || !isHermesUninstallLockEntry(*entry) {
			continue
		}
		candidates = append(candidates, entry)
		if entry.TargetPath != "" && samePath(entry.TargetPath, expectedTarget) {
			return entry, nil
		}
	}
	if len(candidates) == 1 {
		return candidates[0], nil
	}
	if len(candidates) > 1 {
		return nil, fmt.Errorf("Hermes skill %q has multiple lock entries; refusing ambiguous uninstall", name)
	}
	return nil, nil
}

func isHermesUninstallLockEntry(entry config.LockEntry) bool {
	agent := strings.TrimSpace(entry.Agent)
	if strings.EqualFold(agent, "hermes") {
		return true
	}
	if agent != "" {
		return false
	}
	return strings.TrimSpace(entry.Ownership) != "" || strings.TrimSpace(entry.InstallMode) != "" || strings.TrimSpace(entry.TargetPath) != ""
}

func verifyUninstallChecksum(entry config.LockEntry, sourcePath, targetPath string) error {
	expected := strings.TrimPrefix(strings.TrimSpace(entry.Checksum), "sha256:")
	if expected == "" {
		return fmt.Errorf("Hermes skill %q has no checksum; refusing destructive uninstall", entry.Name)
	}
	if err := verifyUninstallPathChecksum(entry.Name, sourcePath, expected, "source"); err != nil {
		return err
	}
	if targetPath != "" && !samePath(sourcePath, targetPath) && !isSymlinkPath(targetPath) {
		if _, err := os.Lstat(targetPath); err == nil {
			if err := verifyUninstallPathChecksum(entry.Name, targetPath, expected, "target"); err != nil {
				return err
			}
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("cannot inspect Hermes skill %q target before uninstall: %w", entry.Name, err)
		}
	}
	return nil
}

func verifyUninstallPathChecksum(name, path, expected, label string) error {
	actual, err := directoryChecksum(path)
	if err != nil {
		return fmt.Errorf("failed to checksum Hermes skill %q %s before uninstall: %w", name, label, err)
	}
	actual = strings.TrimPrefix(strings.TrimSpace(actual), "sha256:")
	if actual != expected {
		return fmt.Errorf("Hermes skill %q %s has local changes; refusing uninstall", name, label)
	}
	return nil
}

func pathWithinDir(path, dir string) bool {
	path = filepath.Clean(path)
	dir = filepath.Clean(dir)
	rel, err := filepath.Rel(dir, path)
	if err != nil {
		return false
	}
	return rel == "." || (rel != "" && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)))
}

func samePath(left, right string) bool {
	return filepath.Clean(left) == filepath.Clean(right)
}
