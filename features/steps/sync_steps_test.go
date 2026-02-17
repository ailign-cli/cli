package steps

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cucumber/godog"
)

func registerSyncSteps(ctx *godog.ScenarioContext, w *testWorld) {
	// Given steps
	ctx.Given(`^a repository with a valid \.ailign\.yml$`, w.aRepoWithValidAilignYml)
	ctx.Given(`^a \.ailign\.yml with targets "([^"]*)" and overlay "([^"]*)"$`, w.aConfigWithTargetsAndOverlay)
	ctx.Given(`^a \.ailign\.yml with targets "([^"]*)" and overlays "([^"]*)"$`, w.aConfigWithTargetsAndOverlays)
	ctx.Given(`^a \.ailign\.yml with targets "([^"]*)" and no local_overlays$`, w.aConfigWithTargetsNoOverlays)
	ctx.Given(`^an overlay file "([^"]*)" containing "([^"]*)"$`, w.anOverlayFileContaining)
	ctx.Given(`^an overlay file "([^"]*)" that is empty$`, w.anOverlayFileThatIsEmpty)
	ctx.Given(`^an overlay file "([^"]*)" containing non-UTF-8 bytes$`, w.anOverlayFileWithNonUTF8)
	ctx.Given(`^the overlay file "([^"]*)" does not exist$`, w.theOverlayFileDoesNotExist)
	ctx.Given(`^the directory "([^"]*)" does not exist$`, w.theDirectoryDoesNotExist)
	ctx.Given(`^a regular file exists at "([^"]*)" containing "([^"]*)"$`, w.aRegularFileExistsAt)
	ctx.Given(`^ailign sync has been run previously$`, w.ailignSyncHasBeenRunPreviously)
	ctx.Given(`^the target file "([^"]*)" is not writable$`, w.theTargetFileIsNotWritable)

	// When steps
	ctx.When(`^the developer runs ailign sync$`, w.theDeveloperRunsAilignSync)
	ctx.When(`^the developer runs ailign sync with "([^"]*)"$`, w.theDeveloperRunsAilignSyncWith)
	ctx.When(`^the overlay file "([^"]*)" is changed to "([^"]*)"$`, w.theOverlayFileIsChangedTo)

	// Then steps
	ctx.Then(`^the hub file "([^"]*)" will be written$`, w.theHubFileWillBeWritten)
	ctx.Then(`^the hub file "([^"]*)" will not exist$`, w.theHubFileWillNotExist)
	ctx.Then(`^symlinks will be created at "([^"]*)"$`, w.symlinksWillBeCreatedAt)
	ctx.Then(`^a symlink will exist at "([^"]*)"$`, w.aSymlinkWillExistAt)
	ctx.Then(`^each target file will contain "([^"]*)"$`, w.eachTargetFileWillContain)
	ctx.Then(`^the target file "([^"]*)" will contain "([^"]*)"$`, w.theTargetFileWillContain)
	ctx.Then(`^the target file "([^"]*)" will contain "([^"]*)" before "([^"]*)"$`, w.theTargetFileWillContainBefore)
	ctx.Then(`^the target file "([^"]*)" will start with "([^"]*)"$`, w.theTargetFileWillStartWith)
	ctx.Then(`^the directory "([^"]*)" will be created$`, w.theDirectoryWillBeCreated)
	ctx.Then(`^"([^"]*)" will be a symlink$`, w.willBeASymlink)
	ctx.Then(`^stdout will contain "([^"]*)"$`, w.stdoutWillContain)
	ctx.Then(`^stdout will be valid JSON$`, w.stdoutWillBeValidJSON)
	ctx.Then(`^the JSON output will contain target "([^"]*)" with status "([^"]*)"$`, w.theJSONOutputWillContainTargetWithStatus)
	ctx.Then(`^it will report a warning containing "([^"]*)" to stderr$`, w.itReportsWarningContaining)
	// "it will report an error containing" and "it will exit with code" are
	// registered in config_parsing_steps_test.go (shared across features).
}

// --- Given steps ---

func (w *testWorld) aRepoWithValidAilignYml() error {
	// Background step — no-op, config is written by subsequent Given steps
	return nil
}

