package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/yeasy/ask/internal/config"
)

func TestHermesImportWritesLockForLocalOnlySkills(t *testing.T) {
	home := t.TempDir()
	hermesHome := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("HERMES_HOME", hermesHome)
	writeHermesListSkill(t, filepath.Join(hermesHome, "skills"), "local-skill", "local-skill", "Local skill", "")

	runImportCommandForTest(t, true, []string{"hermes"}, false, true, nil)

	lock, err := config.LoadGlobalLockFile()
	if err != nil {
		t.Fatalf("LoadGlobalLockFile returned error: %v", err)
	}
	entry := lock.GetEntryForAgent("local-skill", "hermes")
	if entry == nil {
		t.Fatalf("expected local-skill lock entry, got %#v", lock.Skills)
	}
	if entry.Ownership != "imported" || entry.InstallMode != "in-place" || entry.UpdateStrategy != "none" {
		t.Fatalf("entry metadata = %#v", entry)
	}
}

func TestHermesImportDryRunDoesNotWriteLock(t *testing.T) {
	home := t.TempDir()
	hermesHome := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("HERMES_HOME", hermesHome)
	writeHermesListSkill(t, filepath.Join(hermesHome, "skills"), "local-skill", "local-skill", "Local skill", "")

	runImportCommandForTest(t, true, []string{"hermes"}, true, true, nil)

	if _, err := os.Stat(filepath.Join(home, ".ask", "ask.lock")); err == nil {
		t.Fatal("dry-run created ask.lock")
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat ask.lock: %v", err)
	}
}

func runImportCommandForTest(t *testing.T, global bool, agents []string, dryRun bool, importAll bool, args []string) {
	t.Helper()
	cmd := &cobra.Command{Use: "import"}
	cmd.SetOut(os.Stdout)
	cmd.SetErr(os.Stderr)
	cmd.Flags().Bool("global", global, "")
	cmd.Flags().StringSliceP("agent", "a", []string{}, "")
	cmd.Flags().Bool("dry-run", false, "")
	cmd.Flags().Bool("all", false, "")
	_ = cmd.Flags().Set("global", boolString(global))
	_ = cmd.Flags().Set("agent", strings.Join(agents, ","))
	if dryRun {
		_ = cmd.Flags().Set("dry-run", "true")
	}
	if importAll {
		_ = cmd.Flags().Set("all", "true")
	}
	runImport(cmd, args)
}

func boolString(v bool) string {
	if v {
		return "true"
	}
	return "false"
}
