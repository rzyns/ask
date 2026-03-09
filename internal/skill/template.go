package skill

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

// TemplateData holds data for the skill template
type TemplateData struct {
	Name        string
	Description string
	Author      string
}

const skillMDTemplate = `---
name: {{.Name}}
description: {{.Description}}
version: 1.0.0
author: {{.Author}}
tags:
  - agent-skill
---

# {{.Name}}

{{.Description}}

## Usage
Explain how to use this skill here.

## Scripts
- **hello**: Example script

## Resources
- [Reference](references/ref.md)
`

// getGitAuthor attempts to get the author name from git config
// Falls back to "User" if git config is unavailable
func getGitAuthor() string {
	cmd := exec.Command("git", "config", "user.name")
	output, err := cmd.Output()
	if err != nil {
		return "User"
	}
	author := strings.TrimSpace(string(output))
	if author == "" {
		return "User"
	}
	return author
}

// CreateSkillTemplate creates a new skill directory with template files
func CreateSkillTemplate(name, destDir string) error {
	skillDir := filepath.Join(destDir, name)

	// 1. Create directory structure
	dirs := []string{
		skillDir,
		filepath.Join(skillDir, "prompts"),
		filepath.Join(skillDir, "scripts"),
		filepath.Join(skillDir, "references"),
		filepath.Join(skillDir, "assets"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// 2. Create SKILL.md
	tmpl, err := template.New("skill").Parse(skillMDTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	f, err := os.Create(filepath.Join(skillDir, "SKILL.md"))
	if err != nil {
		return fmt.Errorf("failed to create SKILL.md: %w", err)
	}
	defer func() { _ = f.Close() }()

	data := TemplateData{
		Name:        name,
		Description: "A new skill for AI Agents",
		Author:      getGitAuthor(),
	}

	if err := tmpl.Execute(f, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	// 3. Create example script
	scriptContent := fmt.Sprintf("#!/bin/bash\necho \"Hello from %s skill!\"\n", name)
	scriptFile := filepath.Join(skillDir, "scripts", "hello.sh")
	if err := os.WriteFile(scriptFile, []byte(scriptContent), 0755); err != nil {
		return fmt.Errorf("failed to create example script: %w", err)
	}

	// 4. Create example reference
	refContent := `# Reference
This is an example reference file.
`
	if err := os.WriteFile(filepath.Join(skillDir, "references", "ref.md"), []byte(refContent), 0644); err != nil {
		return fmt.Errorf("failed to create example reference: %w", err)
	}

	// 5. Create example prompt
	promptContent := `# Example Prompt

This is an example prompt file. Replace this with your actual prompt content.

## Usage

Describe how this prompt should be used by the AI agent.

## Variables

- Variable1: Description of variable 1
- Variable2: Description of variable 2
`
	if err := os.WriteFile(filepath.Join(skillDir, "prompts", "example.md"), []byte(promptContent), 0644); err != nil {
		return fmt.Errorf("failed to create example prompt: %w", err)
	}

	// 6. Create README.md
	readmeContent := fmt.Sprintf(`# %s

A skill for AI Agents to [describe what this skill does].

## Description

[Add a detailed description of your skill here. Explain what it does, when to use it, and how it benefits users.]

## Features

- [Feature 1]
- [Feature 2]
- [Feature 3]

## Installation

` + "`ask install <your-github-username>/%s`" + `

## Usage

[Explain how to use this skill. Provide examples if applicable.]

### Example

[Provide a concrete example of how to use the skill]

## Requirements

[List any external dependencies, API keys, or environment variables needed]

## Configuration

[Explain any configuration options or environment variables]

### Environment Variables

See `.env.example` for required environment variables.

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Add tests if applicable
4. Submit a pull request

## License

[Specify your license, e.g., MIT, Apache 2.0, etc.]

## Support

For issues and questions, please open an issue on GitHub.
`, name, name)
	if err := os.WriteFile(filepath.Join(skillDir, "README.md"), []byte(readmeContent), 0644); err != nil {
		return fmt.Errorf("failed to create README.md: %w", err)
	}

	// 7. Create .env.example
	envExampleContent := `# Environment variables for ` + name + `

# Example configuration - copy to .env and fill in your values

# API Keys
# API_KEY=your_api_key_here

# Configuration
# DEBUG=false
# TIMEOUT=30
`
	if err := os.WriteFile(filepath.Join(skillDir, ".env.example"), []byte(envExampleContent), 0644); err != nil {
		return fmt.Errorf("failed to create .env.example: %w", err)
	}

	return nil
}
