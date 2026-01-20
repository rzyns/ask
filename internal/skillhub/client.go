package skillhub

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// SearchURL is the endpoint for quick search
const SearchURL = "https://www.skillhub.club/api/search/quick"

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
			Timeout: 10 * time.Second,
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
	defer resp.Body.Close()

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
	skillURL := fmt.Sprintf("https://www.skillhub.club/skills/%s", slug)
	resp, err := c.HTTPClient.Get(skillURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch skill page: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Look for GitHub link in the page content
	// Use regex or string searching.
	// We confirmed with curl that it appears as href="https://github.com/..."
	// A simple regex:
	re := regexp.MustCompile(`href="(https://github\.com/[^"]+)"`)
	matches := re.FindStringSubmatch(string(body))

	if len(matches) > 1 {
		rawURL := matches[1]
		// clean fragments (e.g. #plugins...)
		if idx := strings.Index(rawURL, "#"); idx != -1 {
			rawURL = rawURL[:idx]
		}
		return rawURL, nil
	}

	return "", fmt.Errorf("GitHub URL not found for skill: %s", slug)
}
