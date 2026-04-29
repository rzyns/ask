package hermes

import (
	"testing"

	"github.com/yeasy/ask/internal/config"
)

func TestPlanImportClassifiesAlreadyManagedAndLocalOnly(t *testing.T) {
	installed := []InstalledHermesSkill{
		{Name: "managed", Path: "/tmp/managed", RelativePath: "managed", Ownership: HermesSkillOwnershipASK, Managed: true},
		{Name: "local", Path: "/tmp/local", RelativePath: "local", Ownership: HermesSkillOwnershipNative},
	}
	lock := &config.LockFile{Version: 1, Skills: []config.LockEntry{{Name: "managed", Agent: "hermes"}}}

	plan := PlanImport(installed, lock, nil, true)
	if len(plan) != 2 {
		t.Fatalf("got %d candidates, want 2", len(plan))
	}
	if plan[0].Classification != HermesImportAlreadyManaged || plan[0].Action != "skip" {
		t.Fatalf("managed candidate = %#v", plan[0])
	}
	if plan[1].Classification != HermesImportLocalOnly || plan[1].Action != "import as local" {
		t.Fatalf("local candidate = %#v", plan[1])
	}
}

func TestLockEntryForImportedSkillRecordsLocalOnlyMetadata(t *testing.T) {
	root := t.TempDir()
	writeSkill(t, root, "local-skill", "local-skill", "Local", "0.1.0")
	scanned, err := ScanInstalledSkills(root, InstalledScanOptions{})
	if err != nil {
		t.Fatalf("ScanInstalledSkills returned error: %v", err)
	}
	entry, err := LockEntryForImportedSkill(scanned[0])
	if err != nil {
		t.Fatalf("LockEntryForImportedSkill returned error: %v", err)
	}
	if entry.Agent != "hermes" || entry.Ownership != "imported" || entry.InstallMode != "in-place" || entry.UpdateStrategy != "none" || entry.TargetPath == "" || entry.Checksum == "" {
		t.Fatalf("imported entry missing metadata: %#v", entry)
	}
}
