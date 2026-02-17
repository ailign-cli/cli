package sync

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// WriteHub writes content to the hub file using write-temp-rename.
// The temp file is fsynced before rename for crash safety; note that
// the parent directory is not fsynced, so a power loss after rename
// could lose the update on some filesystems. This is acceptable
// because the hub file is regenerable via "ailign sync".
// Returns "written" if content changed, "unchanged" if identical to existing.
func WriteHub(hubPath string, content []byte) (string, error) {
	// Check if existing content is identical
	existing, err := os.ReadFile(hubPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("reading existing hub file: %w", err)
	}
	if err == nil && bytes.Equal(existing, content) {
		return "unchanged", nil
	}

	// Create directory if needed
	dir := filepath.Dir(hubPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("creating hub directory: %w", err)
	}

	// Atomic write: temp file → fsync → rename
	tmp, err := os.CreateTemp(dir, ".ailign-*.tmp")
	if err != nil {
		return "", fmt.Errorf("creating temp file: %w", err)
	}
	defer os.Remove(tmp.Name()) // cleanup on error path

	if _, err := tmp.Write(content); err != nil {
		tmp.Close()
		return "", fmt.Errorf("writing temp file: %w", err)
	}

	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return "", fmt.Errorf("syncing temp file: %w", err)
	}

	if err := tmp.Close(); err != nil {
		return "", fmt.Errorf("closing temp file: %w", err)
	}

	if err := os.Rename(tmp.Name(), hubPath); err != nil {
		return "", fmt.Errorf("renaming hub file: %w", err)
	}

	return "written", nil
}
