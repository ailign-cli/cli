package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Helper: assert that a ValidationError looks like a proper error (not warning)
// ---------------------------------------------------------------------------

func assertIsValidationError(t *testing.T, ve ValidationError) {
	t.Helper()
	assert.Equal(t, "error", ve.Severity, "expected severity 'error'")
	assert.NotEmpty(t, ve.FieldPath, "FieldPath must be populated")
	assert.NotEmpty(t, ve.Message, "Message must not be empty")
	assert.NotEmpty(t, ve.Remediation, "Remediation must not be empty")
}

// ---------------------------------------------------------------------------
// Validate() tests
// ---------------------------------------------------------------------------

func TestValidate_MultipleValidTargets(t *testing.T) {
	cfg := &Config{
		Targets: []string{"claude", "cursor", "copilot"},
	}

	result := Validate(cfg)

	require.NotNil(t, result)
	assert.True(t, result.Valid)
	assert.Empty(t, result.Errors)
	assert.Empty(t, result.Warnings)
	assert.Equal(t, cfg, result.Config)
}

func TestValidate_SingleValidTarget(t *testing.T) {
	cfg := &Config{
		Targets: []string{"windsurf"},
	}

	result := Validate(cfg)

	require.NotNil(t, result)
	assert.True(t, result.Valid)
	assert.Empty(t, result.Errors)
	assert.Empty(t, result.Warnings)
	assert.Equal(t, cfg, result.Config)
}

func TestValidate_AllFourTargets(t *testing.T) {
	cfg := &Config{
		Targets: []string{"claude", "cursor", "copilot", "windsurf"},
	}

	result := Validate(cfg)

	require.NotNil(t, result)
	assert.True(t, result.Valid)
	assert.Empty(t, result.Errors)
}

func TestValidate_NilTargets(t *testing.T) {
	cfg := &Config{
		Targets: nil,
	}

	result := Validate(cfg)

	require.NotNil(t, result)
	assert.False(t, result.Valid)
	require.NotEmpty(t, result.Errors)

	// At least one error must reference the "targets" field
	found := false
	for _, ve := range result.Errors {
		assertIsValidationError(t, ve)
		if ve.FieldPath == "targets" {
			found = true
		}
	}
	assert.True(t, found, "expected an error with FieldPath 'targets'")
}

func TestValidate_EmptyTargetsArray(t *testing.T) {
	cfg := &Config{
		Targets: []string{},
	}

	result := Validate(cfg)

	require.NotNil(t, result)
	assert.False(t, result.Valid)
	require.NotEmpty(t, result.Errors)

	// Should report a minItems violation on "targets"
	found := false
	for _, ve := range result.Errors {
		assertIsValidationError(t, ve)
		if ve.FieldPath == "targets" {
			found = true
		}
	}
	assert.True(t, found, "expected a minItems error with FieldPath 'targets'")
}

func TestValidate_InvalidTargetName(t *testing.T) {
	cfg := &Config{
		Targets: []string{"vscode"},
	}

	result := Validate(cfg)

	require.NotNil(t, result)
	assert.False(t, result.Valid)
	require.NotEmpty(t, result.Errors)

	// Should report an enum violation pointing at the item
	found := false
	for _, ve := range result.Errors {
		assertIsValidationError(t, ve)
		// Field path should indicate the specific array item, e.g. "targets[0]"
		if assert.Contains(t, ve.FieldPath, "targets") {
			found = true
		}
	}
	assert.True(t, found, "expected an error referencing the invalid target item")
}

func TestValidate_InvalidTargetAmongValid(t *testing.T) {
	cfg := &Config{
		Targets: []string{"claude", "vscode", "cursor"},
	}

	result := Validate(cfg)

	require.NotNil(t, result)
	assert.False(t, result.Valid)
	require.NotEmpty(t, result.Errors)

	for _, ve := range result.Errors {
		assertIsValidationError(t, ve)
	}

	// The error should pinpoint the invalid item, not the valid ones
	foundInvalid := false
	for _, ve := range result.Errors {
		if ve.FieldPath == "targets[1]" {
			foundInvalid = true
			assert.Contains(t, ve.Actual, "vscode")
		}
	}
	assert.True(t, foundInvalid, "expected an error at targets[1] for 'vscode'")
}

