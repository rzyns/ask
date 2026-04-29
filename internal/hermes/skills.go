package hermes

import (
	"net/url"
	"strings"
)

// HermesSourceKind identifies the policy bucket for a Hermes source.
type HermesSourceKind string

const (
	HermesSourceUnknown          HermesSourceKind = "unknown"
	HermesSourceIndex            HermesSourceKind = "index"
	HermesSourceOfficialOptional HermesSourceKind = "official-optional"
	HermesSourceBundled          HermesSourceKind = "bundled"
)

// HermesSourceClassification describes whether ASK should manage a Hermes source.
type HermesSourceClassification struct {
	Kind       HermesSourceKind
	Manageable bool
	Reason     string
}

const hermesSkillsIndexURL = "https://hermes-agent.nousresearch.com/docs/api/skills-index.json"

// ClassifyHermesSource centralizes ASK's bundled-vs-user-installable Hermes policy.
func ClassifyHermesSource(input string) HermesSourceClassification {
	normalized := strings.Trim(strings.TrimSpace(input), "/")
	if strings.EqualFold(normalized, hermesSkillsIndexURL) {
		return HermesSourceClassification{
			Kind:       HermesSourceIndex,
			Manageable: true,
			Reason:     "Hermes skills index is the canonical manageable index source.",
		}
	}

	parts := hermesSourcePathParts(normalized)
	if len(parts) >= 3 && strings.EqualFold(parts[0], "NousResearch") && strings.EqualFold(parts[1], "hermes-agent") {
		switch strings.ToLower(parts[2]) {
		case "skills":
			return HermesSourceClassification{
				Kind:       HermesSourceBundled,
				Manageable: false,
				Reason:     "Sources under NousResearch/hermes-agent/skills are bundled with Hermes Agent and are not managed by ASK.",
			}
		case "optional-skills":
			return HermesSourceClassification{
				Kind:       HermesSourceOfficialOptional,
				Manageable: true,
				Reason:     "Sources under NousResearch/hermes-agent/optional-skills are official optional Hermes skills installable by users.",
			}
		}
	}

	return HermesSourceClassification{
		Kind:       HermesSourceUnknown,
		Manageable: true,
		Reason:     "Source is not a known bundled Hermes source; treat as user-manageable.",
	}
}

func hermesSourcePathParts(input string) []string {
	if input == "" {
		return nil
	}

	if parsed, err := url.Parse(input); err == nil && parsed.Host != "" {
		if !strings.EqualFold(parsed.Host, "github.com") {
			return nil
		}
		return normalizeGitHubPathParts(splitCleanPath(parsed.Path))
	}

	parts := splitCleanPath(input)
	if len(parts) > 0 && strings.EqualFold(parts[0], "github.com") {
		parts = parts[1:]
	}
	return normalizeGitHubPathParts(parts)
}

func normalizeGitHubPathParts(parts []string) []string {
	if len(parts) >= 5 && (strings.EqualFold(parts[2], "tree") || strings.EqualFold(parts[2], "blob")) {
		// Convert GitHub browser URLs from owner/repo/tree/<branch>/<path> to owner/repo/<path>.
		return append([]string{parts[0], trimGitSuffix(parts[1])}, parts[4:]...)
	}
	if len(parts) >= 2 {
		parts[1] = trimGitSuffix(parts[1])
	}
	return parts
}

func splitCleanPath(path string) []string {
	raw := strings.Split(strings.Trim(path, "/"), "/")
	parts := make([]string, 0, len(raw))
	for _, part := range raw {
		part = strings.TrimSpace(part)
		if part != "" {
			parts = append(parts, part)
		}
	}
	return parts
}

func trimGitSuffix(repo string) string {
	return strings.TrimSuffix(repo, ".git")
}
