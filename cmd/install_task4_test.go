package cmd

import (
	"strings"
	"testing"

	"github.com/yeasy/ask/internal/config"
	"github.com/yeasy/ask/internal/github"
	"github.com/yeasy/ask/internal/installer"
)

func TestSkillsSHRepoSearchSelectionQueuesSingleExactSupportedMatch(t *testing.T) {
	repos := []github.Repository{
		{
			Name:       "grill-me-extra",
			Source:     config.RepoTypeSkillsSH,
			InstallRef: "https://github.com/acme/skills/tree/main/skills/grill-me-extra",
			HTMLURL:    "https://github.com/acme/skills/tree/main/skills/grill-me-extra",
			Supported:  true,
		},
		{
			Name:             "grill-me",
			Source:           config.RepoTypeSkillsSH,
			SourceIdentifier: "skill_123",
			UpdateStrategy:   "skills.sh",
			InstallRef:       "https://github.com/mattpocock/skills/tree/main/skills/productivity/grill-me",
			HTMLURL:          "https://github.com/mattpocock/skills/tree/main/skills/productivity/grill-me",
			Supported:        true,
		},
	}
	var expanded, failed []string
	metadata := map[string]installer.InstallSourceMetadata{}

	messages := appendSkillsSHSearchSelection(&expanded, &failed, "skills-sh", "grill-me", repos, metadata)
	if len(messages) != 0 {
		t.Fatalf("unexpected messages: %v", messages)
	}
	if len(expanded) != 1 || expanded[0] != repos[1].InstallRef {
		t.Fatalf("unexpected expanded refs: %v", expanded)
	}
	if len(failed) != 0 {
		t.Fatalf("unexpected failed refs: %v", failed)
	}
	if _, ok := metadata[repos[1].InstallRef]; !ok {
		t.Fatalf("metadata missing under install ref: %#v", metadata)
	}
}

func TestSkillsSHRepoSearchSelectionRejectsAmbiguousSupportedExactMatches(t *testing.T) {
	repos := []github.Repository{
		{Name: "grill-me", Source: config.RepoTypeSkillsSH, InstallRef: "https://github.com/one/skills/tree/main/grill-me", HTMLURL: "https://github.com/one/skills/tree/main/grill-me", Supported: true},
		{Name: "grill-me", Source: config.RepoTypeSkillsSH, InstallRef: "https://github.com/two/skills/tree/main/grill-me", HTMLURL: "https://github.com/two/skills/tree/main/grill-me", Supported: true},
	}
	var expanded, failed []string
	metadata := map[string]installer.InstallSourceMetadata{}

	messages := appendSkillsSHSearchSelection(&expanded, &failed, "skills-sh", "grill-me", repos, metadata)
	if len(expanded) != 0 || len(failed) != 1 || failed[0] != "grill-me" {
		t.Fatalf("ambiguous result should fail without expansion, expanded=%v failed=%v", expanded, failed)
	}
	joined := strings.Join(messages, "\n")
	for _, want := range []string{"ambiguous", repos[0].InstallRef, repos[1].InstallRef} {
		if !strings.Contains(joined, want) {
			t.Fatalf("ambiguous message missing %q:\n%s", want, joined)
		}
	}
}

func TestSkillsSHRepoSearchSelectionReportsUnsupportedOnlyExactMatches(t *testing.T) {
	repos := []github.Repository{{
		Name:              "mintlify",
		Source:            config.RepoTypeSkillsSH,
		UnsupportedReason: "no native ASK resolver for skills.sh entry",
	}}
	var expanded, failed []string
	metadata := map[string]installer.InstallSourceMetadata{}

	messages := appendSkillsSHSearchSelection(&expanded, &failed, "skills-sh", "mintlify", repos, metadata)
	if len(expanded) != 0 || len(failed) != 1 || failed[0] != "mintlify" {
		t.Fatalf("unsupported result should fail without expansion, expanded=%v failed=%v", expanded, failed)
	}
	if len(messages) != 1 || !strings.Contains(messages[0], "not installable") || !strings.Contains(messages[0], repos[0].UnsupportedReason) {
		t.Fatalf("unsupported message missing reason: %v", messages)
	}
}

