package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/yeasy/ask/internal/config"
	"github.com/yeasy/ask/internal/hermes"
)

func TestSkillUninstallHermesASKOwnedRemovesTargetSourceAndTracking(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	hermesHome := filepath.Join(home, "hermes-home")
	t.Setenv("HERMES_HOME", hermesHome)
	source := writeCmdUninstallSkill(t, filepath.Join(home, ".ask", "skills"), "ask-owned")
	checksum, err := hermesDirectoryChecksumForUpdateTest(source)
	if err != nil {
		t.Fatalf("checksum: %v", err)
	}
	target := filepath.Join(hermesHome, "skills", "ask-owned")
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(source, target); err != nil {
		t.Fatal(err)
	}
	lock := &config.LockFile{Version: 1, Skills: []config.LockEntry{{
		Name:        "ask-owned",
		InstalledAt: time.Now().UTC(),
		Agent:       "hermes",
		Ownership:   string(hermes.HermesSkillOwnershipASK),
		InstallMode: "ask-cache",
		SourcePath:  source,
		TargetPath:  target,
		Checksum:    checksum,
	}}}
	if err := lock.SaveGlobal(); err != nil {
		t.Fatalf("SaveGlobal: %v", err)
	}
	cfg := &config.Config{Version: "1.2", Skills: []string{"ask-owned"}, SkillsInfo: []config.SkillInfo{{Name: "ask-owned"}}}
	if err := cfg.SaveByScope(true); err != nil {
		t.Fatalf("SaveByScope: %v", err)
	}

	buf, err := executeUninstallCommandForTest("skill", "uninstall", "ask-owned", "--agent", "hermes", "--global")
	if err != nil {
		t.Fatalf("command returned error: %v output=%s", err, buf.String())
	}
	if _, err := os.Lstat(target); !os.IsNotExist(err) {
		t.Fatalf("target still exists or unexpected error: %v", err)
	}
	if _, err := os.Stat(source); !os.IsNotExist(err) {
		t.Fatalf("source still exists or unexpected error: %v", err)
	}
	loaded, err := config.LoadGlobalLockFile()
	if err != nil {
		t.Fatalf("LoadGlobalLockFile: %v", err)
	}
	if got := loaded.GetEntryForAgent("ask-owned", "hermes"); got != nil {
		t.Fatalf("hermes lock entry still present: %#v", got)
	}
}

func TestSkillUninstallHermesASKOwnedForgetPreservesFiles(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	hermesHome := filepath.Join(home, "hermes-home")
	t.Setenv("HERMES_HOME", hermesHome)
	source := writeCmdUninstallSkill(t, filepath.Join(home, ".ask", "skills"), "ask-owned")
	target := filepath.Join(hermesHome, "skills", "ask-owned")
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(source, target); err != nil {
		t.Fatal(err)
	}
	lock := &config.LockFile{Version: 1, Skills: []config.LockEntry{{
		Name:        "ask-owned",
		Agent:       "hermes",
		Ownership:   string(hermes.HermesSkillOwnershipASK),
		InstallMode: "ask-cache",
		SourcePath:  source,
		TargetPath:  target,
		Checksum:    "stale-for-forget",
	}}}
	if err := lock.SaveGlobal(); err != nil {
		t.Fatalf("SaveGlobal: %v", err)
	}

	buf, err := executeUninstallCommandForTest("skill", "uninstall", "ask-owned", "--agent", "hermes", "--global", "--forget")
	if err != nil {
		t.Fatalf("command returned error: %v output=%s", err, buf.String())
	}
	if _, err := os.Lstat(target); err != nil {
		t.Fatalf("target should remain after forget: %v", err)
	}
	if _, err := os.Stat(source); err != nil {
		t.Fatalf("source should remain after forget: %v", err)
	}
	loaded, err := config.LoadGlobalLockFile()
	if err != nil {
		t.Fatalf("LoadGlobalLockFile: %v", err)
	}
	if got := loaded.GetEntryForAgent("ask-owned", "hermes"); got != nil {
		t.Fatalf("hermes lock entry still present: %#v", got)
	}
}

func TestSkillUninstallHermesImportedRefusesByDefault(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	hermesHome := filepath.Join(home, "hermes-home")
	t.Setenv("HERMES_HOME", hermesHome)
	target := writeCmdUninstallSkill(t, filepath.Join(hermesHome, "skills"), "imported")
	lock := &config.LockFile{Version: 1, Skills: []config.LockEntry{{
		Name:        "imported",
		InstalledAt: time.Now().UTC(),
		Agent:       "hermes",
		Ownership:   string(hermes.HermesSkillOwnershipImported),
		InstallMode: "in-place",
		TargetPath:  target,
	}}}
	if err := lock.SaveGlobal(); err != nil {
		t.Fatalf("SaveGlobal: %v", err)
	}

	buf, err := executeUninstallCommandForTest("skill", "uninstall", "imported", "--agent", "hermes", "--global")
	if err == nil {
		t.Fatalf("expected error output=%s", buf.String())
	}
	if !strings.Contains(err.Error(), "--forget") || !strings.Contains(err.Error(), "--delete-files") {
		t.Fatalf("error %q missing guidance", err.Error())
	}
	if _, statErr := os.Stat(target); statErr != nil {
		t.Fatalf("imported target should remain: %v", statErr)
	}
}

