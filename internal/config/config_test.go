package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	if config.Version != "1.0" {
		t.Errorf("Expected version 1.0, got %s", config.Version)
	}
	if len(config.Skills) != 0 {
		t.Errorf("Expected empty skills list, got %d", len(config.Skills))
	}
	// We now have 9 default repos: community, anthropics, scientific, superpowers, openai, matlab, composio, vercel, skillhub
	if len(config.Repos) != 9 {
		t.Errorf("Expected 9 default repos, got %d", len(config.Repos))
	}
}

func TestDefaultReposConfiguration(t *testing.T) {
	config := DefaultConfig()

	expectedRepos := map[string]struct {
		repoType string
		url      string
	}{
		"community":   {repoType: "topic", url: "agent-skill"},
		"anthropics":  {repoType: "dir", url: "anthropics/skills/skills"},
		"scientific":  {repoType: "dir", url: "K-Dense-AI/claude-scientific-skills/scientific-skills"},
		"superpowers": {repoType: "dir", url: "obra/superpowers/skills"},
		"openai":      {repoType: "dir", url: "openai/skills/skills"},
		"matlab":      {repoType: "dir", url: "matlab/skills/skills"},
		"composio":    {repoType: "dir", url: "ComposioHQ/awesome-claude-skills"},
		"vercel":      {repoType: "dir", url: "vercel-labs/agent-skills"},
		"skillhub":    {repoType: "skillhub", url: "skills"},
	}

	for _, repo := range config.Repos {
		expected, exists := expectedRepos[repo.Name]
		if !exists {
			t.Errorf("Unexpected repo in defaults: %s", repo.Name)
			continue
		}
		if repo.Type != expected.repoType {
			t.Errorf("Repo %s: expected type '%s', got '%s'", repo.Name, expected.repoType, repo.Type)
		}
		if repo.URL != expected.url {
			t.Errorf("Repo %s: expected URL '%s', got '%s'", repo.Name, expected.url, repo.URL)
		}
	}
}

func TestSaveAndLoadConfig(t *testing.T) {
	// Setup temporary file
	tmpFile := "test_ask.yaml"
	defer func() { _ = os.Remove(tmpFile) }()

	// Test Save
	config := DefaultConfig()
	config.Skills = append(config.Skills, "test-skill")

	// Temporarily redirect Save to write to tmpFile by modifying how we use Save
	// Since Save() uses correct "ask.yaml" hardcoded, we should mock or change current dir.
	// For simplicity in this env, we will change directory or just test logic if refactored.
	// Let's refactor Config.Save() in the future to take a path, but for now
	// we'll run this test in a temp dir.

	dir := t.TempDir()
	originalDir, _ := os.Getwd()
	_ = os.Chdir(dir)
	defer func() { _ = os.Chdir(originalDir) }()

	err := config.Save()
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Test Load
	loadedConfig, err := LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if len(loadedConfig.Skills) != 1 || loadedConfig.Skills[0] != "test-skill" {
		t.Errorf("Config persistence failed. Expected [test-skill], got %v", loadedConfig.Skills)
	}
}

func TestAddSkill(t *testing.T) {
	config := DefaultConfig()
	config.AddSkill("skill-a")
	config.AddSkill("skill-b")
	config.AddSkill("skill-a") // Duplicate

	if len(config.Skills) != 2 {
		t.Errorf("AddSkill should handle duplicates. Expected 2 skills, got %d", len(config.Skills))
	}
}

func TestRemoveSkill(t *testing.T) {
	config := DefaultConfig()
	config.AddSkill("skill-a")
	config.AddSkill("skill-b")

	config.RemoveSkill("skill-a")
	if len(config.Skills) != 1 {
		t.Errorf("Expected 1 skill after removal, got %d", len(config.Skills))
	}
	if config.Skills[0] != "skill-b" {
		t.Errorf("Expected skill-b to remain, got %s", config.Skills[0])
	}

	config.RemoveSkill("non-existent")
	if len(config.Skills) != 1 {
		t.Errorf("Removing non-existent skill should not change list size")
	}
}

func TestRemoveSkillInfo(t *testing.T) {
	config := DefaultConfig()
	config.AddSkillInfo(SkillInfo{Name: "skill-a", Description: "Skill A"})
	config.AddSkillInfo(SkillInfo{Name: "skill-b", Description: "Skill B"})

	if len(config.SkillsInfo) != 2 {
		t.Errorf("Expected 2 skill infos, got %d", len(config.SkillsInfo))
	}

	config.RemoveSkillInfo("skill-a")
	if len(config.SkillsInfo) != 1 {
		t.Errorf("Expected 1 skill info after removal, got %d", len(config.SkillsInfo))
	}
	if config.SkillsInfo[0].Name != "skill-b" {
		t.Errorf("Expected skill-b to remain, got %s", config.SkillsInfo[0].Name)
	}

	config.RemoveSkillInfo("non-existent")
	if len(config.SkillsInfo) != 1 {
		t.Errorf("Removing non-existent skill info should not change list size")
	}
}

func TestDefaultToolTargets(t *testing.T) {
	targets := DefaultToolTargets()
	if len(targets) == 0 {
		t.Error("Expected default tool targets, got empty list")
	}
	// Verify "agent" target exists
	foundAgent := false
	for _, target := range targets {
		if target.Name == "agent" {
			foundAgent = true
			break
		}
	}
	if !foundAgent {
		t.Error("Expected 'agent' tool target")
	}
}

func TestDetectExistingToolDirs(t *testing.T) {
	// Setup temp dir
	dir := t.TempDir()

	// Create .claude directory (mocking existing project)
	err := os.Mkdir(filepath.Join(dir, ".claude"), 0755)
	if err != nil {
		t.Fatalf("Failed to create .claude dir: %v", err)
	}

	detected := DetectExistingToolDirs(dir)
	foundClaude := false
	for _, target := range detected {
		if target.Name == "claude" {
			foundClaude = true
			break
		}
	}
	if !foundClaude {
		t.Error("Expected 'claude' to be detected in " + dir)
	}
}
