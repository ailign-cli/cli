package output

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHumanFormatterImplementsFormatter(t *testing.T) {
	var _ Formatter = &HumanFormatter{}
}

func TestHumanFormatSuccess_NoWarnings(t *testing.T) {
	f := &HumanFormatter{}
	result := ValidationResult{
		Valid: true,
		File:  ".ailign.yml",
	}

	got := f.FormatSuccess(result)

	assert.Equal(t, ".ailign.yml: valid\n", got)
}

func TestHumanFormatSuccess_OneWarning(t *testing.T) {
	f := &HumanFormatter{}
	result := ValidationResult{
		Valid: true,
		File:  ".ailign.yml",
		Warnings: []ValidationError{
			{FieldPath: "custom_field", Message: "unrecognized field"},
		},
	}

	got := f.FormatSuccess(result)

	assert.Equal(t, ".ailign.yml: valid (1 warning)\n", got)
}

func TestHumanFormatSuccess_MultipleWarnings(t *testing.T) {
	f := &HumanFormatter{}
	result := ValidationResult{
		Valid: true,
		File:  ".ailign.yml",
		Warnings: []ValidationError{
			{FieldPath: "custom_field", Message: "unrecognized field"},
			{FieldPath: "another_field", Message: "unrecognized field"},
			{FieldPath: "third_field", Message: "unrecognized field"},
		},
	}

	got := f.FormatSuccess(result)

	assert.Equal(t, ".ailign.yml: valid (3 warnings)\n", got)
}

func TestHumanFormatErrors_SingleError_WithActual(t *testing.T) {
	f := &HumanFormatter{}
	result := ValidationResult{
		Valid: false,
		File:  ".ailign.yml",
		Errors: []ValidationError{
			{
				FieldPath:   "targets[0]",
				Message:     "unknown target",
				Expected:    "one of: claude, cursor, copilot, windsurf",
				Actual:      "vscode",
				Remediation: "Use a supported target name",
			},
		},
	}

	got := f.FormatErrors(result)

	expected := `Error: .ailign.yml validation failed

  targets[0]: unknown target
    Expected: one of: claude, cursor, copilot, windsurf
    Found: vscode
    Fix: Use a supported target name

1 error found
`
	assert.Equal(t, expected, got)
}

func TestHumanFormatErrors_SingleError_WithoutActual(t *testing.T) {
	f := &HumanFormatter{}
	result := ValidationResult{
		Valid: false,
		File:  ".ailign.yml",
		Errors: []ValidationError{
			{
				FieldPath:   "targets",
				Message:     "required field missing",
				Expected:    "array of target names (claude, cursor, copilot, windsurf)",
				Actual:      "",
				Remediation: `Add a "targets" field with at least one target`,
			},
		},
	}

	got := f.FormatErrors(result)

	expected := `Error: .ailign.yml validation failed

  targets: required field missing
    Expected: array of target names (claude, cursor, copilot, windsurf)
    Fix: Add a "targets" field with at least one target

1 error found
`
	assert.Equal(t, expected, got)
}

func TestHumanFormatErrors_MultipleErrors(t *testing.T) {
	f := &HumanFormatter{}
	result := ValidationResult{
		Valid: false,
		File:  ".ailign.yml",
		Errors: []ValidationError{
			{
				FieldPath:   "targets",
				Message:     "required field missing",
				Expected:    "array of target names (claude, cursor, copilot, windsurf)",
				Remediation: `Add a "targets" field with at least one target`,
			},
			{
				FieldPath:   "version",
				Message:     "invalid value",
				Expected:    "1",
				Actual:      "99",
				Remediation: `Set version to "1"`,
			},
		},
	}

	got := f.FormatErrors(result)

	expected := `Error: .ailign.yml validation failed

  targets: required field missing
    Expected: array of target names (claude, cursor, copilot, windsurf)
    Fix: Add a "targets" field with at least one target

  version: invalid value
    Expected: 1
    Found: 99
    Fix: Set version to "1"

2 errors found
`
	assert.Equal(t, expected, got)
}

func TestHumanFormatWarnings_SingleWarning(t *testing.T) {
	f := &HumanFormatter{}
	result := ValidationResult{
		Valid: true,
		File:  ".ailign.yml",
		Warnings: []ValidationError{
			{
				FieldPath:   "custom_field",
				Message:     "unrecognized field",
				Remediation: "Remove it or check for typos",
			},
		},
	}

	got := f.FormatWarnings(result)

	expected := `Warning: .ailign.yml has warnings

  custom_field: unrecognized field
    Fix: Remove it or check for typos

1 warning found
`
	assert.Equal(t, expected, got)
}

func TestHumanFormatWarnings_MultipleWarnings(t *testing.T) {
	f := &HumanFormatter{}
	result := ValidationResult{
		Valid: true,
		File:  ".ailign.yml",
		Warnings: []ValidationError{
			{
				FieldPath:   "custom_field",
				Message:     "unrecognized field",
				Remediation: "Remove it or check for typos",
			},
			{
				FieldPath:   "extra",
				Message:     "unrecognized field",
				Remediation: "Remove it or check for typos",
			},
		},
	}

	got := f.FormatWarnings(result)

	expected := `Warning: .ailign.yml has warnings

  custom_field: unrecognized field
    Fix: Remove it or check for typos

  extra: unrecognized field
    Fix: Remove it or check for typos

2 warnings found
`
	assert.Equal(t, expected, got)
}

func TestHumanFormatWarnings_WithExpectedAndActual(t *testing.T) {
	f := &HumanFormatter{}
	result := ValidationResult{
		Valid: true,
		File:  ".ailign.yml",
		Warnings: []ValidationError{
			{
				FieldPath:   "version",
				Message:     "outdated version",
				Expected:    "2",
				Actual:      "1",
				Remediation: "Consider upgrading to version 2",
			},
		},
	}

	got := f.FormatWarnings(result)

	expected := `Warning: .ailign.yml has warnings

  version: outdated version
    Expected: 2
    Found: 1
    Fix: Consider upgrading to version 2

1 warning found
`
	assert.Equal(t, expected, got)
}

func TestHumanFormatWarnings_WithExpectedButNoActual(t *testing.T) {
	f := &HumanFormatter{}
	result := ValidationResult{
		Valid: true,
		File:  ".ailign.yml",
		Warnings: []ValidationError{
			{
				FieldPath:   "description",
				Message:     "recommended field missing",
				Expected:    "a short description of the project",
				Remediation: "Add a description field",
			},
		},
	}

	got := f.FormatWarnings(result)

	expected := `Warning: .ailign.yml has warnings

  description: recommended field missing
    Expected: a short description of the project
    Fix: Add a description field

1 warning found
`
	assert.Equal(t, expected, got)
}
