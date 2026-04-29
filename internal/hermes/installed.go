package hermes

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/yeasy/ask/internal/config"
	"github.com/yeasy/ask/internal/skill"
)

// HermesSkillOwnership describes who owns or manages a visible Hermes skill.
type HermesSkillOwnership string

const (
	HermesSkillOwnershipASK      HermesSkillOwnership = "ask"
	HermesSkillOwnershipImported HermesSkillOwnership = "imported"
	HermesSkillOwnershipNative   HermesSkillOwnership = "hermes-native"
	HermesSkillOwnershipBundled  HermesSkillOwnership = "bundled"
)

// InstalledHermesSkill is metadata discovered from a Hermes skills directory.
type InstalledHermesSkill struct {
	Name           string
	Description    string
	Version        string
	Path           string
	RelativePath   string
	Ownership      HermesSkillOwnership
	Managed        bool
	Source         string
	UpdateStrategy string
}

// InstalledScanOptions configures ScanInstalledSkills.
type InstalledScanOptions struct {
	LockFile *config.LockFile
	MaxDepth int
}

const defaultInstalledScanMaxDepth = 5

// ScanInstalledSkills discovers Hermes skill directories without mutating them.
func ScanInstalledSkills(skillsDir string, opts InstalledScanOptions) ([]InstalledHermesSkill, error) {
	if strings.TrimSpace(skillsDir) == "" {
		return nil, nil
	}
	rootInfo, err := os.Lstat(skillsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	if !rootInfo.IsDir() || rootInfo.Mode()&os.ModeSymlink != 0 {
		return nil, nil
	}

	maxDepth := opts.MaxDepth
	if maxDepth <= 0 {
		maxDepth = defaultInstalledScanMaxDepth
	}
	var out []InstalledHermesSkill
	if err := scanInstalledDir(skillsDir, skillsDir, 0, maxDepth, &out); err != nil {
		return nil, err
	}
	applyLockMetadata(out, opts.LockFile)
	sort.Slice(out, func(i, j int) bool {
		if out[i].RelativePath == out[j].RelativePath {
			return out[i].Name < out[j].Name
		}
		return out[i].RelativePath < out[j].RelativePath
	})
	return out, nil
}

func scanInstalledDir(root, dir string, depth, maxDepth int, out *[]InstalledHermesSkill) error {
	info, err := os.Lstat(dir)
	if err != nil {
		return err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return nil
	}
	if isHiddenDir(info.Name()) && dir != root {
		return nil
	}

	if skill.FindSkillMD(dir) {
		installed, err := installedSkillFromDir(root, dir)
		if err != nil {
			return err
		}
		*out = append(*out, installed)
		return nil
	}
	if depth >= maxDepth {
		return nil
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if !entry.IsDir() || isHiddenDir(entry.Name()) {
			continue
		}
		entryInfo, err := os.Lstat(filepath.Join(dir, entry.Name()))
		if err != nil {
			return err
		}
		if entryInfo.Mode()&os.ModeSymlink != 0 {
			continue
		}
		if err := scanInstalledDir(root, filepath.Join(dir, entry.Name()), depth+1, maxDepth, out); err != nil {
			return err
		}
	}
	return nil
}

func installedSkillFromDir(root, dir string) (InstalledHermesSkill, error) {
	meta, err := skill.ParseSkillMD(dir)
	if err != nil {
		return InstalledHermesSkill{}, err
	}
	rel, err := filepath.Rel(root, dir)
	if err != nil {
		return InstalledHermesSkill{}, err
	}
	rel = filepath.ToSlash(rel)
	name := strings.TrimSpace(meta.Name)
	if name == "" {
		name = filepath.Base(dir)
	}
	installed := InstalledHermesSkill{
		Name:           name,
		Description:    meta.Description,
		Version:        meta.Version,
		Path:           dir,
		RelativePath:   rel,
		Ownership:      HermesSkillOwnershipNative,
		Managed:        false,
		Source:         "local",
		UpdateStrategy: "none",
	}
	return installed, nil
}

func applyLockMetadata(skills []InstalledHermesSkill, lockFile *config.LockFile) {
	locks := lockedSkillsByName(lockFile)
	if len(locks) == 0 {
		return
	}
	nameCounts := make(map[string]int, len(skills))
	for _, installed := range skills {
		nameCounts[installed.Name]++
	}
	for i := range skills {
		locked, ok := locks[skills[i].Name]
		if !ok || nameCounts[skills[i].Name] != 1 || skills[i].RelativePath != skills[i].Name {
			continue
		}
		skills[i].Ownership = HermesSkillOwnershipASK
		skills[i].Managed = true
		skills[i].Source = locked.Source
		if skills[i].Version == "" {
			skills[i].Version = locked.Version
		}
		if strings.TrimSpace(locked.URL) != "" {
			skills[i].UpdateStrategy = "git"
		}
	}
}

func lockedSkillsByName(lockFile *config.LockFile) map[string]config.LockEntry {
	locks := make(map[string]config.LockEntry)
	if lockFile == nil {
		return locks
	}
	for _, locked := range lockFile.Skills {
		name := strings.TrimSpace(locked.Name)
		if name != "" {
			locks[name] = locked
		}
	}
	return locks
}

func isHiddenDir(name string) bool {
	return strings.HasPrefix(name, ".")
}
