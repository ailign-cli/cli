package steps

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ailign/cli/internal/config"
	"github.com/cucumber/godog"
)

func registerConfigParsingSteps(ctx *godog.ScenarioContext, w *testWorld) {
	// Given steps (past tense) — config file setup, shared across features
	ctx.Given(`^a repository contained a valid \.ailign\.yml with targets "([^"]*)"$`, w.aRepoWithValidConfig)
	ctx.Given(`^a repository had no \.ailign\.yml file$`, w.aRepoWithNoConfig)
	ctx.Given(`^a repository contained an empty \.ailign\.yml file$`, w.aRepoWithEmptyConfig)

	// When steps (present tense)
	ctx.When(`^the CLI parses the configuration$`, w.theCLIParsesConfig)
	ctx.When(`^the CLI attempts to load configuration$`, w.theCLIAttemptsToLoadConfig)
	ctx.When(`^the CLI attempts to parse it$`, w.theCLIAttemptsToParse)

	// Then steps (future tense) — parsing results, some shared across features
	ctx.Then(`^the targets will be loaded successfully$`, w.targetsLoadedSuccessfully)
	ctx.Then(`^the loaded targets will be "([^"]*)"$`, w.theLoadedTargetsAre)
	ctx.Then(`^it will report an error containing "([^"]*)" to stderr$`, w.itReportsErrorContaining)
	ctx.Then(`^it will exit with code (\d+)$`, w.itExitsWithCode)
	ctx.Then(`^it will report a validation error about missing "([^"]*)" field$`, w.itReportsValidationErrorAboutMissing)
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
	if w.exitCode != -1 {
		if w.exitCode != code {
			return fmt.Errorf("expected exit code %d, got %d", code, w.exitCode)
		}
		return nil
	}
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
