package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ReposCache manages local git repository cache for skill discovery
type ReposCache struct {
	baseDir string
}

// RepoInfo represents cached repository metadata
type RepoInfo struct {
	Name      string    `json:"name"`
	URL       string    `json:"url"`
	LocalPath string    `json:"local_path"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SkillEntry represents a skill found in local cache
type SkillEntry struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Path        string `json:"path"`
	RepoName    string `json:"repo_name"`
}

// NewReposCache creates a new repos cache instance
func NewReposCache() (*ReposCache, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home dir: %w", err)
	}
	baseDir := filepath.Join(homeDir, ".ask", "repos")

	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create repos cache dir: %w", err)
	}

	return &ReposCache{baseDir: baseDir}, nil
}

// GetReposCacheDir returns the repos cache directory path
func GetReposCacheDir() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".ask", "repos")
}

// HasRepo checks if a repository is cached locally
func (c *ReposCache) HasRepo(repoName string) bool {
	repoPath := filepath.Join(c.baseDir, sanitizeRepoName(repoName))
	_, err := os.Stat(repoPath)
	return err == nil
}

// CloneOrPull clones a repo if not exists, or pulls if exists
func (c *ReposCache) CloneOrPull(repoURL, repoName string) error {
	repoPath := filepath.Join(c.baseDir, sanitizeRepoName(repoName))

	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		// Clone with depth=1 for speed
		fmt.Printf("  Cloning %s...\n", repoName)
		cmd := exec.Command("git", "clone", "--depth=1", repoURL, repoPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	// Pull latest
	fmt.Printf("  Updating %s...\n", repoName)
	cmd := exec.Command("git", "-C", repoPath, "pull", "--ff-only")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// ListSkills lists all skills in a cached repo
func (c *ReposCache) ListSkills(repoName string) ([]SkillEntry, error) {
	repoPath := filepath.Join(c.baseDir, sanitizeRepoName(repoName))
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("repo %s not cached", repoName)
	}

	var skills []SkillEntry

	// Walk the repo looking for SKILL.md files
	err := filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		// Skip .git directory
		if info.IsDir() && info.Name() == ".git" {
			return filepath.SkipDir
		}
		// Look for SKILL.md
		if !info.IsDir() && strings.ToUpper(info.Name()) == "SKILL.MD" {
			skillDir := filepath.Dir(path)
			skillName := filepath.Base(skillDir)

			// Try to extract description from SKILL.md
			description := extractDescription(path)

			skills = append(skills, SkillEntry{
				Name:        skillName,
				Description: description,
				Path:        skillDir,
				RepoName:    repoName,
			})
		}
		return nil
	})

	return skills, err
}

// SearchSkills searches for skills matching keyword in all cached repos
func (c *ReposCache) SearchSkills(keyword string) ([]SkillEntry, error) {
	keyword = strings.ToLower(keyword)
	var results []SkillEntry

	// List all cached repos
	entries, err := os.ReadDir(c.baseDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		skills, err := c.ListSkills(entry.Name())
		if err != nil {
			continue
		}

		for _, skill := range skills {
			if strings.Contains(strings.ToLower(skill.Name), keyword) ||
				strings.Contains(strings.ToLower(skill.Description), keyword) {
				results = append(results, skill)
			}
		}
	}

	return results, nil
}

// GetCachedRepos returns list of cached repo names
func (c *ReposCache) GetCachedRepos() []string {
	var repos []string
	entries, err := os.ReadDir(c.baseDir)
	if err != nil {
		return repos
	}
	for _, entry := range entries {
		if entry.IsDir() && entry.Name() != ".git" {
			repos = append(repos, entry.Name())
		}
	}
	return repos
}

// sanitizeRepoName converts owner/repo to owner-repo for filesystem
func sanitizeRepoName(name string) string {
	return strings.ReplaceAll(name, "/", "-")
}

// extractDescription reads SKILL.md and extracts description from frontmatter
func extractDescription(skillMDPath string) string {
	data, err := os.ReadFile(skillMDPath)
	if err != nil {
		return ""
	}
	content := string(data)

	// Check for YAML frontmatter
	if !strings.HasPrefix(content, "---") {
		return ""
	}

	// Find end of frontmatter
	endIdx := strings.Index(content[3:], "---")
	if endIdx == -1 {
		return ""
	}

	frontmatter := content[3 : endIdx+3]
	lines := strings.Split(frontmatter, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "description:") {
			desc := strings.TrimPrefix(line, "description:")
			desc = strings.TrimSpace(desc)
			desc = strings.Trim(desc, "\"'")
			return desc
		}
	}

	return ""
}

// SaveIndex saves the current repo index to disk
func (c *ReposCache) SaveIndex() error {
	indexPath := filepath.Join(c.baseDir, "index.json")
	repos := c.GetCachedRepos()

	var repoInfos []RepoInfo
	for _, repo := range repos {
		repoPath := filepath.Join(c.baseDir, repo)
		info, _ := os.Stat(repoPath)
		repoInfos = append(repoInfos, RepoInfo{
			Name:      repo,
			LocalPath: repoPath,
			UpdatedAt: info.ModTime(),
		})
	}

	data, err := json.MarshalIndent(repoInfos, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(indexPath, data, 0644)
}
