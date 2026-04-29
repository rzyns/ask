package hermes

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/yeasy/ask/internal/config"
)

func TestPlanUninstallASKOwnedHermesSkillRemovesTargetSourceAndTracking(t *testing.T) {
	root := t.TempDir()
	skillsDir := filepath.Join(root, "hermes", "skills")
	sourceRoot := filepath.Join(root, ".ask", "skills")
	source := writeUpdateSkill(t, sourceRoot, "ask-owned", "body")
	checksum, err := directoryChecksum(source)
	if err != nil {
		t.Fatalf("checksum: %v", err)
	}
	target := filepath.Join(skillsDir, "ask-owned")
	if err := os.MkdirAll(skillsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(source, target); err != nil {
		t.Fatal(err)
	}
	lock := &config.LockFile{Version: 1, Skills: []config.LockEntry{{
		Name:        "ask-owned",
		InstalledAt: time.Now().UTC(),
		Agent:       "hermes",
		Ownership:   string(HermesSkillOwnershipASK),
		InstallMode: "ask-cache",
		SourcePath:  source,
		TargetPath:  target,
		Checksum:    checksum,
	}}}

	decision, err := PlanUninstall(UninstallOptions{LockFile: lock, Name: "ask-owned", SkillsDir: skillsDir, SourceDir: sourceRoot})
	if err != nil {
		t.Fatalf("PlanUninstall: %v", err)
	}
	if !decision.RemoveTarget || !decision.RemoveSource || !decision.RemoveTracking {
		t.Fatalf("decision = %#v", decision)
	}
	if decision.TargetPath != target || decision.SourcePath != source {
		t.Fatalf("decision paths = %#v", decision)
	}
}

func TestPlanUninstallRefusesDirtyASKOwnedSource(t *testing.T) {
	root := t.TempDir()
	skillsDir := filepath.Join(root, "hermes", "skills")
	sourceRoot := filepath.Join(root, ".ask", "skills")
	source := writeUpdateSkill(t, sourceRoot, "dirty", "clean")
	checksum, err := directoryChecksum(source)
	if err != nil {
		t.Fatalf("checksum: %v", err)
	}
	if err := os.WriteFile(filepath.Join(source, "changed.txt"), []byte("dirty"), 0o644); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(skillsDir, "dirty")
	if err := os.MkdirAll(skillsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(source, target); err != nil {
		t.Fatal(err)
	}
	lock := &config.LockFile{Version: 1, Skills: []config.LockEntry{{
		Name:        "dirty",
		Agent:       "hermes",
		Ownership:   string(HermesSkillOwnershipASK),
		InstallMode: "ask-cache",
		SourcePath:  source,
		TargetPath:  target,
		Checksum:    checksum,
	}}}

	_, err = PlanUninstall(UninstallOptions{LockFile: lock, Name: "dirty", SkillsDir: skillsDir, SourceDir: sourceRoot})
	if err == nil || !strings.Contains(err.Error(), "local changes") {
		t.Fatalf("expected dirty refusal, got %v", err)
	}
}

func TestPlanUninstallRefusesASKOwnedSourceOutsideASKSkillsDir(t *testing.T) {
	root := t.TempDir()
	skillsDir := filepath.Join(root, "hermes", "skills")
	source := writeUpdateSkill(t, filepath.Join(root, "elsewhere"), "ask-owned", "body")
	checksum, err := directoryChecksum(source)
	if err != nil {
		t.Fatalf("checksum: %v", err)
	}
	target := filepath.Join(skillsDir, "ask-owned")
	if err := os.MkdirAll(skillsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(source, target); err != nil {
		t.Fatal(err)
	}
	lock := &config.LockFile{Version: 1, Skills: []config.LockEntry{{
		Name:        "ask-owned",
		Agent:       "hermes",
		Ownership:   string(HermesSkillOwnershipASK),
		InstallMode: "ask-cache",
		SourcePath:  source,
		TargetPath:  target,
		Checksum:    checksum,
	}}}

	_, err = PlanUninstall(UninstallOptions{LockFile: lock, Name: "ask-owned", SkillsDir: skillsDir, SourceDir: filepath.Join(root, ".ask", "skills")})
	if err == nil || !strings.Contains(err.Error(), "outside ASK skills dir") {
		t.Fatalf("expected unsafe source refusal, got %v", err)
	}
}

func TestPlanUninstallASKOwnedForgetRemovesTrackingOnly(t *testing.T) {
	root := t.TempDir()
	skillsDir := filepath.Join(root, "hermes", "skills")
	sourceRoot := filepath.Join(root, ".ask", "skills")
	source := writeUpdateSkill(t, sourceRoot, "ask-owned", "body")
	target := filepath.Join(skillsDir, "ask-owned")
	lock := &config.LockFile{Version: 1, Skills: []config.LockEntry{{
		Name:        "ask-owned",
		Agent:       "hermes",
		Ownership:   string(HermesSkillOwnershipASK),
		InstallMode: "ask-cache",
		SourcePath:  source,
		TargetPath:  target,
		Checksum:    "stale-is-irrelevant-for-forget",
	}}}

	decision, err := PlanUninstall(UninstallOptions{LockFile: lock, Name: "ask-owned", SkillsDir: skillsDir, SourceDir: sourceRoot, Forget: true})
	if err != nil {
		t.Fatalf("PlanUninstall: %v", err)
	}
	if decision.RemoveTarget || decision.RemoveSource || !decision.RemoveTracking {
		t.Fatalf("decision = %#v", decision)
	}
}

func TestPlanUninstallLegacyHermesEntryWithoutAgent(t *testing.T) {
	root := t.TempDir()
	skillsDir := filepath.Join(root, "hermes", "skills")
	sourceRoot := filepath.Join(root, ".ask", "skills")
	source := writeUpdateSkill(t, sourceRoot, "legacy", "body")
	checksum, err := directoryChecksum(source)
	if err != nil {
		t.Fatalf("checksum: %v", err)
	}
	target := filepath.Join(skillsDir, "legacy")
	if err := os.MkdirAll(skillsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(source, target); err != nil {
		t.Fatal(err)
	}
	lock := &config.LockFile{Version: 1, Skills: []config.LockEntry{{
		Name:        "legacy",
		Ownership:   string(HermesSkillOwnershipASK),
		InstallMode: "ask-cache",
		SourcePath:  source,
		TargetPath:  target,
		Checksum:    checksum,
	}}}

	decision, err := PlanUninstall(UninstallOptions{LockFile: lock, Name: "legacy", SkillsDir: skillsDir, SourceDir: sourceRoot})
	if err != nil {
		t.Fatalf("PlanUninstall: %v", err)
	}
	if !decision.RemoveTarget || !decision.RemoveSource || !decision.RemoveTracking {
		t.Fatalf("decision = %#v", decision)
	}
}

func TestPlanUninstallImportedHermesSkillRequiresExplicitChoice(t *testing.T) {
	root := t.TempDir()
	skillsDir := filepath.Join(root, "hermes", "skills")
	target := writeUpdateSkill(t, skillsDir, "imported", "body")
	lock := &config.LockFile{Version: 1, Skills: []config.LockEntry{{
		Name:        "imported",
		InstalledAt: time.Now().UTC(),
		Agent:       "hermes",
		Ownership:   string(HermesSkillOwnershipImported),
		InstallMode: "in-place",
		TargetPath:  target,
	}}}

	_, err := PlanUninstall(UninstallOptions{LockFile: lock, Name: "imported", SkillsDir: skillsDir})
	if err == nil || !strings.Contains(err.Error(), "--forget") || !strings.Contains(err.Error(), "--delete-files") {
		t.Fatalf("expected explicit-choice error, got %v", err)
	}
}

func TestPlanUninstallImportedForgetRemovesTrackingOnly(t *testing.T) {
	root := t.TempDir()
	skillsDir := filepath.Join(root, "hermes", "skills")
	target := writeUpdateSkill(t, skillsDir, "imported", "body")
	lock := &config.LockFile{Version: 1, Skills: []config.LockEntry{{
		Name:        "imported",
		InstalledAt: time.Now().UTC(),
		Agent:       "hermes",
		Ownership:   string(HermesSkillOwnershipImported),
		InstallMode: "in-place",
		TargetPath:  target,
	}}}

	decision, err := PlanUninstall(UninstallOptions{LockFile: lock, Name: "imported", SkillsDir: skillsDir, Forget: true})
	if err != nil {
		t.Fatalf("PlanUninstall: %v", err)
	}
	if decision.RemoveTarget || decision.RemoveSource || !decision.RemoveTracking {
		t.Fatalf("decision = %#v", decision)
	}
}

func TestPlanUninstallImportedDeleteFilesRemovesTargetAndTracking(t *testing.T) {
	root := t.TempDir()
	skillsDir := filepath.Join(root, "hermes", "skills")
	target := writeUpdateSkill(t, skillsDir, "imported", "body")
	lock := &config.LockFile{Version: 1, Skills: []config.LockEntry{{
		Name:        "imported",
		InstalledAt: time.Now().UTC(),
		Agent:       "hermes",
		Ownership:   string(HermesSkillOwnershipImported),
		InstallMode: "in-place",
		TargetPath:  target,
	}}}

	decision, err := PlanUninstall(UninstallOptions{LockFile: lock, Name: "imported", SkillsDir: skillsDir, DeleteFiles: true})
	if err != nil {
		t.Fatalf("PlanUninstall: %v", err)
	}
	if !decision.RemoveTarget || decision.RemoveSource || !decision.RemoveTracking {
		t.Fatalf("decision = %#v", decision)
	}
}

func TestPlanUninstallRefusesBundledHermesSkill(t *testing.T) {
	root := t.TempDir()
	skillsDir := filepath.Join(root, "hermes", "skills")
	target := writeUpdateSkill(t, skillsDir, "bundled", "body")
	lock := &config.LockFile{Version: 1, Skills: []config.LockEntry{{
		Name:        "bundled",
		InstalledAt: time.Now().UTC(),
		Agent:       "hermes",
		Ownership:   string(HermesSkillOwnershipBundled),
		InstallMode: "in-place",
		TargetPath:  target,
	}}}

	_, err := PlanUninstall(UninstallOptions{LockFile: lock, Name: "bundled", SkillsDir: skillsDir, DeleteFiles: true})
	if err == nil || !strings.Contains(err.Error(), "bundled") {
		t.Fatalf("expected bundled refusal, got %v", err)
	}
}

func TestPlanUninstallRefusesUnmanagedHermesNativeSkill(t *testing.T) {
	root := t.TempDir()
	skillsDir := filepath.Join(root, "hermes", "skills")
	writeUpdateSkill(t, skillsDir, "native", "body")

	_, err := PlanUninstall(UninstallOptions{LockFile: &config.LockFile{Version: 1}, Name: "native", SkillsDir: skillsDir})
	if err == nil || !strings.Contains(err.Error(), "unmanaged") {
		t.Fatalf("expected unmanaged refusal, got %v", err)
	}
}
