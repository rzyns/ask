package github

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ParseBrowserURL parses a GitHub browser URL and extracts components
// Input: https://github.com/owner/repo/tree/branch/path/to/skill
// Returns: repoURL, branch, subDir, skillName, ok
func ParseBrowserURL(url string) (repoURL, branch, subDir, skillName string, ok bool) {
	// Remove trailing slashes
	url = strings.TrimSuffix(url, "/")

	// Check if it contains /tree/ (GitHub browser URL format)
	if !strings.Contains(url, "/tree/") {
		return "", "", "", "", false
	}

	// Pattern: https://github.com/owner/repo/tree/branch/path
	parts := strings.SplitN(url, "/tree/", 2)
	if len(parts) != 2 {
		return "", "", "", "", false
	}

	repoURL = parts[0] + ".git"

	// Split branch and path
	branchAndPath := parts[1]
	pathParts := strings.SplitN(branchAndPath, "/", 2)
	branch = pathParts[0]

	if len(pathParts) > 1 {
		subDir = pathParts[1]
		// Skill name is the last component of the path
		skillName = filepath.Base(subDir)
	} else {
		// No subdir, use repo name from URL
		urlParts := strings.Split(parts[0], "/")
		skillName = urlParts[len(urlParts)-1]
	}

	return repoURL, branch, subDir, skillName, true
}

// ParseRepoURL parses a GitHub repository URL to extract owner and repo name
// Supports formats:
// - owner/repo
// - https://github.com/owner/repo
// - https://github.com/owner/repo.git
// - git@github.com:owner/repo.git
func ParseRepoURL(url string) (owner, repo string, err error) {
	url = strings.TrimSpace(url)
	url = strings.TrimSuffix(url, "/")
	url = strings.TrimSuffix(url, ".git")

	// Handle git@github.com:owner/repo
	if strings.HasPrefix(url, "git@") {
		parts := strings.Split(url, ":")
		if len(parts) != 2 {
			return "", "", fmt.Errorf("invalid git url: %s", url)
		}
		url = parts[1]
	}

	// Handle https://github.com/owner/repo
	if strings.HasPrefix(url, "https://") || strings.HasPrefix(url, "http://") {
		// Just strip protocol and domain
		parts := strings.Split(url, "github.com/")
		if len(parts) == 2 {
			url = parts[1]
		} else {
			// If we couldn't strip github.com from an http(s) URL, it's not a valid GitHub repo URL for us
			return "", "", fmt.Errorf("invalid repo URL (must be github.com): %s", url)
		}
	}

	// Split owner/repo
	parts := strings.Split(url, "/")
	if len(parts) >= 2 {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("invalid repo format (expected owner/repo): %s", url)
}
