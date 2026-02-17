package sync

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// EnsureSymlink tests (T017)
// ---------------------------------------------------------------------------

func TestEnsureSymlink_CreateNew(t *testing.T) {
	dir := resolveDir(t)
	hubPath := filepath.Join(dir, ".ailign", "instructions.md")
	linkPath := filepath.Join(dir, ".cursorrules")

	require.NoError(t, os.MkdirAll(filepath.Dir(hubPath), 0755))
	require.NoError(t, os.WriteFile(hubPath, []byte("content"), 0644))

	status, err := EnsureSymlink(linkPath, hubPath)
	require.NoError(t, err)
	assert.Equal(t, "created", status)

	// Verify symlink resolves to hub
	resolved, err := filepath.EvalSymlinks(linkPath)
	require.NoError(t, err)
	assert.Equal(t, hubPath, resolved)
}

func TestEnsureSymlink_ExistingCorrectSymlink(t *testing.T) {
	dir := t.TempDir()
	hubPath := filepath.Join(dir, ".ailign", "instructions.md")
	linkPath := filepath.Join(dir, ".cursorrules")

	require.NoError(t, os.MkdirAll(filepath.Dir(hubPath), 0755))
	require.NoError(t, os.WriteFile(hubPath, []byte("content"), 0644))

	// Create correct symlink first
	_, err := EnsureSymlink(linkPath, hubPath)
	require.NoError(t, err)

	// Call again — should detect existing correct symlink
	status, err := EnsureSymlink(linkPath, hubPath)
	require.NoError(t, err)
	assert.Equal(t, "exists", status)
}

func TestEnsureSymlink_ExistingWrongSymlink(t *testing.T) {
	dir := resolveDir(t)
	hubPath := filepath.Join(dir, ".ailign", "instructions.md")
	otherPath := filepath.Join(dir, "other.md")
	linkPath := filepath.Join(dir, ".cursorrules")

	require.NoError(t, os.MkdirAll(filepath.Dir(hubPath), 0755))
	require.NoError(t, os.WriteFile(hubPath, []byte("content"), 0644))
	require.NoError(t, os.WriteFile(otherPath, []byte("other"), 0644))

	// Create wrong symlink
	require.NoError(t, os.Symlink(otherPath, linkPath))

	status, err := EnsureSymlink(linkPath, hubPath)
	require.NoError(t, err)
	assert.Equal(t, "replaced", status)

	// Verify now points to hub
	resolved, err := filepath.EvalSymlinks(linkPath)
	require.NoError(t, err)
	assert.Equal(t, hubPath, resolved)
}

func TestEnsureSymlink_ReplaceRegularFile(t *testing.T) {
	dir := t.TempDir()
	hubPath := filepath.Join(dir, ".ailign", "instructions.md")
	linkPath := filepath.Join(dir, ".cursorrules")

	require.NoError(t, os.MkdirAll(filepath.Dir(hubPath), 0755))
	require.NoError(t, os.WriteFile(hubPath, []byte("content"), 0644))
	require.NoError(t, os.WriteFile(linkPath, []byte("regular file"), 0644))

	status, err := EnsureSymlink(linkPath, hubPath)
	require.NoError(t, err)
	assert.Equal(t, "replaced", status)

	// Verify it's now a symlink
	info, err := os.Lstat(linkPath)
	require.NoError(t, err)
	assert.True(t, info.Mode()&os.ModeSymlink != 0, "should be a symlink")
}

func TestEnsureSymlink_CreateDirectory(t *testing.T) {
	dir := t.TempDir()
	hubPath := filepath.Join(dir, ".ailign", "instructions.md")
	linkPath := filepath.Join(dir, ".claude", "instructions.md")

	require.NoError(t, os.MkdirAll(filepath.Dir(hubPath), 0755))
	require.NoError(t, os.WriteFile(hubPath, []byte("content"), 0644))

	// .claude/ doesn't exist yet
	status, err := EnsureSymlink(linkPath, hubPath)
	require.NoError(t, err)
	assert.Equal(t, "created", status)
}

func TestEnsureSymlink_RelativePath(t *testing.T) {
	dir := t.TempDir()
	hubPath := filepath.Join(dir, ".ailign", "instructions.md")
	linkPath := filepath.Join(dir, ".claude", "instructions.md")

	require.NoError(t, os.MkdirAll(filepath.Dir(hubPath), 0755))
	require.NoError(t, os.WriteFile(hubPath, []byte("content"), 0644))

	_, err := EnsureSymlink(linkPath, hubPath)
	require.NoError(t, err)

	// Read the raw symlink target — should be relative
	rawTarget, err := os.Readlink(linkPath)
	require.NoError(t, err)
	assert.False(t, filepath.IsAbs(rawTarget), "symlink should be relative, got: %s", rawTarget)
}

func TestEnsureSymlink_PermissionError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission test not reliable on Windows")
	}

	dir := t.TempDir()
	hubPath := filepath.Join(dir, ".ailign", "instructions.md")
	readonlyDir := filepath.Join(dir, "readonly")
	linkPath := filepath.Join(readonlyDir, ".cursorrules")

	require.NoError(t, os.MkdirAll(filepath.Dir(hubPath), 0755))
	require.NoError(t, os.WriteFile(hubPath, []byte("content"), 0644))
	require.NoError(t, os.MkdirAll(readonlyDir, 0555))
	defer os.Chmod(readonlyDir, 0755)

	_, err := EnsureSymlink(linkPath, hubPath)
	require.Error(t, err)
}

// resolveDir returns t.TempDir() with symlinks resolved (macOS: /var → /private/var).
func resolveDir(t *testing.T) string {
	t.Helper()
	dir, err := filepath.EvalSymlinks(t.TempDir())
	require.NoError(t, err)
	return dir
}
