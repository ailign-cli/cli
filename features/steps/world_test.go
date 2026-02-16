package steps

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
}

func (w *testWorld) reset() {
	w.cfg = nil
	w.loadErr = nil
	w.result = nil
	w.stdout = ""
	w.stderr = ""
	w.exitCode = -1
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
