package repository

import (
	"testing"

	"github.com/yeasy/ask/internal/config"
)

func TestCandidateToRepositoryPreservesSkillsSHInstallabilityAndProvenance(t *testing.T) {
	supported := SkillCandidate{
		Name:             "pdf",
		FullName:         "pdf",
		Description:      "PDF helper",
		Source:           config.RepoTypeSkillsSH,
		SourceIdentifier: "skill_123",
		UpdateStrategy:   "skills.sh",
		Install: InstallRef{
			Kind:  InstallRefGitHubPath,
			Value: "https://github.com/acme/skills/tree/main/pdf",
		},
		Stars:     42,
		PageURL:   "https://skills.sh/pdf",
		Supported: true,
	}

	repo := candidateToRepository(supported)
	if repo.Source != config.RepoTypeSkillsSH || repo.SourceIdentifier != "skill_123" || repo.UpdateStrategy != "skills.sh" {
		t.Fatalf("skills.sh provenance not preserved: %#v", repo)
	}
	if repo.HTMLURL != supported.Install.Value {
		t.Fatalf("expected native install ref %q, got %q", supported.Install.Value, repo.HTMLURL)
	}
	if !repo.Supported || repo.UnsupportedReason != "" || repo.PageURL != supported.PageURL {
		t.Fatalf("installability metadata not preserved: %#v", repo)
	}

	unsupported := supported
	unsupported.Name = "archive-only"
	unsupported.Install = InstallRef{Kind: InstallRefUnsupported}
	unsupported.Supported = false
	unsupported.UnsupportedReason = "skills.sh artifact type is not natively installable yet"

	repo = candidateToRepository(unsupported)
	if repo.HTMLURL != "" {
		t.Fatalf("unsupported skills.sh repository should not expose install ref, got %q", repo.HTMLURL)
	}
	if repo.Supported || repo.UnsupportedReason != unsupported.UnsupportedReason || repo.PageURL != unsupported.PageURL {
		t.Fatalf("unsupported metadata not preserved: %#v", repo)
	}
}
