package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadFromFile_ValidConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".ailign.yml")
	_ = os.WriteFile(path, []byte("targets:\n  - claude\n  - cursor\n"), 0644)

	cfg, err := LoadFromFile(path)
	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, []string{"claude", "cursor"}, cfg.Targets)
}

func TestLoadFromFile_SingleTarget(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".ailign.yml")
	_ = os.WriteFile(path, []byte("targets:\n  - copilot\n"), 0644)

	cfg, err := LoadFromFile(path)
	require.NoError(t, err)
	assert.Equal(t, []string{"copilot"}, cfg.Targets)
}

func TestLoadFromFile_AllTargets(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".ailign.yml")
	_ = os.WriteFile(path, []byte("targets:\n  - claude\n  - cursor\n  - copilot\n  - windsurf\n"), 0644)

	cfg, err := LoadFromFile(path)
	require.NoError(t, err)
	assert.Len(t, cfg.Targets, 4)
}

func TestLoadFromFile_MissingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".ailign.yml")

	cfg, err := LoadFromFile(path)
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "not found")
}

func TestLoadFromFile_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".ailign.yml")
	_ = os.WriteFile(path, []byte(""), 0644)

	cfg, err := LoadFromFile(path)
	require.NoError(t, err)
	require.NotNil(t, cfg)
	// Empty file parses to zero-value Config (nil targets)
	assert.Nil(t, cfg.Targets)
}

func TestLoadFromFile_YAMLParseError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".ailign.yml")
	_ = os.WriteFile(path, []byte("targets:\n  - claude\n bad: [yaml\n"), 0644)

	cfg, err := LoadFromFile(path)
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "parsing")
}

func TestLoadFromFile_TabsAsIndentation(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".ailign.yml")
	// YAML spec forbids tabs for indentation; parser correctly rejects them
	_ = os.WriteFile(path, []byte("targets:\n\t- claude\n\t- cursor\n"), 0644)

	cfg, err := LoadFromFile(path)
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "parsing")
}

func TestLoadFromFile_TabsInValues(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".ailign.yml")
	// Tabs within quoted scalar values are valid YAML (only indentation tabs are forbidden)
	_ = os.WriteFile(path, []byte("targets:\n  - \"clau\tde\"\n"), 0644)

	cfg, err := LoadFromFile(path)
	require.NoError(t, err)
	assert.Equal(t, []string{"clau\tde"}, cfg.Targets)
}

func TestLoadFromFile_UnicodeBOM(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".ailign.yml")
	// UTF-8 BOM + valid YAML
	bom := []byte{0xEF, 0xBB, 0xBF}
	content := append(bom, []byte("targets:\n  - claude\n")...)
	_ = os.WriteFile(path, content, 0644)

	cfg, err := LoadFromFile(path)
	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, []string{"claude"}, cfg.Targets)
}

func TestLoadFromFile_Symlink(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink test not reliable on Windows without elevated privileges")
	}

	dir := t.TempDir()
	realPath := filepath.Join(dir, "real-config.yml")
	linkPath := filepath.Join(dir, ".ailign.yml")
	_ = os.WriteFile(realPath, []byte("targets:\n  - windsurf\n"), 0644)
	require.NoError(t, os.Symlink(realPath, linkPath))

	cfg, err := LoadFromFile(linkPath)
	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, []string{"windsurf"}, cfg.Targets)
}

func TestLoadFromFile_PermissionError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("file permission test not reliable on Windows")
	}

	dir := t.TempDir()
	path := filepath.Join(dir, ".ailign.yml")
	_ = os.WriteFile(path, []byte("targets:\n  - claude\n"), 0000)

	cfg, err := LoadFromFile(path)
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "reading")
}

func TestLoadFromFile_UnknownFieldsPreserved(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".ailign.yml")
	_ = os.WriteFile(path, []byte("targets:\n  - claude\ncustom_field: hello\n"), 0644)

	// Loader should parse without error even with unknown fields
	// (unknown field detection is a validator concern, not loader)
	cfg, err := LoadFromFile(path)
	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, []string{"claude"}, cfg.Targets)
}

// ---------------------------------------------------------------------------
// LoadAndValidate() tests
// ---------------------------------------------------------------------------

func TestLoadAndValidate_ValidConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".ailign.yml")
	_ = os.WriteFile(path, []byte("targets:\n  - claude\n  - cursor\n"), 0644)

	result := LoadAndValidate(path)

	require.NotNil(t, result)
	assert.True(t, result.Valid)
	assert.Empty(t, result.Errors)
	assert.Empty(t, result.Warnings)
	require.NotNil(t, result.Config)
	assert.Equal(t, []string{"claude", "cursor"}, result.Config.Targets)
}

func TestLoadAndValidate_MissingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".ailign.yml")

	result := LoadAndValidate(path)

	require.NotNil(t, result)
	assert.False(t, result.Valid)
	require.NotEmpty(t, result.Errors)
	assert.Contains(t, result.Errors[0].Message, "not found")
	assert.Contains(t, result.Errors[0].Remediation, "ailign init")
}

func TestLoadAndValidate_PermissionError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("file permission test not reliable on Windows")
	}

	dir := t.TempDir()
	path := filepath.Join(dir, ".ailign.yml")
	_ = os.WriteFile(path, []byte("targets:\n  - claude\n"), 0000)

	result := LoadAndValidate(path)

	require.NotNil(t, result)
	assert.False(t, result.Valid)
	require.NotEmpty(t, result.Errors)
	assert.Contains(t, result.Errors[0].Message, "reading config")
}

func TestLoadAndValidate_YAMLParseError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".ailign.yml")
	_ = os.WriteFile(path, []byte("targets:\n  - claude\n bad: [yaml\n"), 0644)

	result := LoadAndValidate(path)

	require.NotNil(t, result)
	assert.False(t, result.Valid)
	require.NotEmpty(t, result.Errors)
	assert.Contains(t, result.Errors[0].Message, "parsing config")
}

func TestLoadAndValidate_InvalidConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".ailign.yml")
	_ = os.WriteFile(path, []byte("targets:\n  - vscode\n"), 0644)

	result := LoadAndValidate(path)

	require.NotNil(t, result)
	assert.False(t, result.Valid)
	require.NotEmpty(t, result.Errors)
}

func TestLoadAndValidate_UnknownFieldsAsWarnings(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".ailign.yml")
	_ = os.WriteFile(path, []byte("targets:\n  - claude\ncustom_field: value\n"), 0644)

	result := LoadAndValidate(path)

	require.NotNil(t, result)
	assert.True(t, result.Valid)
	require.NotEmpty(t, result.Warnings)
	assert.Equal(t, "custom_field", result.Warnings[0].FieldPath)
	assert.Equal(t, "warning", result.Warnings[0].Severity)
}

func TestLoadAndValidate_UnicodeBOM(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".ailign.yml")
	bom := []byte{0xEF, 0xBB, 0xBF}
	content := append(bom, []byte("targets:\n  - claude\n")...)
	_ = os.WriteFile(path, content, 0644)

	result := LoadAndValidate(path)

	require.NotNil(t, result)
	assert.True(t, result.Valid)
	require.NotNil(t, result.Config)
	assert.Equal(t, []string{"claude"}, result.Config.Targets)
}

func TestLoadAndValidate_EmptyTargets(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".ailign.yml")
	_ = os.WriteFile(path, []byte("targets: []\n"), 0644)

	result := LoadAndValidate(path)

	require.NotNil(t, result)
	assert.False(t, result.Valid)
	require.NotEmpty(t, result.Errors)
}
