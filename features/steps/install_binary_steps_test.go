package steps

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/cucumber/godog"
)

// installScriptState holds state specific to install script scenarios.
type installScriptState struct {
	env         map[string]string
	scriptOut   string
	scriptErr   string
	scriptExit  int
	archivePath string
	checksumTxt string
	tempDirs    []string // tracks temp dirs for cleanup
}

func registerInstallBinarySteps(ctx *godog.ScenarioContext, w *testWorld) {
	is := &installScriptState{env: make(map[string]string), tempDirs: []string{}}

	ctx.After(func(ctx2 context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		for _, d := range is.tempDirs {
			os.RemoveAll(d)
		}
		return ctx2, nil
	})

	// Given steps (past tense)
	ctx.Given(`^the developer sets INSTALL_DIR to "([^"]*)"$`, is.setsInstallDir)
	ctx.Given(`^the developer sets AILIGN_VERSION to "([^"]*)"$`, is.setsAilignVersion)
	ctx.Given(`^the developer sets INSTALL_DIR to a directory not in PATH$`, is.setsInstallDirNotInPath)
	ctx.Given(`^the developer is on an unsupported OS$`, is.isOnUnsupportedOS)
	ctx.Given(`^a release archive has been downloaded$`, is.aReleaseArchiveHasBeenDownloaded)
	ctx.Given(`^the checksums\.txt file has been downloaded$`, is.theChecksumsTxtFileHasBeenDownloaded)

	// When steps (present tense)
	ctx.When(`^the developer runs the install script$`, func() error {
		return is.runsInstallScript(w)
	})
	ctx.When(`^the checksum is verified against the archive$`, is.theChecksumIsVerifiedAgainstTheArchive)

	// Then steps (future tense)
	ctx.Then(`^the ailign binary will be at "([^"]*)"$`, is.binaryWillBeAt)
	ctx.Then(`^running "([^"]*)" will contain "([^"]*)"$`, func(cmd, expected string) error {
		return is.runningWillContain(w, cmd, expected)
	})
	ctx.Then(`^the output will contain a warning about PATH$`, is.outputContainsPathWarning)
	ctx.Then(`^the script will exit with an error$`, is.scriptExitedWithError)
	ctx.Then(`^the error message will list supported platforms$`, is.errorListsSupportedPlatforms)
	ctx.Then(`^the checksum will match$`, is.checksumWillMatch)
}

// --- Given steps ---

func (is *installScriptState) setsInstallDir(dir string) error {
	is.env["INSTALL_DIR"] = dir
	return nil
}

func (is *installScriptState) setsAilignVersion(version string) error {
	is.env["AILIGN_VERSION"] = version
	return nil
}

func (is *installScriptState) setsInstallDirNotInPath() error {
	dir, err := os.MkdirTemp("", "ailign-test-not-in-path-*")
	if err != nil {
		return fmt.Errorf("creating temp install dir: %w", err)
	}
	is.tempDirs = append(is.tempDirs, dir)
	is.env["INSTALL_DIR"] = dir
	return nil
}

func (is *installScriptState) isOnUnsupportedOS() error {
	// Override OS detection by setting env var the script checks
	is.env["AILIGN_OS_OVERRIDE"] = "freebsd"
	return nil
}

func (is *installScriptState) aReleaseArchiveHasBeenDownloaded() error {
	// Create a dummy tar.gz archive with a fake binary
	tmpDir, err := os.MkdirTemp("", "ailign-checksum-test-*")
	if err != nil {
		return err
	}
	is.tempDirs = append(is.tempDirs, tmpDir)

	binPath := filepath.Join(tmpDir, "ailign")
	if err := os.WriteFile(binPath, []byte("#!/bin/sh\necho ailign v0.1.0"), 0755); err != nil {
		return err
	}

	archivePath := filepath.Join(tmpDir, "ailign_0.1.0_darwin_arm64.tar.gz")
	cmd := exec.Command("tar", "czf", archivePath, "-C", tmpDir, "ailign")
	if output, cmdErr := cmd.CombinedOutput(); cmdErr != nil {
		return fmt.Errorf("failed to create archive: %s\n%s", cmdErr, output)
	}

	is.archivePath = archivePath
	return nil
}

func detectChecksumTool() (string, []string, error) {
	if _, err := exec.LookPath("sha256sum"); err == nil {
		return "sha256sum", []string{}, nil
	}
	if _, err := exec.LookPath("shasum"); err == nil {
		return "shasum", []string{"-a", "256"}, nil
	}
	return "", nil, fmt.Errorf("neither sha256sum nor shasum found in PATH")
}

