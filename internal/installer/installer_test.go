package installer

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yeasy/ask/internal/config"
)

func TestInstall_InvalidInput(t *testing.T) {
	opts := InstallOptions{
		Global: false,
		Agents: []string{},
		Config: nil,
	}

	// Empty input
	err := Install("", opts)
	assert.Error(t, err)
	if err != nil {
		assert.Contains(t, err.Error(), "could not determine skill name")
	}

	// Input that resolves to empty name (e.g. "/")
	err = Install("/", opts)
	assert.Error(t, err)
}

func TestInstall_NilConfig(t *testing.T) {
	// Install with nil config should not panic
	opts := InstallOptions{
		Global: false,
		Agents: []string{},
		Config: nil,
	}

	err := Install("some-skill", opts)
	// Should fail gracefully (no config means no targets), not panic
	assert.Error(t, err)
}

func TestInstall_SpecialCharactersInInput(t *testing.T) {
	opts := InstallOptions{
		Config: &config.Config{},
	}

	// Path traversal attempt
	err := Install("../../etc/passwd", opts)
	assert.Error(t, err)

	// Very long input
	longInput := strings.Repeat("a", 300)
	err = Install(longInput, opts)
	assert.Error(t, err)
}

func TestResolveDirectInstallTarget(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  installTarget
		ok    bool
	}{
		{
			name:  "owner repo shorthand",
			input: "owner/repo",
			want: installTarget{
				repoURL:       "https://github.com/owner/repo",
				skillName:     "repo",
				originalInput: "owner/repo",
				input:         "owner/repo",
			},
			ok: true,
		},
		{
			name:  "owner repo subdirectory shorthand",
			input: "owner/repo/path/to/skill",
			want: installTarget{
				repoURL:       "https://github.com/owner/repo.git",
				subDir:        "path/to/skill",
				skillName:     "skill",
				originalInput: "owner/repo/path/to/skill",
				input:         "owner/repo/path/to/skill",
			},
			ok: true,
		},
		{
			name:  "github browser tree url",
			input: "https://github.com/owner/repo/tree/main/path/to/skill",
			want: installTarget{
				repoURL:       "https://github.com/owner/repo.git",
				subDir:        "path/to/skill",
				skillName:     "skill",
				branch:        "main",
				originalInput: "https://github.com/owner/repo/tree/main/path/to/skill",
				input:         "https://github.com/owner/repo/tree/main/path/to/skill",
			},
			ok: true,
		},
		{
			name:  "direct https git url",
			input: "https://example.com/team/example-skill.git",
			want: installTarget{
				repoURL:       "https://example.com/team/example-skill.git",
				skillName:     "example-skill",
				originalInput: "https://example.com/team/example-skill.git",
				input:         "https://example.com/team/example-skill.git",
			},
			ok: true,
		},
		{
			name:  "direct ssh git url",
			input: "git@github.com:owner/ssh-skill.git",
			want: installTarget{
				repoURL:       "git@github.com:owner/ssh-skill.git",
				skillName:     "ssh-skill",
				originalInput: "git@github.com:owner/ssh-skill.git",
				input:         "git@github.com:owner/ssh-skill.git",
			},
			ok: true,
		},
		{
			name:  "version suffix",
			input: "owner/repo@v1.2.3",
			want: installTarget{
				repoURL:       "https://github.com/owner/repo",
				skillName:     "repo",
				version:       "v1.2.3",
				originalInput: "owner/repo@v1.2.3",
				input:         "owner/repo",
			},
			ok: true,
		},
		{
			name:  "bare skill name is not direct",
			input: "browser-use",
			want: installTarget{
				originalInput: "browser-use",
				input:         "browser-use",
			},
			ok: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok, err := resolveDirectInstallTarget(tt.input)
			assert.NoError(t, err)
			assert.Equal(t, tt.ok, ok)
			assert.Equal(t, tt.want, got)
		})
	}
}
