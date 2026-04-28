package cmd

import (
	"bytes"
	"strings"
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

func TestRootHelpShowsSubcommandDetails(t *testing.T) {
	rootCmd.SetArgs([]string{"--help"})

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)

	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("root command failed: %v", err)
	}

	output := buf.String()

	// Verify subcommand details section is present
	subcommandDetails := []string{
		"Skill Commands (ask skill <command>):",
		"Repository Commands (ask repo <command>):",
		"Supported Agents:",
		"search",
		"install",
		"uninstall",
		"Claude",
		"Cursor",
		"antigravity",
	}

	for _, detail := range subcommandDetails {
		if !bytes.Contains([]byte(output), []byte(detail)) {
			t.Errorf("expected help to contain subcommand detail '%s'", detail)
		}
	}

	// Verify subcommand details appear AFTER Flags section
	flagsIndex := bytes.Index([]byte(output), []byte("Flags:"))
	skillCmdsIndex := bytes.Index([]byte(output), []byte("Skill Commands"))

	if flagsIndex == -1 {
		t.Error("expected help to contain 'Flags:' section")
	}
	if skillCmdsIndex == -1 {
		t.Error("expected help to contain 'Skill Commands' section")
	}
	if flagsIndex > skillCmdsIndex {
		t.Error("expected 'Skill Commands' to appear AFTER 'Flags:' section")
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

func TestSubcommandHelpDoesNotShowRootCatalog(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		contains []string
	}{
		{
			name: "init short help",
			args: []string{"init", "-h"},
			contains: []string{
				"Initialize a new Agent Skills Kit project",
				"Usage:\n  ask init [flags]",
				"-y, --yes",
			},
		},
		{
			name: "init long help",
			args: []string{"init", "--help"},
			contains: []string{
				"Initialize a new Agent Skills Kit project",
				"Usage:\n  ask init [flags]",
				"-y, --yes",
			},
		},
		{
			name: "nested skill install help",
			args: []string{"skill", "install", "--help"},
			contains: []string{
				"Download and install skills into agent-specific directories",
				"Usage:\n  ask skill install [url...] [flags]",
				"--min-score string",
			},
		},
		{
			name: "nested repo add help",
			args: []string{"repo", "add", "--help"},
			contains: []string{
				"Add a GitHub repository as a skill source",
				"Usage:\n  ask repo add <owner/repo|URL> [flags]",
				"--sync",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := rootCmd
			cmd.SetArgs(tt.args)

			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)

			if err := cmd.Execute(); err != nil {
				t.Fatalf("command help failed: %v", err)
			}

			output := buf.String()
			for _, want := range tt.contains {
				if !strings.Contains(output, want) {
					t.Errorf("expected help output to contain %q\noutput:\n%s", want, output)
				}
			}

			for _, forbidden := range []string{
				"Skill Commands (ask skill <command>):",
				"Repository Commands (ask repo <command>):",
				"System Commands:",
				"Supported Agents:",
			} {
				if strings.Contains(output, forbidden) {
					t.Errorf("expected subcommand help not to contain root catalog section %q\noutput:\n%s", forbidden, output)
				}
			}
		})
	}
}
