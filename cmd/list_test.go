package cmd

import (
	"testing"
)

func TestListConfiguration(t *testing.T) {
	if listCmd.Flags().Lookup("all") == nil {
		t.Error("listCmd missing 'all' flag")
	}
	if listCmd.Flags().Lookup("agent") == nil {
		t.Error("listCmd missing 'agent' flag")
	}

	if listRootCmd.Flags().Lookup("all") == nil {
		t.Error("listRootCmd missing 'all' flag")
	}
}

func TestListRootCommand(t *testing.T) {
	if listRootCmd.Use != "list" {
		t.Errorf("Expected use 'list', got '%s'", listRootCmd.Use)
	}
}
