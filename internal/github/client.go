package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/yeasy/ask/internal/cache"
)

const (
	// Default topic to search for agent skills
	SkillTopic = "agent-skill"
	APIURL     = "https://api.github.com/search/repositories"
)

// Global cache instance
var searchCache *cache.Cache

// OfflineMode controls whether to skip network requests
var OfflineMode = false

// SetOffline sets the offline mode
func SetOffline(offline bool) {
	OfflineMode = offline
}

func init() {
	// Initialize cache with default settings
	var err error
	searchCache, err = cache.New("", cache.DefaultTTL)
	if err != nil {
		// Cache is optional, continue without it
		searchCache = nil
	}
}

type SearchResult struct {
	TotalCount int          `json:"total_count"`
	Items      []Repository `json:"items"`
}

type Repository struct {
	Name            string    `json:"name"`
	FullName        string    `json:"full_name"`
	Description     string    `json:"description"`
	HTMLURL         string    `json:"html_url"`
	StargazersCount int       `json:"stargazers_count"`
	CloneURL        string    `json:"clone_url"`
	UpdatedAt       time.Time `json:"updated_at"`
	Source          string    `json:"-"` // Source name (e.g., "community", "anthropics")
	Owner           struct {
		Login string `json:"login"`
	} `json:"owner"`
}

// SearchTopic searches GitHub for repositories with a specific topic and keyword
func SearchTopic(topic, keyword string) ([]Repository, error) {
	cacheKey := fmt.Sprintf("topic:%s:%s", topic, keyword)

	// Try cache first
	// In offline mode, we MUST find it in cache or return error
	if searchCache != nil {
		var cached []Repository
		if searchCache.Get(cacheKey, &cached) {
			return cached, nil
		}
	}

	if OfflineMode {
		return nil, fmt.Errorf("offline mode: data not found in cache")
	}

	// Construct query: topic:<topic> <keyword>
	q := fmt.Sprintf("topic:%s %s", topic, keyword)

	params := url.Values{}
	params.Add("q", q)
	params.Add("sort", "stars")
	params.Add("order", "desc")

	req, err := http.NewRequest("GET", APIURL+"?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "ask-cli")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status: %d", resp.StatusCode)
	}

	var result SearchResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	// Cache the result
	if searchCache != nil {
		_ = searchCache.Set(cacheKey, result.Items)
	}

	return result.Items, nil
}

// Content represents a file or directory in a GitHub repository
type Content struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	HTMLURL string `json:"html_url"`
}

// SearchDir searches a specific directory in a GitHub repository for subdirectories (skills)
func SearchDir(owner, repo, path string) ([]Repository, error) {
	cacheKey := fmt.Sprintf("dir:%s/%s/%s", owner, repo, path)

	// Try cache first
	if searchCache != nil {
		var cached []Repository
		if searchCache.Get(cacheKey, &cached) {
			return cached, nil
		}
	}

	if OfflineMode {
		return nil, fmt.Errorf("offline mode: data not found in cache")
	}

	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", owner, repo, path)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "ask-cli")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status: %d", resp.StatusCode)
	}

	var contents []Content
	if err := json.NewDecoder(resp.Body).Decode(&contents); err != nil {
		return nil, err
	}

	// Fetch repo details for stars
	repoDetails, err := fetchRepoDetails(owner, repo)
	stars := 0
	if err == nil {
		stars = repoDetails.StargazersCount
	}

	var skills []Repository
	for _, item := range contents {
		if item.Type == "dir" {
			skills = append(skills, Repository{
				Name:            item.Name,
				FullName:        fmt.Sprintf("%s/%s/%s/%s", owner, repo, path, item.Name),
				Description:     "Skill from " + owner + "/" + repo,
				HTMLURL:         item.HTMLURL,
				StargazersCount: stars,
				CloneURL:        repoDetails.CloneURL,
			})
		}
	}

	// Cache the result
	if searchCache != nil {
		_ = searchCache.Set(cacheKey, skills)
	}

	return skills, nil
}

func fetchRepoDetails(owner, repo string) (*Repository, error) {
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s", owner, repo)
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "ask-cli")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status: %d", resp.StatusCode)
	}

	var repoInfo Repository
	if err := json.NewDecoder(resp.Body).Decode(&repoInfo); err != nil {
		return nil, err
	}
	return &repoInfo, nil
}
