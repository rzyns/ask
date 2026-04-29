package hermes

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yeasy/ask/internal/config"
)

func TestUninstallSkillRefusesImportedWithoutExplicitFlag(t *testing.T) {
	lock := &config.LockFile{Version: 1, Skills: []config.LockEntry{{Name: "local", Agent: "hermes", Ownership: "imported", InstallMode: "in-place", TargetPath: t.TempDir()}}}
	_, err := UninstallSkill(lock, "local", UninstallOptions{})
	if err == nil {
		t.Fatal("expected refusal for imported skill without flags")
	}
	if len(lock.Skills) != 1 {
		t.Fatalf("lock entry removed despite refusal")
	}
}

func TestUninstallSkillForgetImportedLeavesFiles(t *testing.T) {
	target := t.TempDir()
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
	root := t.TempDir()
	source := filepath.Join(root, "source")
	target := filepath.Join(root, "target")
	if err := os.MkdirAll(source, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatal(err)
	}
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
