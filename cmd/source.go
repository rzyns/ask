package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yeasy/ask/internal/config"
)

// sourceCmd represents the source command
var sourceCmd = &cobra.Command{
	Use:   "source",
	Short: "Manage skill sources",
	Long:  `Add, remove, or list skill repository sources.`,
}

// sourceAddCmd represents the source add command
var sourceAddCmd = &cobra.Command{
	Use:   "add [username/repo] [path]",
	Short: "Add a skill repository source",
	Long: `Add a GitHub repository as a skill source.
Format: username/repo [optional-path]

Examples:
  ask source add anthropics/skills skills
  ask source add my-org/my-skills
  ask source add browser-use/browser-use`,
	Args: cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		input := args[0]

		// Parse username/repo format
		parts := strings.Split(input, "/")
		if len(parts) != 2 {
			fmt.Println("Error: Invalid format. Use: username/repo")
			os.Exit(1)
		}

		owner := parts[0]
		repo := parts[1]

		// Optional path within repo
		path := ""
		if len(args) > 1 {
			path = args[1]
		}

		fmt.Printf("Validating repository %s/%s...\n", owner, repo)

		// Validate repository exists and is a valid skills repo
		valid, sourceType, detectedPath := validateSkillsRepo(owner, repo, path)
		if !valid {
			fmt.Println("Error: Repository does not appear to be a valid skills repository.")
			fmt.Println("A valid skills repo should contain:")
			fmt.Println("  - A 'skills/' directory with skill folders, or")
			fmt.Println("  - Skills directly at root with SKILL.md files")
			os.Exit(1)
		}

		// Load or create config
		cfg, err := config.LoadConfig()
		if err != nil {
			if os.IsNotExist(err) {
				def := config.DefaultConfig()
				cfg = &def
			} else {
				fmt.Printf("Error loading config: %v\n", err)
				os.Exit(1)
			}
		}

		// Create source entry
		sourceName := repo
		sourceURL := fmt.Sprintf("%s/%s", owner, repo)
		if detectedPath != "" {
			sourceURL = fmt.Sprintf("%s/%s/%s", owner, repo, detectedPath)
		}

		// Check if source already exists
		for _, s := range cfg.Sources {
			if s.URL == sourceURL || s.Name == sourceName {
				fmt.Printf("Source '%s' already exists.\n", sourceName)
				return
			}
		}

		// Add source
		newSource := config.Source{
			Name: sourceName,
			Type: sourceType,
			URL:  sourceURL,
		}
		cfg.Sources = append(cfg.Sources, newSource)

		if err := cfg.Save(); err != nil {
			fmt.Printf("Error saving config: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✓ Added source '%s' (type: %s)\n", sourceName, sourceType)
		fmt.Printf("  URL: %s\n", sourceURL)
	},
}

// sourceListCmd represents the source list command
var sourceListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configured skill sources",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadConfig()
		if err != nil {
			if os.IsNotExist(err) {
				def := config.DefaultConfig()
				cfg = &def
			} else {
				fmt.Printf("Error loading config: %v\n", err)
				os.Exit(1)
			}
		}

		if len(cfg.Sources) == 0 {
			fmt.Println("No sources configured.")
			return
		}

		fmt.Println("Configured Sources:")
		for _, s := range cfg.Sources {
			fmt.Printf("  %s (%s): %s\n", s.Name, s.Type, s.URL)
		}
	},
}

// sourceRemoveCmd represents the source remove command
var sourceRemoveCmd = &cobra.Command{
	Use:   "remove [name]",
	Short: "Remove a skill source",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}

		found := false
		newSources := []config.Source{}
		for _, s := range cfg.Sources {
			if s.Name == name {
				found = true
				continue
			}
			newSources = append(newSources, s)
		}

		if !found {
			fmt.Printf("Source '%s' not found.\n", name)
			os.Exit(1)
		}

		cfg.Sources = newSources
		if err := cfg.Save(); err != nil {
			fmt.Printf("Error saving config: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Removed source '%s'\n", name)
	},
}

// validateSkillsRepo checks if a GitHub repo is a valid skills repository
func validateSkillsRepo(owner, repo, path string) (bool, string, string) {
	// First, check if the repo exists
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s", owner, repo)
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != 200 {
		return false, "", ""
	}
	resp.Body.Close()

	// Check for common skills directory patterns
	pathsToCheck := []string{}
	if path != "" {
		pathsToCheck = append(pathsToCheck, path)
	} else {
		pathsToCheck = append(pathsToCheck, "skills", "src", "")
	}

	for _, p := range pathsToCheck {
		contentsURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents", owner, repo)
		if p != "" {
			contentsURL = fmt.Sprintf("%s/%s", contentsURL, p)
		}

		resp, err := http.Get(contentsURL)
		if err != nil || resp.StatusCode != 200 {
			continue
		}
		defer resp.Body.Close()

		var contents []struct {
			Name string `json:"name"`
			Type string `json:"type"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&contents); err != nil {
			continue
		}

		// Check if this looks like a skills directory
		hasSkills := false
		for _, item := range contents {
			// Look for SKILL.md files or directories that could be skills
			if item.Type == "dir" {
				// Could be a skill directory
				hasSkills = true
				break
			}
			if item.Name == "SKILL.md" {
				hasSkills = true
				break
			}
		}

		if hasSkills {
			return true, "dir", p
		}
	}

	// If no skills found in subdirs, check if repo itself has topic
	// For topic-based repos like those tagged with agent-skill
	return true, "dir", ""
}

func init() {
	rootCmd.AddCommand(sourceCmd)
	sourceCmd.AddCommand(sourceAddCmd)
	sourceCmd.AddCommand(sourceListCmd)
	sourceCmd.AddCommand(sourceRemoveCmd)
}
