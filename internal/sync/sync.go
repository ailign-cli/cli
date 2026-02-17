package sync

import (
	"fmt"
	"path/filepath"

	"github.com/ailign/cli/internal/config"
	"github.com/ailign/cli/internal/target"
)

const hubRelPath = ".ailign/instructions.md"

// Sync composes overlays and syncs to all configured targets.
// Returns a SyncResult with per-target outcomes. Partial failures
// (e.g., one target's symlink fails) are captured in LinkResult,
// not as an overall error.
func Sync(baseDir string, cfg *config.Config, registry *target.Registry, opts SyncOptions) (*SyncResult, error) {
	if len(cfg.LocalOverlays) == 0 {
		return nil, fmt.Errorf("no local_overlays configured in .ailign.yml")
	}

	// Normalize to absolute path so derived paths satisfy EnsureSymlink's contract
	absBase, err := filepath.Abs(baseDir)
	if err != nil {
		return nil, fmt.Errorf("resolving base directory: %w", err)
	}
	baseDir = absBase

	hubPath := filepath.Join(baseDir, hubRelPath)

	// Compose overlays
	composed, err := ComposeOverlays(baseDir, cfg.LocalOverlays)
	if err != nil {
		return nil, err
	}

	// Write or check hub file
	var hubStatus string
	if opts.DryRun {
		hubStatus, err = CheckHubStatus(hubPath, composed.Content)
	} else {
		hubStatus, err = WriteHub(hubPath, composed.Content)
	}
	if err != nil {
		if opts.DryRun {
			return nil, fmt.Errorf("checking hub file: %w", err)
		}
		return nil, fmt.Errorf("writing hub file: %w", err)
	}

	result := &SyncResult{
		DryRun:    opts.DryRun,
		HubPath:   hubPath,
		HubStatus: hubStatus,
		Links:     make([]LinkResult, 0, len(cfg.Targets)),
		Warnings:  composed.Warnings,
	}

	// Create or check symlinks per target
	for _, targetName := range cfg.Targets {
		tgt, ok := registry.Get(targetName)
		if !ok {
			result.Links = append(result.Links, LinkResult{
				Target:   targetName,
				LinkPath: "",
				Status:   "error",
				Error:    fmt.Sprintf("unknown target: %s", targetName),
			})
			continue
		}

		linkPath := filepath.Join(baseDir, tgt.InstructionPath())
		var status string
		if opts.DryRun {
			status, err = CheckSymlinkStatus(linkPath, hubPath)
		} else {
			status, err = EnsureSymlink(linkPath, hubPath)
		}

		link := LinkResult{
			Target:   targetName,
			LinkPath: tgt.InstructionPath(),
		}
		if err != nil {
			link.Status = "error"
			link.Error = err.Error()
		} else {
			link.Status = status
		}
		result.Links = append(result.Links, link)
	}

	return result, nil
}
