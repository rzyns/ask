package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Repo represents a skill repository
type Repo struct {
	Name string `yaml:"name"`
	Type string `yaml:"type"` // "topic" or "dir"
	URL  string `yaml:"url"`  // GitHub topic or "owner/repo/path"
}

// ToolTarget represents a supported AI coding tool
type ToolTarget struct {
	Name      string `yaml:"name"`
	SkillsDir string `yaml:"skills_dir"`
	Enabled   bool   `yaml:"enabled"`
}

// DefaultToolTargets returns the supported AI coding tools
func DefaultToolTargets() []ToolTarget {
	targets := []ToolTarget{}
	// Add default Agent skills
	targets = append(targets, ToolTarget{
		Name:      "agent",
		SkillsDir: DefaultSkillsDir,
		Enabled:   true,
	})

	// Add supported agents
	for _, name := range GetSupportedAgentNames() {
		agentType, _ := ResolveAgentType(name)
		config := SupportedAgents[agentType]
		targets = append(targets, ToolTarget{
			Name:      name,
			SkillsDir: config.ProjectDir,
			Enabled:   true,
		})
	}
	return targets
}

// SkillInfo represents an installed skill with metadata
type SkillInfo struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description,omitempty"`
	URL         string `yaml:"url,omitempty"`
}

// Config represents the structure of ask.yaml
type Config struct {
	Version     string       `yaml:"version"`
	SkillsDir   string       `yaml:"skills_dir,omitempty"`   // Skills installation directory (default: .agent/skills)
	ToolTargets []ToolTarget `yaml:"tool_targets,omitempty"` // Target AI tools for skill installation
	Skills      []string     `yaml:"skills,omitempty"`       // Legacy: simple list of skill names
	SkillsInfo  []SkillInfo  `yaml:"skills_info,omitempty"`  // New: skills with metadata
	Repos       []Repo       `yaml:"repos,omitempty"`
}

const DefaultSkillsDir = ".agent/skills"

// Global installation paths
const GlobalConfigDirName = ".ask"
const GlobalConfigFileName = "config.yaml"
const GlobalSkillsDirName = "skills"
const GlobalLockFileName = "ask.lock"

// GetGlobalConfigDir returns the global config directory path (~/.ask)
func GetGlobalConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, GlobalConfigDirName)
}

// GetGlobalConfigPath returns the global config file path (~/.ask/config.yaml)
func GetGlobalConfigPath() string {
	return filepath.Join(GetGlobalConfigDir(), GlobalConfigFileName)
}

// GetGlobalSkillsDir returns the global skills directory path (~/.ask/skills)
func GetGlobalSkillsDir() string {
	return filepath.Join(GetGlobalConfigDir(), GlobalSkillsDirName)
}

// GetGlobalLockPath returns the global lock file path (~/.ask/ask.lock)
func GetGlobalLockPath() string {
	return filepath.Join(GetGlobalConfigDir(), GlobalLockFileName)
}

// EnsureGlobalDirExists creates the global config directory if it doesn't exist
func EnsureGlobalDirExists() error {
	globalDir := GetGlobalConfigDir()
	if globalDir == "" {
		return fmt.Errorf("could not determine home directory")
	}
	return os.MkdirAll(globalDir, 0755)
}

// GetSkillsDir returns the skills directory based on global flag
func GetSkillsDirByScope(global bool) string {
	if global {
		return GetGlobalSkillsDir()
	}
	return DefaultSkillsDir
}

// GetSkillsDir returns the skills directory, using default if not set
func (c *Config) GetSkillsDir() string {
	if c.SkillsDir == "" {
		return DefaultSkillsDir
	}
	return c.SkillsDir
}

// DefaultConfig returns the default configuration
func DefaultConfig() Config {
	return Config{
		Version: "1.0",
		Skills:  []string{},
		Repos: []Repo{
			{
				Name: "community",
				Type: "topic",
				URL:  "agent-skill",
			},
			{
				Name: "anthropics",
				Type: "dir",
				URL:  "anthropics/skills/skills",
			},
			{
				Name: "mcp-servers",
				Type: "dir",
				URL:  "modelcontextprotocol/servers/src",
			},
			{
				Name: "scientific",
				Type: "dir",
				URL:  "K-Dense-AI/claude-scientific-skills/scientific-skills",
			},
			{
				Name: "superpowers",
				Type: "dir",
				URL:  "obra/superpowers/skills",
			},
			{
				Name: "openai",
				Type: "dir",
				URL:  "openai/skills/skills",
			},
		},
	}
}

// LoadConfig loads the current ask.yaml configuration
func LoadConfig() (*Config, error) {
	data, err := os.ReadFile("ask.yaml")
	if err != nil {
		return nil, err
	}
	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	// Merge default repos with existing (add missing defaults)
	defaultRepos := DefaultConfig().Repos
	existingNames := make(map[string]bool)
	for _, r := range config.Repos {
		existingNames[r.Name] = true
	}
	for _, dr := range defaultRepos {
		if !existingNames[dr.Name] {
			config.Repos = append(config.Repos, dr)
		}
	}

	return &config, nil
}

// Save saves the configuration to ask.yaml
func (c *Config) Save() error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile("ask.yaml", data, 0644)
}

// RemoveSkill removes a skill from the configuration
func (c *Config) RemoveSkill(skillName string) {
	for i, s := range c.Skills {
		if s == skillName {
			c.Skills = append(c.Skills[:i], c.Skills[i+1:]...)
			return
		}
	}
}

