package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/yeasy/ask/internal/config"
	"github.com/yeasy/ask/internal/github"
)

// searchCmd represents the search command
var searchCmd = &cobra.Command{
	Use:   "search [keyword]",
	Short: "Search for skills on GitHub",
	Long: `Search GitHub repositories with the 'agent-skill' topic. 
You can provide an optional keyword to filter results (e.g. 'browser', 'python').`,
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

		// Search sources in parallel
		type searchResult struct {
			source string
			repos  []github.Repository
			err    error
		}

		results := make(chan searchResult, len(cfg.Sources))

		for _, source := range cfg.Sources {
			go func(s config.Source) {
				var repos []github.Repository
				var err error

				if s.Type == "topic" {
					repos, err = github.SearchTopic(s.URL, keyword)
				} else if s.Type == "dir" {
					parts := strings.Split(s.URL, "/")
					if len(parts) >= 3 {
						owner := parts[0]
						repo := parts[1]
						path := strings.Join(parts[2:], "/")
						repos, err = github.SearchDir(owner, repo, path)

						// Filter client-side by keyword
						if err == nil && keyword != "" {
							var filtered []github.Repository
							for _, r := range repos {
								if strings.Contains(strings.ToLower(r.Name), strings.ToLower(keyword)) {
									filtered = append(filtered, r)
								}
							}
							repos = filtered
						}
					}
				}

				results <- searchResult{source: s.Name, repos: repos, err: err}
			}(source)
		}

		var allRepos []github.Repository
		for i := 0; i < len(cfg.Sources); i++ {
			result := <-results
			fmt.Printf("Scanning source: %s...\n", result.source)
			if result.err != nil {
				fmt.Printf("  Error: %v\n", result.err)
				continue
			}
			allRepos = append(allRepos, result.repos...)
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tDESCRIPTION\tSTARS\tURL")
		for _, repo := range allRepos {
			// Truncate description if too long
			desc := repo.Description
			if len(desc) > 50 {
				desc = desc[:47] + "..."
			}
			fmt.Fprintf(w, "%s\t%s\t%d\t%s\n", repo.FullName, desc, repo.StargazersCount, repo.HTMLURL)
		}
		w.Flush()
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
}