func (w *testWorld) aConfigWithTargetsAndOverlay(targets, overlay string) error {
	return w.writeConfigWithOverlays(targets, overlay)
}

func (w *testWorld) aConfigWithTargetsAndOverlays(targets, overlays string) error {
	return w.writeConfigWithOverlays(targets, overlays)
}

func (w *testWorld) aConfigWithTargetsNoOverlays(targets string) error {
	return w.writeConfigWithTargets(targets)
}

func (w *testWorld) anOverlayFileContaining(name, content string) error {
	return w.writeOverlayFile(name, content)
}

func (w *testWorld) anOverlayFileThatIsEmpty(name string) error {
	return w.writeOverlayFile(name, "")
}

func (w *testWorld) anOverlayFileWithNonUTF8(name string) error {
	fullPath := filepath.Join(w.dir, name)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return err
	}
	// Invalid UTF-8 sequence
	return os.WriteFile(fullPath, []byte{0xff, 0xfe, 0x80, 0x81}, 0644)
}

func (w *testWorld) theOverlayFileDoesNotExist(_ string) error {
	// No-op: we simply don't create it
	return nil
}

func (w *testWorld) theDirectoryDoesNotExist(dir string) error {
	// Ensure it doesn't exist
	_ = os.RemoveAll(filepath.Join(w.dir, dir))
	return nil
}

func (w *testWorld) aRegularFileExistsAt(path, content string) error {
	fullPath := filepath.Join(w.dir, path)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return err
	}
	return os.WriteFile(fullPath, []byte(content), 0644)
}

func (w *testWorld) ailignSyncHasBeenRunPreviously() error {
	w.executeCommand([]string{"sync"})
	if w.exitCode != 0 {
		return fmt.Errorf("initial sync failed: %s", w.stderr)
	}
	// Reset output for the next sync
	w.stdout = ""
	w.stderr = ""
	w.exitCode = -1
	return nil
}

func (w *testWorld) theTargetFileIsNotWritable(path string) error {
	// Make the parent directory of the target path read-only
	fullPath := filepath.Join(w.dir, path)
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.Chmod(dir, 0555)
}

// --- When steps ---

func (w *testWorld) theDeveloperRunsAilignSync() error {
	w.executeCommand([]string{"sync"})
	return nil
}

func (w *testWorld) theDeveloperRunsAilignSyncWith(flags string) error {
	args := []string{"sync"}
	args = append(args, strings.Fields(flags)...)
	w.executeCommand(args)
	return nil
}

func (w *testWorld) theOverlayFileIsChangedTo(name, content string) error {
	return w.writeOverlayFile(name, content)
}

// --- Then steps ---

func (w *testWorld) theHubFileWillBeWritten(hubPath string) error {
	fullPath := filepath.Join(w.dir, hubPath)
	if _, err := os.Stat(fullPath); err != nil {
		return fmt.Errorf("hub file %s does not exist: %w", hubPath, err)
	}
	return nil
}

func (w *testWorld) theHubFileWillNotExist(hubPath string) error {
	fullPath := filepath.Join(w.dir, hubPath)
	if _, err := os.Stat(fullPath); err == nil {
		return fmt.Errorf("hub file %s exists but should not", hubPath)
	}
	return nil
}

func (w *testWorld) symlinksWillBeCreatedAt(paths string) error {
	for _, p := range strings.Split(paths, ",") {
		p = strings.TrimSpace(p)
		fullPath := filepath.Join(w.dir, p)
		info, err := os.Lstat(fullPath)
		if err != nil {
			return fmt.Errorf("expected symlink at %s but file does not exist: %w", p, err)
		}
		if info.Mode()&os.ModeSymlink == 0 {
			return fmt.Errorf("expected %s to be a symlink but it is not", p)
		}
	}
	return nil
}

