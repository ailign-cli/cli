package sync

import (
	"fmt"
	"os"
	"path/filepath"
)

// EnsureSymlink creates or verifies a symlink from linkPath pointing to hubPath.
// Both paths must be absolute. The symlink uses a relative path for portability.
// Returns status: "created", "exists" (already correct), "replaced".
func EnsureSymlink(linkPath string, hubPath string) (string, error) {
	// Compute relative path from link's directory to hub
	relTarget, err := filepath.Rel(filepath.Dir(linkPath), hubPath)
	if err != nil {
		return "", fmt.Errorf("computing relative path: %w", err)
	}

	// Check existing state at linkPath
	info, err := os.Lstat(linkPath)
	if err == nil {
		// Something exists at linkPath
		if info.Mode()&os.ModeSymlink != 0 {
			// It's a symlink — check if it points to the right place
			existingTarget, err := os.Readlink(linkPath)
			if err == nil && existingTarget == relTarget {
				return "exists", nil
			}
			// Wrong symlink — remove and recreate
			if err := os.Remove(linkPath); err != nil {
				return "", fmt.Errorf("removing existing symlink: %w", err)
			}
			return createSymlink(linkPath, relTarget, "replaced")
		}
		// Regular file — remove and replace with symlink
		if err := os.Remove(linkPath); err != nil {
			return "", fmt.Errorf("removing existing file: %w", err)
		}
		return createSymlink(linkPath, relTarget, "replaced")
	}

	// Nothing exists — create directory if needed and create symlink
	return createSymlink(linkPath, relTarget, "created")
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
