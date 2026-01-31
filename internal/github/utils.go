package github

import (
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
