package repository

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func createFile(t *testing.T, baseDir, path, content string) {
	fullPath := filepath.Join(baseDir, path)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		t.Fatalf("failed to mkdir: %v", err)
	}
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
}

func TestScanSkills(t *testing.T) {
	// Create repo dir
	repoDir, err := os.MkdirTemp("", "test-repo-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(repoDir) }()

	// Create structure:
	// .curated/skill-a/SKILL.md
	// .experimental/skill-b/SKILL.md
	// normal-skill/SKILL.md
	// ignore-me/file.txt (no SKILL.md)

	createFile(t, repoDir, ".curated/skill-a/SKILL.md", "name: skill-a\ndescription: A curated skill")
	createFile(t, repoDir, ".experimental/skill-b/SKILL.md", "name: skill-b\ndescription: An experimental skill")
	createFile(t, repoDir, "normal-skill/SKILL.md", "name: normal-skill\ndescription: A normal skill")
	createFile(t, repoDir, "ignore-me/file.txt", "nothing here")

	// Run ScanSkills
	owner := "testowner"
	repo := "testrepo"
	skills, err := ScanSkills(repoDir, "", owner, repo)
	if err != nil {
		t.Fatalf("ScanSkills failed: %v", err)
	}

	// Verify results
	if len(skills) != 3 {
		t.Errorf("Expected 3 skills, got %d", len(skills))
	}

	foundSkills := make(map[string]bool)
	for _, s := range skills {
		foundSkills[s.Name] = true

		// Check install arg matches expectation
		// Should be testowner/testrepo/<path>
		// E.g. testowner/testrepo/.curated/skill-a
		if !strings.HasPrefix(s.HTMLURL, owner+"/"+repo+"/") {
			t.Errorf("Unexpected install arg: %s", s.HTMLURL)
		}
	}

	if !foundSkills["skill-a"] {
		t.Error("Missing skill-a")
	}
	if !foundSkills["skill-b"] {
		t.Error("Missing skill-b")
	}
	if !foundSkills["normal-skill"] {
		t.Error("Missing normal-skill")
	}
}

func TestScanSkills_WithSubDir(t *testing.T) {
	// Create repo dir
	repoDir, err := os.MkdirTemp("", "test-repo-sub-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(repoDir) }()

	// Structure:
	// skills/
	//   .curated/skill-a/SKILL.md
	//   skill-c/SKILL.md

	createFile(t, repoDir, "skills/.curated/skill-a/SKILL.md", "name: skill-a")
	createFile(t, repoDir, "skills/skill-c/SKILL.md", "name: skill-c")

	owner := "testowner"
	repo := "testrepo"
	subPath := "skills"

	skills, err := ScanSkills(repoDir, subPath, owner, repo)
	if err != nil {
		t.Fatalf("ScanSkills failed: %v", err)
	}

	if len(skills) != 2 {
		t.Errorf("Expected 2 skills, got %d", len(skills))
		for _, s := range skills {
			t.Logf("Found: %s -> %s", s.Name, s.HTMLURL)
		}
	}

	foundSkills := make(map[string]bool)
	for _, s := range skills {
		foundSkills[s.Name] = true
		// Expect install path to include subdir
		// e.g. testowner/testrepo/skills/.curated/skill-a
		expectedPrefix := owner + "/" + repo + "/" + subPath + "/"
		if !strings.HasPrefix(s.HTMLURL, expectedPrefix) {
			t.Errorf("Skill %s install arg %s does not start with %s", s.Name, s.HTMLURL, expectedPrefix)
		}
	}

	if !foundSkills["skill-a"] {
		t.Error("Missing skill-a")
	}
	if !foundSkills["skill-c"] {
		t.Error("Missing skill-c")
	}
}
