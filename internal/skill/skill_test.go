package skill

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseSkillMDWithFrontmatter(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()
	skillDir := filepath.Join(tmpDir, "test-skill")
	if err := os.Mkdir(skillDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Create a SKILL.md with frontmatter
	skillMD := `---
name: test-skill
description: A test skill for testing
version: 1.0.0
author: Test Author
tags:
  - test
  - example
dependencies:
  - python
---

# Test Skill

This is a test skill for unit testing.
`
	skillMDPath := filepath.Join(skillDir, "SKILL.md")
	if err := os.WriteFile(skillMDPath, []byte(skillMD), 0644); err != nil {
		t.Fatalf("Failed to write SKILL.md: %v", err)
	}

	// Parse the skill
	meta, err := ParseSkillMD(skillDir)
	if err != nil {
		t.Fatalf("Failed to parse SKILL.md: %v", err)
	}

	// Verify metadata
	if meta.Name != "test-skill" {
		t.Errorf("Expected name 'test-skill', got '%s'", meta.Name)
	}
	if meta.Description != "A test skill for testing" {
		t.Errorf("Expected description 'A test skill for testing', got '%s'", meta.Description)
	}
	if meta.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", meta.Version)
	}
	if meta.Author != "Test Author" {
		t.Errorf("Expected author 'Test Author', got '%s'", meta.Author)
	}
	if len(meta.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(meta.Tags))
	}
	if len(meta.Dependencies) != 1 {
		t.Errorf("Expected 1 dependency, got %d", len(meta.Dependencies))
	}
}

func TestParseSkillMDWithoutFrontmatter(t *testing.T) {
	tmpDir := t.TempDir()
	skillDir := filepath.Join(tmpDir, "test-skill")
	if err := os.Mkdir(skillDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Create a SKILL.md without frontmatter
	skillMD := `# Browser Use

A skill for browser automation using Playwright.

## Features
- Web scraping
`
	skillMDPath := filepath.Join(skillDir, "SKILL.md")
	if err := os.WriteFile(skillMDPath, []byte(skillMD), 0644); err != nil {
		t.Fatalf("Failed to write SKILL.md: %v", err)
	}

	// Parse the skill
	meta, err := ParseSkillMD(skillDir)
	if err != nil {
		t.Fatalf("Failed to parse SKILL.md: %v", err)
	}

	// Verify metadata extracted from content
	if meta.Name != "Browser Use" {
		t.Errorf("Expected name 'Browser Use', got '%s'", meta.Name)
	}
	if meta.Description != "A skill for browser automation using Playwright." {
		t.Errorf("Expected description from content, got '%s'", meta.Description)
	}
}

func TestFindSkillMD(t *testing.T) {
	tmpDir := t.TempDir()
	skillDir := filepath.Join(tmpDir, "test-skill")
	if err := os.Mkdir(skillDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Test without SKILL.md
	if FindSkillMD(skillDir) {
		t.Error("Expected FindSkillMD to return false, got true")
	}

	// Create SKILL.md
	skillMDPath := filepath.Join(skillDir, "SKILL.md")
	if err := os.WriteFile(skillMDPath, []byte("# Test"), 0644); err != nil {
		t.Fatalf("Failed to write SKILL.md: %v", err)
	}

	// Test with SKILL.md
	if !FindSkillMD(skillDir) {
		t.Error("Expected FindSkillMD to return true, got false")
	}
}

func TestCreateSkillTemplate(t *testing.T) {
	tmpDir := t.TempDir()
	skillName := "my-test-skill"

	// Create skill template
	err := CreateSkillTemplate(skillName, tmpDir)
	if err != nil {
		t.Fatalf("Failed to create skill template: %v", err)
	}

	skillDir := filepath.Join(tmpDir, skillName)

	// Verify directory structure
	expectedDirs := []string{
		skillDir,
		filepath.Join(skillDir, "scripts"),
		filepath.Join(skillDir, "references"),
		filepath.Join(skillDir, "assets"),
	}

	for _, dir := range expectedDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Errorf("Expected directory %s to exist", dir)
		}
	}

	// Verify SKILL.md exists
	skillMDPath := filepath.Join(skillDir, "SKILL.md")
	if _, err := os.Stat(skillMDPath); os.IsNotExist(err) {
		t.Error("Expected SKILL.md to exist")
	}

	// Verify script exists
	scriptPath := filepath.Join(skillDir, "scripts", "hello.sh")
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		t.Error("Expected hello.sh script to exist")
	}

	// Verify reference exists
	refPath := filepath.Join(skillDir, "references", "ref.md")
	if _, err := os.Stat(refPath); os.IsNotExist(err) {
		t.Error("Expected ref.md to exist")
	}

	// Parse and verify SKILL.md metadata
	meta, err := ParseSkillMD(skillDir)
	if err != nil {
		t.Fatalf("Failed to parse generated SKILL.md: %v", err)
	}

	if meta.Name != skillName {
		t.Errorf("Expected name '%s', got '%s'", skillName, meta.Name)
	}
	if meta.Description == "" {
		t.Error("Expected description to be set")
	}
	if meta.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", meta.Version)
	}
}

func TestGetGitAuthor(t *testing.T) {
	author := getGitAuthor()

	// Should return a non-empty string (either from git config or "User")
	if author == "" {
		t.Error("Expected getGitAuthor to return a non-empty string")
	}

	// The author should be either from git config or the fallback "User"
	// We can't test the exact value as it depends on the environment
	t.Logf("Git author: %s", author)
}
