package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/yeasy/ask/internal/config"
)

const (
	defaultSkillsSHBaseURL = "https://skills.sh"
	skillsSHSearchLimit    = "10"
	skillsSHFetchPerPage   = "100"
)

var (
	skillsSHHTTPClient                 = &http.Client{Timeout: 30 * time.Second}
	skillsSHMaxBodyBytes               = int64(5 << 20)
	skillsSHGitHubAPIBaseURL           = "https://api.github.com"
	resolveSkillsSHGitHubSkillPathFunc = resolveSkillsSHGitHubSkillPath
)

func searchSkillsSHSource(ctx context.Context, repo config.Repo, keyword string) ([]SkillCandidate, error) {
	query := url.Values{}
	query.Set("q", keyword)
	query.Set("limit", skillsSHSearchLimit)
	body, err := doSkillsSHPublicRequest(ctx, repo, "/api/search", query)
	if err != nil {
		return nil, err
	}
	candidates, err := parseSkillsSHLegacySearchCandidatesTrusted(body)
	if err != nil {
		return nil, fmt.Errorf("parse skills.sh response: %w", err)
	}
	return candidates, nil
}

func fetchSkillsSHSource(repo config.Repo) ([]SkillCandidate, error) {
	return fetchSkillsSHSourceContext(context.Background(), repo)
}

func fetchSkillsSHSourceContext(ctx context.Context, repo config.Repo) ([]SkillCandidate, error) {
	if skillsSHToken(repo) == "" {
		return nil, fmt.Errorf("skills.sh full catalog listing requires SKILLS_SH_API_KEY or repo token; public skills.sh supports search only")
	}
	query := url.Values{}
	query.Set("view", "all-time")
	query.Set("page", "0")
	query.Set("per_page", skillsSHFetchPerPage)
	body, err := doSkillsSHAuthenticatedRequest(ctx, repo, "/api/v1/skills", query)
	if err != nil {
		return nil, err
	}
	candidates, err := parseSkillsSHV1Candidates(body)
	if err != nil {
		return nil, fmt.Errorf("parse skills.sh response: %w", err)
	}
	return candidates, nil
}

func skillsSHToken(repo config.Repo) string {
	if strings.TrimSpace(repo.Token) != "" {
		return strings.TrimSpace(repo.Token)
	}
	return strings.TrimSpace(os.Getenv("SKILLS_SH_API_KEY"))
}

func skillsSHBaseURL(repo config.Repo) (*url.URL, error) {
	raw := strings.TrimSpace(repo.URL)
	if raw == "" {
		raw = defaultSkillsSHBaseURL
	}
	base, err := url.Parse(raw)
	if err != nil || base.Scheme == "" || base.Host == "" {
		return nil, fmt.Errorf("invalid skills.sh base URL")
	}
	return base, nil
}

func doSkillsSHPublicRequest(ctx context.Context, repo config.Repo, path string, query url.Values) ([]byte, error) {
	return doSkillsSHRequest(ctx, repo, path, query, "")
}

func doSkillsSHAuthenticatedRequest(ctx context.Context, repo config.Repo, path string, query url.Values) ([]byte, error) {
	token := skillsSHToken(repo)
	return doSkillsSHRequest(ctx, repo, path, query, token)
}

func doSkillsSHRequest(ctx context.Context, repo config.Repo, path string, query url.Values, token string) ([]byte, error) {
	base, err := skillsSHBaseURL(repo)
	if err != nil {
		return nil, err
	}
	u := *base
	u.Path = strings.TrimRight(base.Path, "/") + path
	u.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "ask-cli")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := skillsSHHTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	limited := io.LimitReader(resp.Body, skillsSHMaxBodyBytes+1)
	body, readErr := io.ReadAll(limited)
	if readErr != nil {
		return nil, fmt.Errorf("read skills.sh response: %w", readErr)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, skillsSHStatusError(resp, body, token)
	}
	if int64(len(body)) > skillsSHMaxBodyBytes {
		return nil, fmt.Errorf("skills.sh response body too large (limit %d bytes)", skillsSHMaxBodyBytes)
	}
	return body, nil
}

