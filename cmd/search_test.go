package cmd

import (
	"testing"

	"github.com/yeasy/ask/internal/config"
)

func TestSearchConfiguration(t *testing.T) {
	if searchCmd.Flags().Lookup("local") == nil {
		t.Error("searchCmd missing 'local' flag")
	}
	if searchCmd.Flags().Lookup("remote") == nil {
		t.Error("searchCmd missing 'remote' flag")
	}
	if searchCmd.Flags().Lookup("min-stars") == nil {
		t.Error("searchCmd missing 'min-stars' flag")
	}

	if searchRootCmd.Flags().Lookup("local") == nil {
		t.Error("searchRootCmd missing 'local' flag")
	}
}

func TestSearchRootCommand(t *testing.T) {
	if searchRootCmd.Use != "search [keyword]" {
		t.Errorf("Expected use 'search [keyword]', got '%s'", searchRootCmd.Use)
	}
	// Note: We can't easily compare function pointers in Go for Run,
	// but we can assume it's wired if the object exists.
}

func TestFilterRemoteOnlyReposForCachedSearch(t *testing.T) {
	repos := []config.Repo{
		{Name: "anthropics", Type: config.RepoTypeDir},
		{Name: "featured", Type: config.RepoTypeRegistry},
		{Name: "hermes-index", Type: config.RepoTypeHermes},
	}

	filtered := remoteReposAfterLocalCacheHit(repos)
	if len(filtered) != 1 {
		t.Fatalf("expected 1 uncached Hermes repo, got %d: %#v", len(filtered), filtered)
	}
	if filtered[0].Name != "hermes-index" {
		t.Fatalf("unexpected remote repo filter result: %#v", filtered)
	}
}

func TestRepoDisplayURLKeepsAbsoluteHermesIndexURL(t *testing.T) {
	repo := config.Repo{Type: config.RepoTypeHermes, URL: "https://hermes-agent.nousresearch.com/docs/api/skills-index.json"}
	want := "https://hermes-agent.nousresearch.com/docs/api/skills-index.json"
	if got := repoDisplayURL(repo); got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}
