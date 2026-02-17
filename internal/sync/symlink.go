package sync

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// EnsureSymlink creates or verifies a symlink from linkPath pointing to hubPath.
// Both paths must be absolute. The symlink uses a relative path for portability.
// Returns status: "created", "exists" (already correct), "replaced".
func EnsureSymlink(linkPath string, hubPath string) (string, error) {
	if !filepath.IsAbs(linkPath) {
		return "", fmt.Errorf("linkPath must be absolute, got: %s", linkPath)
	}
	if !filepath.IsAbs(hubPath) {
		return "", fmt.Errorf("hubPath must be absolute, got: %s", hubPath)
	}

	// Compute relative path from link's directory to hub
	relTarget, err := filepath.Rel(filepath.Dir(linkPath), hubPath)
	if err != nil {
		return "", fmt.Errorf("computing relative path: %w", err)
	}

	// Check existing state at linkPath
	info, err := os.Lstat(linkPath)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("checking existing path: %w", err)
		}
		// Nothing exists — create directory if needed and create symlink
		return createSymlink(linkPath, relTarget, "created")
	}

	// Something exists at linkPath
	if info.Mode()&os.ModeSymlink != 0 {
		// It's a symlink — check if it points to the right place
		existingTarget, err := os.Readlink(linkPath)
		if err == nil && existingTarget == relTarget {
			return "exists", nil
		}
	}

	// Wrong symlink or regular file — atomically replace
	return replaceSymlink(linkPath, relTarget)
}

func createSymlink(linkPath string, relTarget string, status string) (string, error) {
	dir := filepath.Dir(linkPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("creating directory for symlink: %w", err)
	}

	if err := os.Symlink(relTarget, linkPath); err != nil {
		return "", fmt.Errorf("creating symlink: %w", err)
	}

	return status, nil
}

// replaceSymlink atomically replaces whatever exists at linkPath with a
// symlink pointing to relTarget. It creates a temporary symlink in the
// same directory and renames it into place so the previous file/link
// remains until the new one is ready.
func replaceSymlink(linkPath, relTarget string) (string, error) {
	dir := filepath.Dir(linkPath)
	tmpLink := filepath.Join(dir, ".ailign-symlink-tmp")

	// Remove any stale temp from a previous interrupted run
	_ = os.Remove(tmpLink)

	if err := os.Symlink(relTarget, tmpLink); err != nil {
		return "", fmt.Errorf("creating temporary symlink: %w", err)
	}

	if err := os.Rename(tmpLink, linkPath); err != nil {
		_ = os.Remove(tmpLink)
		return "", fmt.Errorf("replacing symlink: %w", err)
	}

	return "replaced", nil
}
