package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidationError_Fields(t *testing.T) {
	err := ValidationError{
		FieldPath:   "targets[0]",
		Expected:    "one of claude, cursor, copilot, windsurf",
		Actual:      "vscode",
		Message:     "invalid target name",
		Remediation: "Use a supported target name",
		Severity:    "error",
	}
	assert.Equal(t, "targets[0]", err.FieldPath)
	assert.Equal(t, "one of claude, cursor, copilot, windsurf", err.Expected)
	assert.Equal(t, "vscode", err.Actual)
	assert.Equal(t, "invalid target name", err.Message)
	assert.Equal(t, "Use a supported target name", err.Remediation)
	assert.Equal(t, "error", err.Severity)
}

func TestValidationError_WarningSeverity(t *testing.T) {
	err := ValidationError{
		FieldPath:   "custom_field",
		Message:     "unrecognized field",
		Remediation: "Remove it or check for typos",
		Severity:    "warning",
	}
	assert.Equal(t, "warning", err.Severity)
}

func TestValidationError_MissingActual(t *testing.T) {
	err := ValidationError{
		FieldPath: "targets",
		Expected:  "array of target names",
		Actual:    "",
		Message:   "required field missing",
	}
	assert.Empty(t, err.Actual)
}

func TestValidationResult_Valid(t *testing.T) {
	cfg := &Config{Targets: []string{"claude"}}
	result := &ValidationResult{
		Valid:  true,
		Config: cfg,
	}
	assert.True(t, result.Valid)
	assert.Empty(t, result.Errors)
	assert.Empty(t, result.Warnings)
	assert.NotNil(t, result.Config)
}

func TestValidationResult_WithErrors(t *testing.T) {
	result := &ValidationResult{
		Valid: false,
		Errors: []ValidationError{
			{FieldPath: "targets", Message: "required field missing", Severity: "error"},
		},
	}
	assert.False(t, result.Valid)
	assert.Len(t, result.Errors, 1)
	assert.Nil(t, result.Config)
}

func TestValidationResult_WithWarnings(t *testing.T) {
	cfg := &Config{Targets: []string{"claude"}}
	result := &ValidationResult{
		Valid: true,
		Warnings: []ValidationError{
			{FieldPath: "custom", Message: "unrecognized field", Severity: "warning"},
		},
		Config: cfg,
	}
	assert.True(t, result.Valid)
	assert.Empty(t, result.Errors)
	assert.Len(t, result.Warnings, 1)
	assert.NotNil(t, result.Config)
}

func TestValidationResult_ErrorsAndWarnings(t *testing.T) {
	result := &ValidationResult{
		Valid: false,
		Errors: []ValidationError{
			{FieldPath: "targets", Message: "required field missing", Severity: "error"},
		},
		Warnings: []ValidationError{
			{FieldPath: "extra", Message: "unrecognized field", Severity: "warning"},
		},
	}
	assert.False(t, result.Valid)
	assert.Len(t, result.Errors, 1)
	assert.Len(t, result.Warnings, 1)
}