func skillsSHStatusError(resp *http.Response, body []byte, token string) error {
	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("skills.sh API key required; set SKILLS_SH_API_KEY or repo token")
	}
	msg := fmt.Sprintf("skills.sh API error: %d", resp.StatusCode)
	if resp.StatusCode == http.StatusTooManyRequests {
		if retryAfter := strings.TrimSpace(resp.Header.Get("Retry-After")); retryAfter != "" {
			msg += "; Retry-After: " + retryAfter
		}
	}
	if resp.StatusCode == http.StatusServiceUnavailable {
		msg += "; retry later or configure backoff before retrying"
	}
	if snippet := redactSkillsSHSecret(strings.TrimSpace(string(body)), token); snippet != "" {
		if len(snippet) > 200 {
			snippet = snippet[:200]
		}
		msg += ": " + snippet
	}
	return fmt.Errorf("%s", msg)
}

func redactSkillsSHSecret(s, token string) string {
	if strings.TrimSpace(token) == "" {
		return s
	}
	return strings.ReplaceAll(s, token, "[REDACTED]")
}

type skillsSHV1Wrapper struct {
	Data []skillsSHV1Skill `json:"data"`
}

type skillsSHV1Skill struct {
	ID          string `json:"id"`
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Source      string `json:"source"`
	Installs    int    `json:"installs"`
	SourceType  string `json:"sourceType"`
	InstallURL  string `json:"installUrl"`
	URL         string `json:"url"`
	IsDuplicate bool   `json:"isDuplicate"`
	Type        string `json:"type"`
}

type skillsSHLegacySearchWrapper struct {
	Skills []skillsSHLegacySkill `json:"skills"`
}

type skillsSHLegacySkill struct {
	ID       string `json:"id"`
	SkillID  string `json:"skillId"`
	Name     string `json:"name"`
	Installs int    `json:"installs"`
	Source   string `json:"source"`
	URL      string `json:"url"`
}

func parseSkillsSHV1Candidates(body []byte) ([]SkillCandidate, error) {
	var wrapper skillsSHV1Wrapper
	if err := json.Unmarshal(body, &wrapper); err != nil {
		return nil, err
	}
	candidates := make([]SkillCandidate, 0, len(wrapper.Data))
	for _, skill := range wrapper.Data {
		if skill.IsDuplicate {
			continue
		}
		candidates = append(candidates, candidateFromSkillsSHV1(skill))
	}
	return candidates, nil
}

func parseSkillsSHLegacySearchCandidates(body []byte) ([]SkillCandidate, error) {
	return parseSkillsSHLegacySearchCandidatesWithTrust(body, false)
}

func parseSkillsSHLegacySearchCandidatesTrusted(body []byte) ([]SkillCandidate, error) {
	return parseSkillsSHLegacySearchCandidatesWithTrust(body, true)
}

func parseSkillsSHLegacySearchCandidatesWithTrust(body []byte, allowBareOwnerRepo bool) ([]SkillCandidate, error) {
	var wrapper skillsSHLegacySearchWrapper
	if err := json.Unmarshal(body, &wrapper); err != nil {
		return nil, err
	}
	candidates := make([]SkillCandidate, 0, len(wrapper.Skills))
	for _, skill := range wrapper.Skills {
		candidates = append(candidates, candidateFromSkillsSHLegacy(skill, allowBareOwnerRepo))
	}
	return candidates, nil
}

