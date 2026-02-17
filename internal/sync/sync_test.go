package sync

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ailign/cli/internal/config"
	"github.com/ailign/cli/internal/target"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Sync orchestration tests (T018)
// ---------------------------------------------------------------------------

func TestSync_FullFlow(t *testing.T) {
	skipOnWindows(t)
	dir := resolveDir(t)

	writeFile(t, filepath.Join(dir, ".ai-instructions", "base.md"), "Base content.\n")
	writeFile(t, filepath.Join(dir, ".ai-instructions", "project.md"), "Project content.\n")

	cfg := &config.Config{
		Targets:       []string{"claude", "cursor"},
		LocalOverlays: []string{".ai-instructions/base.md", ".ai-instructions/project.md"},
	}
	registry := target.NewDefaultRegistry()

	result, err := Sync(dir, cfg, registry, SyncOptions{})
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, filepath.Join(dir, ".ailign", "instructions.md"), result.HubPath)
	assert.Equal(t, "written", result.HubStatus)
	assert.Len(t, result.Links, 2)

	// Verify hub file exists with composed content
	hubContent, err := os.ReadFile(result.HubPath)
	require.NoError(t, err)
	assert.Contains(t, string(hubContent), "Base content.")
	assert.Contains(t, string(hubContent), "Project content.")

	// Verify symlinks exist and resolve to hub
	for _, link := range result.Links {
		assert.Equal(t, "created", link.Status, "target %s", link.Target)
		linkFullPath := filepath.Join(dir, link.LinkPath)
		resolved, err := filepath.EvalSymlinks(linkFullPath)
		require.NoError(t, err, "symlink should resolve for target %s", link.Target)
		assert.Equal(t, result.HubPath, resolved)
	}
}

func TestSync_MissingOverlayError(t *testing.T) {
	dir := t.TempDir()

	cfg := &config.Config{
		Targets:       []string{"claude"},
		LocalOverlays: []string{"nonexistent.md"},
	}
	registry := target.NewDefaultRegistry()

	_, err := Sync(dir, cfg, registry, SyncOptions{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestSync_NoOverlaysError(t *testing.T) {
	dir := t.TempDir()

	cfg := &config.Config{
		Targets:       []string{"claude"},
		LocalOverlays: nil,
	}
	registry := target.NewDefaultRegistry()

	_, err := Sync(dir, cfg, registry, SyncOptions{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no local_overlays")
}

func TestSync_PartialFailure(t *testing.T) {
	skipOnWindows(t)

	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "base.md"), "Content\n")

	// Create .claude/ as read-only so symlink creation fails for claude
	claudeDir := filepath.Join(dir, ".claude")
	require.NoError(t, os.MkdirAll(claudeDir, 0555))
	defer func() { _ = os.Chmod(claudeDir, 0755) }()

	cfg := &config.Config{
		Targets:       []string{"claude", "cursor"},
		LocalOverlays: []string{"base.md"},
	}
	registry := target.NewDefaultRegistry()

	result, err := Sync(dir, cfg, registry, SyncOptions{})
	// Partial failure returns a result (not an error) with per-target status
	require.NoError(t, err)
	require.NotNil(t, result)

	var errorCount, successCount int
	for _, link := range result.Links {
		if link.Status == "error" {
			errorCount++
			assert.NotEmpty(t, link.Error)
		} else {
			successCount++
		}
	}
	assert.Equal(t, 1, errorCount, "claude should fail (read-only directory)")
	assert.Equal(t, 1, successCount, "cursor should succeed")
}

func TestSync_EmptyOverlayWarning(t *testing.T) {
	skipOnWindows(t)
	dir := resolveDir(t)

	writeFile(t, filepath.Join(dir, "empty.md"), "")

	cfg := &config.Config{
		Targets:       []string{"claude"},
		LocalOverlays: []string{"empty.md"},
	}
	registry := target.NewDefaultRegistry()

	result, err := Sync(dir, cfg, registry, SyncOptions{})
	require.NoError(t, err)
	require.NotNil(t, result)

	require.Len(t, result.Warnings, 1)
	assert.Contains(t, result.Warnings[0], "empty.md")
	assert.Contains(t, result.Warnings[0], "empty")
}

// ---------------------------------------------------------------------------
// Dry-run tests (T036)
// ---------------------------------------------------------------------------

func TestSync_DryRun_NoFilesWritten(t *testing.T) {
	dir := resolveDir(t)
	writeFile(t, filepath.Join(dir, "base.md"), "Content\n")

	cfg := &config.Config{
		Targets:       []string{"claude", "cursor"},
		LocalOverlays: []string{"base.md"},
	}
	registry := target.NewDefaultRegistry()

	result, err := Sync(dir, cfg, registry, SyncOptions{DryRun: true})
	require.NoError(t, err)
	require.NotNil(t, result)

	// Hub file should NOT exist
	_, err = os.Stat(filepath.Join(dir, ".ailign", "instructions.md"))
	assert.True(t, os.IsNotExist(err), "hub file should not be created in dry-run")

	// Symlinks should NOT exist
	for _, link := range result.Links {
		_, err = os.Lstat(filepath.Join(dir, link.LinkPath))
		assert.True(t, os.IsNotExist(err), "symlink %s should not be created in dry-run", link.LinkPath)
	}

	// Result should still report what would happen
	assert.Equal(t, "written", result.HubStatus)
	assert.Len(t, result.Links, 2)
	for _, link := range result.Links {
		assert.Equal(t, "created", link.Status)
	}
}

func TestSync_DryRun_ExistingSymlinksDetected(t *testing.T) {
	skipOnWindows(t)
	dir := resolveDir(t)
	writeFile(t, filepath.Join(dir, "base.md"), "Content\n")

	cfg := &config.Config{
		Targets:       []string{"cursor"},
		LocalOverlays: []string{"base.md"},
	}
	registry := target.NewDefaultRegistry()

	// Run real sync first
	_, err := Sync(dir, cfg, registry, SyncOptions{})
	require.NoError(t, err)

	// Now dry-run should detect existing correct symlinks
	result, err := Sync(dir, cfg, registry, SyncOptions{DryRun: true})
	require.NoError(t, err)

	assert.Equal(t, "unchanged", result.HubStatus)
	require.Len(t, result.Links, 1)
	assert.Equal(t, "exists", result.Links[0].Status)
}

func TestSync_UnknownTarget(t *testing.T) {
	dir := resolveDir(t)

	writeFile(t, filepath.Join(dir, "base.md"), "Content\n")

	cfg := &config.Config{
		Targets:       []string{"nonexistent-tool"},
		LocalOverlays: []string{"base.md"},
	}
	registry := target.NewDefaultRegistry()

	result, err := Sync(dir, cfg, registry, SyncOptions{})
	require.NoError(t, err)
	require.NotNil(t, result)

	require.Len(t, result.Links, 1)
	assert.Equal(t, "error", result.Links[0].Status)
	assert.Contains(t, result.Links[0].Error, "unknown target")
	assert.Empty(t, result.Links[0].LinkPath)
}
