package hermes

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/yeasy/ask/internal/config"
)

// UpdateSkipReason identifies why a Hermes skill cannot be updated safely.
type UpdateSkipReason string

const (
	UpdateSkipUnavailable UpdateSkipReason = "update unavailable"
	UpdateSkipBundled     UpdateSkipReason = "bundled skill"
	UpdateSkipDirty       UpdateSkipReason = "local modifications"
)

// UpdateSourceMetadata carries provenance needed by the command layer when it
// delegates an approved update to the installer.
type UpdateSourceMetadata struct {
	Source           string
	SourceIdentifier string
	UpdateStrategy   string
}

// UpdateOptions configures a conservative Hermes update plan.
type UpdateOptions struct {
	LockFile *config.LockFile
	Names    []string
	Force    bool
}

// UpdateCandidate is an ASK-owned Hermes skill that passed provenance and dirty
// state checks and may be handed to the installer for replacement.
type UpdateCandidate struct {
	Entry          config.LockEntry
	Input          string
	SourceMetadata UpdateSourceMetadata
}

// UpdateSkipped records a Hermes lock entry that was intentionally not updated.
type UpdateSkipped struct {
	Entry  config.LockEntry
	Reason UpdateSkipReason
}

// UpdatePlan summarizes safe Hermes update decisions.
type UpdatePlan struct {
	Updateable []UpdateCandidate
	Skipped    []UpdateSkipped
	Blocked    []UpdateSkipped
}

// PlanUpdate inspects Hermes lock entries and selects only ASK-owned, updateable,
// clean skills. Named updates are strict: missing or unsafe named skills return
// an error so callers do not silently ignore user intent.
func PlanUpdate(opts UpdateOptions) (UpdatePlan, error) {
	lockFile := opts.LockFile
	if lockFile == nil {
		lockFile = &config.LockFile{Version: 1}
	}
	want := make(map[string]bool, len(opts.Names))
	matched := make(map[string]bool, len(opts.Names))
	for _, name := range opts.Names {
		name = strings.TrimSpace(name)
		if name != "" {
			want[name] = true
		}
	}

	plan := UpdatePlan{}
	for _, entry := range lockFile.Skills {
		if !isHermesUpdateLockEntry(entry) {
			continue
		}
		name := strings.TrimSpace(entry.Name)
		if len(want) > 0 && !want[name] {
			continue
		}
		if want[name] {
			matched[name] = true
		}
		candidate, skipped, err := classifyUpdateEntry(entry, opts.Force)
		if err != nil {
			plan.Blocked = append(plan.Blocked, skipped)
			if len(want) > 0 {
				return plan, err
			}
			continue
		}
		if skipped.Reason != "" {
			plan.Skipped = append(plan.Skipped, skipped)
			if len(want) > 0 {
				return plan, fmt.Errorf("Hermes skill %q: %s", entry.Name, skipped.Reason)
			}
			continue
		}
		plan.Updateable = append(plan.Updateable, candidate)
	}
	for name := range want {
		if !matched[name] {
			return plan, fmt.Errorf("Hermes skill %q is not installed for Hermes", name)
		}
	}
	return plan, nil
}

func isHermesUpdateLockEntry(entry config.LockEntry) bool {
	if strings.EqualFold(strings.TrimSpace(entry.Agent), "hermes") {
		return true
	}
	// Legacy pre-agent entries are considered only when they carry Hermes
	// ownership/provenance metadata. Generic package-manager lock entries can
	// share the same name/URL and must not shadow the agent-scoped Hermes entry.
	return strings.TrimSpace(entry.Ownership) != "" || strings.TrimSpace(entry.InstallMode) != "" || strings.TrimSpace(entry.TargetPath) != ""
}

