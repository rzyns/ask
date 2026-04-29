package hermes

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/yeasy/ask/internal/config"
)

// ImportOptions configures conservative adoption of existing Hermes skills.
type ImportOptions struct {
	SkillsDir string
	LockFile  *config.LockFile
	All       bool
	Names     []string
}

// ImportCandidate is a local skill that can be adopted into ask.lock.
type ImportCandidate struct {
	Skill InstalledHermesSkill
	Entry config.LockEntry
}

// ImportResult summarizes a Hermes skill import plan.
type ImportResult struct {
	Importable     []ImportCandidate
	SkippedManaged []InstalledHermesSkill
	SkippedBundled []InstalledHermesSkill
	UnmatchedNames []string
}

// PlanImport scans local Hermes skills and builds conservative in-place lock entries.
func PlanImport(opts ImportOptions) (ImportResult, error) {
	lockFile := opts.LockFile
	installed, err := ScanInstalledSkills(opts.SkillsDir, InstalledScanOptions{LockFile: lockFile})
	if err != nil {
		return ImportResult{}, err
	}
	want := map[string]bool{}
	matched := map[string]bool{}
	var wantedNames []string
	for _, name := range opts.Names {
		name = strings.TrimSpace(name)
		if name != "" {
			want[name] = true
			wantedNames = append(wantedNames, name)
		}
	}
	result := ImportResult{}
	for _, s := range installed {
		if !opts.All && len(want) > 0 && !want[s.Name] {
			continue
		}
		if want[s.Name] {
			matched[s.Name] = true
		}
		if !opts.All && len(want) == 0 {
			continue
		}
		if isBundledHermesSkillPath(s.Path) {
			result.SkippedBundled = append(result.SkippedBundled, s)
			continue
		}
		if lockFile != nil && getHermesLockEntryForTargetPath(lockFile, s.Name, s.Path) != nil {
			result.SkippedManaged = append(result.SkippedManaged, s)
			continue
		}
		checksum, err := directoryChecksum(s.Path)
		if err != nil {
			return ImportResult{}, err
		}
		result.Importable = append(result.Importable, ImportCandidate{Skill: s, Entry: config.LockEntry{
			Name:           s.Name,
			Source:         "local",
			Version:        s.Version,
			InstalledAt:    time.Now().UTC(),
			Agent:          "hermes",
			Ownership:      string(HermesSkillOwnershipImported),
			InstallMode:    "in-place",
			UpdateStrategy: "none",
			TargetPath:     s.Path,
			Checksum:       checksum,
		}})
	}
	for _, name := range wantedNames {
		if !matched[name] {
			result.UnmatchedNames = append(result.UnmatchedNames, name)
		}
	}
	return result, nil
}

// ApplyImport adds planned import entries to the lock file.
func ApplyImport(lockFile *config.LockFile, result ImportResult) {
	for _, candidate := range result.Importable {
		lockFile.AddEntry(candidate.Entry)
	}
}

func getHermesLockEntryForTargetPath(lockFile *config.LockFile, name, targetPath string) *config.LockEntry {
	requestedPath := normalizeInstalledPath(targetPath)
	for i := range lockFile.Skills {
		locked := &lockFile.Skills[i]
		if locked.Name != name || !isHermesLockEntry(*locked) {
			continue
		}
		lockedPath := normalizeInstalledPath(locked.TargetPath)
		if lockedPath == "" || lockedPath == requestedPath {
			return locked
		}
	}
	return nil
}

func isBundledHermesSkillPath(path string) bool {
	parts := splitCleanPath(filepath.ToSlash(path))
	for i := 0; i+2 < len(parts); i++ {
		if strings.EqualFold(parts[i], "NousResearch") && strings.EqualFold(parts[i+1], "hermes-agent") && strings.EqualFold(parts[i+2], "skills") {
			return true
		}
	}
	return ClassifyHermesSource(path).Kind == HermesSourceBundled
}

func directoryChecksum(root string) (string, error) {
	var files []string
	if err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		files = append(files, path)
		return nil
	}); err != nil {
		return "", err
	}
	sort.Strings(files)
	h := sha256.New()
	for _, path := range files {
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return "", err
		}
		_, _ = io.WriteString(h, filepath.ToSlash(rel))
		_, _ = io.WriteString(h, "\x00")
		f, err := os.Open(path)
		if err != nil {
			return "", err
		}
		if _, err := io.Copy(h, f); err != nil {
			_ = f.Close()
			return "", err
		}
		if err := f.Close(); err != nil {
			return "", err
		}
		_, _ = io.WriteString(h, "\x00")
	}
	return fmt.Sprintf("sha256:%s", hex.EncodeToString(h.Sum(nil))), nil
}
