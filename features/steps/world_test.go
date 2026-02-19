package steps

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ailign/cli/internal/cli"
	"github.com/ailign/cli/internal/config"
)

// testWorld is the shared state across all step definitions within a scenario.
type testWorld struct {
	dir      string
	cfg      *config.Config
	loadErr  error
	result   *config.ValidationResult
	stdout   string
	stderr   string
	exitCode int
	binPath  string
}

func (w *testWorld) reset() {
	w.cfg = nil
	w.loadErr = nil
	w.result = nil
	w.stdout = ""
	w.stderr = ""
	w.exitCode = -1
	w.binPath = ""
}

func (w *testWorld) writeConfigWithTargets(targets string) error {
	targetList := strings.Split(targets, ",")
	var yaml strings.Builder
	yaml.WriteString("targets:\n")
	for _, t := range targetList {
		yaml.WriteString(fmt.Sprintf("  - %s\n", strings.TrimSpace(t)))
	}
	return os.WriteFile(filepath.Join(w.dir, ".ailign.yml"), []byte(yaml.String()), 0644)
}

func (w *testWorld) writeConfigWithOverlays(targets, overlays string) error {
	targetList := strings.Split(targets, ",")
	var yaml strings.Builder
	yaml.WriteString("targets:\n")
	for _, t := range targetList {
		yaml.WriteString(fmt.Sprintf("  - %s\n", strings.TrimSpace(t)))
	}
	overlayList := strings.Split(overlays, ",")
	yaml.WriteString("local_overlays:\n")
	for _, o := range overlayList {
		yaml.WriteString(fmt.Sprintf("  - %s\n", strings.TrimSpace(o)))
	}
	return os.WriteFile(filepath.Join(w.dir, ".ailign.yml"), []byte(yaml.String()), 0644)
}

func (w *testWorld) writeOverlayFile(name, content string) error {
	fullPath := filepath.Join(w.dir, name)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return err
	}
	return os.WriteFile(fullPath, []byte(content), 0644)
}

func (w *testWorld) executeCommand(args []string) {
	rootCmd := cli.NewRootCommand()

	stdoutBuf := new(bytes.Buffer)
	stderrBuf := new(bytes.Buffer)
	rootCmd.SetOut(stdoutBuf)
	rootCmd.SetErr(stderrBuf)
	rootCmd.SetArgs(args)

	origDir, _ := os.Getwd()
	_ = os.Chdir(w.dir)
	defer func() { _ = os.Chdir(origDir) }()

	err := rootCmd.Execute()
	w.stdout = stdoutBuf.String()
	w.stderr = stderrBuf.String()
	if err != nil {
		w.exitCode = 2
	} else {
		w.exitCode = 0
	}
}
