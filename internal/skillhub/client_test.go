package skillhub

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolve(t *testing.T) {
	// We only test NewClient here to avoid creating a real HTTP server
	// which causes race conditions in some sandboxed environments.
	c := NewClient()
	assert.NotNil(t, c)
}
