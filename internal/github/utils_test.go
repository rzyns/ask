package github

import (
	"testing"
)

func TestParseBrowserURL(t *testing.T) {
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
			gotRepoURL, gotBranch, gotSubDir, gotSkillName, gotOK := ParseBrowserURL(tt.input)

			if gotOK != tt.wantOK {
				t.Errorf("ParseBrowserURL() ok = %v, want %v", gotOK, tt.wantOK)
				return
			}

			if !tt.wantOK {
				return // No need to check other fields if we expected failure
			}

			if gotRepoURL != tt.wantRepoURL {
				t.Errorf("ParseBrowserURL() repoURL = %v, want %v", gotRepoURL, tt.wantRepoURL)
			}
			if gotBranch != tt.wantBranch {
				t.Errorf("ParseBrowserURL() branch = %v, want %v", gotBranch, tt.wantBranch)
			}
			if gotSubDir != tt.wantSubDir {
				t.Errorf("ParseBrowserURL() subDir = %v, want %v", gotSubDir, tt.wantSubDir)
			}
			if gotSkillName != tt.wantSkillName {
				t.Errorf("ParseBrowserURL() skillName = %v, want %v", gotSkillName, tt.wantSkillName)
			}
		})
	}
}
