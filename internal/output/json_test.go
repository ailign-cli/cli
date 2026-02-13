package output

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

// jsonResult mirrors the expected JSON output structure for unmarshalling.
type jsonResult struct {
	Valid    bool             `json:"valid"`
	Errors   []jsonErrorEntry `json:"errors"`
	Warnings []jsonErrorEntry `json:"warnings"`
	File     string           `json:"file"`
}

// jsonErrorEntry mirrors a single error/warning entry in JSON output.
type jsonErrorEntry struct {
	FieldPath   string  `json:"field_path"`
	Expected    string  `json:"expected"`
	Actual      *string `json:"actual"` // pointer so null decodes as nil
	Message     string  `json:"message"`
	Remediation string  `json:"remediation"`
}

func newJSONFormatter() Formatter {
	return &JSONFormatter{}
}

// --- FormatSuccess ---

func TestJSONFormatterImplementsFormatter(t *testing.T) {
	var _ Formatter = &JSONFormatter{}
}

func TestJSONFormatSuccess_NoWarnings(t *testing.T) {
	f := newJSONFormatter()
	result := ValidationResult{
		Valid:    true,
		Errors:   nil,
		Warnings: nil,
		File:     ".ailign.yml",
	}

	out := f.FormatSuccess(result)

	var parsed jsonResult
	err := json.Unmarshal([]byte(out), &parsed)
	assert.NoError(t, err, "FormatSuccess output must be valid JSON")

	assert.True(t, parsed.Valid)
	assert.Empty(t, parsed.Errors, "errors must be an empty array")
	assert.Empty(t, parsed.Warnings, "warnings must be an empty array")
	assert.Equal(t, ".ailign.yml", parsed.File)
}

func TestJSONFormatSuccess_WithWarnings(t *testing.T) {
	f := newJSONFormatter()
	result := ValidationResult{
		Valid:  true,
		Errors: nil,
		Warnings: []ValidationError{
			{
				FieldPath:   "targets[0].rules",
				Expected:    "non-empty rule set",
				Actual:      "",
				Message:     "target has no rules defined",
				Remediation: "Add at least one rule to the target",
				Severity:    "warning",
			},
		},
		File: ".ailign.yml",
	}

	out := f.FormatSuccess(result)

	var parsed jsonResult
	err := json.Unmarshal([]byte(out), &parsed)
	assert.NoError(t, err, "FormatSuccess output must be valid JSON")

	assert.True(t, parsed.Valid)
	assert.Empty(t, parsed.Errors)
	assert.Len(t, parsed.Warnings, 1)
	assert.Equal(t, "targets[0].rules", parsed.Warnings[0].FieldPath)
	assert.Nil(t, parsed.Warnings[0].Actual, "empty Actual should serialize as JSON null")
	assert.Equal(t, "target has no rules defined", parsed.Warnings[0].Message)
}

// --- FormatErrors ---

func TestJSONFormatErrors_SingleError_ActualAbsent(t *testing.T) {
	f := newJSONFormatter()
	result := ValidationResult{
		Valid: false,
		Errors: []ValidationError{
			{
				FieldPath:   "targets",
				Expected:    "array of target names",
				Actual:      "",
				Message:     "required field missing",
				Remediation: "Add a \"targets\" field with at least one target",
				Severity:    "error",
			},
		},
		Warnings: nil,
		File:     ".ailign.yml",
	}

	out := f.FormatErrors(result)

	var parsed jsonResult
	err := json.Unmarshal([]byte(out), &parsed)
	assert.NoError(t, err, "FormatErrors output must be valid JSON")

	assert.False(t, parsed.Valid)
	assert.Len(t, parsed.Errors, 1)
	assert.Equal(t, "targets", parsed.Errors[0].FieldPath)
	assert.Equal(t, "array of target names", parsed.Errors[0].Expected)
	assert.Nil(t, parsed.Errors[0].Actual, "empty Actual should be JSON null")
	assert.Equal(t, "required field missing", parsed.Errors[0].Message)
	assert.Equal(t, "Add a \"targets\" field with at least one target", parsed.Errors[0].Remediation)
	assert.Empty(t, parsed.Warnings)
	assert.Equal(t, ".ailign.yml", parsed.File)
}

func TestJSONFormatErrors_SingleError_ActualPresent(t *testing.T) {
	f := newJSONFormatter()
	result := ValidationResult{
		Valid: false,
		Errors: []ValidationError{
			{
				FieldPath:   "targets[0]",
				Expected:    "one of: claude, cursor, copilot, windsurf",
				Actual:      "vscode",
				Message:     "unknown target",
				Remediation: "Use a supported target name",
				Severity:    "error",
			},
		},
		Warnings: nil,
		File:     ".ailign.yml",
	}

	out := f.FormatErrors(result)

	var parsed jsonResult
	err := json.Unmarshal([]byte(out), &parsed)
	assert.NoError(t, err)

	assert.False(t, parsed.Valid)
	assert.Len(t, parsed.Errors, 1)
	assert.NotNil(t, parsed.Errors[0].Actual, "non-empty Actual should be a JSON string")
	assert.Equal(t, "vscode", *parsed.Errors[0].Actual)
}

