package installer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yeasy/ask/internal/cache"
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

func TestResolveTwoPartInstallTarget(t *testing.T) {
	t.Run("cache hit uses cached source path and index URL", func(t *testing.T) {
		t.Setenv("HOME", t.TempDir())
		repoRoot := setupCachedSkill(t, "anthropics-skills", "browser-use", "https://github.com/anthropics/skills.git")

		got, ok, err := resolveTwoPartInstallTarget("anthropics-skills/browser-use", InstallOptions{})

		assert.NoError(t, err)
		assert.True(t, ok)
		assert.Equal(t, installTarget{
			repoURL:         "https://github.com/anthropics/skills.git",
			subDir:          "browser-use",
			skillName:       "browser-use",
			originalInput:   "anthropics-skills/browser-use",
			input:           "anthropics-skills/browser-use",
			localSourcePath: filepath.Join(repoRoot, "browser-use"),
		}, got)
	})

	t.Run("configured repo alias with base subdirectory", func(t *testing.T) {
		got, ok, err := resolveTwoPartInstallTarget("anthropics/browser-use", InstallOptions{
			Config: &config.Config{Repos: []config.Repo{{Name: "anthropics", Type: config.RepoTypeDir, URL: "anthropics/skills/skills"}}},
		})

		assert.NoError(t, err)
		assert.True(t, ok)
		assert.Equal(t, installTarget{
			repoURL:       "https://github.com/anthropics/skills.git",
			subDir:        filepath.Join("skills", "browser-use"),
			skillName:     "browser-use",
			originalInput: "anthropics/browser-use",
			input:         "anthropics/browser-use",
		}, got)
	})

	t.Run("standard owner repo fallback", func(t *testing.T) {
		got, ok, err := resolveTwoPartInstallTarget("owner/repo", InstallOptions{})

		assert.NoError(t, err)
		assert.True(t, ok)
		assert.Equal(t, installTarget{
			repoURL:       "https://github.com/owner/repo",
			skillName:     "repo",
			originalInput: "owner/repo",
			input:         "owner/repo",
		}, got)
	})
}

func TestResolveBareCachedSkillInput(t *testing.T) {
	t.Run("single exact cache match resolves to repo skill input", func(t *testing.T) {
		t.Setenv("HOME", t.TempDir())
		setupCachedSkill(t, "anthropics-skills", "browser-use", "https://github.com/anthropics/skills.git")

		got, ok, err := resolveBareCachedSkillInput("browser-use")

		assert.NoError(t, err)
		assert.True(t, ok)
		assert.Equal(t, "anthropics-skills/browser-use", got)
	})

	t.Run("ambiguous exact cache matches return existing ambiguity error", func(t *testing.T) {
		t.Setenv("HOME", t.TempDir())
		setupCachedSkill(t, "repo-one", "browser-use", "https://github.com/acme/one.git")
		setupCachedSkill(t, "repo-two", "browser-use", "https://github.com/acme/two.git")

		got, ok, err := resolveBareCachedSkillInput("browser-use")

		assert.Error(t, err)
		assert.False(t, ok)
		assert.Empty(t, got)
		assert.EqualError(t, err, "ambiguous skill name 'browser-use'. Please specify the repository like 'RepoName/SkillName'")
	})
}

