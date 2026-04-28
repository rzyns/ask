package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/yeasy/ask/internal/config"
	"github.com/yeasy/ask/internal/github"
)

// SearchSkills dispatches search for a configured repository source.
func SearchSkills(ctx context.Context, repo config.Repo, keyword string) ([]github.Repository, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	switch repo.Type {
	case "topic":
		return github.SearchTopic(repo.URL, keyword)
	case "dir":
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

		repos, err := github.SearchDir(owner, repoName, path)
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
	case "registry":
		return FetchSkillsFromRegistry(repo.URL, keyword)
	case "skillhub":
		return FetchSkillsFromSkillHub(keyword, "")
	default:
		return nil, fmt.Errorf("unknown repository type: %s", repo.Type)
	}
}
