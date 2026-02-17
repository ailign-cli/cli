package cli

import (
	"fmt"
	"os"

	"github.com/ailign/cli/internal/output"
	"github.com/ailign/cli/internal/sync"
	"github.com/ailign/cli/internal/target"
	"github.com/spf13/cobra"
)

var dryRunFlag bool

func newSyncCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Sync local instruction files to all configured targets",
		Long:  "Composes local overlay files and creates symlinks from each target's instruction path to the central hub file (.ailign/instructions.md).",
		RunE:  runSync,
	}
	cmd.Flags().BoolVarP(&dryRunFlag, "dry-run", "n", false,
		"Preview changes without modifying any files")
	return cmd
}

func runSync(cmd *cobra.Command, args []string) error {
	cfg := GetConfig()
	if cfg == nil {
		return ErrAlreadyReported
	}

	cwd, err := os.Getwd()
	if err != nil {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %s\n", err)
		return ErrAlreadyReported
	}

	registry := target.NewDefaultRegistry()
	result, err := sync.Sync(cwd, cfg, registry, sync.SyncOptions{DryRun: dryRunFlag})
	if err != nil {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %s\n", err)
		return ErrAlreadyReported
	}

	// Print warnings to stderr
	for _, w := range result.Warnings {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Warning: %s\n", w)
	}

	// Format and print result to stdout
	syncResult := toSyncOutputResult(result, len(cfg.LocalOverlays), dryRunFlag)
	sf := getSyncFormatter(formatFlag)
	_, _ = fmt.Fprint(cmd.OutOrStdout(), sf.FormatSyncResult(syncResult))

	// Check for per-target errors
	for _, link := range result.Links {
		if link.Status == "error" {
			return ErrAlreadyReported
		}
	}

	return nil
}

func toSyncOutputResult(r *sync.SyncResult, overlayCount int, dryRun bool) output.SyncResult {
	links := make([]output.LinkResult, 0, len(r.Links))
	for _, l := range r.Links {
		links = append(links, output.LinkResult{
			Target:   l.Target,
			LinkPath: l.LinkPath,
			Status:   l.Status,
			Error:    l.Error,
		})
	}

	return output.SyncResult{
		DryRun:       dryRun,
		HubPath:      r.HubPath,
		HubStatus:    r.HubStatus,
		Links:        links,
		Warnings:     r.Warnings,
		OverlayCount: overlayCount,
	}
}

func getSyncFormatter(format string) output.SyncFormatter {
	switch format {
	case "json":
		return &output.JSONFormatter{}
	case "human":
		return &output.HumanFormatter{}
	default:
		return &output.HumanFormatter{}
	}
}
