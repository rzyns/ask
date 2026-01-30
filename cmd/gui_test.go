package cmd

import (
	"testing"
)

func TestGuiCommand(t *testing.T) {
	if guiCmd == nil {
		t.Fatal("guiCmd is nil")
	}

	if guiCmd.Use != "gui" {
		t.Errorf("guiCmd.Use = %s; want gui", guiCmd.Use)
	}

	// Verify it's added to root
	found := false
	for _, c := range rootCmd.Commands() {
		if c == guiCmd {
			found = true
			break
		}
	}
	if !found {
		t.Error("guiCmd not added to rootCmd")
	}
}