func classifyUpdateEntry(entry config.LockEntry, force bool) (UpdateCandidate, UpdateSkipped, error) {
	if isBundledHermesSkillPath(entry.TargetPath) || strings.EqualFold(strings.TrimSpace(entry.Ownership), string(HermesSkillOwnershipBundled)) {
		skipped := UpdateSkipped{Entry: entry, Reason: UpdateSkipBundled}
		return UpdateCandidate{}, skipped, fmt.Errorf("Hermes skill %q is bundled and cannot be managed by ASK", entry.Name)
	}
	if strings.TrimSpace(entry.Ownership) != string(HermesSkillOwnershipASK) {
		return UpdateCandidate{}, UpdateSkipped{Entry: entry, Reason: UpdateSkipUnavailable}, nil
	}
	strategy := strings.TrimSpace(entry.UpdateStrategy)
	if strategy != "hermes-index" && strategy != "git" {
		return UpdateCandidate{}, UpdateSkipped{Entry: entry, Reason: UpdateSkipUnavailable}, nil
	}
	if !force {
		if err := verifyCleanChecksum(entry); err != nil {
			skipped := UpdateSkipped{Entry: entry, Reason: UpdateSkipDirty}
			return UpdateCandidate{}, skipped, err
		}
	}
	input := updateInputForEntry(entry)
	if input == "" {
		return UpdateCandidate{}, UpdateSkipped{Entry: entry, Reason: UpdateSkipUnavailable}, nil
	}
	return UpdateCandidate{
		Entry: entry,
		Input: input,
		SourceMetadata: UpdateSourceMetadata{
			Source:           entry.Source,
			SourceIdentifier: entry.SourceIdentifier,
			UpdateStrategy:   entry.UpdateStrategy,
		},
	}, UpdateSkipped{}, nil
}

func updateInputForEntry(entry config.LockEntry) string {
	identifier := strings.Trim(strings.TrimSpace(entry.SourceIdentifier), "/")
	url := strings.TrimSpace(entry.URL)
	if strings.TrimSpace(entry.UpdateStrategy) == "hermes-index" {
		if identifier != "" && len(strings.Split(identifier, "/")) <= 2 && strings.TrimSpace(entry.Source) == config.RepoTypeHermes {
			return identifier
		}
		if isConcreteHermesSkillURL(url) {
			return normalizeHermesGitHubPathURL(url)
		}
		return ""
	}
	if identifier != "" && len(strings.Split(identifier, "/")) <= 2 {
		return identifier
	}
	if url != "" {
		return normalizeHermesGitHubPathURL(url)
	}
	return identifier
}

func isConcreteHermesSkillURL(url string) bool {
	const prefix = "https://github.com/NousResearch/hermes-agent/"
	if !strings.HasPrefix(url, prefix) {
		return false
	}
	return strings.Contains(url, "/optional-skills/") || strings.Contains(url, "/skills/") || strings.HasPrefix(strings.TrimPrefix(url, prefix), "optional-skills/") || strings.HasPrefix(strings.TrimPrefix(url, prefix), "skills/")
}

func normalizeHermesGitHubPathURL(url string) string {
	const prefix = "https://github.com/NousResearch/hermes-agent/"
	if strings.HasPrefix(url, prefix+"optional-skills/") || strings.HasPrefix(url, prefix+"skills/") {
		return prefix + "tree/main/" + strings.TrimPrefix(url, prefix)
	}
	return url
}

func verifyCleanChecksum(entry config.LockEntry) error {
	expected := strings.TrimSpace(entry.Checksum)
	if expected == "" {
		return fmt.Errorf("Hermes skill %q has no checksum; refusing update without --force", entry.Name)
	}
	paths := checksumVerificationPaths(entry)
	if len(paths) == 0 {
		return fmt.Errorf("Hermes skill %q has no source or target path; refusing update without --force", entry.Name)
	}
	expected = strings.TrimPrefix(expected, "sha256:")
	for _, path := range paths {
		actual, err := directoryChecksum(path)
		if err != nil {
			return fmt.Errorf("Hermes skill %q checksum failed: %w", entry.Name, err)
		}
		actual = strings.TrimPrefix(actual, "sha256:")
		if actual != expected {
			return fmt.Errorf("Hermes skill %q has local modifications; refusing update without --force", entry.Name)
		}
	}
	return nil
}

func checksumVerificationPaths(entry config.LockEntry) []string {
	var paths []string
	if source := strings.TrimSpace(entry.SourcePath); source != "" {
		paths = append(paths, source)
	}
	target := strings.TrimSpace(entry.TargetPath)
	if target == "" || isSymlinkPath(target) {
		return paths
	}
	if len(paths) > 0 && filepath.Clean(paths[0]) == filepath.Clean(target) {
		return paths
	}
	return append(paths, target)
}

func isSymlinkPath(path string) bool {
	info, err := os.Lstat(path)
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeSymlink != 0
}
