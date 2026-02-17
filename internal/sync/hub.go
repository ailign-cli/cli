package sync

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
)

// WriteHub atomically writes content to the hub file using write-temp-rename.
// Returns "written" if content changed, "unchanged" if identical to existing.
func WriteHub(hubPath string, content []byte) (string, error) {
	// Check if existing content is identical
	existing, err := os.ReadFile(hubPath)
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
