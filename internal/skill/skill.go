package skill

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// SkillMeta represents metadata parsed from SKILL.md
type SkillMeta struct {
	Name         string   `yaml:"name"`
	Description  string   `yaml:"description"`
	Version      string   `yaml:"version"`
	Author       string   `yaml:"author"`
	Dependencies []string `yaml:"dependencies"`
	Tags         []string `yaml:"tags"`
}

// ParseSkillMD parses a SKILL.md file and extracts frontmatter metadata
func ParseSkillMD(skillPath string) (*SkillMeta, error) {
	skillMDPath := filepath.Join(skillPath, "SKILL.md")

	file, err := os.Open(skillMDPath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	scanner := bufio.NewScanner(file)
	var frontmatter []string
	inFrontmatter := false
	lineCount := 0

	for scanner.Scan() {
		line := scanner.Text()
		lineCount++

		// Check for frontmatter delimiters
		if strings.TrimSpace(line) == "---" {
			if !inFrontmatter && lineCount <= 2 {
				inFrontmatter = true
				continue
			} else if inFrontmatter {
				// End of frontmatter
				break
			}
		}

		if inFrontmatter {
			frontmatter = append(frontmatter, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if len(frontmatter) == 0 {
		// No frontmatter found, try to extract info from file content
		return parseFromContent(skillMDPath)
	}

	var meta SkillMeta
	yamlContent := strings.Join(frontmatter, "\n")
	if err := yaml.Unmarshal([]byte(yamlContent), &meta); err != nil {
		return nil, err
	}

	return &meta, nil
}

// parseFromContent attempts to extract metadata from SKILL.md content
func parseFromContent(skillMDPath string) (*SkillMeta, error) {
	content, err := os.ReadFile(skillMDPath)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(content), "\n")
	meta := &SkillMeta{}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Look for title (first H1)
		if strings.HasPrefix(line, "# ") && meta.Name == "" {
			meta.Name = strings.TrimPrefix(line, "# ")
		}
		// Look for description (first paragraph after title)
		if meta.Name != "" && meta.Description == "" && !strings.HasPrefix(line, "#") && line != "" {
			meta.Description = line
			break
		}
	}

	return meta, nil
}

// FindSkillMD checks if a skill has a SKILL.md file
func FindSkillMD(skillPath string) bool {
	skillMDPath := filepath.Join(skillPath, "SKILL.md")
	_, err := os.Stat(skillMDPath)
	return err == nil
}
