package target

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Registry tests (T013)
// ---------------------------------------------------------------------------

func TestRegistry_Register_And_Get(t *testing.T) {
	r := NewRegistry()
	r.Register(Claude{})

	got, ok := r.Get("claude")
	assert.True(t, ok)
	assert.Equal(t, "claude", got.Name())
	assert.Equal(t, ".claude/instructions.md", got.InstructionPath())
}

func TestRegistry_Get_Unknown(t *testing.T) {
	r := NewRegistry()

	_, ok := r.Get("vscode")
	assert.False(t, ok)
}

func TestRegistry_IsValid(t *testing.T) {
	r := NewDefaultRegistry()

	assert.True(t, r.IsValid("claude"))
	assert.True(t, r.IsValid("cursor"))
	assert.True(t, r.IsValid("copilot"))
	assert.True(t, r.IsValid("windsurf"))
	assert.False(t, r.IsValid("vscode"))
	assert.False(t, r.IsValid(""))
	assert.False(t, r.IsValid("Claude"))
}

func TestRegistry_KnownTargets(t *testing.T) {
	r := NewDefaultRegistry()

	targets := r.KnownTargets()
	assert.Len(t, targets, 4)
	// Sorted alphabetically
	assert.Equal(t, []string{"claude", "copilot", "cursor", "windsurf"}, targets)
}

func TestRegistry_KnownTargets_ReturnsNewSlice(t *testing.T) {
	r := NewDefaultRegistry()

	a := r.KnownTargets()
	b := r.KnownTargets()
	a[0] = "modified"
	assert.NotEqual(t, a[0], b[0], "KnownTargets should return a new slice each time")
}

func TestRegistry_DuplicateRegistration(t *testing.T) {
	r := NewRegistry()
	r.Register(Claude{})
	r.Register(Claude{}) // re-register same target

	targets := r.KnownTargets()
	assert.Len(t, targets, 1, "duplicate registration should overwrite, not duplicate")
}

func TestNewDefaultRegistry_HasAllTargets(t *testing.T) {
	r := NewDefaultRegistry()

	for _, name := range []string{"claude", "cursor", "copilot", "windsurf"} {
		got, ok := r.Get(name)
		require.True(t, ok, "default registry should contain %q", name)
		assert.Equal(t, name, got.Name())
		assert.NotEmpty(t, got.InstructionPath(), "InstructionPath for %q should not be empty", name)
	}
}

// ---------------------------------------------------------------------------
// Package-level convenience function tests
// ---------------------------------------------------------------------------

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

// ---------------------------------------------------------------------------
// Schema-registry invariant: registered targets match schema enum (T013)
// ---------------------------------------------------------------------------

func TestSchemaRegistryInvariant(t *testing.T) {
	// Read the embedded schema from disk (relative to test working dir)
	schemaBytes, err := os.ReadFile("../config/schema.json")
	require.NoError(t, err, "failed to read schema.json")

	var schema struct {
		Properties struct {
			Targets struct {
				Items struct {
					Enum []string `json:"enum"`
				} `json:"items"`
			} `json:"targets"`
		} `json:"properties"`
	}
	err = json.Unmarshal(schemaBytes, &schema)
	require.NoError(t, err, "failed to parse schema JSON")

	schemaTargets := schema.Properties.Targets.Items.Enum
	require.NotEmpty(t, schemaTargets, "schema should have target enum values")

	registryTargets := NewDefaultRegistry().KnownTargets()

	// Sort schema targets for comparison
	sortStrings(schemaTargets)

	assert.Equal(t, schemaTargets, registryTargets,
		"registry targets must match schema enum — update both when adding a target")
}