func TestValidate_DuplicateTargets(t *testing.T) {
	cfg := &Config{
		Targets: []string{"claude", "claude"},
	}

	result := Validate(cfg)

	require.NotNil(t, result)
	assert.False(t, result.Valid)
	require.NotEmpty(t, result.Errors)

	found := false
	for _, ve := range result.Errors {
		assertIsValidationError(t, ve)
		if ve.FieldPath == "targets" {
			found = true
		}
	}
	assert.True(t, found, "expected a uniqueItems error with FieldPath 'targets'")
}

func TestValidate_MultipleErrorsAtOnce(t *testing.T) {
	// Invalid target name AND duplicate -- both should be reported
	cfg := &Config{
		Targets: []string{"vscode", "vscode"},
	}

	result := Validate(cfg)

	require.NotNil(t, result)
	assert.False(t, result.Valid)
	// We expect at least 2 errors: enum violation(s) + uniqueItems violation
	assert.GreaterOrEqual(t, len(result.Errors), 2,
		"expected at least 2 errors (enum + uniqueItems)")

	for _, ve := range result.Errors {
		assertIsValidationError(t, ve)
	}
}

func TestValidate_InvalidResultHasNilConfig(t *testing.T) {
	cfg := &Config{Targets: nil}

	result := Validate(cfg)

	require.NotNil(t, result)
	assert.False(t, result.Valid)
	// When validation fails, Config in the result should be nil to prevent
	// callers from accidentally using an invalid config.
	assert.Nil(t, result.Config)
}

func TestValidate_ValidResultHasConfig(t *testing.T) {
	cfg := &Config{Targets: []string{"claude"}}

	result := Validate(cfg)

	require.NotNil(t, result)
	assert.True(t, result.Valid)
	assert.Equal(t, cfg, result.Config)
}

// ---------------------------------------------------------------------------
// DetectUnknownFields() tests
// ---------------------------------------------------------------------------

func TestDetectUnknownFields_NoUnknownFields(t *testing.T) {
	rawYAML := []byte("targets:\n  - claude\n  - cursor\n")

	warnings := DetectUnknownFields(rawYAML)

	assert.Empty(t, warnings)
}

func TestDetectUnknownFields_OneUnknownField(t *testing.T) {
	rawYAML := []byte("targets:\n  - claude\ncustom_field: hello\n")

	warnings := DetectUnknownFields(rawYAML)

	require.Len(t, warnings, 1)
	assert.Equal(t, "warning", warnings[0].Severity)
	assert.Equal(t, "custom_field", warnings[0].FieldPath)
	assert.NotEmpty(t, warnings[0].Message)
	assert.NotEmpty(t, warnings[0].Remediation)
}

func TestDetectUnknownFields_MultipleUnknownFields(t *testing.T) {
	rawYAML := []byte("targets:\n  - claude\nextra_one: foo\nextra_two: bar\n")

	warnings := DetectUnknownFields(rawYAML)

	require.Len(t, warnings, 2)

	fieldPaths := make([]string, len(warnings))
	for i, w := range warnings {
		assert.Equal(t, "warning", w.Severity)
		assert.NotEmpty(t, w.Message)
		assert.NotEmpty(t, w.Remediation)
		fieldPaths[i] = w.FieldPath
	}
	assert.Contains(t, fieldPaths, "extra_one")
	assert.Contains(t, fieldPaths, "extra_two")
}

func TestDetectUnknownFields_OnlyUnknownFields(t *testing.T) {
	// YAML with no known fields at all
	rawYAML := []byte("foo: 1\nbar: 2\n")

	warnings := DetectUnknownFields(rawYAML)

	require.Len(t, warnings, 2)
	for _, w := range warnings {
		assert.Equal(t, "warning", w.Severity)
		assert.NotEmpty(t, w.FieldPath)
		assert.NotEmpty(t, w.Message)
		assert.NotEmpty(t, w.Remediation)
	}
}

func TestDetectUnknownFields_EmptyYAML(t *testing.T) {
	rawYAML := []byte("")

	warnings := DetectUnknownFields(rawYAML)

	assert.Empty(t, warnings)
}

func TestDetectUnknownFields_WarningsAreNotErrors(t *testing.T) {
	rawYAML := []byte("targets:\n  - claude\nunknown: value\n")

	warnings := DetectUnknownFields(rawYAML)

	require.NotEmpty(t, warnings)
	for _, w := range warnings {
		assert.Equal(t, "warning", w.Severity,
			"DetectUnknownFields should produce warnings, not errors")
	}
}
