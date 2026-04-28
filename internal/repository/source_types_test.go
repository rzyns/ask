package repository

import (
	"context"
	"reflect"
	"testing"

	"github.com/yeasy/ask/internal/config"
	"github.com/yeasy/ask/internal/github"
)

func TestRepositoryCandidateAdaptersPreserveGitHubPathResult(t *testing.T) {
	repo := github.Repository{
		Name:            "browser-use",
		FullName:        "anthropics/skills",
		Description:     "desc",
		HTMLURL:         "anthropics/skills/skills/browser-use",
		StargazersCount: 42,
		Source:          "anthropics",
	}

	candidate := repositoryToCandidate(repo)
	if candidate.Name != repo.Name || candidate.FullName != repo.FullName || candidate.Description != repo.Description || candidate.Source != repo.Source || candidate.Stars != repo.StargazersCount {
		t.Fatalf("candidate did not preserve repository fields: %#v", candidate)
	}
	if candidate.Install.Kind != InstallRefGitHubPath || candidate.Install.Value != repo.HTMLURL {
		t.Fatalf("unexpected install ref: %#v", candidate.Install)
	}

	got := candidatesToRepositories([]SkillCandidate{candidate})
	if !reflect.DeepEqual(got, []github.Repository{repo}) {
		t.Fatalf("round trip changed repository:\n got: %#v\nwant: %#v", got, []github.Repository{repo})
	}
}

func TestRepositoryCandidateAdaptersPreserveSkillHubSlugResult(t *testing.T) {
	repo := github.Repository{
		Name:            "foo",
		FullName:        "foo-slug",
		Description:     "desc",
		HTMLURL:         "foo-slug",
		StargazersCount: 7,
		Source:          config.RepoTypeSkillHub,
	}

	candidate := repositoryToCandidate(repo)
	if candidate.Install.Kind != InstallRefSlug || candidate.Install.Value != repo.HTMLURL {
		t.Fatalf("unexpected install ref: %#v", candidate.Install)
	}

	got := candidatesToRepositories([]SkillCandidate{candidate})
	if !reflect.DeepEqual(got, []github.Repository{repo}) {
		t.Fatalf("round trip changed repository:\n got: %#v\nwant: %#v", got, []github.Repository{repo})
	}
}

func TestSourceDispatcherUsesCandidatesInternallyWhilePublicSearchReturnsRepositories(t *testing.T) {
	var source repositorySource = sourceFuncSet{
		search: func(context.Context, config.Repo, string) ([]SkillCandidate, error) {
			return []SkillCandidate{{
				Name:        "candidate-skill",
				FullName:    "owner/repo",
				Description: "desc",
				Source:      config.RepoTypeTopic,
				Stars:       3,
				Install:     InstallRef{Kind: InstallRefGitHubPath, Value: "owner/repo/skills/candidate-skill"},
			}}, nil
		},
		fetch: func(config.Repo) ([]SkillCandidate, error) {
			return []SkillCandidate{{Name: "fetch-candidate", Install: InstallRef{Kind: InstallRefGitHubPath, Value: "owner/repo/fetch-candidate"}}}, nil
		},
	}

	candidates, err := source.Search(context.Background(), config.Repo{Type: config.RepoTypeTopic}, "candidate")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(candidates) != 1 || candidates[0].Name != "candidate-skill" {
		t.Fatalf("unexpected candidates: %#v", candidates)
	}

	repos := candidatesToRepositories(candidates)
	want := []github.Repository{{
		Name:            "candidate-skill",
		FullName:        "owner/repo",
		Description:     "desc",
		HTMLURL:         "owner/repo/skills/candidate-skill",
		StargazersCount: 3,
		Source:          config.RepoTypeTopic,
	}}
	if !reflect.DeepEqual(repos, want) {
		t.Fatalf("public adapter changed repositories:\n got: %#v\nwant: %#v", repos, want)
	}
}
