package steps

import (
	"context"
	"os"
	"testing"

	"github.com/cucumber/godog"
)

func TestFeatures(t *testing.T) {
	format := "pretty"
	if output := os.Getenv("CUCUMBER_REPORT"); output != "" {
		format = "cucumber:" + output
	}

	suite := godog.TestSuite{
		ScenarioInitializer: InitializeScenario,
		Options: &godog.Options{
			Format:   format,
			Paths:    []string{"../"},
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
			if removeErr := os.RemoveAll(w.dir); removeErr != nil {
				return ctx, removeErr
			}
		}
		return ctx, nil
	})

	registerConfigParsingSteps(ctx, w)
	registerSchemaValidationSteps(ctx, w)
}
