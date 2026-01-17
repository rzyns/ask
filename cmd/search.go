package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/yeasy/ask/internal/config"
	"github.com/yeasy/ask/internal/github"
	"github.com/yeasy/ask/internal/ui"
)

// searchCmd represents the search command
var searchCmd = &cobra.Command{
	Use:   "search [keyword]",
	Short: "Search for skills on GitHub",
	Long: `Search GitHub repositories with the 'agent-skill' topic. 
You can provide an optional keyword to filter results (e.g. 'browser', 'python').`,
	Example: `  # Search for browser-related skills
  ask skill search browser
  
  # Search for Python skills
  ask skill search python
  
  # Search all available skills
  ask skill search`,
	Run: func(cmd *cobra.Command, args []string) {
		keyword := ""
		if len(args) > 0 {
			keyword = strings.Join(args, " ")
		}

		fmt.Printf("Searching for skills matching '%s'...\n", keyword)

		// Load config or use default
		cfg, err := config.LoadConfig()
		if err != nil {
			// It's okay if config doesn't exist, use default
			def := config.DefaultConfig()
			cfg = &def
		}

		// Build a set of installed skills for marking
		installedSkills := make(map[string]bool)
		for _, s := range cfg.Skills {
			installedSkills[s] = true
		}
		for _, s := range cfg.SkillsInfo {
			installedSkills[s.Name] = true
		}

		// Create progress bar for scanning sources
		bar := ui.NewProgressBar(len(cfg.Repos), "Scanning sources")

		// Search sources in parallel
		type searchResult struct {
			source string
			repos  []github.Repository
			err    error
		}

		results := make(chan searchResult, len(cfg.Repos))

		for _, repo := range cfg.Repos {
			go func(r config.Repo) {
				var repos []github.Repository
				var err error

				switch r.Type {
				case "topic":
					repos, err = github.SearchTopic(r.URL, keyword)
				case "dir":
					parts := strings.Split(r.URL, "/")
					if len(parts) >= 3 {
						owner := parts[0]
						repoName := parts[1]
						path := strings.Join(parts[2:], "/")
						repos, err = github.SearchDir(owner, repoName, path)

						// Filter client-side by keyword
						if err == nil && keyword != "" {
							var filtered []github.Repository
							for _, rp := range repos {
								if strings.Contains(strings.ToLower(rp.Name), strings.ToLower(keyword)) {
									filtered = append(filtered, rp)
								}
							}
							repos = filtered
						}
					}
				}

				// Set source name for each repo
				for i := range repos {
					repos[i].Source = r.Name
				}

				results <- searchResult{source: r.Name, repos: repos, err: err}
			}(repo)
		}

		var allRepos []github.Repository
		var errors []string
		for i := 0; i < len(cfg.Repos); i++ {
			result := <-results
			_ = bar.Add(1)
			if result.err != nil {
				errors = append(errors, fmt.Sprintf("%s: %v", result.source, result.err))
				continue
			}
			allRepos = append(allRepos, result.repos...)
		}

		fmt.Println()
		if len(errors) > 0 {
			fmt.Println("Warning: Some sources failed to load:")
			for _, err := range errors {
				fmt.Printf("  - %s\n", err)
			}
			fmt.Println()
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		_, _ = fmt.Fprintln(w, "NAME\tSOURCE\tINSTALLED\tSTARS\tDESCRIPTION")
		for _, repo := range allRepos {
			// Truncate description if too long
			desc := repo.Description
			if len(desc) > 40 {
				desc = desc[:37] + "..."
			}

			// Check if installed
			installed := ""
			if installedSkills[repo.Name] {
				installed = "✓"
			}

			// Format stars (use "-" for dir-based where stars are parent repo)
			stars := fmt.Sprintf("%d", repo.StargazersCount)
			if repo.StargazersCount == 0 {
				stars = "-"
			}

			_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", repo.Name, repo.Source, installed, stars, desc)
		}
		_ = w.Flush()

		fmt.Printf("\nFound %d skills.\n", len(allRepos))
	},
}

func init() {
	skillCmd.AddCommand(searchCmd)
}
