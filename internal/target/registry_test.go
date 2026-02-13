package target

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValid_KnownTargets(t *testing.T) {
	known := []string{"claude", "cursor", "copilot", "windsurf"}
	for _, name := range known {
		assert.True(t, IsValid(name), "expected %q to be a valid target", name)
	}
}

func TestIsValid_UnknownTargets(t *testing.T) {
	unknown := []string{"vscode", "vim", "emacs", "", "Claude", "CURSOR"}
	for _, name := range unknown {
		assert.False(t, IsValid(name), "expected %q to be an invalid target", name)
	}
}

func TestKnownTargets_ReturnsAllTargets(t *testing.T) {
	targets := KnownTargets()
	assert.Len(t, targets, 4)
	assert.Contains(t, targets, "claude")
	assert.Contains(t, targets, "cursor")
	assert.Contains(t, targets, "copilot")
	assert.Contains(t, targets, "windsurf")
}

func TestKnownTargets_ReturnsNewSlice(t *testing.T) {
	a := KnownTargets()
	b := KnownTargets()
	a[0] = "modified"
	assert.NotEqual(t, a[0], b[0], "KnownTargets should return a new slice each time")
}
