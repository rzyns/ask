package installer

import (
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

func TestInstall_Ambiguous(_ *testing.T) {
	// This test depends on cache state, so we might need to mock cache
	// or we accept that without cache it might try to define it as repo
	// For now, let's test basic validation logic
}

func TestInstallOptions(t *testing.T) {
	cfg := &config.Config{}
	opts := InstallOptions{
		Global: true,
		Agents: []string{"claude"},
		Config: cfg,
	}

	assert.True(t, opts.Global)
	assert.Len(t, opts.Agents, 1)
	assert.Equal(t, "claude", opts.Agents[0])
	assert.Equal(t, cfg, opts.Config)
}
