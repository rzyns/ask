package skill

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanDirectory(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "skill_scan_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create a dummy skill structure
	// tempDir/
	//   skill1/
	//     SKILL.md
	//   group/
	//     skill2/
	//       SKILL.md
	//   not_a_skill/
	//     other.txt

	skill1Dir := filepath.Join(tempDir, "skill1")
	if err := os.MkdirAll(skill1Dir, 0755); err != nil {
		t.Fatalf("Failed to create skill1 dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(skill1Dir, "SKILL.md"), []byte("---\nname: skill1\n---\n"), 0644); err != nil {
		t.Fatalf("Failed to create skill1 SKILL.md: %v", err)
	}

	skill2Dir := filepath.Join(tempDir, "group", "skill2")
	if err := os.MkdirAll(skill2Dir, 0755); err != nil {
		t.Fatalf("Failed to create skill2 dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(skill2Dir, "SKILL.md"), []byte("---\nname: skill2\n---\n"), 0644); err != nil {
		t.Fatalf("Failed to create skill2 SKILL.md: %v", err)
	}

	notSkillDir := filepath.Join(tempDir, "not_a_skill")
	if err := os.MkdirAll(notSkillDir, 0755); err != nil {
		t.Fatalf("Failed to create not_a_skill dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(notSkillDir, "other.txt"), []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create other.txt: %v", err)
	}

	// Test ScanDirectory
	results, err := ScanDirectory(tempDir, 3)
	if err != nil {
		t.Fatalf("ScanDirectory failed: %v", err)
	}

	// Expect 2 skills
	if len(results) != 2 {
		t.Errorf("Expected 2 skills, got %d", len(results))
	}

	// Verify details
	foundSkill1 := false
	foundSkill2 := false

	for _, s := range results {
		if s.Meta != nil && s.Meta.Name == "skill1" {
			foundSkill1 = true
		}
		if s.Meta != nil && s.Meta.Name == "skill2" {
			foundSkill2 = true
		}
	}

	if !foundSkill1 {
		t.Error("Did not find skill1")
	}
	if !foundSkill2 {
		t.Error("Did not find skill2")
	}
}

func TestScanDirectory_DepthLimit(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "skill_scan_depth")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create a skill at depth 2
	shallowSkill := filepath.Join(tempDir, "level1", "shallow-skill")
	if err := os.MkdirAll(shallowSkill, 0755); err != nil {
		t.Fatalf("Failed to create shallow skill dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(shallowSkill, "SKILL.md"), []byte("---\nname: shallow\n---\n"), 0644); err != nil {
		t.Fatalf("Failed to create SKILL.md: %v", err)
	}

	// Create a skill at depth 5 (should be excluded with limit 3)
	deepSkill := filepath.Join(tempDir, "level1", "level2", "level3", "level4", "deep-skill")
	if err := os.MkdirAll(deepSkill, 0755); err != nil {
		t.Fatalf("Failed to create deep skill dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(deepSkill, "SKILL.md"), []byte("---\nname: deep\n---\n"), 0644); err != nil {
		t.Fatalf("Failed to create SKILL.md: %v", err)
	}

	// Test with depth limit 3
	results, err := ScanDirectory(tempDir, 3)
	if err != nil {
		t.Fatalf("ScanDirectory failed: %v", err)
	}

	// Should only find the shallow skill
	if len(results) != 1 {
		t.Errorf("Expected 1 skill with depth limit 3, got %d", len(results))
	}

	if len(results) > 0 && results[0].Meta != nil && results[0].Meta.Name != "shallow" {
		t.Errorf("Expected to find 'shallow' skill, got %q", results[0].Meta.Name)
	}
}

func TestScanDirectory_HiddenDirectories(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "skill_scan_hidden")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create a skill in a hidden directory (should be skipped)
	hiddenSkill := filepath.Join(tempDir, ".hidden", "secret-skill")
	if err := os.MkdirAll(hiddenSkill, 0755); err != nil {
		t.Fatalf("Failed to create hidden skill dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(hiddenSkill, "SKILL.md"), []byte("---\nname: secret\n---\n"), 0644); err != nil {
		t.Fatalf("Failed to create SKILL.md: %v", err)
	}

	// Create a visible skill
	visibleSkill := filepath.Join(tempDir, "visible-skill")
	if err := os.MkdirAll(visibleSkill, 0755); err != nil {
		t.Fatalf("Failed to create visible skill dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(visibleSkill, "SKILL.md"), []byte("---\nname: visible\n---\n"), 0644); err != nil {
		t.Fatalf("Failed to create SKILL.md: %v", err)
	}

	results, err := ScanDirectory(tempDir, 3)
	if err != nil {
		t.Fatalf("ScanDirectory failed: %v", err)
	}

	// Should only find the visible skill
	if len(results) != 1 {
		t.Errorf("Expected 1 skill (hidden should be skipped), got %d", len(results))
	}

	if len(results) > 0 && results[0].Meta != nil && results[0].Meta.Name != "visible" {
		t.Errorf("Expected to find 'visible' skill, got %q", results[0].Meta.Name)
	}
}

func TestScanDirectory_Symlinks(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "skill_scan_symlink")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create a real skill
	realSkill := filepath.Join(tempDir, "real-skill")
	if err := os.MkdirAll(realSkill, 0755); err != nil {
		t.Fatalf("Failed to create skill dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(realSkill, "SKILL.md"), []byte("---\nname: real\n---\n"), 0644); err != nil {
		t.Fatalf("Failed to create SKILL.md: %v", err)
	}

	// Create a symlink to the skill directory (should be skipped for security)
	symlinkPath := filepath.Join(tempDir, "symlink-skill")
	if err := os.Symlink(realSkill, symlinkPath); err != nil {
		t.Skipf("Cannot create symlinks on this system: %v", err)
	}

	results, err := ScanDirectory(tempDir, 3)
	if err != nil {
		t.Fatalf("ScanDirectory failed: %v", err)
	}

	// Should only find the real skill (symlink should be skipped)
	if len(results) != 1 {
		t.Errorf("Expected 1 skill (symlink should be skipped), got %d", len(results))
	}
}

func TestScanDirectory_EmptyDirectory(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "skill_scan_empty")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	results, err := ScanDirectory(tempDir, 3)
	if err != nil {
		t.Fatalf("ScanDirectory failed: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("Expected 0 skills in empty directory, got %d", len(results))
	}
}

func TestScanDirectory_NonexistentPath(t *testing.T) {
	_, err := ScanDirectory("/nonexistent/path/12345", 3)
	if err == nil {
		t.Error("Expected error for nonexistent path")
	}
}
