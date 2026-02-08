package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPromptCommand_SingleSkill(t *testing.T) {
	tmpDir := t.TempDir()
	skillDir := filepath.Join(tmpDir, "test-skill")
	if err := os.Mkdir(skillDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create valid SKILL.md
	skillMD := `---
name: test-skill
description: A test skill for prompt generation
---
# Test Skill
`
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(skillMD), 0644); err != nil {
		t.Fatal(err)
	}

	// Build available skills
	paths := []string{skillDir}
	availableSkills := AvailableSkills{}

	for _, path := range paths {
		meta, err := parseSkillMDForTest(path)
		if err != nil {
			t.Fatalf("Failed to parse skill: %v", err)
		}
		entry := SkillEntry{
			Name:        meta.name,
			Description: meta.description,
			Location:    filepath.Join(path, "SKILL.md"),
		}
		availableSkills.Skills = append(availableSkills.Skills, entry)
	}

	if len(availableSkills.Skills) != 1 {
		t.Errorf("Expected 1 skill, got %d", len(availableSkills.Skills))
	}

	if availableSkills.Skills[0].Name != "test-skill" {
		t.Errorf("Expected name 'test-skill', got '%s'", availableSkills.Skills[0].Name)
	}
}

// Helper to parse SKILL.md for testing
type testMeta struct {
	name        string
	description string
}

func parseSkillMDForTest(skillPath string) (*testMeta, error) {
	content, err := os.ReadFile(filepath.Join(skillPath, "SKILL.md"))
	if err != nil {
		return nil, err
	}

	meta := &testMeta{}
	lines := strings.Split(string(content), "\n")
	inFrontmatter := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "---" {
			if !inFrontmatter {
				inFrontmatter = true
				continue
			} else {
				break
			}
		}
		if inFrontmatter {
			if strings.HasPrefix(line, "name:") {
				meta.name = strings.TrimSpace(strings.TrimPrefix(line, "name:"))
			}
			if strings.HasPrefix(line, "description:") {
				meta.description = strings.TrimSpace(strings.TrimPrefix(line, "description:"))
			}
		}
	}

	return meta, nil
}

func TestContainsPath(t *testing.T) {
	paths := []string{"/a/b/c", "/d/e/f"}

	if !containsPath(paths, "/a/b/c") {
		t.Error("Expected to find /a/b/c")
	}

	if containsPath(paths, "/x/y/z") {
		t.Error("Did not expect to find /x/y/z")
	}
}
