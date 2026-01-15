package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Source represents a skill source
type Source struct {
	Name string `yaml:"name"`
	Type string `yaml:"type"` // "topic" or "dir"
	URL  string `yaml:"url"`  // GitHub topic or "owner/repo/path"
}

// SkillInfo represents an installed skill with metadata
type SkillInfo struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description,omitempty"`
	URL         string `yaml:"url,omitempty"`
}

// Config represents the structure of ask.yaml
type Config struct {
	Version    string      `yaml:"version"`
	Skills     []string    `yaml:"skills,omitempty"`      // Legacy: simple list of skill names
	SkillsInfo []SkillInfo `yaml:"skills_info,omitempty"` // New: skills with metadata
	Sources    []Source    `yaml:"sources,omitempty"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() Config {
	return Config{
		Version: "1.0",
		Skills:  []string{},
		Sources: []Source{
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

	// Populate default sources if missing
	if len(config.Sources) == 0 {
		config.Sources = DefaultConfig().Sources
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
