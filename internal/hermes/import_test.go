package hermes

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yeasy/ask/internal/config"
)

func TestPlanImportHermesSkillsSkipsManagedAndBundled(t *testing.T) {
	root := t.TempDir()
	writeImportSkill(t, root, "native", "native")
	writeImportSkill(t, root, "managed", "managed")
	bundledRoot := filepath.Join(t.TempDir(), "NousResearch", "hermes-agent", "skills")
	writeImportSkill(t, bundledRoot, "memory", "memory")

	lock := &config.LockFile{Version: 1, Skills: []config.LockEntry{{Name: "managed", Agent: "hermes"}}}
	res, err := PlanImport(ImportOptions{SkillsDir: root, LockFile: lock, All: true})
	if err != nil {
		t.Fatalf("PlanImport returned error: %v", err)
	}
	if len(res.Importable) != 1 || res.Importable[0].Entry.Name != "native" {
		t.Fatalf("importable = %#v, want only native", res.Importable)
	}
	if got := res.Importable[0].Entry; got.Ownership != "imported" || got.InstallMode != "in-place" || got.UpdateStrategy != "none" || got.TargetPath == "" || got.Checksum == "" || got.Agent != "hermes" {
		t.Fatalf("entry metadata = %#v", got)
	}

	bundled, err := PlanImport(ImportOptions{SkillsDir: bundledRoot, LockFile: &config.LockFile{Version: 1}, All: true})
	if err != nil {
		t.Fatalf("PlanImport bundled returned error: %v", err)
	}
	if len(bundled.Importable) != 0 || len(bundled.SkippedBundled) != 1 {
		t.Fatalf("bundled result = %#v, want skipped bundled", bundled)
	}
}

func TestPlanImportNamedOnly(t *testing.T) {
	root := t.TempDir()
	writeImportSkill(t, root, "one", "one")
	writeImportSkill(t, root, "two", "two")
	res, err := PlanImport(ImportOptions{SkillsDir: root, LockFile: &config.LockFile{Version: 1}, Names: []string{"two"}})
	if err != nil {
		t.Fatalf("PlanImport returned error: %v", err)
	}
	if len(res.Importable) != 1 || res.Importable[0].Entry.Name != "two" {
		t.Fatalf("importable = %#v, want only two", res.Importable)
	}
}

func TestPlanImportDoesNotSkipNonHermesLockedSkill(t *testing.T) {
	root := t.TempDir()
	writeImportSkill(t, root, "shared", "shared")
	lock := &config.LockFile{Version: 1, Skills: []config.LockEntry{{Name: "shared", Agent: "claude"}}}

	res, err := PlanImport(ImportOptions{SkillsDir: root, LockFile: lock, All: true})
	if err != nil {
		t.Fatalf("PlanImport returned error: %v", err)
	}
	if len(res.Importable) != 1 || res.Importable[0].Entry.Name != "shared" {
		t.Fatalf("importable = %#v, want shared despite non-Hermes lock entry", res.Importable)
	}
	if len(res.SkippedManaged) != 0 {
		t.Fatalf("skipped managed = %#v, want none for non-Hermes lock entry", res.SkippedManaged)
	}
}

func TestPlanImportDoesNotSkipLegacyNonHermesLockEntry(t *testing.T) {
	root := t.TempDir()
	writeImportSkill(t, root, "shared", "shared")
	lock := &config.LockFile{Version: 1, Skills: []config.LockEntry{{Name: "shared", Source: "legacy-non-hermes"}}}

	res, err := PlanImport(ImportOptions{SkillsDir: root, LockFile: lock, All: true})
	if err != nil {
		t.Fatalf("PlanImport returned error: %v", err)
	}
	if len(res.Importable) != 1 || res.Importable[0].Entry.Name != "shared" {
		t.Fatalf("importable = %#v, want shared despite legacy non-Hermes lock entry", res.Importable)
	}
	if len(res.SkippedManaged) != 0 {
		t.Fatalf("skipped managed = %#v, want none for legacy non-Hermes lock entry", res.SkippedManaged)
	}
}

func TestApplyImportKeepsDuplicateHermesNamesDistinctByTargetPath(t *testing.T) {
	root := t.TempDir()
	writeImportSkill(t, root, "one", "shared")
	writeImportSkill(t, root, "two", "shared")
	lock := &config.LockFile{Version: 1}

	res, err := PlanImport(ImportOptions{SkillsDir: root, LockFile: lock, All: true})
	if err != nil {
		t.Fatalf("PlanImport returned error: %v", err)
	}
	if len(res.Importable) != 2 {
		t.Fatalf("importable = %#v, want two same-name skills", res.Importable)
	}
	ApplyImport(lock, res)
	if len(lock.Skills) != 2 {
		t.Fatalf("lock entries = %#v, want two entries distinguished by target path", lock.Skills)
	}
	for _, entry := range lock.Skills {
		if entry.Name != "shared" || entry.Agent != "hermes" || entry.TargetPath == "" {
			t.Fatalf("entry = %#v, want shared/hermes with target path", entry)
		}
	}
}

func TestImportChecksumIsDeterministic(t *testing.T) {
	root := t.TempDir()
	writeImportSkill(t, root, "native", "native")
	res1, err := PlanImport(ImportOptions{SkillsDir: root, LockFile: &config.LockFile{Version: 1}, All: true})
	if err != nil {
		t.Fatal(err)
	}
	res2, err := PlanImport(ImportOptions{SkillsDir: root, LockFile: &config.LockFile{Version: 1}, All: true})
	if err != nil {
		t.Fatal(err)
	}
	if res1.Importable[0].Entry.Checksum != res2.Importable[0].Entry.Checksum {
		t.Fatalf("checksums differ: %q vs %q", res1.Importable[0].Entry.Checksum, res2.Importable[0].Entry.Checksum)
	}
}

func writeImportSkill(t *testing.T, root, rel, name string) {
	t.Helper()
	dir := filepath.Join(root, rel)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("---\nname: "+name+"\ndescription: test\n---\nbody\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}