func TestSkillUninstallHermesImportedForgetPreservesFiles(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	hermesHome := filepath.Join(home, "hermes-home")
	t.Setenv("HERMES_HOME", hermesHome)
	target := writeCmdUninstallSkill(t, filepath.Join(hermesHome, "skills"), "imported")
	lock := &config.LockFile{Version: 1, Skills: []config.LockEntry{{
		Name:        "imported",
		InstalledAt: time.Now().UTC(),
		Agent:       "hermes",
		Ownership:   string(hermes.HermesSkillOwnershipImported),
		InstallMode: "in-place",
		TargetPath:  target,
	}}}
	if err := lock.SaveGlobal(); err != nil {
		t.Fatalf("SaveGlobal: %v", err)
	}

	buf, err := executeUninstallCommandForTest("skill", "uninstall", "imported", "--agent", "hermes", "--global", "--forget")
	if err != nil {
		t.Fatalf("command returned error: %v output=%s", err, buf.String())
	}
	if _, statErr := os.Stat(target); statErr != nil {
		t.Fatalf("imported target should remain after forget: %v", statErr)
	}
	loaded, err := config.LoadGlobalLockFile()
	if err != nil {
		t.Fatalf("LoadGlobalLockFile: %v", err)
	}
	if got := loaded.GetEntryForAgent("imported", "hermes"); got != nil {
		t.Fatalf("hermes lock entry still present: %#v", got)
	}
}

func TestSkillUninstallHermesImportedDeleteFilesRemovesTarget(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	hermesHome := filepath.Join(home, "hermes-home")
	t.Setenv("HERMES_HOME", hermesHome)
	target := writeCmdUninstallSkill(t, filepath.Join(hermesHome, "skills"), "imported")
	lock := &config.LockFile{Version: 1, Skills: []config.LockEntry{{
		Name:        "imported",
		InstalledAt: time.Now().UTC(),
		Agent:       "hermes",
		Ownership:   string(hermes.HermesSkillOwnershipImported),
		InstallMode: "in-place",
		TargetPath:  target,
	}}}
	if err := lock.SaveGlobal(); err != nil {
		t.Fatalf("SaveGlobal: %v", err)
	}

	buf, err := executeUninstallCommandForTest("skill", "uninstall", "imported", "--agent", "hermes", "--global", "--delete-files")
	if err != nil {
		t.Fatalf("command returned error: %v output=%s", err, buf.String())
	}
	if _, statErr := os.Stat(target); !os.IsNotExist(statErr) {
		t.Fatalf("imported target still exists or unexpected error: %v", statErr)
	}
}

func TestSkillUninstallHermesRejectsNonBasenameSkillName(t *testing.T) {
	buf, err := executeUninstallCommandForTest("skill", "uninstall", "../anything", "--agent", "hermes", "--global")
	if err == nil {
		t.Fatalf("expected error output=%s", buf.String())
	}
	if !strings.Contains(err.Error(), "invalid Hermes skill name") {
		t.Fatalf("error %q missing invalid-name guidance", err.Error())
	}
}

func TestSkillUninstallHermesRejectsForgetAndDeleteFilesTogether(t *testing.T) {
	buf, err := executeUninstallCommandForTest("skill", "uninstall", "anything", "--agent", "hermes", "--global", "--forget", "--delete-files")
	if err == nil {
		t.Fatalf("expected error output=%s", buf.String())
	}
	if !strings.Contains(err.Error(), "mutually exclusive") {
		t.Fatalf("error %q missing mutual exclusion", err.Error())
	}
}

func executeUninstallCommandForTest(args ...string) (*bytes.Buffer, error) {
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs(args)
	_ = uninstallCmd.Flags().Set("agent", "")
	_ = uninstallCmd.Flags().Set("all", "false")
	_ = uninstallCmd.Flags().Set("forget", "false")
	_ = uninstallCmd.Flags().Set("delete-files", "false")
	_ = rootCmd.PersistentFlags().Set("global", "false")
	return &buf, rootCmd.Execute()
}

func writeCmdUninstallSkill(t *testing.T, root, name string) string {
	t.Helper()
	dir := filepath.Join(root, name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("---\nname: "+name+"\ndescription: test\n---\nbody\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}
