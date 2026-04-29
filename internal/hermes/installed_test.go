package hermes

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/yeasy/ask/internal/config"
)

func TestScanInstalledSkillsFindsNestedSkillsAndIgnoresHiddenAndNonSkills(t *testing.T) {
	root := t.TempDir()
	writeSkill(t, root, "gitnexus-explorer", "gitnexus-explorer", "Explore GitNexus", "1.2.3")
	writeSkill(t, root, filepath.Join("research", "gitnexus-explorer"), "research-gitnexus", "Research nested", "2.0.0")
	writeSkill(t, root, filepath.Join("category-without-skill", "child"), "child-skill", "Nested child", "")
	writeSkill(t, root, filepath.Join(".hub", "hidden-skill"), "hidden", "Hidden", "9.9.9")
	mustWriteFile(t, filepath.Join(root, ".hub", "taps.json"), "{}")
	mustWriteFile(t, filepath.Join(root, "non-skill", "README.md"), "not a skill")

	got, err := ScanInstalledSkills(root, InstalledScanOptions{})
	if err != nil {
		t.Fatalf("ScanInstalledSkills returned error: %v", err)
	}

	wantRel := []string{"category-without-skill/child", "gitnexus-explorer", "research/gitnexus-explorer"}
	if len(got) != len(wantRel) {
		t.Fatalf("got %d skills %#v, want %d", len(got), got, len(wantRel))
	}
	for i, want := range wantRel {
		if got[i].RelativePath != want {
			t.Fatalf("skill[%d].RelativePath = %q, want %q (all: %#v)", i, got[i].RelativePath, want, got)
		}
		if got[i].Path != filepath.Join(root, filepath.FromSlash(want)) {
			t.Fatalf("skill[%d].Path = %q, want %q", i, got[i].Path, filepath.Join(root, filepath.FromSlash(want)))
		}
		if got[i].Ownership != HermesSkillOwnershipNative || got[i].Managed || got[i].Source != "local" || got[i].UpdateStrategy != "none" {
			t.Fatalf("unknown skill ownership/managed/source/update = %q/%v/%q/%q, want hermes-native/false/local/none", got[i].Ownership, got[i].Managed, got[i].Source, got[i].UpdateStrategy)
		}
	}
	if got[0].Name != "child-skill" || got[0].Description != "Nested child" {
		t.Fatalf("nested child metadata = %#v", got[0])
	}
	if got[1].Name != "gitnexus-explorer" || got[1].Description != "Explore GitNexus" || got[1].Version != "1.2.3" {
		t.Fatalf("top-level metadata = %#v", got[1])
	}
}

func TestScanInstalledSkillsUsesDirectoryBasenameWhenNameMissing(t *testing.T) {
	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "nameless", "SKILL.md"), "---\ndescription: No name here\nversion: 0.1.0\n---\n# Ignored\n")

	got, err := ScanInstalledSkills(root, InstalledScanOptions{})
	if err != nil {
		t.Fatalf("ScanInstalledSkills returned error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d skills, want 1", len(got))
	}
	if got[0].Name != "nameless" {
		t.Fatalf("Name = %q, want directory basename", got[0].Name)
	}
}

func TestScanInstalledSkillsSkipsSymlinkDirectories(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation may require privileges on Windows")
	}
	root := t.TempDir()
	outside := t.TempDir()
	writeSkill(t, outside, "external", "external", "Outside", "1.0.0")
	if err := os.Symlink(filepath.Join(outside, "external"), filepath.Join(root, "linked")); err != nil {
		t.Fatalf("create symlink: %v", err)
	}

	got, err := ScanInstalledSkills(root, InstalledScanOptions{})
	if err != nil {
		t.Fatalf("ScanInstalledSkills returned error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("got %#v, want symlink dir skipped", got)
	}
}