func TestJSONFormatErrors_MultipleErrors(t *testing.T) {
	f := newJSONFormatter()
	result := ValidationResult{
		Valid: false,
		Errors: []ValidationError{
			{
				FieldPath:   "targets",
				Expected:    "array of target names",
				Actual:      "",
				Message:     "required field missing",
				Remediation: "Add a \"targets\" field",
				Severity:    "error",
			},
			{
				FieldPath:   "version",
				Expected:    "1",
				Actual:      "2",
				Message:     "unsupported version",
				Remediation: "Use version 1",
				Severity:    "error",
			},
		},
		Warnings: nil,
		File:     ".ailign.yml",
	}

	out := f.FormatErrors(result)

	var parsed jsonResult
	err := json.Unmarshal([]byte(out), &parsed)
	assert.NoError(t, err)

	assert.False(t, parsed.Valid)
	assert.Len(t, parsed.Errors, 2)

	assert.Equal(t, "targets", parsed.Errors[0].FieldPath)
	assert.Nil(t, parsed.Errors[0].Actual)

	assert.Equal(t, "version", parsed.Errors[1].FieldPath)
	assert.NotNil(t, parsed.Errors[1].Actual)
	assert.Equal(t, "2", *parsed.Errors[1].Actual)
}

// --- FormatWarnings ---

func TestJSONFormatWarnings(t *testing.T) {
	f := newJSONFormatter()
	result := ValidationResult{
		Valid:  true,
		Errors: nil,
		Warnings: []ValidationError{
			{
				FieldPath:   "targets[0].rules",
				Expected:    "non-empty rule set",
				Actual:      "",
				Message:     "target has no rules defined",
				Remediation: "Add at least one rule",
				Severity:    "warning",
			},
			{
				FieldPath:   "targets[1].description",
				Expected:    "non-empty string",
				Actual:      "",
				Message:     "missing description",
				Remediation: "Add a description field",
				Severity:    "warning",
			},
		},
		File: ".ailign.yml",
	}

	out := f.FormatWarnings(result)

	var parsed jsonResult
	err := json.Unmarshal([]byte(out), &parsed)
	assert.NoError(t, err, "FormatWarnings output must be valid JSON")

	assert.True(t, parsed.Valid)
	assert.Empty(t, parsed.Errors)
	assert.Len(t, parsed.Warnings, 2)
	assert.Equal(t, "targets[0].rules", parsed.Warnings[0].FieldPath)
	assert.Equal(t, "targets[1].description", parsed.Warnings[1].FieldPath)
	assert.Equal(t, ".ailign.yml", parsed.File)
}

// --- JSON structure validity ---

func TestJSON_PrettyPrintedWithTwoSpaceIndent(t *testing.T) {
	f := newJSONFormatter()
	result := ValidationResult{
		Valid:    true,
		Errors:   nil,
		Warnings: nil,
		File:     ".ailign.yml",
	}

	out := f.FormatSuccess(result)

	assert.Contains(t, out, "\n")
	assert.Contains(t, out, "  \"valid\"", "JSON must be indented with 2 spaces")
}

func TestJSON_ErrorsAndWarningsAlwaysArrays(t *testing.T) {
	f := newJSONFormatter()
	result := ValidationResult{
		Valid:    true,
		Errors:   nil,
		Warnings: nil,
		File:     ".ailign.yml",
	}

	out := f.FormatSuccess(result)

	// Parse into a raw map to verify the JSON types of errors/warnings.
	var raw map[string]json.RawMessage
	err := json.Unmarshal([]byte(out), &raw)
	assert.NoError(t, err)

	// errors and warnings must start with '[' (array), never be null.
	assert.Equal(t, byte('['), raw["errors"][0], "errors must be a JSON array, not null")
	assert.Equal(t, byte('['), raw["warnings"][0], "warnings must be a JSON array, not null")
}

func TestJSON_FieldNamesUseSnakeCase(t *testing.T) {
	f := newJSONFormatter()
	result := ValidationResult{
		Valid: false,
		Errors: []ValidationError{
			{
				FieldPath:   "x",
				Expected:    "y",
				Actual:      "z",
				Message:     "m",
				Remediation: "r",
				Severity:    "error",
			},
		},
		File: "f.yml",
	}

	out := f.FormatErrors(result)

	assert.Contains(t, out, "\"field_path\"")
	assert.NotContains(t, out, "\"FieldPath\"")
	assert.NotContains(t, out, "\"fieldPath\"")
}

func TestJSON_SeverityOmitted(t *testing.T) {
	f := newJSONFormatter()
	result := ValidationResult{
		Valid: false,
		Errors: []ValidationError{
			{
				FieldPath:   "x",
				Expected:    "y",
				Actual:      "",
				Message:     "m",
				Remediation: "r",
				Severity:    "error",
			},
		},
		File: "f.yml",
	}

	out := f.FormatErrors(result)

	assert.NotContains(t, out, "severity", "severity should not appear in JSON output")
}
