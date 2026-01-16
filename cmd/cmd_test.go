package cmd

import (
	"bytes"
	"testing"
)

func TestRootCommand(t *testing.T) {
	// Reset args before test
	rootCmd.SetArgs([]string{"--help"})

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)

	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("root command failed: %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Error("expected help output, got empty string")
	}

	// Verify key sections are present in help
	expectedSections := []string{
		"skill",
		"repo",
		"init",
	}

	for _, section := range expectedSections {
		if !bytes.Contains([]byte(output), []byte(section)) {
			t.Errorf("expected help to contain '%s'", section)
		}
	}
}

func TestSkillCommandHelp(t *testing.T) {
	// Create a new root command for testing to avoid state pollution
	cmd := rootCmd
	cmd.SetArgs([]string{"skill", "--help"})

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err != nil {
		t.Errorf("skill command failed: %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Error("expected skill help output, got empty string")
	}

	// Verify at least some subcommands are listed
	// Note: We check for a few key ones rather than all to make test less brittle
	expectedCommands := []string{
		"search",
		"install",
		"list",
	}

	for _, cmd := range expectedCommands {
		if !bytes.Contains([]byte(output), []byte(cmd)) {
			t.Errorf("expected skill help to contain '%s' command", cmd)
		}
	}
}

func TestRepoCommandHelp(t *testing.T) {
	cmd := rootCmd
	cmd.SetArgs([]string{"repo", "--help"})

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err != nil {
		t.Errorf("repo command failed: %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Error("expected repo help output, got empty string")
	}

	// Verify subcommands are listed
	expectedCommands := []string{
		"add",
		"list",
	}

	for _, cmd := range expectedCommands {
		if !bytes.Contains([]byte(output), []byte(cmd)) {
			t.Errorf("expected repo help to contain '%s' command", cmd)
		}
	}
}

func TestInitCommandHelp(t *testing.T) {
	cmd := rootCmd
	cmd.SetArgs([]string{"init", "--help"})

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err != nil {
		t.Errorf("init command help failed: %v", err)
	}

	output := buf.String()
	if !bytes.Contains([]byte(output), []byte("Initialize")) {
		t.Error("expected init help to contain 'Initialize'")
	}
}
