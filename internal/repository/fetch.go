package repository

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/yeasy/ask/internal/config"
	"github.com/yeasy/ask/internal/git"
	"github.com/yeasy/ask/internal/github"
	"github.com/yeasy/ask/internal/skill"
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
}

// FetchSkillsViaGit clones a repo and discovers skills locally (no API needed)
func FetchSkillsViaGit(repo config.Repo) ([]github.Repository, error) {
	if repo.Type != "dir" {
		return nil, fmt.Errorf("git fetch only supports 'dir' type repos")
	}

	parts := strings.Split(repo.URL, "/")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid repository URL format: %s", repo.URL)
	}

	owner := parts[0]
	repoName := parts[1]
	subPath := ""
	if len(parts) > 2 {
		subPath = strings.Join(parts[2:], "/")
	}

	cloneURL := fmt.Sprintf("https://github.com/%s/%s.git", owner, repoName)

	// Create temp dir
	tempDir, err := os.MkdirTemp("", "ask-discovery-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Clone repo
	// Using generic Clone (depth 1)
	if err := git.Clone(cloneURL, tempDir); err != nil {
		return nil, fmt.Errorf("failed to clone repo: %w", err)
	}

	// Scan for skills
	searchPath := tempDir
	if subPath != "" {
		searchPath = filepath.Join(tempDir, subPath)
	}

	entries, err := os.ReadDir(searchPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read skills directory: %w", err)
	}

	var skills []github.Repository

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		skillDir := filepath.Join(searchPath, entry.Name())
		if skill.FindSkillMD(skillDir) {
			// Found a skill, extract metadata
			var desc string
			if meta, err := skill.ParseSkillMD(skillDir); err == nil && meta != nil {
				desc = meta.Description
			}

			// Construct installation argument (owner/repo/path/to/skill)
			skillRelPath := subPath
			if skillRelPath != "" {
				skillRelPath = skillRelPath + "/" + entry.Name()
			} else {
				skillRelPath = entry.Name()
			}
			installArg := fmt.Sprintf("%s/%s/%s", owner, repoName, skillRelPath)

			skills = append(skills, github.Repository{
				Name:        entry.Name(),
				Description: desc,
				HTMLURL:     installArg, // Passing the install arg string
			})
		}
	}

	return skills, nil
}