func TestInstall_UsesResolverHelperSeams(t *testing.T) {
	t.Run("two-part input installs from resolved target", func(t *testing.T) {
		t.Setenv("HOME", t.TempDir())
		source := setupLocalSkillSource(t, "resolved-skill")
		called := false
		oldTwoPart := resolveTwoPartInstallTargetForInstall
		resolveTwoPartInstallTargetForInstall = func(input string, opts InstallOptions) (installTarget, bool, error) {
			called = true
			assert.Equal(t, "repo/resolved-skill", input)
			return installTarget{repoURL: "", skillName: "resolved-skill", localSourcePath: source, originalInput: input, input: input}, true, nil
		}
		defer func() { resolveTwoPartInstallTargetForInstall = oldTwoPart }()

		err := Install("repo/resolved-skill", InstallOptions{Global: true, SkipScore: true})

		assert.NoError(t, err)
		assert.True(t, called)
	})

	t.Run("bare cached input recurses through resolved two-part input", func(t *testing.T) {
		t.Setenv("HOME", t.TempDir())
		source := setupLocalSkillSource(t, "browser-use")
		bareCalled := false
		twoPartCalled := false
		oldBare := resolveBareCachedSkillInputForInstall
		oldTwoPart := resolveTwoPartInstallTargetForInstall
		resolveBareCachedSkillInputForInstall = func(input string) (string, bool, error) {
			bareCalled = true
			assert.Equal(t, "browser-use", input)
			return "anthropics-skills/browser-use", true, nil
		}
		resolveTwoPartInstallTargetForInstall = func(input string, opts InstallOptions) (installTarget, bool, error) {
			twoPartCalled = true
			assert.Equal(t, "anthropics-skills/browser-use", input)
			assert.Equal(t, 1, opts.depth)
			return installTarget{repoURL: "", skillName: "browser-use", localSourcePath: source, originalInput: input, input: input}, true, nil
		}
		defer func() {
			resolveBareCachedSkillInputForInstall = oldBare
			resolveTwoPartInstallTargetForInstall = oldTwoPart
		}()

		err := Install("browser-use", InstallOptions{Global: true, SkipScore: true})

		assert.NoError(t, err)
		assert.True(t, bareCalled)
		assert.True(t, twoPartCalled)
	})
}

func TestInstall_HermesAgentWritesProvenanceLockMetadata(t *testing.T) {
	home := t.TempDir()
	hermesHome := filepath.Join(t.TempDir(), "hermes-home")
	t.Setenv("HOME", home)
	t.Setenv("HERMES_HOME", hermesHome)
	source := setupLocalSkillSource(t, "gitnexus-explorer")

	oldTwoPart := resolveTwoPartInstallTargetForInstall
	resolveTwoPartInstallTargetForInstall = func(input string, opts InstallOptions) (installTarget, bool, error) {
		assert.Equal(t, "official/gitnexus-explorer", input)
		return installTarget{
			repoURL:         "",
			skillName:       "gitnexus-explorer",
			localSourcePath: source,
			originalInput:   input,
			input:           input,
			sourceMetadata: &InstallSourceMetadata{
				Source:           config.RepoTypeHermes,
				SourceIdentifier: "official/research/gitnexus-explorer",
				UpdateStrategy:   "hermes-index",
			},
		}, true, nil
	}
	defer func() { resolveTwoPartInstallTargetForInstall = oldTwoPart }()

	err := Install("official/gitnexus-explorer", InstallOptions{
		Global:    true,
		Agents:    []string{"hermes"},
		SkipScore: true,
	})

	assert.NoError(t, err)
	lockFile, lockErr := config.LoadGlobalLockFile()
	assert.NoError(t, lockErr)
	entry := lockFile.GetEntryForAgentTargetPath("gitnexus-explorer", "hermes", filepath.Join(hermesHome, "skills", "gitnexus-explorer"))
	if assert.NotNil(t, entry) {
		assert.Equal(t, "hermes", entry.Agent)
		assert.Equal(t, "ask", entry.Ownership)
		assert.Equal(t, "ask-cache", entry.InstallMode)
		assert.Equal(t, "hermes-index", entry.UpdateStrategy)
		assert.Equal(t, "official/research/gitnexus-explorer", entry.SourceIdentifier)
		assert.Equal(t, filepath.Join(home, ".ask", "skills", "gitnexus-explorer"), entry.SourcePath)
		assert.Equal(t, filepath.Join(hermesHome, "skills", "gitnexus-explorer"), entry.TargetPath)
		assert.True(t, strings.HasPrefix(entry.Checksum, "sha256:"), entry.Checksum)
	}
}

