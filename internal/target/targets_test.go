package target

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// Per-target implementation tests (T014)
// ---------------------------------------------------------------------------

func TestClaude_Name(t *testing.T) {
	assert.Equal(t, "claude", Claude{}.Name())
}

func TestClaude_InstructionPath(t *testing.T) {
	assert.Equal(t, ".claude/instructions.md", Claude{}.InstructionPath())
}

func TestCursor_Name(t *testing.T) {
	assert.Equal(t, "cursor", Cursor{}.Name())
}

func TestCursor_InstructionPath(t *testing.T) {
	assert.Equal(t, ".cursorrules", Cursor{}.InstructionPath())
}

func TestCopilot_Name(t *testing.T) {
	assert.Equal(t, "copilot", Copilot{}.Name())
}

func TestCopilot_InstructionPath(t *testing.T) {
	assert.Equal(t, ".github/copilot-instructions.md", Copilot{}.InstructionPath())
}

func TestWindsurf_Name(t *testing.T) {
	assert.Equal(t, "windsurf", Windsurf{}.Name())
}

func TestWindsurf_InstructionPath(t *testing.T) {
	assert.Equal(t, ".windsurfrules", Windsurf{}.InstructionPath())
}

func TestAllTargets_ImplementInterface(t *testing.T) {
	// Compile-time check that all types implement Target
	var targets []Target
	targets = append(targets, Claude{}, Cursor{}, Copilot{}, Windsurf{})

	for _, tgt := range targets {
		assert.NotEmpty(t, tgt.Name(), "Name() should not be empty")
		assert.NotEmpty(t, tgt.InstructionPath(), "InstructionPath() should not be empty")
	}
}
