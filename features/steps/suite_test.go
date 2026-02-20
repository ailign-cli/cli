package steps

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/cucumber/godog"
)

func TestFeatures(t *testing.T) {
	format := "pretty"
	if output := os.Getenv("CUCUMBER_REPORT"); output != "" {
		format = "cucumber:" + output
	} else if output := os.Getenv("JUNIT_REPORT"); output != "" {
		format = "junit:" + output
	}

	tags := "~@wip && ~@ci"
	if envTags := os.Getenv("GODOG_TAGS"); envTags != "" {
		tags = envTags
	}

	suite := godog.TestSuite{
		ScenarioInitializer: InitializeScenario,
		Options: &godog.Options{
			Format:   format,
			Paths:    []string{"../"},
			Tags:     tags,
			TestingT: t,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	w := &testWorld{}

	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		w.dir, _ = os.MkdirTemp("", "ailign-bdd-*")
		w.reset()
		return ctx, nil
	})

	ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		if w.dir != "" {
			// Restore write permissions (some tests make dirs read-only)
			_ = filepath.WalkDir(w.dir, func(path string, d os.DirEntry, err error) error {
				if err != nil {
					return nil
				}
				if d.IsDir() {
					_ = os.Chmod(path, 0755)
				}
				return nil
			})
			if removeErr := os.RemoveAll(w.dir); removeErr != nil {
				return ctx, removeErr
			}
		}
		return ctx, nil
	})

	registerConfigParsingSteps(ctx, w)
	registerSchemaValidationSteps(ctx, w)
	registerSyncSteps(ctx, w)
	registerPreviewSteps(ctx, w)
	registerInstallGoSteps(ctx, w)
	registerInstallBinarySteps(ctx, w)
	registerInstallDocsSteps(ctx, w)
}
