package cmd

import (
	"testing"
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