// RemoveSkillInfo removes skill metadata from the configuration
func (c *Config) RemoveSkillInfo(skillName string) {
	for i, s := range c.SkillsInfo {
		if s.Name == skillName {
			c.SkillsInfo = append(c.SkillsInfo[:i], c.SkillsInfo[i+1:]...)
			return
		}
	}
}

// AddSkill adds a skill to the configuration if it doesn't exist
func (c *Config) AddSkill(skillName string) {
	for _, s := range c.Skills {
		if s == skillName {
			return
		}
	}
	c.Skills = append(c.Skills, skillName)
}

// AddSkillInfo adds a skill with metadata to the configuration
func (c *Config) AddSkillInfo(info SkillInfo) {
	// Check if skill already exists
	for i, s := range c.SkillsInfo {
		if s.Name == info.Name {
			// Update existing
			c.SkillsInfo[i] = info
			return
		}
	}
	c.SkillsInfo = append(c.SkillsInfo, info)

	// Also add to legacy Skills list for backward compatibility
	c.AddSkill(info.Name)
}

// GetSkillInfo returns skill info by name
func (c *Config) GetSkillInfo(name string) *SkillInfo {
	for _, s := range c.SkillsInfo {
		if s.Name == name {
			return &s
		}
	}
	return nil
}

// CreateDefaultConfig creates a new ask.yaml in the current directory
func CreateDefaultConfig() error {
	config := DefaultConfig()
	return config.Save()
}

// LoadGlobalConfig loads the global config file (~/.ask/config.yaml)
func LoadGlobalConfig() (*Config, error) {
	configPath := GetGlobalConfigPath()
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return default config if doesn't exist
			cfg := DefaultConfig()
			return &cfg, nil
		}
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	// Merge default repos with existing (add missing defaults)
	defaultRepos := DefaultConfig().Repos
	existingNames := make(map[string]bool)
	for _, r := range config.Repos {
		existingNames[r.Name] = true
	}
	for _, dr := range defaultRepos {
		if !existingNames[dr.Name] {
			config.Repos = append(config.Repos, dr)
		}
	}

	return &config, nil
}

// SaveGlobal saves the configuration to the global config file (~/.ask/config.yaml)
func (c *Config) SaveGlobal() error {
	if err := EnsureGlobalDirExists(); err != nil {
		return err
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(GetGlobalConfigPath(), data, 0644)
}

// LoadConfigByScope loads config based on global flag
func LoadConfigByScope(global bool) (*Config, error) {
	if global {
		return LoadGlobalConfig()
	}
	return LoadConfig()
}

// SaveByScope saves config based on global flag
func (c *Config) SaveByScope(global bool) error {
	if global {
		return c.SaveGlobal()
	}
	return c.Save()
}

// GetToolTargets returns the configured tool targets, or defaults if none configured
func (c *Config) GetToolTargets() []ToolTarget {
	if len(c.ToolTargets) > 0 {
		return c.ToolTargets
	}
	return DefaultToolTargets()
}

// GetEnabledToolTargets returns only the enabled tool targets
func (c *Config) GetEnabledToolTargets() []ToolTarget {
	var enabled []ToolTarget
	for _, t := range c.GetToolTargets() {
		if t.Enabled {
			enabled = append(enabled, t)
		}
	}
	return enabled
}

// GetEnabledSkillsDirs returns all enabled skill directories
func (c *Config) GetEnabledSkillsDirs() []string {
	var dirs []string
	for _, t := range c.GetEnabledToolTargets() {
		dirs = append(dirs, t.SkillsDir)
	}
	return dirs
}

// DetectExistingToolDirs detects which AI tool directories already exist in the project
func DetectExistingToolDirs(projectDir string) []ToolTarget {
	var detected []ToolTarget
	for _, t := range DefaultToolTargets() {
		// Check if the tool's parent directory exists (e.g., .claude, .cursor)
		toolDir := filepath.Dir(t.SkillsDir)
		if _, err := os.Stat(filepath.Join(projectDir, toolDir)); err == nil {
			detected = append(detected, t)
		}
	}
	return detected
}

// GetActiveSkillsDirs returns skill directories that exist or should be created
// If specific tool directories exist, only those are returned; otherwise returns all enabled
func (c *Config) GetActiveSkillsDirs(projectDir string) []string {
	detected := DetectExistingToolDirs(projectDir)
	if len(detected) > 0 {
		var dirs []string
		for _, t := range detected {
			if t.Enabled {
				dirs = append(dirs, t.SkillsDir)
			}
		}
		return dirs
	}
	// No specific tool detected, use default
	return []string{c.GetSkillsDir()}
}

// GetToolTargetByName returns a tool target by name
func (c *Config) GetToolTargetByName(name string) *ToolTarget {
	for _, t := range c.GetToolTargets() {
		if t.Name == name {
			return &t
		}
	}
	return nil
}

// ParseToolTargetFlags parses a comma-separated list of tool names into directories
func (c *Config) ParseToolTargetFlags(targetFlags string) []string {
	if targetFlags == "" {
		return nil
	}
	var dirs []string
	for _, name := range splitAndTrim(targetFlags, ",") {
		if t := c.GetToolTargetByName(name); t != nil && t.Enabled {
			dirs = append(dirs, t.SkillsDir)
		}
	}
	return dirs
}

// splitAndTrim splits a string and trims whitespace from each part
func splitAndTrim(s, sep string) []string {
	parts := strings.Split(s, sep)
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
