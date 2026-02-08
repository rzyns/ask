package cmd

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/yeasy/ask/internal/config"
	"github.com/yeasy/ask/internal/skill"
)

// AvailableSkills represents the XML structure for agent prompt integration
type AvailableSkills struct {
	XMLName xml.Name     `xml:"available_skills"`
	Skills  []SkillEntry `xml:"skill"`
}

// SkillEntry represents a single skill in the XML output
type SkillEntry struct {
	Name        string `xml:"name"`
	Description string `xml:"description"`
	Location    string `xml:"location"`
}

var promptOutputFile string

// promptCmd represents the prompt command
var promptCmd = &cobra.Command{
	Use:   "prompt [paths...]",
	Short: "Generate XML skill listing for agent prompts",
	Long: `Generate <available_skills> XML block for agent system prompts.

This format is recommended for Anthropic's models and follows the
Agent Skills specification at https://agentskills.io/specification.

If no paths are specified, scans installed skills in default locations.

Examples:
  ask skill prompt                          # Scan all installed skills
  ask skill prompt .agent/skills/pdf        # Single skill
  ask skill prompt ./skills/a ./skills/b    # Multiple skills`,
	Run: runPrompt,
}

func init() {
	promptCmd.Flags().StringVarP(&promptOutputFile, "output", "o", "", "Write XML to file instead of stdout")
}

func runPrompt(_ *cobra.Command, args []string) {
	var skillPaths []string

	if len(args) > 0 {
		// Use provided paths
		for _, arg := range args {
			absPath, err := filepath.Abs(arg)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error resolving path %s: %v\n", arg, err)
				continue
			}
			if skill.FindSkillMD(absPath) {
				skillPaths = append(skillPaths, absPath)
			} else {
				fmt.Fprintf(os.Stderr, "Warning: %s is not a valid skill (no SKILL.md found)\n", arg)
			}
		}
	} else {
		// Scan default locations
		skillPaths = discoverInstalledSkills()
	}

	if len(skillPaths) == 0 {
		fmt.Fprintln(os.Stderr, "No skills found.")
		os.Exit(1)
	}

	// Build XML structure
	availableSkills := AvailableSkills{}
	for _, path := range skillPaths {
		meta, err := skill.ParseSkillMD(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to parse %s: %v\n", path, err)
			continue
		}

		entry := SkillEntry{
			Name:        meta.Name,
			Description: meta.Description,
			Location:    filepath.Join(path, "SKILL.md"),
		}
		availableSkills.Skills = append(availableSkills.Skills, entry)
	}

	// Generate XML
	output, err := xml.MarshalIndent(availableSkills, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating XML: %v\n", err)
		os.Exit(1)
	}

	xmlContent := string(output)

	if promptOutputFile != "" {
		if err := os.WriteFile(promptOutputFile, []byte(xmlContent), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing to %s: %v\n", promptOutputFile, err)
			os.Exit(1)
		}
		fmt.Printf("XML written to %s\n", promptOutputFile)
	} else {
		fmt.Println(xmlContent)
	}
}

// discoverInstalledSkills finds all installed skills in known locations
func discoverInstalledSkills() []string {
	var paths []string
	cwd, _ := os.Getwd()
	home, _ := os.UserHomeDir()

	// Locations to scan
	searchDirs := []string{
		filepath.Join(cwd, config.DefaultSkillsDir),
		filepath.Join(cwd, "skills"),
	}

	// Add agent-specific directories
	for _, agentConfig := range config.SupportedAgents {
		searchDirs = append(searchDirs, filepath.Join(cwd, agentConfig.ProjectDir))
		if home != "" {
			searchDirs = append(searchDirs, filepath.Join(home, agentConfig.GlobalDir))
		}
	}

	// Add global skills directory
	if home != "" {
		searchDirs = append(searchDirs, filepath.Join(home, ".ask", "skills"))
	}

	// Deduplicate and scan
	seen := make(map[string]bool)
	for _, dir := range searchDirs {
		if seen[dir] {
			continue
		}
		seen[dir] = true

		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			skillPath := filepath.Join(dir, entry.Name())
			if skill.FindSkillMD(skillPath) {
				// Use absolute path for deduplication
				absPath, _ := filepath.Abs(skillPath)
				if !containsPath(paths, absPath) {
					paths = append(paths, absPath)
				}
			}
		}
	}

	return paths
}

func containsPath(paths []string, target string) bool {
	for _, p := range paths {
		if p == target {
			return true
		}
	}
	return false
}
