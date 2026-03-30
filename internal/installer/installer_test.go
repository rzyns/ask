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
