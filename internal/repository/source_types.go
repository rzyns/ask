package repository

import (
	"github.com/yeasy/ask/internal/config"
	"github.com/yeasy/ask/internal/github"
)

type InstallRefKind string

const (
	InstallRefGitHubPath  InstallRefKind = "github_path"
	InstallRefSlug        InstallRefKind = "slug"
	InstallRefUnsupported InstallRefKind = "unsupported"
)

type InstallRef struct {
	Kind  InstallRefKind
	Value string
}

type SkillCandidate struct {
	Name              string
	FullName          string
	Description       string
	Source            string
	SourceIdentifier  string
	UpdateStrategy    string
	Install           InstallRef
	Stars             int
	PageURL           string
	Supported         bool
	UnsupportedReason string
}

func repositoryToCandidate(repo github.Repository) SkillCandidate {
	kind := InstallRefGitHubPath
	if repo.Source == config.RepoTypeSkillHub {
		kind = InstallRefSlug
	}
	installValue := repo.HTMLURL
	if repo.InstallRef != "" {
		installValue = repo.InstallRef
	}
	return SkillCandidate{
		Name:             repo.Name,
		FullName:         repo.FullName,
		Description:      repo.Description,
		Source:           repo.Source,
		SourceIdentifier: repo.SourceIdentifier,
		UpdateStrategy:   repo.UpdateStrategy,
		Install: InstallRef{
			Kind:  kind,
			Value: installValue,
		},
		Stars:             repo.StargazersCount,
		PageURL:           repo.PageURL,
		Supported:         repo.Supported,
		UnsupportedReason: repo.UnsupportedReason,
	}
}

func repositoriesToCandidates(repos []github.Repository) []SkillCandidate {
	if repos == nil {
		return nil
	}
	candidates := make([]SkillCandidate, 0, len(repos))
	for _, repo := range repos {
		candidates = append(candidates, repositoryToCandidate(repo))
	}
	return candidates
}

func candidateToRepository(candidate SkillCandidate) github.Repository {
	return github.Repository{
		Name:              candidate.Name,
		FullName:          candidate.FullName,
		Description:       candidate.Description,
		HTMLURL:           candidate.Install.Value,
		InstallRef:        candidate.Install.Value,
		StargazersCount:   candidate.Stars,
		Source:            candidate.Source,
		SourceIdentifier:  candidate.SourceIdentifier,
		UpdateStrategy:    candidate.UpdateStrategy,
		PageURL:           candidate.PageURL,
		Supported:         candidate.Supported,
		UnsupportedReason: candidate.UnsupportedReason,
	}
}

func candidatesToRepositories(candidates []SkillCandidate) []github.Repository {
	if candidates == nil {
		return nil
	}
	repos := make([]github.Repository, 0, len(candidates))
	for _, candidate := range candidates {
		repos = append(repos, candidateToRepository(candidate))
	}
	return repos
}