func (is *installScriptState) theChecksumsTxtFileHasBeenDownloaded() error {
	if is.archivePath == "" {
		return fmt.Errorf("no archive path set")
	}

	tool, baseArgs, err := detectChecksumTool()
	if err != nil {
		return fmt.Errorf("selecting checksum tool: %w", err)
	}

	args := append(baseArgs, is.archivePath)
	cmd := exec.Command(tool, args...)
	output, cmdErr := cmd.Output()
	if cmdErr != nil {
		return fmt.Errorf("failed to compute checksum using %s: %w", tool, cmdErr)
	}

	// Write checksums.txt with the same format GoReleaser uses
	checksumFile := filepath.Join(filepath.Dir(is.archivePath), "checksums.txt")
	// Format: <hash>  <filename> (two spaces, basename only)
	parts := strings.Fields(string(output))
	if len(parts) == 0 {
		return fmt.Errorf("failed to parse checksum output %q: expected '<hash>  <filename>' format", strings.TrimSpace(string(output)))
	}
	content := fmt.Sprintf("%s  %s\n", parts[0], filepath.Base(is.archivePath))
	if writeErr := os.WriteFile(checksumFile, []byte(content), 0644); writeErr != nil {
		return writeErr
	}
	is.checksumTxt = checksumFile
	return nil
}

// --- When steps ---

func (is *installScriptState) runsInstallScript(w *testWorld) error {
	scriptPath := filepath.Join(findRepoRoot(), "install.sh")
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return fmt.Errorf("install.sh not found at %s — not yet implemented", scriptPath)
	}

	cmd := exec.Command("sh", scriptPath)
	cmd.Env = append(os.Environ(), "AILIGN_TEST_MODE=1")
	for k, v := range is.env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	is.scriptOut = stdout.String()
	is.scriptErr = stderr.String()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			is.scriptExit = exitErr.ExitCode()
		} else {
			is.scriptExit = 1
		}
	} else {
		is.scriptExit = 0
	}

	// Expose to testWorld for shared "Then" steps
	w.stdout = is.scriptOut
	w.stderr = is.scriptErr
	w.exitCode = is.scriptExit

	return nil
}

func (is *installScriptState) theChecksumIsVerifiedAgainstTheArchive() error {
	if is.archivePath == "" || is.checksumTxt == "" {
		return fmt.Errorf("archive or checksums.txt not set")
	}

	tool, baseArgs, err := detectChecksumTool()
	if err != nil {
		return fmt.Errorf("selecting checksum tool: %w", err)
	}

	args := append(baseArgs, "-c", "--ignore-missing", is.checksumTxt)
	cmd := exec.Command(tool, args...)
	cmd.Dir = filepath.Dir(is.archivePath)
	output, err := cmd.CombinedOutput()
	is.scriptOut = string(output)
	if err != nil {
		is.scriptExit = 1
		is.scriptErr = string(output)
	} else {
		is.scriptExit = 0
	}
	return nil
}

// --- Then steps ---

func (is *installScriptState) runningWillContain(w *testWorld, command, expected string) error {
	// Determine binary path: INSTALL_DIR env, install script output, or built binary
	var binPath string
	if is.env["INSTALL_DIR"] != "" {
		binPath = filepath.Join(is.env["INSTALL_DIR"], "ailign")
	} else if w.binPath != "" {
		binPath = w.binPath
	} else {
		// Default install location used by install.sh in test mode
		binPath = filepath.Join(os.Getenv("HOME"), ".local", "bin", "ailign")
		if _, err := os.Stat(binPath); os.IsNotExist(err) {
			binPath = "/usr/local/bin/ailign"
		}
	}

	// Parse the command string into binary + args, substituting the resolved binary path
	parts := strings.Fields(command)
	var args []string
	if len(parts) > 1 {
		args = parts[1:]
	}

	cmd := exec.Command(binPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to run %s: %s\n%s", command, err, output)
	}
	if !strings.Contains(string(output), expected) {
		return fmt.Errorf("expected output of %q to contain %q, got: %s", command, expected, output)
	}
	return nil
}

func (is *installScriptState) binaryWillBeAt(expectedPath string) error {
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		return fmt.Errorf("expected binary at %s but it does not exist", expectedPath)
	}
	return nil
}

func (is *installScriptState) outputContainsPathWarning() error {
	combined := is.scriptOut + is.scriptErr
	if strings.Contains(strings.ToLower(combined), "path") {
		return nil
	}
	return fmt.Errorf("expected PATH warning in output, got:\nstdout: %s\nstderr: %s", is.scriptOut, is.scriptErr)
}

func (is *installScriptState) scriptExitedWithError() error {
	if is.scriptExit != 0 {
		return nil
	}
	return fmt.Errorf("expected non-zero exit code, got 0")
}

func (is *installScriptState) errorListsSupportedPlatforms() error {
	combined := is.scriptOut + is.scriptErr
	if strings.Contains(combined, "linux") && strings.Contains(combined, "darwin") {
		return nil
	}
	return fmt.Errorf("expected error to list supported platforms (linux, darwin), got:\n%s", combined)
}

func (is *installScriptState) checksumWillMatch() error {
	if is.scriptExit != 0 {
		return fmt.Errorf("checksum verification failed: %s", is.scriptErr)
	}
	if strings.Contains(is.scriptOut, "OK") {
		return nil
	}
	return fmt.Errorf("expected checksum to match (OK), got: %s", is.scriptOut)
}

// findRepoRoot walks up from the working directory to find the repo root (contains go.mod).
func findRepoRoot() string {
	dir, _ := os.Getwd()
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "."
		}
		dir = parent
	}
}