func TestSkillsSHRepoSearchSelectionReportsNoExactMatch(t *testing.T) {
	repos := []github.Repository{{Name: "grill-me-extra", Source: config.RepoTypeSkillsSH, Supported: true, InstallRef: "https://github.com/acme/skills/tree/main/grill-me-extra", HTMLURL: "https://github.com/acme/skills/tree/main/grill-me-extra"}}
	var expanded, failed []string
	metadata := map[string]installer.InstallSourceMetadata{}

	messages := appendSkillsSHSearchSelection(&expanded, &failed, "skills-sh", "grill-me", repos, metadata)
	if len(expanded) != 0 || len(failed) != 1 || failed[0] != "grill-me" {
		t.Fatalf("no exact result should fail without expansion, expanded=%v failed=%v", expanded, failed)
	}
	if len(messages) != 1 || !strings.Contains(messages[0], "not found") || !strings.Contains(messages[0], "skills-sh") {
		t.Fatalf("not-found message unclear: %v", messages)
	}
}

func TestRepoInstallSelectionRejectsUnsupportedSkillsSHEntry(t *testing.T) {
	repo := github.Repository{
		Name:              "mintlify",
		Source:            config.RepoTypeSkillsSH,
		Supported:         false,
		UnsupportedReason: "no native ASK resolver for skills.sh entry",
	}
	var expanded, failed []string
	metadata := map[string]installer.InstallSourceMetadata{}

	ok, msg := appendInstallableRepoSkill(&expanded, &failed, repo, metadata)
	if ok {
		t.Fatal("unsupported skills.sh entry was queued for installation")
	}
	if len(expanded) != 0 || len(failed) != 1 || failed[0] != "mintlify" {
		t.Fatalf("unexpected selection state expanded=%v failed=%v", expanded, failed)
	}
	if !strings.Contains(msg, "mintlify") || !strings.Contains(msg, repo.UnsupportedReason) {
		t.Fatalf("unclear unsupported message: %q", msg)
	}
}

func TestRepoInstallSelectionRecordsSkillsSHProvenanceForNativeRef(t *testing.T) {
	repo := github.Repository{
		Name:             "pdf",
		HTMLURL:          "https://github.com/acme/skills/tree/main/pdf",
		Source:           config.RepoTypeSkillsSH,
		SourceIdentifier: "skill_123",
		UpdateStrategy:   "skills.sh",
		Supported:        true,
	}
	var expanded, failed []string
	metadata := map[string]installer.InstallSourceMetadata{}

	ok, msg := appendInstallableRepoSkill(&expanded, &failed, repo, metadata)
	if !ok || msg != "" {
		t.Fatalf("supported skills.sh entry not queued ok=%v msg=%q", ok, msg)
	}
	if len(expanded) != 1 || expanded[0] != repo.HTMLURL || len(failed) != 0 {
		t.Fatalf("unexpected selection state expanded=%v failed=%v", expanded, failed)
	}
	if _, ok := metadata[repo.HTMLURL]; !ok {
		t.Fatalf("metadata not recorded under native ref %q: %#v", repo.HTMLURL, metadata)
	}
}

func TestRepoNameInstallExpansionSkipsUnsupportedSkillsSHEntries(t *testing.T) {
	repos := []github.Repository{
		{
			Name:              "mintlify",
			Source:            config.RepoTypeSkillsSH,
			UnsupportedReason: "no native ASK resolver for skills.sh entry",
		},
		{
			Name:             "pdf",
			HTMLURL:          "https://github.com/acme/skills/tree/main/pdf",
			Source:           config.RepoTypeSkillsSH,
			SourceIdentifier: "skill_123",
			UpdateStrategy:   "skills.sh",
			Supported:        true,
		},
	}
	var expanded, failed []string
	metadata := map[string]installer.InstallSourceMetadata{}
	var warnings []string

	for _, repo := range repos {
		ok, msg := appendInstallableRepoSkill(&expanded, &failed, repo, metadata)
		if !ok {
			warnings = append(warnings, msg)
		}
	}

	if len(expanded) != 1 || expanded[0] != repos[1].HTMLURL {
		t.Fatalf("unsupported entry should not be expanded for installer: %v", expanded)
	}
	if len(failed) != 1 || failed[0] != "mintlify" {
		t.Fatalf("unsupported entry should be marked failed, got %v", failed)
	}
	if len(warnings) != 1 || !strings.Contains(warnings[0], repos[0].UnsupportedReason) {
		t.Fatalf("unsupported warning missing reason: %v", warnings)
	}
	if _, ok := metadata[repos[1].HTMLURL]; !ok {
		t.Fatalf("supported native ref metadata missing: %#v", metadata)
	}
	if _, ok := metadata[repos[0].HTMLURL]; ok {
		t.Fatalf("unsupported entry recorded metadata under empty ref: %#v", metadata)
	}
}