func TestScanInstalledSkillsHonorsMaxDepthAndMissingDir(t *testing.T) {
	root := t.TempDir()
	writeSkill(t, root, filepath.Join("a", "b"), "too-deep", "Too deep", "")

	got, err := ScanInstalledSkills(root, InstalledScanOptions{MaxDepth: 1})
	if err != nil {
		t.Fatalf("ScanInstalledSkills returned error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("got %#v, want no skills past max depth", got)
	}

	got, err = ScanInstalledSkills(filepath.Join(root, "does-not-exist"), InstalledScanOptions{})
	if err != nil {
		t.Fatalf("missing dir error = %v, want nil", err)
	}
	if len(got) != 0 {
		t.Fatalf("missing dir got %#v, want empty", got)
	}
}

func TestScanInstalledSkillsMarksLockfileBackedEntriesAsASKManaged(t *testing.T) {
	root := t.TempDir()
	writeSkill(t, root, "gitnexus-explorer", "gitnexus-explorer", "Explore GitNexus", "")
	writeSkill(t, root, "native", "native", "Native", "3.0.0")

	lock := &config.LockFile{Skills: []config.LockEntry{{Name: "gitnexus-explorer", Source: "NousResearch/hermes-agent/optional-skills/gitnexus-explorer", Version: "1.2.3", URL: "https://github.com/NousResearch/hermes-agent.git"}}}
	got, err := ScanInstalledSkills(root, InstalledScanOptions{LockFile: lock})
	if err != nil {
		t.Fatalf("ScanInstalledSkills returned error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d skills %#v, want 2", len(got), got)
	}
	managed := got[0]
	if managed.Name != "gitnexus-explorer" {
		t.Fatalf("first skill = %#v, want gitnexus-explorer sorted before native", managed)
	}
	if managed.Ownership != HermesSkillOwnershipASK || !managed.Managed || managed.Source != lock.Skills[0].Source || managed.Version != "1.2.3" || managed.UpdateStrategy != "git" {
		t.Fatalf("managed skill = %#v, want ASK managed with lock metadata", managed)
	}
	if got[1].Ownership != HermesSkillOwnershipNative || got[1].Managed || got[1].Source != "local" || got[1].Version != "3.0.0" || got[1].UpdateStrategy != "none" {
		t.Fatalf("native skill = %#v, want native unmanaged with metadata version", got[1])
	}
}

func TestScanInstalledSkillsOmitsBundledDistributionCopies(t *testing.T) {
	root := t.TempDir()
	bundled := t.TempDir()
	writeSkill(t, bundled, filepath.Join("research", "arxiv"), "arxiv", "Bundled arXiv", "1.0.0")
	writeSkill(t, root, filepath.Join("research", "arxiv"), "arxiv", "Bundled arXiv", "1.0.0")
	writeSkill(t, root, filepath.Join("research", "gitnexus-explorer"), "gitnexus-explorer", "User installed", "2.0.0")

	got, err := ScanInstalledSkills(root, InstalledScanOptions{BundledSkillsDir: bundled})
	if err != nil {
		t.Fatalf("ScanInstalledSkills returned error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d skills %#v, want only non-bundled skill", len(got), got)
	}
	if got[0].Name != "gitnexus-explorer" || got[0].RelativePath != "research/gitnexus-explorer" {
		t.Fatalf("remaining skill = %#v, want gitnexus-explorer", got[0])
	}
}

func TestScanInstalledSkillsKeepsModifiedBundledPathCopy(t *testing.T) {
	root := t.TempDir()
	bundled := t.TempDir()
	writeSkill(t, bundled, filepath.Join("software-development", "systematic-debugging"), "systematic-debugging", "Bundled", "1.0.0")
	writeSkill(t, root, filepath.Join("software-development", "systematic-debugging"), "systematic-debugging", "Local override", "1.0.0")

	got, err := ScanInstalledSkills(root, InstalledScanOptions{BundledSkillsDir: bundled})
	if err != nil {
		t.Fatalf("ScanInstalledSkills returned error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d skills %#v, want modified local copy retained", len(got), got)
	}
	if got[0].Name != "systematic-debugging" || got[0].Description != "Local override" {
		t.Fatalf("skill = %#v, want modified local copy", got[0])
	}
}

func TestScanInstalledSkillsIgnoresNonHermesLockEntries(t *testing.T) {
	root := t.TempDir()
	writeSkill(t, root, "shared", "shared", "Shared name", "1.0.0")

	lock := &config.LockFile{Skills: []config.LockEntry{{Name: "shared", Agent: "claude", Source: "claude-source", Version: "9.9.9"}}}
	got, err := ScanInstalledSkills(root, InstalledScanOptions{LockFile: lock})
	if err != nil {
		t.Fatalf("ScanInstalledSkills returned error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d skills %#v, want 1", len(got), got)
	}
	if got[0].Ownership != HermesSkillOwnershipNative || got[0].Managed || got[0].Source != "local" || got[0].Version != "1.0.0" || got[0].UpdateStrategy != "none" {
		t.Fatalf("skill = %#v, want native classification unaffected by non-Hermes lock entry", got[0])
	}
}

func TestScanInstalledSkillsIgnoresLegacyNonHermesLockEntries(t *testing.T) {
	root := t.TempDir()
	writeSkill(t, root, "shared", "shared", "Shared name", "1.0.0")

	lock := &config.LockFile{Skills: []config.LockEntry{{Name: "shared", Source: "legacy-non-hermes", Version: "9.9.9"}}}
	got, err := ScanInstalledSkills(root, InstalledScanOptions{LockFile: lock})
	if err != nil {
		t.Fatalf("ScanInstalledSkills returned error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d skills %#v, want 1", len(got), got)
	}
	if got[0].Ownership != HermesSkillOwnershipNative || got[0].Managed || got[0].Source != "local" || got[0].Version != "1.0.0" || got[0].UpdateStrategy != "none" {
		t.Fatalf("skill = %#v, want native classification unaffected by legacy non-Hermes lock entry", got[0])
	}
}

func TestScanInstalledSkillsDoesNotInferLockfileOwnershipFromNameAlone(t *testing.T) {
	root := t.TempDir()
	writeSkill(t, root, filepath.Join("research", "gitnexus-explorer"), "gitnexus-explorer", "Native duplicate name", "")

	lock := &config.LockFile{Skills: []config.LockEntry{{Name: "gitnexus-explorer", Source: "hermes-index", Version: "1.2.3", URL: "https://github.com/NousResearch/hermes-agent.git"}}}
	got, err := ScanInstalledSkills(root, InstalledScanOptions{LockFile: lock})
	if err != nil {
		t.Fatalf("ScanInstalledSkills returned error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d skills %#v, want 1", len(got), got)
	}
	if got[0].Ownership != HermesSkillOwnershipNative || got[0].Managed || got[0].Source != "local" || got[0].UpdateStrategy != "none" {
		t.Fatalf("nested same-name skill = %#v, want conservative native classification", got[0])
	}
}

func TestScanInstalledSkillsMatchesImportedNestedSkillByTargetPath(t *testing.T) {
	root := t.TempDir()
	nestedRel := filepath.Join("research", "shared")
	writeSkill(t, root, nestedRel, "shared", "Imported nested", "")
	targetPath := filepath.Join(root, nestedRel)

	lock := &config.LockFile{Skills: []config.LockEntry{{
		Name:           "shared",
		Agent:          "hermes",
		Ownership:      string(HermesSkillOwnershipImported),
		Source:         "local",
		UpdateStrategy: "none",
		TargetPath:     targetPath,
		Version:        "2.0.0",
	}}}
	got, err := ScanInstalledSkills(root, InstalledScanOptions{LockFile: lock})
	if err != nil {
		t.Fatalf("ScanInstalledSkills returned error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d skills %#v, want 1", len(got), got)
	}
	if got[0].Ownership != HermesSkillOwnershipImported || got[0].Managed || got[0].Source != "local" || got[0].Version != "2.0.0" || got[0].UpdateStrategy != "none" {
		t.Fatalf("nested imported skill = %#v, want imported lock metadata matched by target path", got[0])
	}
}

func TestScanInstalledSkillsDoesNotMarkDuplicateNamesAsASKManaged(t *testing.T) {
	root := t.TempDir()
	writeSkill(t, root, "gitnexus-explorer", "gitnexus-explorer", "Top level", "")
	writeSkill(t, root, filepath.Join("research", "gitnexus-explorer"), "gitnexus-explorer", "Nested duplicate", "")

	lock := &config.LockFile{Skills: []config.LockEntry{{Name: "gitnexus-explorer", Source: "hermes-index", Version: "1.2.3", URL: "https://github.com/NousResearch/hermes-agent.git"}}}
	got, err := ScanInstalledSkills(root, InstalledScanOptions{LockFile: lock})
	if err != nil {
		t.Fatalf("ScanInstalledSkills returned error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d skills %#v, want 2", len(got), got)
	}
	for _, installed := range got {
		if installed.Managed || installed.Ownership != HermesSkillOwnershipNative {
			t.Fatalf("duplicate-name skill was marked managed: %#v", installed)
		}
	}
}

func writeSkill(t *testing.T, root, rel, name, description, version string) {
	t.Helper()
	content := "---\nname: " + name + "\ndescription: " + description + "\n"
	if version != "" {
		content += "version: " + version + "\n"
	}
	content += "---\n# " + name + "\n"
	mustWriteFile(t, filepath.Join(root, rel, "SKILL.md"), content)
}

func mustWriteFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