func candidateFromSkillsSHV1(skill skillsSHV1Skill) SkillCandidate {
	c := baseSkillsSHCandidate(skill.Name, skill.Description, skill.URL, skill.Installs, firstNonEmpty(skill.ID, skill.Source, skill.Slug))
	if skill.Type == "skill-md" || skill.Type == "archive" || skill.SourceType == "skill-md" || skill.SourceType == "archive" {
		markUnsupported(&c, "skills.sh artifact type is not natively installable yet")
		return c
	}
	if strings.EqualFold(skill.SourceType, "github") {
		if strings.TrimSpace(skill.InstallURL) != "" {
			if ref, ok := resolveSkillsSHGitHubRef(skill.InstallURL); ok {
				markSupported(&c, ref)
				return c
			}
			markUnsupported(&c, "invalid GitHub installUrl for skills.sh entry")
			return c
		}
		if ref, ok := resolveSkillsSHGitHubSource(skill.Source, true); ok {
			markSupported(&c, ref)
			return c
		}
	}
	markUnsupported(&c, "no native ASK resolver for skills.sh entry")
	return c
}

func candidateFromSkillsSHLegacy(skill skillsSHLegacySkill, allowBareOwnerRepo bool) SkillCandidate {
	c := baseSkillsSHCandidate(skill.Name, "", skill.URL, skill.Installs, firstNonEmpty(skill.ID, skill.SkillID, skill.Source))
	if allowBareOwnerRepo {
		if ref, reason := resolveSkillsSHGitHubSkillPathFunc(skill.Source, skill.SkillID, skill.Name); ref != "" {
			markSupported(&c, ref)
			return c
		} else if reason != "" {
			markUnsupported(&c, reason)
			return c
		}
	}
	if ref, ok := resolveSkillsSHGitHubSource(skill.Source, allowBareOwnerRepo); ok {
		markSupported(&c, ref)
		return c
	}
	markUnsupported(&c, "no native ASK resolver for legacy skills.sh entry")
	return c
}

func baseSkillsSHCandidate(name, description, pageURL string, installs int, identifier string) SkillCandidate {
	return SkillCandidate{
		Name:             name,
		FullName:         name,
		Description:      description,
		PageURL:          pageURL,
		Stars:            installs,
		Source:           config.RepoTypeSkillsSH,
		SourceIdentifier: identifier,
		UpdateStrategy:   "skills.sh",
	}
}

func markSupported(c *SkillCandidate, ref string) {
	c.Supported = true
	c.Install = InstallRef{Kind: InstallRefGitHubPath, Value: ref}
}

func markUnsupported(c *SkillCandidate, reason string) {
	c.Supported = false
	c.UnsupportedReason = reason
	c.Install = InstallRef{Kind: InstallRefUnsupported}
}

func resolveSkillsSHGitHubSource(source string, allowBareOwnerRepo bool) (string, bool) {
	source = strings.TrimSpace(source)
	if source == "" {
		return "", false
	}
	if strings.HasPrefix(source, "https://") || strings.HasPrefix(source, "http://") {
		return resolveSkillsSHGitHubRef(source)
	}
	if strings.HasPrefix(source, "github.com/") {
		return resolveSkillsSHGitHubRef("https://" + source)
	}
	if !allowBareOwnerRepo {
		return "", false
	}
	parts := strings.Split(strings.Trim(source, "/"), "/")
	if len(parts) >= 2 && !strings.Contains(parts[0], ".") {
		return resolveSkillsSHGitHubRef("https://github.com/" + strings.Join(parts, "/"))
	}
	return "", false
}

func resolveSkillsSHGitHubRef(raw string) (string, bool) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.Scheme != "https" || u.Hostname() != "github.com" {
		return "", false
	}
	parts := strings.Split(strings.Trim(u.EscapedPath(), "/"), "/")
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		return "", false
	}
	allowed := len(parts) == 2 || (len(parts) >= 5 && parts[2] == "tree" && parts[3] != "")
	if !allowed {
		return "", false
	}
	return fmt.Sprintf("https://github.com/%s", strings.Join(parts, "/")), true
}

type skillsSHGitHubRepo struct {
	DefaultBranch string `json:"default_branch"`
}

type skillsSHGitTree struct {
	Tree []struct {
		Path string `json:"path"`
		Type string `json:"type"`
	} `json:"tree"`
}

