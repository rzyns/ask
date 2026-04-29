package hermes

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/yeasy/ask/internal/config"
)

func TestPlanUpdateAllowsCleanASKManagedHermesIndexSkill(t *testing.T) {
	root := t.TempDir()
	source := writeUpdateSkill(t, root, "gitnexus-explorer", "old")
	checksum, err := directoryChecksum(source)
	if err != nil {
		t.Fatalf("checksum: %v", err)
	}
	lock := &config.LockFile{Version: 1, Skills: []config.LockEntry{{
		Name:        "gitnexus-explorer",
		URL:         "https://github.com/NousResearch/hermes-agent/optional-skills/research/gitnexus-explorer",
		InstalledAt: time.Now().UTC(),
	}, {
		Name:             "gitnexus-explorer",
		Source:           config.RepoTypeHermes,
		URL:              "https://github.com/NousResearch/hermes-agent/optional-skills/gitnexus-explorer",
		InstalledAt:      time.Now().UTC(),
		Agent:            "hermes",
		Ownership:        string(HermesSkillOwnershipASK),
		InstallMode:      "ask-cache",
		UpdateStrategy:   "hermes-index",
		SourceIdentifier: "official/gitnexus-explorer",
		SourcePath:       source,
		TargetPath:       source,
		Checksum:         checksum,
	}}}

	plan, err := PlanUpdate(UpdateOptions{LockFile: lock, Names: []string{"gitnexus-explorer"}})
	if err != nil {
		t.Fatalf("PlanUpdate: %v", err)
	}
	if len(plan.Updateable) != 1 {
		t.Fatalf("expected one updateable skill, got %#v", plan)
	}
	candidate := plan.Updateable[0]
	if candidate.Entry.Name != "gitnexus-explorer" || candidate.Input != "official/gitnexus-explorer" {
		t.Fatalf("candidate = %#v", candidate)
	}
	if candidate.SourceMetadata.Source != config.RepoTypeHermes || candidate.SourceMetadata.UpdateStrategy != "hermes-index" {
		t.Fatalf("source metadata = %#v", candidate.SourceMetadata)
	}
}

func TestPlanUpdateUsesURLForNestedHermesIndexIdentifier(t *testing.T) {
	root := t.TempDir()
	source := writeUpdateSkill(t, root, "gitnexus-explorer", "old")
	checksum, err := directoryChecksum(source)
	if err != nil {
		t.Fatalf("checksum: %v", err)
	}
	lock := &config.LockFile{Version: 1, Skills: []config.LockEntry{{
		Name:             "gitnexus-explorer",
		URL:              "https://github.com/NousResearch/hermes-agent/optional-skills/research/gitnexus-explorer",
		InstalledAt:      time.Now().UTC(),
		Agent:            "hermes",
		Ownership:        string(HermesSkillOwnershipASK),
		InstallMode:      "ask-cache",
		UpdateStrategy:   "hermes-index",
		SourceIdentifier: "official/research/gitnexus-explorer",
		SourcePath:       source,
		TargetPath:       source,
		Checksum:         checksum,
	}}}

	plan, err := PlanUpdate(UpdateOptions{LockFile: lock, Names: []string{"gitnexus-explorer"}})
	if err != nil {
		t.Fatalf("PlanUpdate: %v", err)
	}
	if got, want := plan.Updateable[0].Input, "https://github.com/NousResearch/hermes-agent/tree/main/optional-skills/research/gitnexus-explorer"; got != want {
		t.Fatalf("input = %q, want %q", got, want)
	}
}

func TestPlanUpdateHermesIndexRefusesUntrustedTwoSegmentIdentifier(t *testing.T) {
	root := t.TempDir()
	source := writeUpdateSkill(t, root, "repo", "old")
	checksum, err := directoryChecksum(source)
	if err != nil {
		t.Fatalf("checksum: %v", err)
	}
	lock := &config.LockFile{Version: 1, Skills: []config.LockEntry{{
		Name:             "repo",
		URL:              "https://github.com/someone/repo",
		InstalledAt:      time.Now().UTC(),
		Agent:            "hermes",
		Source:           "github",
		Ownership:        string(HermesSkillOwnershipASK),
		InstallMode:      "ask-cache",
		UpdateStrategy:   "hermes-index",
		SourceIdentifier: "someone/repo",
		SourcePath:       source,
		TargetPath:       source,
		Checksum:         checksum,
	}}}

	plan, err := PlanUpdate(UpdateOptions{LockFile: lock, Names: []string{"repo"}})
	if err == nil {
		t.Fatalf("expected unavailable error, plan=%#v", plan)
	}
	if !strings.Contains(err.Error(), "update unavailable") {
		t.Fatalf("error %q missing update unavailable", err.Error())
	}
}

