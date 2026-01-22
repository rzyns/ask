package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/yeasy/ask/internal/cache"
	"github.com/yeasy/ask/internal/config"
	"github.com/yeasy/ask/internal/github"
	"github.com/yeasy/ask/internal/repository"
	"github.com/yeasy/ask/internal/ui"
)

// searchCmd represents the search command
var searchCmd = &cobra.Command{
	Use:   "search [keyword]",
	Short: "Search for skills on GitHub",
	Long: `Search for skills matching the keyword. 

By default, uses local cache if available (fastest), otherwise fetches from remote.
Use --local to force local-only search.
Use --remote to force remote API search.`,
	Example: `  # Search (local-first, then remote)
  ask skill search mcp
  
  # Force local cache only (offline, fastest)
  ask skill search mcp --local
  
  # Force remote API (latest data)
  ask skill search mcp --remote`,
	Run: func(cmd *cobra.Command, args []string) {
		keyword := ""
		if len(args) > 0 {
			keyword = strings.Join(args, " ")
		}

		forceLocal, _ := cmd.Flags().GetBool("local")
		forceRemote, _ := cmd.Flags().GetBool("remote")

		// Load config or use default
		cfg, err := config.LoadConfig()
		if err != nil {
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

		var allRepos []github.Repository
		var errors []string
		var searchSource string

		// Check local cache first (unless --remote is specified)
		if !forceRemote {
			reposCache, err := cache.NewReposCache()
			if err == nil {
				cachedRepos := reposCache.GetCachedRepos()
				if len(cachedRepos) > 0 || forceLocal {
					fmt.Printf("Searching local cache for '%s'...\n", keyword)
					skills, _ := reposCache.SearchSkills(keyword)

					for _, skill := range skills {
						allRepos = append(allRepos, github.Repository{
							Name:        skill.Name,
							Description: skill.Description,
							Source:      "local:" + skill.RepoName,
						})
					}
					searchSource = "local"

					if len(allRepos) > 0 || forceLocal {
						// Display results from local cache
						displaySearchResults(allRepos, installedSkills, searchSource)
						if forceLocal && len(allRepos) == 0 {
							fmt.Println("\nTip: Run 'ask repo sync' to populate local cache.")
						}
						return
					}
					// No local results and not forced local, fall through to remote
				}
			}
		}

		// Remote search
		fmt.Printf("Searching for skills matching '%s'...\n", keyword)
		searchSource = "remote"

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
				case "skillhub":
					repos, err = repository.FetchSkillsFromSkillHub(keyword, "")
				}

				// Set source name for each repo
				for i := range repos {
					repos[i].Source = r.Name
				}

				results <- searchResult{source: r.Name, repos: repos, err: err}
			}(repo)
		}

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
			for _, errMsg := range errors {
				fmt.Printf("  - %s\n", errMsg)
			}
			fmt.Println()
		}

		displaySearchResults(allRepos, installedSkills, searchSource)
	},
}

func displaySearchResults(repos []github.Repository, installedSkills map[string]bool, source string) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "NAME\tSOURCE\tINSTALLED\tSTARS\tDESCRIPTION")
	for _, repo := range repos {
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

		// Format stars (use "-" for local or dir-based)
		stars := fmt.Sprintf("%d", repo.StargazersCount)
		if repo.StargazersCount == 0 {
			stars = "-"
		}

		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", repo.Name, repo.Source, installed, stars, desc)
	}
	_ = w.Flush()

	fmt.Printf("\nFound %d skills.\n", len(repos))
	if source == "local" {
		fmt.Println("(from local cache - run 'ask repo sync' to update)")
	}
}

func init() {
	skillCmd.AddCommand(searchCmd)
	searchCmd.Flags().Bool("local", false, "search only local cache (offline)")
	searchCmd.Flags().Bool("remote", false, "force remote API search")
}
