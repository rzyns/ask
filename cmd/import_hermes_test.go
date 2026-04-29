package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yeasy/ask/internal/config"
)

func TestSkillImportHermesRequiresAgent(t *testing.T) {
	for _, args := range [][]string{{"skill", "import", "--all"}, {"skill", "import", "--agent", "claude", "--all"}, {"skill", "import", "--agent", "hermes,claude", "--all"}} {
		buf, err := executeImportCommandForTest(args...)
		if err == nil {
			t.Fatalf("%v returned nil error", args)
		}
		if !strings.Contains(buf.String(), "--agent hermes") {
			t.Fatalf("output %q missing helpful agent error", buf.String())
		}
	}
}

func TestSkillImportHermesDryRunDoesNotWriteLock(t *testing.T) {
	workdir, restore := setupImportCmdWorkdir(t)
	defer restore()
	writeCmdImportSkill(t, filepath.Join(workdir, ".hermes", "skills"), "native", "native")

	buf, err := executeImportCommandForTest("skill", "import", "--agent", "hermes", "--dry-run", "--all")
	if err != nil {
		t.Fatalf("command returned error: %v output=%s", err, buf.String())
	}
	if !strings.Contains(buf.String(), "native") || !strings.Contains(buf.String(), "dry-run") {
		t.Fatalf("output %q missing dry-run importable skill", buf.String())
	}
	if _, err := os.Stat(filepath.Join(workdir, config.LockFileName)); !os.IsNotExist(err) {
		t.Fatalf("dry-run wrote lockfile or unexpected stat error: %v", err)
	}
}

func TestSkillImportHermesAllWritesImportedLock(t *testing.T) {
	workdir, restore := setupImportCmdWorkdir(t)
	defer restore()
	writeCmdImportSkill(t, filepath.Join(workdir, ".hermes", "skills"), "native", "native")

	buf, err := executeImportCommandForTest("skill", "import", "--agent", "hermes", "--all")
	if err != nil {
		t.Fatalf("command returned error: %v output=%s", err, buf.String())
	}
	lock, err := config.LoadLockFile()
	if err != nil {
		t.Fatalf("LoadLockFile: %v", err)
	}
	entry := lock.GetEntry("native")
	if entry == nil {
		t.Fatalf("native not in lock: %#v", lock.Skills)
	}
	if entry.Agent != "hermes" || entry.Ownership != "imported" || entry.InstallMode != "in-place" || entry.UpdateStrategy != "none" || entry.TargetPath == "" || entry.Checksum == "" {
		t.Fatalf("entry = %#v", entry)
	}
}

func TestSkillImportHermesRejectsAllWithNames(t *testing.T) {
	buf, err := executeImportCommandForTest("skill", "import", "--agent", "hermes", "--all", "native")
	if err == nil {
		t.Fatalf("command returned nil error output=%s", buf.String())
	}
	if !strings.Contains(err.Error(), "either --all or specific skill") {
		t.Fatalf("error %q missing --all/name conflict", err.Error())
	}
}

func TestSkillImportHermesNamedMissingReturnsError(t *testing.T) {
	workdir, restore := setupImportCmdWorkdir(t)
	defer restore()
	writeCmdImportSkill(t, filepath.Join(workdir, ".hermes", "skills"), "native", "native")

	buf, err := executeImportCommandForTest("skill", "import", "--agent", "hermes", "missing")
	if err == nil {
		t.Fatalf("command returned nil error output=%s", buf.String())
	}
	if !strings.Contains(err.Error(), "not found: missing") {
		t.Fatalf("error %q missing not found skill", err.Error())
	}
	if _, statErr := os.Stat(filepath.Join(workdir, config.LockFileName)); !os.IsNotExist(statErr) {
		t.Fatalf("missing named import wrote lockfile or unexpected stat error: %v", statErr)
	}
}

func executeImportCommandForTest(args ...string) (*bytes.Buffer, error) {
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs(args)
	_ = importCmd.Flags().Set("agent", "")
	_ = importCmd.Flags().Set("all", "false")
	_ = importCmd.Flags().Set("dry-run", "false")
	_ = importCmd.Flags().Set("global", "false")
	return &buf, rootCmd.Execute()
}

func setupImportCmdWorkdir(t *testing.T) (string, func()) {
	t.Helper()
	workdir := t.TempDir()
	oldwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(workdir); err != nil {
		t.Fatal(err)
	}
	return workdir, func() { _ = os.Chdir(oldwd) }
}

func writeCmdImportSkill(t *testing.T, root, rel, name string) {
	t.Helper()
	dir := filepath.Join(root, rel)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("---\nname: "+name+"\ndescription: test\n---\nbody\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}
