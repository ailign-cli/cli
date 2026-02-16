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