func resolveSkillsSHGitHubSkillPath(source, skillID, name string) (string, string) {
	ownerRepo, ok := skillsSHBareOwnerRepo(source)
	if !ok {
		return "", ""
	}
	branch, ok := fetchSkillsSHGitHubDefaultBranch(ownerRepo)
	if !ok {
		return "", "no GitHub skill path found for skills.sh entry"
	}
	paths, ok := fetchSkillsSHGitHubTreePaths(ownerRepo, branch)
	if !ok {
		return "", "no GitHub skill path found for skills.sh entry"
	}
	return resolveSkillsSHGitHubSkillPathFromTreePaths(ownerRepo, skillID, name, paths, branch)
}

func fetchSkillsSHGitHubDefaultBranch(ownerRepo string) (string, bool) {
	req, err := newSkillsSHGitHubAPIRequest("/repos/"+ownerRepo, nil)
	if err != nil {
		return "", false
	}
	resp, err := skillsSHHTTPClient.Do(req)
	if err != nil {
		return "", false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", false
	}
	var repo skillsSHGitHubRepo
	if err := json.NewDecoder(io.LimitReader(resp.Body, skillsSHMaxBodyBytes)).Decode(&repo); err != nil {
		return "", false
	}
	branch := strings.TrimSpace(repo.DefaultBranch)
	return branch, branch != ""
}

func fetchSkillsSHGitHubTreePaths(ownerRepo, branch string) ([]string, bool) {
	query := url.Values{}
	query.Set("recursive", "1")
	req, err := newSkillsSHGitHubAPIRequest("/repos/"+ownerRepo+"/git/trees/"+url.PathEscape(branch), query)
	if err != nil {
		return nil, false
	}
	resp, err := skillsSHHTTPClient.Do(req)
	if err != nil {
		return nil, false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, false
	}
	var tree skillsSHGitTree
	if err := json.NewDecoder(io.LimitReader(resp.Body, skillsSHMaxBodyBytes)).Decode(&tree); err != nil {
		return nil, false
	}
	paths := make([]string, 0, len(tree.Tree))
	for _, item := range tree.Tree {
		paths = append(paths, item.Path)
	}
	return paths, true
}

func newSkillsSHGitHubAPIRequest(path string, query url.Values) (*http.Request, error) {
	base, err := url.Parse(skillsSHGitHubAPIBaseURL)
	if err != nil || base.Scheme == "" || base.Host == "" {
		return nil, fmt.Errorf("invalid GitHub API base URL")
	}
	u := *base
	u.Path = strings.TrimRight(base.Path, "/") + path
	u.RawQuery = query.Encode()
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "ask-cli")
	return req, nil
}

func skillsSHBareOwnerRepo(source string) (string, bool) {
	parts := strings.Split(strings.Trim(strings.TrimSpace(source), "/"), "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" || strings.Contains(parts[0], ".") || strings.Contains(parts[1], ".") {
		return "", false
	}
	return parts[0] + "/" + parts[1], true
}

func resolveSkillsSHGitHubSkillPathFromTreePaths(ownerRepo, skillID, name string, paths []string, branch string) (string, string) {
	wanted := map[string]bool{}
	if strings.TrimSpace(skillID) != "" {
		wanted[strings.TrimSpace(skillID)] = true
	}
	if strings.TrimSpace(name) != "" {
		wanted[strings.TrimSpace(name)] = true
	}
	matches := []string{}
	for _, p := range paths {
		p = strings.Trim(p, "/")
		if !strings.EqualFold(lastPathSegment(p), "SKILL.md") {
			continue
		}
		dir := strings.TrimSuffix(p, "/"+lastPathSegment(p))
		if wanted[lastPathSegment(dir)] {
			matches = append(matches, dir)
		}
	}
	if len(matches) == 0 {
		return "", "no GitHub skill path found for skills.sh entry"
	}
	if len(matches) > 1 {
		return "", "ambiguous GitHub skill path for skills.sh entry"
	}
	return fmt.Sprintf("https://github.com/%s/tree/%s/%s", ownerRepo, branch, matches[0]), ""
}

func lastPathSegment(p string) string {
	parts := strings.Split(strings.Trim(p, "/"), "/")
	return parts[len(parts)-1]
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
