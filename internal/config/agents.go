package config

import (
	"os"
	"path/filepath"
)

// AgentType represents a supported AI coding agent
type AgentType string

const (
	AgentClaude      AgentType = "claude"
	AgentCursor      AgentType = "cursor"
	AgentCodex       AgentType = "codex"
	AgentOpenCode    AgentType = "opencode"
	AgentAntigravity AgentType = "antigravity"
	AgentGemini      AgentType = "gemini"
	AgentCopilot     AgentType = "copilot"
	AgentWindsurf    AgentType = "windsurf"
	AgentAmp         AgentType = "amp"
	AgentGoose       AgentType = "goose"
	AgentKilo        AgentType = "kilo"
	AgentKiro        AgentType = "kiro"
	AgentRoo         AgentType = "roo"
	AgentTrae        AgentType = "trae"
	AgentDroid       AgentType = "droid"
	AgentClawdBot    AgentType = "clawdbot"
	AgentNeovate     AgentType = "neovate"
)

// AgentConfig holds directory paths for an agent
type AgentConfig struct {
	Name       string   // Display name
	ProjectDir string   // Project-level skills directory (e.g., ".claude/skills")
	GlobalDir  string   // User-level skills directory (e.g., "~/.claude/skills")
	Aliases    []string // Alternative names for this agent
}

// SupportedAgents maps agent types to their configurations
var SupportedAgents = map[AgentType]AgentConfig{
	AgentClaude: {
		Name:       "Claude",
		ProjectDir: ".claude/skills",
		GlobalDir:  ".claude/skills", // Will be prefixed with home dir
		Aliases:    []string{"claude-code"},
	},
	AgentCursor: {
		Name:       "Cursor",
		ProjectDir: ".cursor/skills",
		GlobalDir:  ".cursor/skills",
		Aliases:    []string{},
	},
	AgentCodex: {
		Name:       "Codex",
		ProjectDir: ".codex/skills",
		GlobalDir:  ".codex/skills",
		Aliases:    []string{"openai-codex"},
	},
	AgentOpenCode: {
		Name:       "OpenCode",
		ProjectDir: ".opencode/skills",
		GlobalDir:  ".config/opencode/skills",
		Aliases:    []string{},
	},
	AgentAntigravity: {
		Name:       "Antigravity",
		ProjectDir: ".agent/skills",
		GlobalDir:  ".gemini/antigravity/skills",
		Aliases:    []string{"gemini-antigravity"},
	},
	AgentGemini: {
		Name:       "Gemini CLI",
		ProjectDir: ".gemini/skills",
		GlobalDir:  ".gemini/skills",
		Aliases:    []string{"gemini-cli"},
	},
	AgentCopilot: {
		Name:       "GitHub Copilot",
		ProjectDir: ".github/skills",
		GlobalDir:  ".copilot/skills",
		Aliases:    []string{"github-copilot"},
	},
	AgentWindsurf: {
		Name:       "Windsurf",
		ProjectDir: ".windsurf/skills",
		GlobalDir:  ".codeium/windsurf/skills",
		Aliases:    []string{},
	},
	AgentAmp: {
		Name:       "Amp",
		ProjectDir: ".agents/skills",
		GlobalDir:  ".config/agents/skills",
		Aliases:    []string{},
	},
	AgentGoose: {
		Name:       "Goose",
		ProjectDir: ".goose/skills",
		GlobalDir:  ".config/goose/skills",
		Aliases:    []string{},
	},
	AgentKilo: {
		Name:       "Kilo",
		ProjectDir: ".kilocode/skills",
		GlobalDir:  ".kilocode/skills",
		Aliases:    []string{"kilocode"},
	},
	AgentKiro: {
		Name:       "Kiro",
		ProjectDir: ".kiro/skills",
		GlobalDir:  ".kiro/skills",
		Aliases:    []string{"kiro-cli"},
	},
	AgentRoo: {
		Name:       "Roo",
		ProjectDir: ".roo/skills",
		GlobalDir:  ".roo/skills",
		Aliases:    []string{},
	},
	AgentTrae: {
		Name:       "Trae",
		ProjectDir: ".trae/skills",
		GlobalDir:  ".trae/skills",
		Aliases:    []string{},
	},
	AgentDroid: {
		Name:       "Droid",
		ProjectDir: ".factory/skills",
		GlobalDir:  ".factory/skills",
		Aliases:    []string{},
	},
	AgentClawdBot: {
		Name:       "ClawdBot",
		ProjectDir: "skills",
		GlobalDir:  ".clawdbot/skills",
		Aliases:    []string{},
	},
	AgentNeovate: {
		Name:       "Neovate",
		ProjectDir: ".neovate/skills",
		GlobalDir:  ".neovate/skills",
		Aliases:    []string{},
	},
}

// GetSupportedAgentNames returns a list of all supported agent type names
func GetSupportedAgentNames() []string {
	names := make([]string, 0, len(SupportedAgents))
	for agent := range SupportedAgents {
		names = append(names, string(agent))
	}
	return names
}

// IsValidAgent checks if the given agent name is supported
func IsValidAgent(name string) bool {
	// Check direct match
	if _, ok := SupportedAgents[AgentType(name)]; ok {
		return true
	}
	// Check aliases
	for _, config := range SupportedAgents {
		for _, alias := range config.Aliases {
			if alias == name {
				return true
			}
		}
	}
	return false
}

// ResolveAgentType resolves an agent name (including aliases) to its AgentType
func ResolveAgentType(name string) (AgentType, bool) {
	// Check direct match
	if _, ok := SupportedAgents[AgentType(name)]; ok {
		return AgentType(name), true
	}
	// Check aliases
	for agentType, config := range SupportedAgents {
		for _, alias := range config.Aliases {
			if alias == name {
				return agentType, true
			}
		}
	}
	return "", false
}

// GetAgentSkillsDir returns the skills directory for a specific agent
// If global is true, returns the user-level directory (e.g., ~/.claude/skills)
// Otherwise returns the project-level directory (e.g., .claude/skills)
func GetAgentSkillsDir(agent AgentType, global bool) (string, error) {
	config, ok := SupportedAgents[agent]
	if !ok {
		return "", nil
	}

	if global {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, config.GlobalDir), nil
	}

	return config.ProjectDir, nil
}

// GetAllAgentSkillsDirs returns all possible skill directories for discovery
// Returns both project-level and global directories for all supported agents
func GetAllAgentSkillsDirs() []string {
	dirs := make([]string, 0)

	// Add default ASK directory
	dirs = append(dirs, DefaultSkillsDir)

	// Add project-level directories
	for _, config := range SupportedAgents {
		dirs = append(dirs, config.ProjectDir)
	}

	// Add global directories
	home, err := os.UserHomeDir()
	if err == nil {
		for _, config := range SupportedAgents {
			dirs = append(dirs, filepath.Join(home, config.GlobalDir))
		}
	}

	return dirs
}
