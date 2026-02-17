package steps

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cucumber/godog"
)

func registerPreviewSteps(ctx *godog.ScenarioContext, w *testWorld) {
	ctx.Then(`^no symlinks will be created$`, w.noSymlinksWillBeCreated)
	ctx.Then(`^the JSON output will have "dry_run" set to true$`, w.theJSONOutputWillHaveDryRunTrue)
}

func (w *testWorld) noSymlinksWillBeCreated() error {
	// Check common symlink paths that sync would create
	for _, path := range []string{
		".claude/instructions.md",
		".cursorrules",
		".github/copilot-instructions.md",
		".windsurfrules",
	} {
		fullPath := filepath.Join(w.dir, path)
		info, err := os.Lstat(fullPath)
		if err != nil {
			continue // doesn't exist, good
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("symlink exists at %s but should not in dry-run", path)
		}
	}
	return nil
}

func (w *testWorld) theJSONOutputWillHaveDryRunTrue() error {
	var parsed struct {
		DryRun bool `json:"dry_run"`
	}
	if err := json.Unmarshal([]byte(w.stdout), &parsed); err != nil {
		return fmt.Errorf("cannot parse JSON: %w\nstdout: %s", err, w.stdout)
	}
	if !parsed.DryRun {
		return fmt.Errorf("expected dry_run=true, got false\nstdout: %s", w.stdout)
	}
	return nil
}
