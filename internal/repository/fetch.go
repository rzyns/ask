package repository

import (
	"fmt"
	"strings"

	"github.com/yeasy/ask/internal/config"
	"github.com/yeasy/ask/internal/github"
)

// FetchSkills returns a list of skills available in the given repository
func FetchSkills(repo config.Repo) ([]github.Repository, error) {
	switch repo.Type {
	case "topic":
		return github.SearchTopic(repo.URL, "")
	case "dir":
		parts := strings.Split(repo.URL, "/")
		if len(parts) >= 2 {
			owner := parts[0]
			name := parts[1]
			path := ""
			if len(parts) >= 3 {
				path = strings.Join(parts[2:], "/")
			}
			return github.SearchDir(owner, name, path)
		}
		return nil, fmt.Errorf("invalid repository URL format: %s", repo.URL)
	default:
		return nil, fmt.Errorf("unknown repository type: %s", repo.Type)
	}
}
