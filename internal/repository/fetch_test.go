package repository

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yeasy/ask/internal/config"
	"github.com/yeasy/ask/internal/github"
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

func TestFetchSkillsUnknownTypeReturnsError(t *testing.T) {
	_, err := FetchSkills(config.Repo{Type: "bogus"})
	if err == nil {
		t.Fatal("expected error for unknown repository type")
	}
	if got := err.Error(); !strings.Contains(got, "unknown repository type: bogus") {
		t.Fatalf("expected unknown type error, got %q", got)
	}
}

func TestFetchSkillsDirInvalidURLReturnsError(t *testing.T) {
	_, err := FetchSkills(config.Repo{
		Type: config.RepoTypeDir,
		URL:  "owneronly",
	})
	if err == nil {
		t.Fatal("expected error for invalid dir repository URL")
	}
	if got := err.Error(); !strings.Contains(got, "invalid repository URL format: owneronly") {
		t.Fatalf("expected invalid repository URL format error, got %q", got)
	}
}

func TestFetchSkillsDirRejectsBundledHermesSkills(t *testing.T) {
	origFetchViaGit := fetchSkillsViaGitFunc
	origSearchDir := searchDirFunc
	t.Cleanup(func() {
		fetchSkillsViaGitFunc = origFetchViaGit
		searchDirFunc = origSearchDir
	})
	fetchSkillsViaGitFunc = func(config.Repo) ([]github.Repository, error) {
		t.Fatal("bundled Hermes dir source should be rejected before git fetch")
		return nil, nil
	}
	searchDirFunc = func(_, _, _ string) ([]github.Repository, error) {
		t.Fatal("bundled Hermes dir source should be rejected before GitHub search fallback")
		return nil, nil
	}

	_, err := FetchSkills(config.Repo{
		Type: config.RepoTypeDir,
		URL:  "NousResearch/hermes-agent/skills",
	})
	if err == nil {
		t.Fatal("expected bundled Hermes skills error")
	}
	if got := err.Error(); !strings.Contains(got, "bundled Hermes skills") {
		t.Fatalf("expected bundled Hermes skills error, got %q", got)
	}
}

func TestFetchSkillsViaGitNonDirReturnsError(t *testing.T) {
	_, err := FetchSkillsViaGit(config.Repo{Type: config.RepoTypeRegistry})
	if err == nil {
		t.Fatal("expected error for non-dir repository type")
	}
	if got := err.Error(); !strings.Contains(got, "git fetch only supports 'dir' type repos") {
		t.Fatalf("expected non-dir git fetch error, got %q", got)
	}
}

func TestFetchSkillsRegistryDispatchReturnsRegistryEntries(t *testing.T) {
	config.SetOffline(false)

	index := validRegistryIndex()
	data, err := json.Marshal(index)
	if err != nil {
		t.Fatalf("failed to marshal test index: %v", err)
	}

	_, cleanup := setupTestServer(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	})
	defer cleanup()

	results, err := FetchSkills(config.Repo{
		Type: config.RepoTypeRegistry,
		URL:  "owner/repo/registry/index.json",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	if results[0].Name != "code-review" {
		t.Fatalf("expected first registry entry code-review, got %q", results[0].Name)
	}
	if results[1].Name != "docker-helper" {
		t.Fatalf("expected second registry entry docker-helper, got %q", results[1].Name)
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