func (w *testWorld) aSymlinkWillExistAt(path string) error {
	fullPath := filepath.Join(w.dir, path)
	info, err := os.Lstat(fullPath)
	if err != nil {
		return fmt.Errorf("expected symlink at %s: %w", path, err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		return fmt.Errorf("expected %s to be a symlink", path)
	}
	return nil
}

func (w *testWorld) eachTargetFileWillContain(expected string) error {
	// Read all symlink targets via the hub file
	hubPath := filepath.Join(w.dir, ".ailign", "instructions.md")
	content, err := os.ReadFile(hubPath)
	if err != nil {
		return fmt.Errorf("cannot read hub file: %w", err)
	}
	if !strings.Contains(string(content), expected) {
		return fmt.Errorf("hub file does not contain %q, got: %s", expected, string(content))
	}
	return nil
}

func (w *testWorld) theTargetFileWillContain(path, expected string) error {
	fullPath := filepath.Join(w.dir, path)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return fmt.Errorf("cannot read %s: %w", path, err)
	}
	if !strings.Contains(string(content), expected) {
		return fmt.Errorf("%s does not contain %q, got: %s", path, expected, string(content))
	}
	return nil
}

func (w *testWorld) theTargetFileWillContainBefore(path, first, second string) error {
	fullPath := filepath.Join(w.dir, path)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return fmt.Errorf("cannot read %s: %w", path, err)
	}
	s := string(content)
	firstIdx := strings.Index(s, first)
	secondIdx := strings.Index(s, second)
	if firstIdx == -1 {
		return fmt.Errorf("%s does not contain %q", path, first)
	}
	if secondIdx == -1 {
		return fmt.Errorf("%s does not contain %q", path, second)
	}
	if firstIdx >= secondIdx {
		return fmt.Errorf("in %s, %q (at %d) should appear before %q (at %d)", path, first, firstIdx, second, secondIdx)
	}
	return nil
}

func (w *testWorld) theTargetFileWillStartWith(path, prefix string) error {
	fullPath := filepath.Join(w.dir, path)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return fmt.Errorf("cannot read %s: %w", path, err)
	}
	if !strings.HasPrefix(string(content), prefix) {
		return fmt.Errorf("%s does not start with %q, starts with: %q", path, prefix, string(content)[:min(len(content), 50)])
	}
	return nil
}

func (w *testWorld) theDirectoryWillBeCreated(dir string) error {
	fullPath := filepath.Join(w.dir, dir)
	info, err := os.Stat(fullPath)
	if err != nil {
		return fmt.Errorf("directory %s does not exist: %w", dir, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%s exists but is not a directory", dir)
	}
	return nil
}

func (w *testWorld) willBeASymlink(path string) error {
	fullPath := filepath.Join(w.dir, path)
	info, err := os.Lstat(fullPath)
	if err != nil {
		return fmt.Errorf("%s does not exist: %w", path, err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		return fmt.Errorf("%s is not a symlink (mode: %s)", path, info.Mode())
	}
	return nil
}

func (w *testWorld) stdoutWillContain(expected string) error {
	if !strings.Contains(w.stdout, expected) {
		return fmt.Errorf("stdout does not contain %q, got: %s", expected, w.stdout)
	}
	return nil
}

func (w *testWorld) stdoutWillBeValidJSON() error {
	var raw json.RawMessage
	if err := json.Unmarshal([]byte(w.stdout), &raw); err != nil {
		return fmt.Errorf("stdout is not valid JSON: %w\nstdout: %s", err, w.stdout)
	}
	return nil
}

func (w *testWorld) theJSONOutputWillContainTargetWithStatus(target, status string) error {
	var result struct {
		Links []struct {
			Target string `json:"target"`
			Status string `json:"status"`
		} `json:"links"`
	}
	if err := json.Unmarshal([]byte(w.stdout), &result); err != nil {
		return fmt.Errorf("cannot parse JSON: %w", err)
	}
	for _, l := range result.Links {
		if l.Target == target && l.Status == status {
			return nil
		}
	}
	return fmt.Errorf("JSON output does not contain target %q with status %q, got: %s", target, status, w.stdout)
}

func (w *testWorld) itReportsWarningContaining(substring string) error {
	if !strings.Contains(w.stderr, substring) {
		return fmt.Errorf("stderr does not contain %q, got: %s", substring, w.stderr)
	}
	return nil
}

