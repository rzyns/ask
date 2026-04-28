package repository

import (
	"encoding/json"
	"io"
	"net/url"
	"path"
	"strings"

	"github.com/yeasy/ask/internal/config"
)

type hermesIndex struct {
	Skills []hermesIndexSkill `json:"skills"`
}

type hermesIndexSkill struct {
	ID               string `json:"id"`
	Slug             string `json:"slug"`
	Name             string `json:"name"`
	Description      string `json:"description"`
	Source           string `json:"source"`
	ResolvedGitHubID string `json:"resolved_github_id"`
	Repo             string `json:"repo"`
	Path             string `json:"path"`
	GitHub           string `json:"github"`
	GitHubURL        string `json:"github_url"`
	URL              string `json:"url"`
}

func parseHermesIndex(r io.Reader) ([]hermesIndexSkill, error) {
	var index hermesIndex
	decoder := json.NewDecoder(r)
	if err := decoder.Decode(&index); err != nil {
		return nil, err
	}
	return index.Skills, nil
}

func hermesIndexSkillsToCandidates(skills []hermesIndexSkill, keyword string) []SkillCandidate {
	if skills == nil {
		return nil
	}

	candidates := make([]SkillCandidate, 0, len(skills))
	for _, skill := range skills {
		githubPath, ok := hermesGitHubPath(skill)
		if !ok {
			continue
		}

		name := strings.TrimSpace(skill.Name)
		if name == "" {
			name = path.Base(githubPath)
		}

		candidate := SkillCandidate{
			Name:        name,
			FullName:    githubPath,
			Description: skill.Description,
			Source:      config.RepoTypeHermes,
			Install: InstallRef{
				Kind:  InstallRefGitHubPath,
				Value: githubPath,
			},
			Stars: 0,
		}
		if !hermesCandidateMatches(candidate, keyword) {
			continue
		}
		candidates = append(candidates, candidate)
	}
	return candidates
}

func hermesGitHubPath(skill hermesIndexSkill) (string, bool) {
	if githubPath, ok := normalizeHermesBareGitHubPath(skill.ResolvedGitHubID); ok {
		return githubPath, true
	}
	if githubPath, ok := normalizeHermesRepoPath(skill.Repo, skill.Path); ok {
		return githubPath, true
	}
	if githubPath, ok := normalizeHermesGitHubRef(skill.GitHub, true); ok {
		return githubPath, true
	}
	if githubPath, ok := normalizeHermesGitHubRef(skill.GitHubURL, true); ok {
		return githubPath, true
	}
	return normalizeHermesGitHubRef(skill.URL, false)
}

func normalizeHermesRepoPath(repo, skillPath string) (string, bool) {
	repo = strings.TrimSpace(repo)
	skillPath = strings.Trim(strings.TrimSpace(skillPath), "/")
	if repo == "" {
		return normalizeHermesBareGitHubPath(skillPath)
	}

	var repoPath string
	var ok bool
	if strings.Contains(repo, "://") {
		repoPath, ok = normalizeHermesGitHubURL(repo)
	} else {
		repoPath, ok = normalizeHermesBareGitHubPath(repo)
	}
	if !ok {
		return "", false
	}
	if skillPath == "" {
		return repoPath, true
	}
	return normalizeHermesBareGitHubPath(repoPath + "/" + skillPath)
}

func normalizeHermesGitHubRef(value string, allowBare bool) (string, bool) {
	if githubPath, ok := normalizeHermesGitHubURL(value); ok {
		return githubPath, true
	}
	if allowBare {
		return normalizeHermesBareGitHubPath(value)
	}
	return "", false
}

func normalizeHermesGitHubURL(value string) (string, bool) {
	value = strings.TrimSpace(value)
	if value == "" || !strings.Contains(value, "://") {
		return "", false
	}
	u, err := url.Parse(value)
	if err != nil || !strings.EqualFold(u.Hostname(), "github.com") {
		return "", false
	}
	segments := compactPathSegments(strings.Split(strings.Trim(u.Path, "/"), "/"))
	if len(segments) < 2 {
		return "", false
	}
	if len(segments) >= 4 && (segments[2] == "tree" || segments[2] == "blob") {
		segments = append(segments[:2], segments[4:]...)
	}
	return normalizeHermesBareGitHubPath(strings.Join(segments, "/"))
}

func normalizeHermesBareGitHubPath(value string) (string, bool) {
	value = strings.Trim(strings.TrimSpace(value), "/")
	value = strings.TrimSuffix(value, ".git")
	if value == "" || strings.Contains(value, "://") || strings.Contains(value, "@") {
		return "", false
	}
	parts := compactPathSegments(strings.Split(value, "/"))
	if len(parts) < 2 {
		return "", false
	}
	for _, part := range parts {
		if strings.ContainsAny(part, `\\:`) {
			return "", false
		}
	}
	return strings.Join(parts, "/"), true
}

func compactPathSegments(parts []string) []string {
	compact := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			compact = append(compact, part)
		}
	}
	return compact
}

func hermesCandidateMatches(candidate SkillCandidate, keyword string) bool {
	keyword = strings.ToLower(strings.TrimSpace(keyword))
	if keyword == "" {
		return true
	}
	fields := []string{candidate.Name, candidate.Description, candidate.FullName}
	for _, field := range fields {
		if strings.Contains(strings.ToLower(field), keyword) {
			return true
		}
	}
	return false
}
