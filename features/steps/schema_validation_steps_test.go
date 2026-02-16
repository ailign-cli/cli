package steps

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ailign/cli/internal/cli"
	"github.com/ailign/cli/internal/config"
	"github.com/cucumber/godog"
)

func registerSchemaValidationSteps(ctx *godog.ScenarioContext, w *testWorld) {
	// Given steps (past tense) — validation-specific config setup
	ctx.Given(`^a \.ailign\.yml was configured with UTF-8 BOM and targets "([^"]*)"$`, w.aConfigWithBOM)
	ctx.Given(`^a \.ailign\.yml was configured with targets "([^"]*)"$`, w.aConfigWithTargets)
	ctx.Given(`^a \.ailign\.yml was configured with targets "([^"]*)" and unknown field "([^"]*)"$`, w.aConfigWithTargetsAndUnknownField)
	ctx.Given(`^a \.ailign\.yml was configured with no targets field$`, w.aConfigWithNoTargetsField)

	// When steps (present tense)
	ctx.When(`^the CLI validates the configuration$`, w.theCLIValidatesConfig)
	ctx.When(`^the developer runs ailign validate$`, w.theDeveloperRunsValidate)

	// Then steps (future tense) — validation-specific assertions
	ctx.Then(`^it will report an error at field path "([^"]*)"$`, w.itReportsErrorAtFieldPath)
	ctx.Then(`^the error will suggest valid targets "([^"]*)"$`, w.theErrorSuggestsValidTargets)
	ctx.Then(`^it will report at least (\d+) errors$`, w.itReportsAtLeastNErrors)
	ctx.Then(`^all errors will include field paths and remediation$`, w.allErrorsIncludeFieldPathsAndRemediation)
	ctx.Then(`^it will report a warning about "([^"]*)"$`, w.itReportsWarningAbout)
	ctx.Then(`^validation will succeed$`, w.validationSucceeds)
	ctx.Then(`^it will report success to stdout$`, w.itReportsSuccessToStdout)
	ctx.Then(`^it will report errors to stderr$`, w.itReportsErrorsToStderr)
	ctx.Then(`^it will report an error about duplicate targets$`, w.itReportsErrorAboutDuplicateTargets)
}

// --- Given steps ---

func (w *testWorld) aConfigWithBOM(targets string) error {
	targetList := strings.Split(targets, ",")
	var buf bytes.Buffer
	buf.Write([]byte{0xEF, 0xBB, 0xBF})
	buf.WriteString("targets:\n")
	for _, t := range targetList {
		buf.WriteString(fmt.Sprintf("  - %s\n", strings.TrimSpace(t)))
	}
	return os.WriteFile(filepath.Join(w.dir, ".ailign.yml"), buf.Bytes(), 0644)
}

func (w *testWorld) aConfigWithTargets(targets string) error {
	return w.writeConfigWithTargets(targets)
}

func (w *testWorld) aConfigWithTargetsAndUnknownField(targets, unknownField string) error {
	targetList := strings.Split(targets, ",")
	var yaml strings.Builder
	yaml.WriteString("targets:\n")
	for _, t := range targetList {
		yaml.WriteString(fmt.Sprintf("  - %s\n", strings.TrimSpace(t)))
	}
	yaml.WriteString(fmt.Sprintf("%s: some_value\n", unknownField))
	return os.WriteFile(filepath.Join(w.dir, ".ailign.yml"), []byte(yaml.String()), 0644)
}

func (w *testWorld) aConfigWithNoTargetsField() error {
	return os.WriteFile(filepath.Join(w.dir, ".ailign.yml"), []byte("some_field: hello\n"), 0644)
}

// --- When steps ---

func (w *testWorld) theCLIValidatesConfig() error {
	cfgPath := filepath.Join(w.dir, ".ailign.yml")
	w.result = config.LoadAndValidate(cfgPath)
	return nil
}

