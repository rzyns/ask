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

		var allRepos []github.Repository

		for _, source := range cfg.Sources {
			fmt.Printf("Scanning source: %s (%s)...\n", source.Name, source.Type)

			var repos []github.Repository
			var err error

			if source.Type == "topic" {
				repos, err = github.SearchTopic(source.URL, keyword)
			} else if source.Type == "dir" {
				parts := strings.Split(source.URL, "/")
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
				} else {
					fmt.Printf("Invalid source URL: %s\n", source.URL)
				}
			}

			if err != nil {
				fmt.Printf("Error searching source %s: %v\n", source.Name, err)
				continue
			}

			allRepos = append(allRepos, repos...)
		}

		if len(allRepos) == 0 {
			fmt.Println("No skills found.")
			return
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
