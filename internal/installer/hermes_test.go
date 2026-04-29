package installer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yeasy/ask/internal/config"
)

func TestInstallHermesRecordsAgentScopedProvenance(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	hermesHome := t.TempDir()
	t.Setenv("HERMES_HOME", hermesHome)
	source := setupLocalSkillSource(t, "hermes-skill")

	oldTwoPart := resolveTwoPartInstallTargetForInstall
	resolveTwoPartInstallTargetForInstall = func(input string, opts InstallOptions) (installTarget, bool, error) {
		assert.Equal(t, "repo/hermes-skill", input)
		return installTarget{repoURL: "https://example.test/repo.git", skillName: "hermes-skill", localSourcePath: source, originalInput: input, input: input}, true, nil
	}
	defer func() { resolveTwoPartInstallTargetForInstall = oldTwoPart }()

	err := Install("repo/hermes-skill", InstallOptions{Global: true, Agents: []string{"hermes"}, SkipScore: true})
	assert.NoError(t, err)

	lockFile, err := config.LoadGlobalLockFile()
	assert.NoError(t, err)
	entry := lockFile.GetEntryForAgent("hermes-skill", "hermes")
	if assert.NotNil(t, entry) {
		assert.Equal(t, "ask", entry.Ownership)
		assert.Equal(t, "ask-cache", entry.InstallMode)
		assert.Equal(t, "git", entry.UpdateStrategy)
		assert.NotEmpty(t, entry.SourcePath)
		assert.Equal(t, filepath.Join(hermesHome, "skills", "hermes-skill"), entry.TargetPath)
		assert.Contains(t, entry.Checksum, "sha256:")
	}
	if _, err := os.Lstat(filepath.Join(hermesHome, "skills", "hermes-skill")); err != nil {
		t.Fatalf("expected Hermes target to exist: %v", err)
	}
}

func TestInstallHermesRejectsBundledSource(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("HERMES_HOME", t.TempDir())

	err := Install("NousResearch/hermes-agent/skills/core", InstallOptions{Global: true, Agents: []string{"hermes"}, SkipScore: true})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "bundled")
}