func TestInstall_HermesAgentInfersCachedOfficialOptionalProvenance(t *testing.T) {
	home := t.TempDir()
	hermesHome := filepath.Join(t.TempDir(), "hermes-home")
	t.Setenv("HOME", home)
	t.Setenv("HERMES_HOME", hermesHome)
	setupCachedSkill(t, "official", "gitnexus-explorer", "NousResearch/hermes-agent/optional-skills/research")

	err := Install("official/gitnexus-explorer", InstallOptions{
		Global:    true,
		Agents:    []string{"hermes"},
		SkipScore: true,
	})

	assert.NoError(t, err)
	lockFile, lockErr := config.LoadGlobalLockFile()
	assert.NoError(t, lockErr)
	entry := lockFile.GetEntryForAgentTargetPath("gitnexus-explorer", "hermes", filepath.Join(hermesHome, "skills", "gitnexus-explorer"))
	if assert.NotNil(t, entry) {
		assert.Equal(t, config.RepoTypeHermes, entry.Source)
		assert.Equal(t, "official/gitnexus-explorer", entry.SourceIdentifier)
		assert.Equal(t, "hermes-index", entry.UpdateStrategy)
		assert.Equal(t, filepath.Join(home, ".ask", "skills", "gitnexus-explorer"), entry.SourcePath)
		assert.Equal(t, filepath.Join(hermesHome, "skills", "gitnexus-explorer"), entry.TargetPath)
		assert.True(t, strings.HasPrefix(entry.Checksum, "sha256:"), entry.Checksum)
	}
}

func TestDirectoryChecksumIsDeterministicAndIgnoresGitAndSymlinks(t *testing.T) {
	rootA := t.TempDir()
	rootB := t.TempDir()
	writeChecksumFixture(t, rootA)
	writeChecksumFixtureReverse(t, rootB)

	checksumA, err := directoryChecksum(rootA)
	assert.NoError(t, err)
	checksumB, err := directoryChecksum(rootB)
	assert.NoError(t, err)
	assert.Equal(t, checksumA, checksumB)

	assert.NoError(t, os.WriteFile(filepath.Join(rootB, ".git", "ignored"), []byte("changed"), 0644))
	if err := os.Symlink(filepath.Join(rootB, "missing-target"), filepath.Join(rootB, "ignored-link")); err == nil {
		checksumAfterIgnoredChanges, checksumErr := directoryChecksum(rootB)
		assert.NoError(t, checksumErr)
		assert.Equal(t, checksumA, checksumAfterIgnoredChanges)
	}
}

func writeChecksumFixture(t *testing.T, root string) {
	t.Helper()
	assert.NoError(t, os.MkdirAll(filepath.Join(root, "nested"), 0755))
	assert.NoError(t, os.WriteFile(filepath.Join(root, "b.txt"), []byte("bravo"), 0644))
	assert.NoError(t, os.WriteFile(filepath.Join(root, "nested", "a.txt"), []byte("alpha"), 0644))
	assert.NoError(t, os.MkdirAll(filepath.Join(root, ".git"), 0755))
	assert.NoError(t, os.WriteFile(filepath.Join(root, ".git", "ignored"), []byte("ignored"), 0644))
}

func writeChecksumFixtureReverse(t *testing.T, root string) {
	t.Helper()
	assert.NoError(t, os.MkdirAll(filepath.Join(root, ".git"), 0755))
	assert.NoError(t, os.WriteFile(filepath.Join(root, ".git", "ignored"), []byte("different ignored content"), 0644))
	assert.NoError(t, os.MkdirAll(filepath.Join(root, "nested"), 0755))
	assert.NoError(t, os.WriteFile(filepath.Join(root, "nested", "a.txt"), []byte("alpha"), 0644))
	assert.NoError(t, os.WriteFile(filepath.Join(root, "b.txt"), []byte("bravo"), 0644))
}

func setupLocalSkillSource(t *testing.T, skillName string) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), skillName)
	assert.NoError(t, os.MkdirAll(dir, 0755))
	assert.NoError(t, os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("---\nname: "+skillName+"\ndescription: test skill\n---\n"), 0644))
	return dir
}

func setupCachedSkill(t *testing.T, repoName, skillName, repoURL string) string {
	t.Helper()
	reposCache, err := cache.NewReposCache()
	assert.NoError(t, err)

	repoRoot := filepath.Join(os.Getenv("HOME"), ".ask", "repos", repoName)
	assert.NoError(t, os.MkdirAll(filepath.Join(repoRoot, skillName), 0755))
	assert.NoError(t, os.WriteFile(filepath.Join(repoRoot, skillName, "SKILL.md"), []byte("---\nname: "+skillName+"\n---\n"), 0644))

	assert.NoError(t, reposCache.SaveIndexWithStars(map[string]int{repoName: 0}, map[string]string{repoName: repoURL}))
	return repoRoot
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
