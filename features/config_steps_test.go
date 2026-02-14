package features

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ailign/cli/internal/cli"
	"github.com/ailign/cli/internal/config"
	"github.com/cucumber/godog"
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

func registerConfigParsingSteps(ctx *godog.ScenarioContext) {
	w := &testWorld{}

	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		w.dir, _ = os.MkdirTemp("", "ailign-bdd-*")
		w.cfg = nil
		w.loadErr = nil
		w.result = nil
		w.stdout = ""
		w.stderr = ""
		w.exitCode = -1
		return ctx, nil
	})

	ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		if w.dir != "" {
			os.RemoveAll(w.dir)
		}
		return ctx, nil
	})

	// Given steps (past tense)
	ctx.Given(`^a repository contained a valid \.ailign\.yml with targets "([^"]*)"$`, w.aRepoWithValidConfig)
	ctx.Given(`^a repository had no \.ailign\.yml file$`, w.aRepoWithNoConfig)
	ctx.Given(`^a repository contained an empty \.ailign\.yml file$`, w.aRepoWithEmptyConfig)
	ctx.Given(`^a \.ailign\.yml was configured with UTF-8 BOM and targets "([^"]*)"$`, w.aConfigWithBOM)
	ctx.Given(`^a \.ailign\.yml was configured with targets "([^"]*)"$`, w.aConfigWithTargets)
	ctx.Given(`^a \.ailign\.yml was configured with targets "([^"]*)" and unknown field "([^"]*)"$`, w.aConfigWithTargetsAndUnknownField)
	ctx.Given(`^a \.ailign\.yml was configured with no targets field$`, w.aConfigWithNoTargetsField)

	// When steps
	ctx.When(`^the CLI parses the configuration$`, w.theCLIParsesConfig)
	ctx.When(`^the CLI attempts to load configuration$`, w.theCLIAttemptsToLoadConfig)
	ctx.When(`^the CLI attempts to parse it$`, w.theCLIAttemptsToParse)
	ctx.When(`^the CLI validates the configuration$`, w.theCLIValidatesConfig)
	ctx.When(`^the developer runs ailign validate$`, w.theDeveloperRunsValidate)

	// Then steps (future tense)
	ctx.Then(`^the targets will be loaded successfully$`, w.targetsLoadedSuccessfully)
	ctx.Then(`^the loaded targets will be "([^"]*)"$`, w.theLoadedTargetsAre)
	ctx.Then(`^it will report an error containing "([^"]*)" to stderr$`, w.itReportsErrorContaining)
	ctx.Then(`^it will exit with code (\d+)$`, w.itExitsWithCode)
	ctx.Then(`^it will report a validation error about missing "([^"]*)" field$`, w.itReportsValidationErrorAboutMissing)
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

func (w *testWorld) aRepoWithValidConfig(targets string) error {
	return w.writeConfigWithTargets(targets)
}

func (w *testWorld) aRepoWithNoConfig() error {
	return nil
}

func (w *testWorld) aRepoWithEmptyConfig() error {
	return os.WriteFile(filepath.Join(w.dir, ".ailign.yml"), []byte(""), 0644)
}

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

func (w *testWorld) theCLIParsesConfig() error {
	cfgPath := filepath.Join(w.dir, ".ailign.yml")
	w.cfg, w.loadErr = config.LoadFromFile(cfgPath)
	return nil
}

func (w *testWorld) theCLIAttemptsToLoadConfig() error {
	cfgPath := filepath.Join(w.dir, ".ailign.yml")
	w.cfg, w.loadErr = config.LoadFromFile(cfgPath)
	return nil
}

func (w *testWorld) theCLIAttemptsToParse() error {
	cfgPath := filepath.Join(w.dir, ".ailign.yml")
	w.cfg, w.loadErr = config.LoadFromFile(cfgPath)
	if w.loadErr == nil && w.cfg != nil {
		w.result = config.Validate(w.cfg)
	}
	return nil
}

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
	os.Chdir(w.dir)
	defer os.Chdir(origDir)

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

func (w *testWorld) targetsLoadedSuccessfully() error {
	if w.loadErr != nil {
		return fmt.Errorf("expected no error but got: %v", w.loadErr)
	}
	if w.cfg == nil {
		return fmt.Errorf("expected config to be loaded but it is nil")
	}
	return nil
}

func (w *testWorld) theLoadedTargetsAre(expected string) error {
	if w.cfg == nil {
		return fmt.Errorf("config is nil")
	}
	expectedTargets := strings.Split(expected, ",")
	if len(w.cfg.Targets) != len(expectedTargets) {
		return fmt.Errorf("expected %d targets, got %d", len(expectedTargets), len(w.cfg.Targets))
	}
	for i, t := range expectedTargets {
		if w.cfg.Targets[i] != strings.TrimSpace(t) {
			return fmt.Errorf("expected target[%d] = %q, got %q", i, strings.TrimSpace(t), w.cfg.Targets[i])
		}
	}
	return nil
}

func (w *testWorld) itReportsErrorContaining(substring string) error {
	if w.loadErr == nil {
		return fmt.Errorf("expected an error but got none")
	}
	if !strings.Contains(w.loadErr.Error(), substring) {
		return fmt.Errorf("expected error to contain %q, got: %v", substring, w.loadErr)
	}
	return nil
}

func (w *testWorld) itExitsWithCode(code int) error {
	// If a CLI command was executed, use its exit code
	if w.exitCode != -1 {
		if w.exitCode != code {
			return fmt.Errorf("expected exit code %d, got %d", code, w.exitCode)
		}
		return nil
	}
	// Otherwise infer from library-level results
	if code == 0 {
		if w.loadErr != nil {
			return fmt.Errorf("expected exit code 0 (no error) but got error: %v", w.loadErr)
		}
		if w.result != nil && !w.result.Valid {
			return fmt.Errorf("expected exit code 0 but validation failed")
		}
	} else if code == 2 {
		hasError := w.loadErr != nil || (w.result != nil && !w.result.Valid)
		if !hasError {
			return fmt.Errorf("expected exit code 2 (error) but no error occurred")
		}
	}
	return nil
}

func (w *testWorld) itReportsValidationErrorAboutMissing(field string) error {
	if w.result == nil {
		return fmt.Errorf("no validation result available")
	}
	if w.result.Valid {
		return fmt.Errorf("expected validation to fail but it passed")
	}
	for _, e := range w.result.Errors {
		if strings.Contains(strings.ToLower(e.Message), "missing") ||
			strings.Contains(strings.ToLower(e.Message), "required") {
			if strings.Contains(e.FieldPath, field) ||
				strings.Contains(strings.ToLower(e.Remediation), field) {
				return nil
			}
		}
	}
	return fmt.Errorf("expected a validation error about missing %q field, got errors: %+v", field, w.result.Errors)
}

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

// --- helpers ---

func (w *testWorld) writeConfigWithTargets(targets string) error {
	targetList := strings.Split(targets, ",")
	var yaml strings.Builder
	yaml.WriteString("targets:\n")
	for _, t := range targetList {
		yaml.WriteString(fmt.Sprintf("  - %s\n", strings.TrimSpace(t)))
	}
	return os.WriteFile(filepath.Join(w.dir, ".ailign.yml"), []byte(yaml.String()), 0644)
}
