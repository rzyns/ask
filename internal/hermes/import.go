package hermes

import (
	"strings"
	"time"

	"github.com/yeasy/ask/internal/config"
)

// ImportCandidate describes one installed Hermes skill and the import action ASK
// would take for it.
type ImportCandidate struct {
	Skill          InstalledHermesSkill
	Classification string
	Action         string
}

const (
	HermesImportAlreadyManaged = "already-managed"
	HermesImportLocalOnly      = "local-only"
)

// PlanImport classifies installed Hermes skills for import/adoption.
func PlanImport(installed []InstalledHermesSkill, lockFile *config.LockFile, names []string, importAll bool) []ImportCandidate {
	want := make(map[string]bool, len(names))
	for _, name := range names {
		want[strings.TrimSpace(name)] = true
	}
	var out []ImportCandidate
	for _, s := range installed {
		if !importAll && len(want) > 0 && !want[s.Name] && !want[s.RelativePath] {
			continue
		}
		candidate := ImportCandidate{Skill: s}
		if lockFile != nil && lockFile.GetEntryForAgent(s.Name, string(config.AgentHermes)) != nil {
			candidate.Classification = HermesImportAlreadyManaged
			candidate.Action = "skip"
		} else if s.Managed || s.Ownership == HermesSkillOwnershipASK {
			candidate.Classification = HermesImportAlreadyManaged
			candidate.Action = "skip"
		} else {
			candidate.Classification = HermesImportLocalOnly
			candidate.Action = "import as local"
		}
		out = append(out, candidate)
	}
	return out
}

// LockEntryForImportedSkill creates the lock metadata for a local-only in-place
// Hermes skill import.
func LockEntryForImportedSkill(s InstalledHermesSkill) (config.LockEntry, error) {
	checksum, err := ChecksumSkillDir(s.Path)
	if err != nil {
		return config.LockEntry{}, err
	}
	return config.LockEntry{
		Name:           s.Name,
		Agent:          string(config.AgentHermes),
		Source:         "local",
		URL:            "",
		Version:        s.Version,
		InstalledAt:    time.Now(),
		Ownership:      string(HermesSkillOwnershipImported),
		InstallMode:    "in-place",
		UpdateStrategy: "none",
		TargetPath:     s.Path,
		SourcePath:     s.Path,
		Checksum:       checksum,
	}, nil
}
