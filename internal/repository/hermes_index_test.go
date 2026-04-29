package repository

import (
	"strings"
	"testing"

	"github.com/yeasy/ask/internal/config"
)

func TestParseHermesIndexMapsGitHubResolvedIDToCandidate(t *testing.T) {
	index := `{
		"generated_at": "2026-01-02T03:04:05Z",
		"skill_count": 2,
		"version": "1",
		"skills": [
			{
				"description": "Run controlled marketing experiments",
				"source": "skills.sh",
				"resolved_github_id": "coreyhaines31/marketingskills/skills/ab-test-setup"
			},
			{
				"name": "Ignored local item",
				"description": "No GitHub path here",
				"source": "clawhub",
				"id": "clawhub-only"
			}
		]
	}`

	skills, err := parseHermesIndex(strings.NewReader(index))
	if err != nil {
		t.Fatalf("parseHermesIndex returned error: %v", err)
	}

	candidates := hermesIndexSkillsToCandidates(skills, "")
	if len(candidates) != 1 {
		t.Fatalf("expected 1 GitHub-resolvable candidate, got %d: %#v", len(candidates), candidates)
	}

	candidate := candidates[0]
	if candidate.Name != "ab-test-setup" {
		t.Fatalf("expected basename fallback name, got %q", candidate.Name)
	}
	if candidate.FullName != "coreyhaines31/marketingskills/skills/ab-test-setup" {
		t.Fatalf("unexpected full name: %q", candidate.FullName)
	}
	if candidate.Description != "Run controlled marketing experiments" {
		t.Fatalf("unexpected description: %q", candidate.Description)
	}
	if candidate.Source != config.RepoTypeHermes {
		t.Fatalf("expected source %q, got %q", config.RepoTypeHermes, candidate.Source)
	}
	if candidate.Install.Kind != InstallRefGitHubPath || candidate.Install.Value != "coreyhaines31/marketingskills/skills/ab-test-setup" {
		t.Fatalf("unexpected install ref: %#v", candidate.Install)
	}
	if candidate.Stars != 0 {
		t.Fatalf("expected zero stars, got %d", candidate.Stars)
	}
}

func TestHermesIndexCandidateMappingUsesExplicitNameAndRepoPath(t *testing.T) {
	skills := []hermesIndexSkill{{
		Name:        "Custom Name",
		Description: "A GitHub entry with split repo/path fields",
		Repo:        "owner/repo",
		Path:        "skills/custom-skill",
	}}

	candidates := hermesIndexSkillsToCandidates(skills, "")
	if len(candidates) != 1 {
		t.Fatalf("expected 1 candidate, got %d", len(candidates))
	}
	if candidates[0].Name != "Custom Name" {
		t.Fatalf("expected explicit name, got %q", candidates[0].Name)
	}
	if candidates[0].Install.Value != "owner/repo/skills/custom-skill" {
		t.Fatalf("expected joined repo/path install value, got %q", candidates[0].Install.Value)
	}
}

func TestHermesIndexCandidateMappingResolvesOfficialPathToHermesRepo(t *testing.T) {
	skills := []hermesIndexSkill{{
		Name:        "gitnexus-explorer",
		Description: "Index a codebase with GitNexus",
		Source:      "official",
		Identifier:  "official/research/gitnexus-explorer",
		Path:        "research/gitnexus-explorer",
	}}

	candidates := hermesIndexSkillsToCandidates(skills, "nexus")
	if len(candidates) != 1 {
		t.Fatalf("expected official Hermes candidate, got %d: %#v", len(candidates), candidates)
	}
	want := "NousResearch/hermes-agent/optional-skills/research/gitnexus-explorer"
	if candidates[0].FullName != want || candidates[0].Install.Value != want {
		t.Fatalf("expected official install path %q, got full=%q install=%q", want, candidates[0].FullName, candidates[0].Install.Value)
	}
}

func TestHermesIndexCandidateMappingSkipsBundledHermesSkills(t *testing.T) {
	skills := []hermesIndexSkill{{
		Name:             "core-skill",
		ResolvedGitHubID: "NousResearch/hermes-agent/skills/core-skill",
	}}

	candidates := hermesIndexSkillsToCandidates(skills, "")
	if len(candidates) != 0 {
		t.Fatalf("expected bundled Hermes skill to be skipped, got %#v", candidates)
	}
}

func TestHermesIndexCandidateMappingFiltersKeywordCaseInsensitively(t *testing.T) {
	skills := []hermesIndexSkill{
		{Name: "Alpha", Description: "First", ResolvedGitHubID: "owner/repo/skills/alpha"},
		{Name: "Beta", Description: "Marketing automation", ResolvedGitHubID: "owner/repo/skills/beta"},
	}

	candidates := hermesIndexSkillsToCandidates(skills, "market")
	if len(candidates) != 1 {
		t.Fatalf("expected 1 keyword match, got %d: %#v", len(candidates), candidates)
	}
	if candidates[0].Name != "Beta" {
		t.Fatalf("expected Beta match, got %q", candidates[0].Name)
	}
}

func TestHermesIndexCandidateMappingRejectsAmbiguousNonGitHubRefs(t *testing.T) {
	skills := []hermesIndexSkill{
		{Name: "External URL", URL: "https://example.com/owner/repo"},
		{Name: "Slug URL", URL: "owner/repo"},
		{Name: "Hosted Repo", Repo: "https://example.com/owner/repo", Path: "skills/external"},
		{Name: "GitHub URL", URL: "https://github.com/owner/repo/tree/main/skills/github-url"},
	}

	candidates := hermesIndexSkillsToCandidates(skills, "")
	if len(candidates) != 1 {
		t.Fatalf("expected only the GitHub URL candidate, got %d: %#v", len(candidates), candidates)
	}
	if candidates[0].Name != "GitHub URL" {
		t.Fatalf("expected GitHub URL candidate, got %q", candidates[0].Name)
	}
	if candidates[0].Install.Value != "owner/repo/skills/github-url" {
		t.Fatalf("expected normalized GitHub URL path, got %q", candidates[0].Install.Value)
	}
}

func TestParseHermesIndexRejectsMalformedJSON(t *testing.T) {
	_, err := parseHermesIndex(strings.NewReader(`{"skills": [`))
	if err == nil {
		t.Fatal("expected malformed JSON error")
	}
}
