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
	scriptContent := `#!/bin/bash
echo "Hello from {{.Name}} skill!"
`
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

	return nil
}
