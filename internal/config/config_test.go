package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_TargetsField(t *testing.T) {
	cfg := Config{
		Targets: []string{"claude", "cursor"},
	}
	assert.Equal(t, []string{"claude", "cursor"}, cfg.Targets)
}

func TestConfig_EmptyTargets(t *testing.T) {
	cfg := Config{}
	assert.Nil(t, cfg.Targets)
}

func TestConfig_SingleTarget(t *testing.T) {
	cfg := Config{
		Targets: []string{"copilot"},
	}
	assert.Len(t, cfg.Targets, 1)
	assert.Equal(t, "copilot", cfg.Targets[0])
}

func TestConfig_AllTargets(t *testing.T) {
	cfg := Config{
		Targets: []string{"claude", "cursor", "copilot", "windsurf"},
	}
	assert.Len(t, cfg.Targets, 4)
}

// ---------------------------------------------------------------------------
// LocalOverlays field tests (T005)
// ---------------------------------------------------------------------------

func TestConfig_LocalOverlays_Single(t *testing.T) {
	cfg := Config{
		Targets:       []string{"claude"},
		LocalOverlays: []string{".ai-instructions/base.md"},
	}
	assert.Equal(t, []string{".ai-instructions/base.md"}, cfg.LocalOverlays)
}

func TestConfig_LocalOverlays_Multiple(t *testing.T) {
	cfg := Config{
		Targets: []string{"claude", "cursor"},
		LocalOverlays: []string{
			".ai-instructions/base.md",
			".ai-instructions/project-context.md",
		},
	}
	assert.Len(t, cfg.LocalOverlays, 2)
	assert.Equal(t, ".ai-instructions/base.md", cfg.LocalOverlays[0])
	assert.Equal(t, ".ai-instructions/project-context.md", cfg.LocalOverlays[1])
}

func TestConfig_LocalOverlays_Absent(t *testing.T) {
	cfg := Config{
		Targets: []string{"claude"},
	}
	assert.Nil(t, cfg.LocalOverlays)
}

func TestConfig_LocalOverlays_EmptySlice(t *testing.T) {
	cfg := Config{
		Targets:       []string{"claude"},
		LocalOverlays: []string{},
	}
	assert.Empty(t, cfg.LocalOverlays)
	assert.NotNil(t, cfg.LocalOverlays)
}
