package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yeasy/ask/internal/cache"
	"github.com/yeasy/ask/internal/config"
	"github.com/yeasy/ask/internal/github"
	"github.com/yeasy/ask/internal/ui"
)

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use:   "sync [repo-name]",
	Short: "Sync skill repositories to local cache",
	Long: `Clone or update skill repositories to local cache (~/.ask/repos/).
This enables fast offline skill discovery without GitHub API rate limits.

If no repo name is specified, syncs all configured repositories.`,
	Example: `  ask repo sync              # Sync all configured repos
  ask repo sync anthropics   # Sync only anthropics repo
  ask repo sync openai       # Sync only openai repo`,
	Run: func(_ *cobra.Command, args []string) {
		reposCache, err := cache.NewReposCache()
		if err != nil {
			fmt.Printf("Error initializing repos cache: %v\n", err)
			os.Exit(1)
		}

		// Load config to get repo list
		cfg, err := config.LoadConfig()
		if err != nil {
			// Use default config if not initialized
			def := config.DefaultConfig()
			cfg = &def
		}

		// Filter repos if specific name provided
		var targetRepos []config.Repo
		if len(args) > 0 {
			repoName := strings.ToLower(args[0])
			for _, repo := range cfg.Repos {
				if strings.ToLower(repo.Name) == repoName {
					targetRepos = append(targetRepos, repo)
					break
				}
			}
			if len(targetRepos) == 0 {
				fmt.Printf("Repository '%s' not found in configuration.\n", args[0])
				fmt.Println("Available repos:")
				for _, r := range cfg.Repos {
					fmt.Printf("  - %s\n", r.Name)
				}
				os.Exit(1)
			}
		} else {
			// Sync all repos except topic-based ones
			for _, repo := range cfg.Repos {
				if repo.Type == "dir" {
					targetRepos = append(targetRepos, repo)
				}
			}
		}

		if len(targetRepos) == 0 {
			fmt.Println("No repositories to sync.")
			return
		}

		fmt.Printf("Syncing %d repositories to ~/.ask/repos/...\n\n", len(targetRepos))

		successCount := 0
		starCounts := make(map[string]int)
		repoURLs := make(map[string]string)
		for _, repo := range targetRepos {
			repoURL := buildRepoURL(repo.URL)
			repoName := repo.Name
			if repoName == "" {
				repoName = buildRepoName(repo.URL)
			}

			repoURLs[repoName] = repoURL

			err := reposCache.CloneOrPull(repoURL, repoName)
			if err != nil {
				ui.Warn(fmt.Sprintf("  ✗ Failed to sync %s: %v", repo.Name, err))
			} else {
				ui.Debug(fmt.Sprintf("  ✓ Synced %s", repo.Name))
				successCount++

				// Fetch star count from GitHub API
				parts := strings.Split(repo.URL, "/")
				if len(parts) >= 2 {
					repoDetails, err := github.FetchRepoDetails(parts[0], parts[1])
					if err == nil {
						starCounts[repoName] = repoDetails.StargazersCount
					}
				}
			}
		}

		fmt.Printf("\nSynced %d/%d repositories.\n", successCount, len(targetRepos))

		// Save index with star counts and URLs
		if err := reposCache.SaveIndexWithStars(starCounts, repoURLs); err != nil {
			ui.Warn(fmt.Sprintf("Failed to save index: %v", err))
		}

		// Show cache location
		ui.Debug(fmt.Sprintf("Local cache: %s", cache.GetReposCacheDir()))
	},
}

// buildRepoURL constructs the git clone URL from repo config
func buildRepoURL(url string) string {
	// Handle owner/repo format
	if !strings.HasPrefix(url, "http") && !strings.HasPrefix(url, "git@") {
		// Extract owner/repo from path like "anthropics/skills/skills"
		parts := strings.Split(url, "/")
		if len(parts) >= 2 {
			return fmt.Sprintf("https://github.com/%s/%s.git", parts[0], parts[1])
		}
		return "https://github.com/" + url + ".git"
	}
	return url
}

// buildRepoName constructs a filesystem-safe name from repo URL
func buildRepoName(url string) string {
	// Handle owner/repo/path format
	parts := strings.Split(url, "/")
	if len(parts) >= 2 {
		return parts[0] + "-" + parts[1]
	}
	return strings.ReplaceAll(url, "/", "-")
}

func init() {
	repoCmd.AddCommand(syncCmd)

	// Register repo name completion
	syncCmd.ValidArgsFunction = completeRepoNames
}
