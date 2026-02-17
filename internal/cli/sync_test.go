package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func skipSyncOnWindows(t *testing.T) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("symlink tests require elevated privileges on Windows")
	}
}

func writeOverlay(t *testing.T, dir, name, content string) {
	t.Helper()
	fullPath := filepath.Join(dir, name)
	require.NoError(t, os.MkdirAll(filepath.Dir(fullPath), 0755))
	require.NoError(t, os.WriteFile(fullPath, []byte(content), 0644))
}

func writeConfigWithOverlays(t *testing.T, dir string, targets, overlays []string) {
	t.Helper()
	var cfg string
	cfg += "targets:\n"
	for _, target := range targets {
		cfg += "  - " + target + "\n"
	}
	if len(overlays) > 0 {
		cfg += "local_overlays:\n"
		for _, overlay := range overlays {
			cfg += "  - " + overlay + "\n"
		}
	}
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".ailign.yml"), []byte(cfg), 0644))
}

// ---------------------------------------------------------------------------
// Sync command: basic success
// ---------------------------------------------------------------------------

func TestSync_BasicSuccess_ExitZero(t *testing.T) {
	skipSyncOnWindows(t)
	dir := t.TempDir()
	writeConfigWithOverlays(t, dir, []string{"claude", "cursor"}, []string{"base.md"})
	writeOverlay(t, dir, "base.md", "Use TypeScript strict mode\n")

	stdout, stderr, exitCode := executeCommand([]string{"sync"}, dir)

	assert.Equal(t, 0, exitCode, "sync should exit 0, stderr: %s", stderr)
	assert.Contains(t, stdout, "claude")
	assert.Contains(t, stdout, "cursor")
	assert.Contains(t, stdout, "Synced")

	// Verify hub file was created
	hubPath := filepath.Join(dir, ".ailign", "instructions.md")
	hubContent, err := os.ReadFile(hubPath)
	require.NoError(t, err, "hub file should exist")
	assert.Contains(t, string(hubContent), "Use TypeScript strict mode")

	// Verify symlinks exist
	for _, linkPath := range []string{".claude/instructions.md", ".cursorrules"} {
		info, err := os.Lstat(filepath.Join(dir, linkPath))
		require.NoError(t, err, "symlink should exist at %s", linkPath)
		assert.True(t, info.Mode()&os.ModeSymlink != 0, "%s should be a symlink", linkPath)
	}
}

// ---------------------------------------------------------------------------
// Sync command: no overlays configured
// ---------------------------------------------------------------------------

func TestSync_NoOverlays_ExitNonZero(t *testing.T) {
	dir := t.TempDir()
	// Config with targets but no local_overlays
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".ailign.yml"),
		[]byte("targets:\n  - claude\n"), 0644))

	_, stderr, exitCode := executeCommand([]string{"sync"}, dir)

	assert.NotEqual(t, 0, exitCode, "sync without overlays should fail")
	assert.Contains(t, stderr, "no local_overlays")
}

// ---------------------------------------------------------------------------
// Sync command: missing overlay file
// ---------------------------------------------------------------------------

func TestSync_MissingOverlay_ExitNonZero(t *testing.T) {
	dir := t.TempDir()
	writeConfigWithOverlays(t, dir, []string{"claude"}, []string{"missing.md"})

	_, stderr, exitCode := executeCommand([]string{"sync"}, dir)

	assert.NotEqual(t, 0, exitCode, "sync with missing overlay should fail")
	assert.Contains(t, stderr, "missing.md")
	assert.Contains(t, stderr, "not found")
}

// ---------------------------------------------------------------------------
// Sync command: JSON format
// ---------------------------------------------------------------------------

func TestSync_JSONFormat_ValidOutput(t *testing.T) {
	skipSyncOnWindows(t)
	dir := t.TempDir()
	writeConfigWithOverlays(t, dir, []string{"claude"}, []string{"base.md"})
	writeOverlay(t, dir, "base.md", "Instructions\n")

	stdout, _, exitCode := executeCommand([]string{"sync", "--format", "json"}, dir)

	assert.Equal(t, 0, exitCode)

	var result struct {
		DryRun bool `json:"dry_run"`
		Hub    struct {
			Path   string `json:"path"`
			Status string `json:"status"`
		} `json:"hub"`
		Links []struct {
			Target   string `json:"target"`
			LinkPath string `json:"link_path"`
			Status   string `json:"status"`
		} `json:"links"`
		Summary struct {
			Total   int `json:"total"`
			Created int `json:"created"`
		} `json:"summary"`
	}
	err := json.Unmarshal([]byte(stdout), &result)
	require.NoError(t, err, "stdout must be valid JSON: %s", stdout)

	assert.False(t, result.DryRun)
	assert.Equal(t, "written", result.Hub.Status)
	assert.Len(t, result.Links, 1)
	assert.Equal(t, "claude", result.Links[0].Target)
	assert.Equal(t, "created", result.Links[0].Status)
	assert.Equal(t, 1, result.Summary.Total)
	assert.Equal(t, 1, result.Summary.Created)
}

// ---------------------------------------------------------------------------
// Sync command: empty overlay warning
// ---------------------------------------------------------------------------

func TestSync_EmptyOverlay_WarningOnStderr(t *testing.T) {
	skipSyncOnWindows(t)
	dir := t.TempDir()
	writeConfigWithOverlays(t, dir, []string{"claude"}, []string{"empty.md"})
	writeOverlay(t, dir, "empty.md", "")

	stdout, stderr, exitCode := executeCommand([]string{"sync"}, dir)

	assert.Equal(t, 0, exitCode, "empty overlay should not fail sync")
	assert.Contains(t, stderr, "empty.md", "stderr should warn about empty overlay")
	assert.Contains(t, stdout, "Synced", "stdout should show success")
}

// ---------------------------------------------------------------------------
// Sync command: missing config file
// ---------------------------------------------------------------------------

func TestSync_MissingConfig_ExitNonZero(t *testing.T) {
	dir := t.TempDir()

	_, stderr, exitCode := executeCommand([]string{"sync"}, dir)

	assert.NotEqual(t, 0, exitCode, "sync without config should fail")
	assert.Contains(t, stderr, "not found")
}
