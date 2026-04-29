package hermes

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yeasy/ask/internal/config"
)

func TestUninstallSkillRefusesImportedWithoutExplicitFlag(t *testing.T) {
	lock := &config.LockFile{Version: 1, Skills: []config.LockEntry{{Name: "local", Agent: "hermes", Ownership: "imported", InstallMode: "in-place", TargetPath: hermesTargetPathForTest(t, "local")}}}
	_, err := UninstallSkill(lock, "local", UninstallOptions{})
	if err == nil {
		t.Fatal("expected refusal for imported skill without flags")
	}
	if len(lock.Skills) != 1 {
		t.Fatalf("lock entry removed despite refusal")
	}
}

func TestUninstallSkillForgetImportedLeavesFiles(t *testing.T) {
	target := hermesTargetPathForTest(t, "local")
	lock := &config.LockFile{Version: 1, Skills: []config.LockEntry{{Name: "local", Agent: "hermes", Ownership: "imported", InstallMode: "in-place", TargetPath: target}}}
	action, err := UninstallSkill(lock, "local", UninstallOptions{Forget: true})
	if err != nil {
		t.Fatalf("UninstallSkill returned error: %v", err)
	}
	if !action.Forgot || action.RemovedFiles || len(lock.Skills) != 0 {
		t.Fatalf("action/lock = %#v/%#v", action, lock.Skills)
	}
	if _, err := os.Stat(target); err != nil {
		t.Fatalf("target removed on forget: %v", err)
	}
}

func TestUninstallSkillASKRemovesTargetAndSource(t *testing.T) {
	source := askSourcePathForTest(t, "managed")
	target := hermesTargetPathForTest(t, "managed")
	lock := &config.LockFile{Version: 1, Skills: []config.LockEntry{{Name: "managed", Agent: "hermes", Ownership: "ask", SourcePath: source, TargetPath: target}}}
	action, err := UninstallSkill(lock, "managed", UninstallOptions{})
	if err != nil {
		t.Fatalf("UninstallSkill returned error: %v", err)
	}
	if !action.Forgot || !action.RemovedFiles || !action.RemovedSource || len(lock.Skills) != 0 {
		t.Fatalf("action/lock = %#v/%#v", action, lock.Skills)
	}
	if _, err := os.Stat(source); !os.IsNotExist(err) {
		t.Fatalf("source still exists or stat failed: %v", err)
	}
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Fatalf("target still exists or stat failed: %v", err)
	}
}

func TestUninstallSkillImportedDeleteFilesRemovesHermesTarget(t *testing.T) {
	target := hermesTargetPathForTest(t, "local")
	lock := &config.LockFile{Version: 1, Skills: []config.LockEntry{{Name: "local", Agent: "hermes", Ownership: "imported", InstallMode: "in-place", TargetPath: target, SourcePath: target}}}

	action, err := UninstallSkill(lock, "local", UninstallOptions{DeleteFiles: true})
	if err != nil {
		t.Fatalf("UninstallSkill returned error: %v", err)
	}
	if !action.Forgot || !action.RemovedFiles || action.RemovedSource || len(lock.Skills) != 0 {
		t.Fatalf("action/lock = %#v/%#v", action, lock.Skills)
	}
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Fatalf("target still exists or stat failed: %v", err)
	}
}

func TestUninstallSkillRejectsMaliciousASKOwnedLockPaths(t *testing.T) {
	hermesHome := t.TempDir()
	t.Setenv("HERMES_HOME", hermesHome)
	outside := filepath.Join(t.TempDir(), "managed")
	mustMkdir(t, outside)
	validTarget := filepath.Join(hermesHome, "skills", "managed")
	validSource := askSourcePathForTest(t, "managed")

	tests := []struct {
		name  string
		entry config.LockEntry
	}{
		{
			name:  "target outside Hermes roots",
			entry: config.LockEntry{Name: "managed", Agent: "hermes", Ownership: "ask", TargetPath: outside, SourcePath: validSource},
		},
		{
			name:  "target basename mismatch",
			entry: config.LockEntry{Name: "managed", Agent: "hermes", Ownership: "ask", TargetPath: filepath.Join(hermesHome, "skills", "other"), SourcePath: validSource},
		},
		{
			name:  "target unclean path",
			entry: config.LockEntry{Name: "managed", Agent: "hermes", Ownership: "ask", TargetPath: hermesHome + string(os.PathSeparator) + "skills" + string(os.PathSeparator) + ".." + string(os.PathSeparator) + "skills" + string(os.PathSeparator) + "managed", SourcePath: validSource},
		},
		{
			name:  "source outside ASK roots",
			entry: config.LockEntry{Name: "managed", Agent: "hermes", Ownership: "ask", TargetPath: validTarget, SourcePath: outside},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lock := &config.LockFile{Version: 1, Skills: []config.LockEntry{tt.entry}}
			_, err := UninstallSkill(lock, "managed", UninstallOptions{})
			if err == nil {
				t.Fatal("expected malicious/stale lock metadata to be rejected")
			}
			if !strings.Contains(err.Error(), "refusing to remove") {
				t.Fatalf("error = %q, want refusing to remove", err.Error())
			}
			if len(lock.Skills) != 1 {
				t.Fatalf("lock entry removed despite rejection")
			}
			if _, statErr := os.Stat(outside); statErr != nil {
				t.Fatalf("outside directory was removed or inaccessible: %v", statErr)
			}
		})
	}
}

func TestUninstallSkillRejectsMaliciousImportedDeleteFilesPath(t *testing.T) {
	hermesHome := t.TempDir()
	t.Setenv("HERMES_HOME", hermesHome)
	outside := filepath.Join(t.TempDir(), "local")
	mustMkdir(t, outside)
	lock := &config.LockFile{Version: 1, Skills: []config.LockEntry{{Name: "local", Agent: "hermes", Ownership: "imported", InstallMode: "in-place", TargetPath: outside, SourcePath: outside}}}

	_, err := UninstallSkill(lock, "local", UninstallOptions{DeleteFiles: true})
	if err == nil {
		t.Fatal("expected malicious imported lock metadata to be rejected")
	}
	if !strings.Contains(err.Error(), "refusing to remove") {
		t.Fatalf("error = %q, want refusing to remove", err.Error())
	}
	if len(lock.Skills) != 1 {
		t.Fatalf("lock entry removed despite rejection")
	}
	if _, err := os.Stat(outside); err != nil {
		t.Fatalf("outside directory was removed or inaccessible: %v", err)
	}
}

func hermesTargetPathForTest(t *testing.T, name string) string {
	t.Helper()
	hermesHome := t.TempDir()
	t.Setenv("HERMES_HOME", hermesHome)
	path := filepath.Join(hermesHome, "skills", name)
	mustMkdir(t, path)
	return path
}

func askSourcePathForTest(t *testing.T, name string) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	path := filepath.Join(home, ".ask", "skills", name)
	mustMkdir(t, path)
	return path
}

func mustMkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
}
