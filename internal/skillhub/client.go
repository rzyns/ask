// Package skillhub provides an interface to search and interact with skill registries.
package skillhub

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/yeasy/ask/internal/config"
)

// SearchURL is the endpoint for quick search
const SearchURL = "https://www.skillhub.club/api/search/quick"

// Pre-compiled regexes for Resolve()
var (
	reGitHubHref    = regexp.MustCompile(`href="(https://github\.com/[^"]+)"`)
	reRepoURLJSON   = regexp.MustCompile(`"repoUrl":"(https://github\.com/[^"]+)"`)
	reRepoURLEscape = regexp.MustCompile(`\\?"repoUrl\\?":\\?"(https://github\.com/[^"\\]+)\\?"`)
)

// Skill represents a skill from SkillHub search
type Skill struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Slug        string   `json:"slug"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
	Author      string   `json:"author"`
	Stars       int      `json:"github_stars"`
	Tags        []string `json:"tags"`
}

type searchResponse struct {
	Skills []Skill `json:"skills"`
}

// Client handles interaction with SkillHub
type Client struct {
	HTTPClient *http.Client
}

// NewClient creates a new SkillHub client
func NewClient() *Client {
	return &Client{
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Search searches for skills on SkillHub
func (c *Client) Search(query string) ([]Skill, error) {
	// If query is empty, use a generic term or handle differently,
	// but the API seems to require a query param or returns empty.
	// For "list all", perhaps we need the catalog API, but that requires auth.
	// Let's rely on search for now.
	if query == "" {
		query = "agent" // default search term if none provided?
	}

	if config.IsOffline() {
		return nil, fmt.Errorf("search is not available in offline mode")
	}

	u, err := url.Parse(SearchURL)
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("q", query)
	q.Set("limit", "50") // Fetch up to 50
	u.RawQuery = q.Encode()

	resp, err := c.HTTPClient.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("SkillHub API returned status: %d", resp.StatusCode)
	}

	var result searchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Skills, nil
}

// Resolve fetches the GitHub URL for a given skill slug
func (c *Client) Resolve(slug string) (string, error) {
	if config.IsOffline() {
		return "", fmt.Errorf("skill resolution is not available in offline mode")
	}
	skillURL := fmt.Sprintf("https://www.skillhub.club/skills/%s", url.PathEscape(slug))
	resp, err := c.HTTPClient.Get(skillURL)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch skill page: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 10*1024*1024)) // 10MB limit
	if err != nil {
		return "", err
	}

	// Look for GitHub link in the page content
	bodyStr := string(body)
	matches := reGitHubHref.FindStringSubmatch(bodyStr)

	if len(matches) > 1 {
		rawURL := matches[1]
		// clean fragments (e.g. #plugins...)
		if idx := strings.Index(rawURL, "#"); idx != -1 {
			rawURL = rawURL[:idx]
		}
		return rawURL, nil
	}

	// Fallback: try to find repoUrl in Next.js hydration data or JSON
	matchesJSON := reRepoURLJSON.FindStringSubmatch(bodyStr)
	if len(matchesJSON) > 1 {
		rawURL := matchesJSON[1]
		return strings.ReplaceAll(rawURL, `\/`, "/"), nil
	}

	// Try one more pattern for escaped JSON
	matchesEscaped := reRepoURLEscape.FindStringSubmatch(bodyStr)
	if len(matchesEscaped) > 1 {
		return matchesEscaped[1], nil
	}

	return "", fmt.Errorf("GitHub URL not found for skill: %s", slug)
}
