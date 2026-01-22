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
		"Antigravity",
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

func TestParseGitHubBrowserURL(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		wantRepoURL   string
		wantBranch    string
		wantSubDir    string
		wantSkillName string
		wantOK        bool
	}{
		{
			name:          "full URL with subdirectory",
			input:         "https://github.com/anthropics/skills/tree/main/skills/mcp-builder",
			wantRepoURL:   "https://github.com/anthropics/skills.git",
			wantBranch:    "main",
			wantSubDir:    "skills/mcp-builder",
			wantSkillName: "mcp-builder",
			wantOK:        true,
		},
		{
			name:          "URL with different branch",
			input:         "https://github.com/owner/repo/tree/develop/path/to/skill",
			wantRepoURL:   "https://github.com/owner/repo.git",
			wantBranch:    "develop",
			wantSubDir:    "path/to/skill",
			wantSkillName: "skill",
			wantOK:        true,
		},
		{
			name:          "URL without subdirectory - just branch",
			input:         "https://github.com/owner/repo/tree/main",
			wantRepoURL:   "https://github.com/owner/repo.git",
			wantBranch:    "main",
			wantSubDir:    "",
			wantSkillName: "repo",
			wantOK:        true,
		},
		{
			name:          "URL with trailing slash",
			input:         "https://github.com/anthropics/skills/tree/main/skills/mcp-builder/",
			wantRepoURL:   "https://github.com/anthropics/skills.git",
			wantBranch:    "main",
			wantSubDir:    "skills/mcp-builder",
			wantSkillName: "mcp-builder",
			wantOK:        true,
		},
		{
			name:   "non-tree URL (regular git URL)",
			input:  "https://github.com/owner/repo.git",
			wantOK: false,
		},
		{
			name:   "shorthand format - not a browser URL",
			input:  "owner/repo/path/to/skill",
			wantOK: false,
		},
		{
			name:   "empty string",
			input:  "",
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRepoURL, gotBranch, gotSubDir, gotSkillName, gotOK := parseGitHubBrowserURL(tt.input)

			if gotOK != tt.wantOK {
				t.Errorf("parseGitHubBrowserURL() ok = %v, want %v", gotOK, tt.wantOK)
				return
			}

			if !tt.wantOK {
				return // No need to check other fields if we expected failure
			}

			if gotRepoURL != tt.wantRepoURL {
				t.Errorf("parseGitHubBrowserURL() repoURL = %v, want %v", gotRepoURL, tt.wantRepoURL)
			}
			if gotBranch != tt.wantBranch {
				t.Errorf("parseGitHubBrowserURL() branch = %v, want %v", gotBranch, tt.wantBranch)
			}
			if gotSubDir != tt.wantSubDir {
				t.Errorf("parseGitHubBrowserURL() subDir = %v, want %v", gotSubDir, tt.wantSubDir)
			}
			if gotSkillName != tt.wantSkillName {
				t.Errorf("parseGitHubBrowserURL() skillName = %v, want %v", gotSkillName, tt.wantSkillName)
			}
		})
	}
}
