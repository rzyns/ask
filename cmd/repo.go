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

// repoCmd represents the repo command
var repoCmd = &cobra.Command{
	Use:   "repo",
	Short: "Manage skill repositories",
	Long:  `Add, remove, or list skill repository sources.`,
	Example: `  # Add a custom repository
  ask repo add my-org/agent-skills
  
  # List all configured sources
  ask repo list
  
  # Remove a source
  ask repo remove my-org`,
}

// repoAddCmd represents the repo add command
var repoAddCmd = &cobra.Command{
	Use:   "add <owner/repo|URL>",
	Short: "Add a skill repository",
	Long: `Add a GitHub repository as a skill source.
Format: owner/repo or full URL

Examples:
  ask repo add anthropics/skills
  ask repo add my-org/my-skills
  ask repo add browser-use/browser-use`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		input := args[0]

		// Parse username/repo format
		parts := strings.Split(input, "/")
		if len(parts) < 2 {
			fmt.Println("Error: Invalid format. Use: owner/repo")
			os.Exit(1)
		}

		owner := parts[0]
		repo := parts[1]

		// Optional path within repo (if more than 2 parts)
		path := ""
		if len(parts) > 2 {
			path = strings.Join(parts[2:], "/")
		}

		fmt.Printf("Validating repository %s/%s...\n", owner, repo)

		// Validate repository exists and is a valid skills repo
		valid, repoType, detectedPath := validateSkillsRepo(owner, repo, path)
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

		// Create repo entry
		repoName := repo
		repoURL := fmt.Sprintf("%s/%s", owner, repo)
		if detectedPath != "" {
			repoURL = fmt.Sprintf("%s/%s/%s", owner, repo, detectedPath)
		}

		// Check if repo already exists
		for _, r := range cfg.Repos {
			if r.URL == repoURL || r.Name == repoName {
				fmt.Printf("Repo '%s' already exists.\n", repoName)
				return
			}
		}

		// Add repo
		newRepo := config.Repo{
			Name: repoName,
			Type: repoType,
			URL:  repoURL,
		}
		cfg.Repos = append(cfg.Repos, newRepo)

		if err := cfg.Save(); err != nil {
			fmt.Printf("Error saving config: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✓ Added repo '%s' (type: %s)\n", repoName, repoType)
		fmt.Printf("  URL: %s\n", repoURL)
	},
}

// repoListCmd represents the repo list command
var repoListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configured skill repositories",
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

		if len(cfg.Repos) == 0 {
			fmt.Println("No repos configured.")
			return
		}

		fmt.Println("Configured Repos:")
		for _, r := range cfg.Repos {
			fmt.Printf("  %s (%s): %s\n", r.Name, r.Type, r.URL)
		}
	},
}

// repoRemoveCmd represents the repo remove command
var repoRemoveCmd = &cobra.Command{
	Use:   "remove [name]",
	Short: "Remove a skill repository",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}

		found := false
		newRepos := []config.Repo{}
		for _, r := range cfg.Repos {
			if r.Name == name {
				found = true
				continue
			}
			newRepos = append(newRepos, r)
		}

		if !found {
			fmt.Printf("Repo '%s' not found.\n", name)
			os.Exit(1)
		}

		cfg.Repos = newRepos
		if err := cfg.Save(); err != nil {
			fmt.Printf("Error saving config: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Removed repo '%s'\n", name)
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
	rootCmd.AddCommand(repoCmd)
	repoCmd.AddCommand(repoAddCmd)
	repoCmd.AddCommand(repoListCmd)
	repoCmd.AddCommand(repoRemoveCmd)
}
