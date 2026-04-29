package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/yeasy/ask/internal/config"
	"github.com/yeasy/ask/internal/github"
	"github.com/yeasy/ask/internal/hermes"
)

const bundledHermesSkillsErrorFormat = "Repository %s contains bundled Hermes skills. ASK does not manage bundled Hermes skills. Use hermes-index for user-installable Hermes skills."

type repositorySource interface {
	Search(ctx context.Context, repo config.Repo, keyword string) ([]SkillCandidate, error)
	Fetch(repo config.Repo) ([]SkillCandidate, error)
}

type sourceFuncSet struct {
	search func(ctx context.Context, repo config.Repo, keyword string) ([]SkillCandidate, error)
	fetch  func(repo config.Repo) ([]SkillCandidate, error)
}

func (s sourceFuncSet) Search(ctx context.Context, repo config.Repo, keyword string) ([]SkillCandidate, error) {
	return s.search(ctx, repo, keyword)
}

func (s sourceFuncSet) Fetch(repo config.Repo) ([]SkillCandidate, error) {
	return s.fetch(repo)
}

var (
	searchTopicFunc             = github.SearchTopic
	searchDirFunc               = github.SearchDir
	fetchSkillsViaGitFunc       = FetchSkillsViaGit
	fetchSkillsFromRegistryFunc = FetchSkillsFromRegistry
	fetchSkillsFromSkillHubFunc = FetchSkillsFromSkillHub
	searchSkillsSHFunc          = searchSkillsSHSource
	fetchSkillsSHFunc           = fetchSkillsSHSource
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
		config.RepoTypeSkillsSH: sourceFuncSet{
			search: searchSkillsSHFunc,
			fetch:  fetchSkillsSHFunc,
		},
		config.RepoTypeHermes: sourceFuncSet{
			search: searchHermesSource,
			fetch:  fetchHermesSource,
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
	candidates, err := source.Search(ctx, repo, keyword)
	if err != nil {
		return nil, err
	}
	return candidatesToRepositories(candidates), nil
}

func searchTopicSource(_ context.Context, repo config.Repo, keyword string) ([]SkillCandidate, error) {
	repos, err := searchTopicFunc(repo.URL, keyword)
	return repositoriesToCandidates(repos), err
}

func fetchTopicSource(repo config.Repo) ([]SkillCandidate, error) {
	repos, err := searchTopicFunc(repo.URL, "")
	return repositoriesToCandidates(repos), err
}

func searchDirSource(_ context.Context, repo config.Repo, keyword string) ([]SkillCandidate, error) {
	if err := rejectBundledHermesDirSource(repo); err != nil {
		return nil, err
	}
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
		return repositoriesToCandidates(repos), err
	}

	// Filter client-side by keyword.
	var filtered []github.Repository
	lowerKeyword := strings.ToLower(keyword)
	for _, rp := range repos {
		if strings.Contains(strings.ToLower(rp.Name), lowerKeyword) {
			filtered = append(filtered, rp)
		}
	}
	return repositoriesToCandidates(filtered), nil
}

func fetchDirSource(repo config.Repo) ([]SkillCandidate, error) {
	if err := rejectBundledHermesDirSource(repo); err != nil {
		return nil, err
	}
	// Try git-based discovery first (recursive and more reliable for deep structures)
	skills, err := fetchSkillsViaGitFunc(repo)
	if err == nil && len(skills) > 0 {
		return repositoriesToCandidates(skills), nil
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
		repos, err := searchDirFunc(owner, name, path)
		return repositoriesToCandidates(repos), err
	}
	return nil, fmt.Errorf("invalid repository URL format: %s", repo.URL)
}

func rejectBundledHermesDirSource(repo config.Repo) error {
	if repo.Type == config.RepoTypeDir && hermes.ClassifyHermesSource(repo.URL).Kind == hermes.HermesSourceBundled {
		return fmt.Errorf(bundledHermesSkillsErrorFormat, repo.URL)
	}
	return nil
}

func searchRegistrySource(_ context.Context, repo config.Repo, keyword string) ([]SkillCandidate, error) {
	repos, err := fetchSkillsFromRegistryFunc(repo.URL, keyword)
	return repositoriesToCandidates(repos), err
}

func fetchRegistrySource(repo config.Repo) ([]SkillCandidate, error) {
	repos, err := fetchSkillsFromRegistryFunc(repo.URL, "")
	return repositoriesToCandidates(repos), err
}

func searchSkillHubSource(_ context.Context, _ config.Repo, keyword string) ([]SkillCandidate, error) {
	repos, err := fetchSkillsFromSkillHubFunc(keyword, "")
	return repositoriesToCandidates(repos), err
}

func fetchSkillHubSource(config.Repo) ([]SkillCandidate, error) {
	repos, err := fetchSkillsFromSkillHubFunc("", "")
	return repositoriesToCandidates(repos), err
}
