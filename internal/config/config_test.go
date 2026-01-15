package config

import (
	"os"
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
	if len(config.Sources) != 2 {
		t.Errorf("Expected 2 default sources, got %d", len(config.Sources))
	}
}

func TestSaveAndLoadConfig(t *testing.T) {
	// Setup temporary file
	tmpFile := "test_ask.yaml"
	defer os.Remove(tmpFile)

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
	os.Chdir(dir)
	defer os.Chdir(originalDir)

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
