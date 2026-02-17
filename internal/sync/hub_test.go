package sync

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// WriteHub tests (T016)
// ---------------------------------------------------------------------------

func TestWriteHub_NewFile(t *testing.T) {
	dir := t.TempDir()
	hubPath := filepath.Join(dir, ".ailign", "instructions.md")

	status, err := WriteHub(hubPath, []byte("hub content"))
	require.NoError(t, err)
	assert.Equal(t, "written", status)

	content, err := os.ReadFile(hubPath)
	require.NoError(t, err)
	assert.Equal(t, "hub content", string(content))
}

func TestWriteHub_CreateDirectory(t *testing.T) {
	dir := t.TempDir()
	hubPath := filepath.Join(dir, ".ailign", "instructions.md")

	// .ailign/ doesn't exist yet
	_, err := WriteHub(hubPath, []byte("content"))
	require.NoError(t, err)

	info, err := os.Stat(filepath.Dir(hubPath))
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestWriteHub_UpdateExisting(t *testing.T) {
	dir := t.TempDir()
	hubPath := filepath.Join(dir, ".ailign", "instructions.md")

	// Write initial content
	_, err := WriteHub(hubPath, []byte("old content"))
	require.NoError(t, err)

	// Update
	status, err := WriteHub(hubPath, []byte("new content"))
	require.NoError(t, err)
	assert.Equal(t, "written", status)

	content, err := os.ReadFile(hubPath)
	require.NoError(t, err)
	assert.Equal(t, "new content", string(content))
}

func TestWriteHub_UnchangedContent(t *testing.T) {
	dir := t.TempDir()
	hubPath := filepath.Join(dir, ".ailign", "instructions.md")

	_, err := WriteHub(hubPath, []byte("same content"))
	require.NoError(t, err)

	status, err := WriteHub(hubPath, []byte("same content"))
	require.NoError(t, err)
	assert.Equal(t, "unchanged", status)
}