func TestPlanUpdateSkipsImportedHermesSkill(t *testing.T) {
	root := t.TempDir()
	target := writeUpdateSkill(t, root, "local-only", "old")
	checksum, err := directoryChecksum(target)
	if err != nil {
		t.Fatalf("checksum: %v", err)
	}
	lock := &config.LockFile{Version: 1, Skills: []config.LockEntry{{
		Name:           "local-only",
		InstalledAt:    time.Now().UTC(),
		Agent:          "hermes",
		Ownership:      string(HermesSkillOwnershipImported),
		InstallMode:    "in-place",
		UpdateStrategy: "none",
		TargetPath:     target,
		Checksum:       checksum,
	}}}

	plan, err := PlanUpdate(UpdateOptions{LockFile: lock})
	if err != nil {
		t.Fatalf("PlanUpdate: %v", err)
	}
	if len(plan.Updateable) != 0 || len(plan.Skipped) != 1 {
		t.Fatalf("plan = %#v", plan)
	}
	if plan.Skipped[0].Reason != UpdateSkipUnavailable {
		t.Fatalf("skip reason = %q", plan.Skipped[0].Reason)
	}
}

func TestPlanUpdateRefusesDirtyCopiedTargetEvenWhenSourceIsClean(t *testing.T) {
	root := t.TempDir()
	source := writeUpdateSkill(t, filepath.Join(root, "source"), "copied", "old")
	target := writeUpdateSkill(t, filepath.Join(root, "target"), "copied", "old")
	checksum, err := directoryChecksum(source)
	if err != nil {
		t.Fatalf("checksum: %v", err)
	}
	if err := os.WriteFile(filepath.Join(target, "target-edit.txt"), []byte("local target edit"), 0o644); err != nil {
		t.Fatalf("target dirty write: %v", err)
	}
	lock := &config.LockFile{Version: 1, Skills: []config.LockEntry{{
		Name:             "copied",
		InstalledAt:      time.Now().UTC(),
		Agent:            "hermes",
		Source:           config.RepoTypeHermes,
		Ownership:        string(HermesSkillOwnershipASK),
		InstallMode:      "ask-cache",
		UpdateStrategy:   "hermes-index",
		SourceIdentifier: "official/copied",
		SourcePath:       source,
		TargetPath:       target,
		Checksum:         checksum,
	}}}

	plan, err := PlanUpdate(UpdateOptions{LockFile: lock, Names: []string{"copied"}})
	if err == nil {
		t.Fatalf("expected dirty copied target to return error, plan=%#v", plan)
	}
	if !strings.Contains(err.Error(), "local modifications") {
		t.Fatalf("error %q missing local modifications", err.Error())
	}
}

func TestPlanUpdateRefusesDirtyASKManagedSkillUnlessForced(t *testing.T) {
	root := t.TempDir()
	source := writeUpdateSkill(t, root, "dirty", "old")
	checksum, err := directoryChecksum(source)
	if err != nil {
		t.Fatalf("checksum: %v", err)
	}
	if err := os.WriteFile(filepath.Join(source, "extra.txt"), []byte("local edit"), 0o644); err != nil {
		t.Fatalf("dirty write: %v", err)
	}
	lock := &config.LockFile{Version: 1, Skills: []config.LockEntry{{
		Name:             "dirty",
		InstalledAt:      time.Now().UTC(),
		Agent:            "hermes",
		Source:           config.RepoTypeHermes,
		Ownership:        string(HermesSkillOwnershipASK),
		InstallMode:      "ask-cache",
		UpdateStrategy:   "hermes-index",
		SourceIdentifier: "official/dirty",
		SourcePath:       source,
		TargetPath:       source,
		Checksum:         checksum,
	}}}

	plan, err := PlanUpdate(UpdateOptions{LockFile: lock, Names: []string{"dirty"}})
	if err == nil {
		t.Fatalf("expected dirty named update to return error, plan=%#v", plan)
	}
	if !strings.Contains(err.Error(), "local modifications") {
		t.Fatalf("error %q missing local modifications", err.Error())
	}

	forced, err := PlanUpdate(UpdateOptions{LockFile: lock, Names: []string{"dirty"}, Force: true})
	if err != nil {
		t.Fatalf("forced PlanUpdate: %v", err)
	}
	if len(forced.Updateable) != 1 || len(forced.Blocked) != 0 {
		t.Fatalf("forced plan = %#v", forced)
	}
}

func TestPlanUpdateNamedMissingReturnsError(t *testing.T) {
	_, err := PlanUpdate(UpdateOptions{LockFile: &config.LockFile{Version: 1}, Names: []string{"missing"}})
	if err == nil || !strings.Contains(err.Error(), "not installed for Hermes") {
		t.Fatalf("expected missing named error, got %v", err)
	}
}

func writeUpdateSkill(t *testing.T, root, name, body string) string {
	t.Helper()
	dir := filepath.Join(root, name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("---\nname: "+name+"\ndescription: test\n---\n"+body+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}