func (w *testWorld) theDeveloperRunsValidate() error {
	rootCmd := cli.NewRootCommand()

	stdoutBuf := new(bytes.Buffer)
	stderrBuf := new(bytes.Buffer)
	rootCmd.SetOut(stdoutBuf)
	rootCmd.SetErr(stderrBuf)
	rootCmd.SetArgs([]string{"validate"})

	origDir, _ := os.Getwd()
	_ = os.Chdir(w.dir)
	defer func() { _ = os.Chdir(origDir) }()

	err := rootCmd.Execute()
	w.exitCode = 0
	if err != nil {
		w.exitCode = 2
	}
	w.stdout = stdoutBuf.String()
	w.stderr = stderrBuf.String()
	return nil
}

// --- Then steps ---

func (w *testWorld) itReportsErrorAtFieldPath(fieldPath string) error {
	if w.result == nil {
		return fmt.Errorf("no validation result available")
	}
	for _, e := range w.result.Errors {
		if e.FieldPath == fieldPath {
			return nil
		}
	}
	var paths []string
	for _, e := range w.result.Errors {
		paths = append(paths, e.FieldPath)
	}
	return fmt.Errorf("expected error at field path %q, got paths: %v", fieldPath, paths)
}

func (w *testWorld) theErrorSuggestsValidTargets(suggestions string) error {
	if w.result == nil {
		return fmt.Errorf("no validation result available")
	}
	for _, e := range w.result.Errors {
		if strings.Contains(e.Remediation, "claude") || strings.Contains(e.Expected, "claude") {
			return nil
		}
	}
	return fmt.Errorf("expected error remediation to suggest valid targets, got: %+v", w.result.Errors)
}

func (w *testWorld) itReportsAtLeastNErrors(n int) error {
	if w.result == nil {
		return fmt.Errorf("no validation result available")
	}
	if len(w.result.Errors) < n {
		return fmt.Errorf("expected at least %d errors, got %d", n, len(w.result.Errors))
	}
	return nil
}

func (w *testWorld) allErrorsIncludeFieldPathsAndRemediation() error {
	if w.result == nil {
		return fmt.Errorf("no validation result available")
	}
	for _, e := range w.result.Errors {
		if e.FieldPath == "" {
			return fmt.Errorf("error missing field path: %+v", e)
		}
		if e.Remediation == "" {
			return fmt.Errorf("error missing remediation: %+v", e)
		}
	}
	return nil
}

func (w *testWorld) itReportsWarningAbout(field string) error {
	if w.result == nil {
		return fmt.Errorf("no validation result available")
	}
	for _, w2 := range w.result.Warnings {
		if strings.Contains(w2.FieldPath, field) || strings.Contains(w2.Message, field) {
			return nil
		}
	}
	return fmt.Errorf("expected warning about %q, got warnings: %+v", field, w.result.Warnings)
}

func (w *testWorld) validationSucceeds() error {
	if w.result == nil {
		return fmt.Errorf("no validation result available")
	}
	if !w.result.Valid {
		return fmt.Errorf("expected validation to succeed but it failed: %+v", w.result.Errors)
	}
	return nil
}

func (w *testWorld) itReportsSuccessToStdout() error {
	if w.exitCode == -1 {
		return fmt.Errorf("command was not executed")
	}
	if v := w.stdout; v == "" || !strings.Contains(v, "valid") {
		return fmt.Errorf("expected stdout to contain 'valid', got: %q", v)
	}
	return nil
}

func (w *testWorld) itReportsErrorsToStderr() error {
	if w.exitCode == -1 {
		return fmt.Errorf("command was not executed")
	}
	if w.stderr == "" {
		return fmt.Errorf("expected stderr output but got empty")
	}
	return nil
}

func (w *testWorld) itReportsErrorAboutDuplicateTargets() error {
	if w.result == nil {
		return fmt.Errorf("no validation result available")
	}
	for _, e := range w.result.Errors {
		if strings.Contains(strings.ToLower(e.Message), "duplicate") {
			return nil
		}
	}
	return fmt.Errorf("expected error about duplicate targets, got: %+v", w.result.Errors)
}
