package steps

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/cucumber/godog"
)

func registerInstallGoSteps(ctx *godog.ScenarioContext, w *testWorld) {
	ctx.Given(`^ailign was installed via "([^"]*)" from tag "([^"]*)"$`, w.ailignWasInstalledViaFromTag)
	ctx.When(`^the developer runs "ailign --version"$`, w.theDeveloperRunsAilignVersion)
	ctx.Then(`^the output will contain "([^"]*)"$`, w.theOutputWillContain)
}

func (w *testWorld) ailignWasInstalledViaFromTag(method, tag string) error {
	// Build the binary with ldflags that simulate the given version tag.
	// The version variable is in cmd/ailign/main.go.
	binPath := filepath.Join(w.dir, "ailign")
	cmd := exec.Command("go", "build",
		"-ldflags", fmt.Sprintf("-X main.version=%s -X main.commit=test", tag),
		"-o", binPath,
		"github.com/ailign/cli/cmd/ailign",
	)
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to build binary: %s\n%s", err, output)
	}
	w.binPath = binPath
	return nil
}

func (w *testWorld) theDeveloperRunsAilignVersion() error {
	if w.binPath == "" {
		return fmt.Errorf("no binary path set — was the binary built?")
	}
	cmd := exec.Command(w.binPath, "--version")
	output, err := cmd.CombinedOutput()
	w.stdout = string(output)
	if err != nil {
		w.exitCode = 2
	} else {
		w.exitCode = 0
	}
	return nil
}

func (w *testWorld) theOutputWillContain(expected string) error {
	if strings.Contains(w.stdout, expected) {
		return nil
	}
	return fmt.Errorf("expected output to contain %q, got: %s", expected, w.stdout)
}
