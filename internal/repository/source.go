package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/yeasy/ask/internal/config"
	"github.com/yeasy/ask/internal/github"
)

type repositorySource interface {
	Search(ctx context.Context, repo config.Repo, keyword string) ([]github.Repository, error)
	Fetch(repo config.Repo) ([]github.Repository, error)
}

type sourceFuncSet struct {
	search func(ctx context.Context, repo config.Repo, keyword string) ([]github.Repository, error)
	fetch  func(repo config.Repo) ([]github.Repository, error)
}

func (s sourceFuncSet) Search(ctx context.Context, repo config.Repo, keyword string) ([]github.Repository, error) {
	return s.search(ctx, repo, keyword)
}

func (s sourceFuncSet) Fetch(repo config.Repo) ([]github.Repository, error) {
	return s.fetch(repo)
}

var (
	searchTopicFunc             = github.SearchTopic
	searchDirFunc               = github.SearchDir
	fetchSkillsViaGitFunc       = FetchSkillsViaGit
	fetchSkillsFromRegistryFunc = FetchSkillsFromRegistry
	fetchSkillsFromSkillHubFunc = FetchSkillsFromSkillHub
)

func sourceForRepo(repo config.Repo) (repositorySource, error) {
	sources := map[string]repositorySource{
		config.RepoTypeTopic: sourceFuncSet{
			search: searchTopicSource,
			fetch:  fetchTopicSource,
		},
		config.RepoTypeDir: sourceFuncSet{
			search: searchDirSource,
			fetch:  fetchDirSource,
		},
		config.RepoTypeRegistry: sourceFuncSet{
			search: searchRegistrySource,
			fetch:  fetchRegistrySource,
		},
		config.RepoTypeSkillHub: sourceFuncSet{
			search: searchSkillHubSource,
			fetch:  fetchSkillHubSource,
		},
	}

	source, ok := sources[repo.Type]
	if !ok {
		return nil, fmt.Errorf("unknown repository type: %s", repo.Type)
	}
	return source, nil
}

// SearchSkills dispatches search for a configured repository source.
func SearchSkills(ctx context.Context, repo config.Repo, keyword string) ([]github.Repository, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	source, err := sourceForRepo(repo)
	if err != nil {
		return nil, err
	}
	return source.Search(ctx, repo, keyword)
}

func searchTopicSource(_ context.Context, repo config.Repo, keyword string) ([]github.Repository, error) {
	return searchTopicFunc(repo.URL, keyword)
}

func fetchTopicSource(repo config.Repo) ([]github.Repository, error) {
	return searchTopicFunc(repo.URL, "")
}

func searchDirSource(_ context.Context, repo config.Repo, keyword string) ([]github.Repository, error) {
	parts := strings.Split(repo.URL, "/")
	if len(parts) < 2 {
		return nil, nil
	}

	owner := parts[0]
	repoName := parts[1]
	path := ""
	if len(parts) > 2 {
		path = strings.Join(parts[2:], "/")
	}

	repos, err := searchDirFunc(owner, repoName, path)
	if err != nil || keyword == "" {
		return repos, err
	}

	// Filter client-side by keyword.
	var filtered []github.Repository
	lowerKeyword := strings.ToLower(keyword)
	for _, rp := range repos {
		if strings.Contains(strings.ToLower(rp.Name), lowerKeyword) {
			filtered = append(filtered, rp)
		}
	}
	return filtered, nil
}

func fetchDirSource(repo config.Repo) ([]github.Repository, error) {
	// Try git-based discovery first (recursive and more reliable for deep structures)
	skills, err := fetchSkillsViaGitFunc(repo)
	if err == nil && len(skills) > 0 {
		return skills, nil
	}

	// Fallback to API if git failed (e.g. no git installed) or found nothing
	parts := strings.Split(repo.URL, "/")
	if len(parts) >= 2 {
		owner := parts[0]
		name := parts[1]
		path := ""
		if len(parts) >= 3 {
			path = strings.Join(parts[2:], "/")
		}
		return searchDirFunc(owner, name, path)
	}
	return nil, fmt.Errorf("invalid repository URL format: %s", repo.URL)
}

func searchRegistrySource(_ context.Context, repo config.Repo, keyword string) ([]github.Repository, error) {
	return fetchSkillsFromRegistryFunc(repo.URL, keyword)
}

func fetchRegistrySource(repo config.Repo) ([]github.Repository, error) {
	return fetchSkillsFromRegistryFunc(repo.URL, "")
}

func searchSkillHubSource(_ context.Context, _ config.Repo, keyword string) ([]github.Repository, error) {
	return fetchSkillsFromSkillHubFunc(keyword, "")
}

func fetchSkillHubSource(config.Repo) ([]github.Repository, error) {
	return fetchSkillsFromSkillHubFunc("", "")
}
