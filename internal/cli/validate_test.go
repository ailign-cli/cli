package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// executeCommand creates a fresh root command, sets the given args, and
// executes it with the working directory changed to dir. It captures stdout
// and stderr via cmd.SetOut/SetErr and returns the output along with an
// exit code (0 for success, non-zero for error).
func executeCommand(args []string, dir string) (stdout string, stderr string, exitCode int) {
	rootCmd := NewRootCommand()

	stdoutBuf := new(bytes.Buffer)
	stderrBuf := new(bytes.Buffer)
	rootCmd.SetOut(stdoutBuf)
	rootCmd.SetErr(stderrBuf)
	rootCmd.SetArgs(args)

	origDir, _ := os.Getwd()
	_ = os.Chdir(dir)
	defer func() { _ = os.Chdir(origDir) }()

	err := rootCmd.Execute()
	code := 0
	if err != nil {
		code = 2
	}
	return stdoutBuf.String(), stderrBuf.String(), code
}

// ---------------------------------------------------------------------------
// 1. Valid .ailign.yml -> exit code 0, stdout contains ".ailign.yml: valid"
// ---------------------------------------------------------------------------

func TestValidate_ValidConfig_ExitZero(t *testing.T) {
	dir := t.TempDir()
	cfgContent := "targets:\n  - claude\n  - cursor\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".ailign.yml"), []byte(cfgContent), 0644))

	stdout, stderr, exitCode := executeCommand([]string{"validate"}, dir)

	assert.Equal(t, 0, exitCode, "valid config should exit 0")
	assert.Contains(t, stdout, ".ailign.yml: valid")
	assert.Empty(t, stderr, "stderr should be empty for valid config without warnings")
}

func TestValidate_ValidConfig_SingleTarget(t *testing.T) {
	dir := t.TempDir()
	cfgContent := "targets:\n  - copilot\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".ailign.yml"), []byte(cfgContent), 0644))

	stdout, _, exitCode := executeCommand([]string{"validate"}, dir)

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout, ".ailign.yml: valid")
}

func TestValidate_ValidConfig_AllTargets(t *testing.T) {
	dir := t.TempDir()
	cfgContent := "targets:\n  - claude\n  - cursor\n  - copilot\n  - windsurf\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".ailign.yml"), []byte(cfgContent), 0644))

	stdout, _, exitCode := executeCommand([]string{"validate"}, dir)

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout, ".ailign.yml: valid")
}

// ---------------------------------------------------------------------------
// 2. Invalid .ailign.yml (bad target) -> exit code non-zero, stderr has error
// ---------------------------------------------------------------------------

func TestValidate_InvalidTarget_ExitNonZero(t *testing.T) {
	dir := t.TempDir()
	cfgContent := "targets:\n  - vscode\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".ailign.yml"), []byte(cfgContent), 0644))

	stdout, stderr, exitCode := executeCommand([]string{"validate"}, dir)

	assert.NotEqual(t, 0, exitCode, "invalid config should exit non-zero")
	assert.NotEmpty(t, stderr, "stderr should contain validation error")
	assert.Contains(t, stderr, "vscode", "stderr should mention the invalid target")
	assert.Empty(t, stdout, "stdout should be empty when validation fails")
}

func TestValidate_InvalidTarget_AmongValid(t *testing.T) {
	dir := t.TempDir()
	cfgContent := "targets:\n  - claude\n  - badtool\n  - cursor\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".ailign.yml"), []byte(cfgContent), 0644))

	_, stderr, exitCode := executeCommand([]string{"validate"}, dir)

	assert.NotEqual(t, 0, exitCode)
	assert.Contains(t, stderr, "badtool")
}

func TestValidate_EmptyTargetsArray_ExitNonZero(t *testing.T) {
	dir := t.TempDir()
	cfgContent := "targets: []\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".ailign.yml"), []byte(cfgContent), 0644))

	_, stderr, exitCode := executeCommand([]string{"validate"}, dir)

	assert.NotEqual(t, 0, exitCode, "empty targets array should be invalid")
	assert.NotEmpty(t, stderr)
}

func TestValidate_MissingTargetsField_ExitNonZero(t *testing.T) {
	dir := t.TempDir()
	// Valid YAML but missing required "targets" field
	cfgContent := "some_field: hello\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".ailign.yml"), []byte(cfgContent), 0644))

	_, stderr, exitCode := executeCommand([]string{"validate"}, dir)

	assert.NotEqual(t, 0, exitCode, "missing targets field should be invalid")
	assert.NotEmpty(t, stderr)
}

func TestValidate_DuplicateTargets_ExitNonZero(t *testing.T) {
	dir := t.TempDir()
	cfgContent := "targets:\n  - claude\n  - claude\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".ailign.yml"), []byte(cfgContent), 0644))

	_, stderr, exitCode := executeCommand([]string{"validate"}, dir)

	assert.NotEqual(t, 0, exitCode, "duplicate targets should be invalid")
	assert.NotEmpty(t, stderr)
}

// ---------------------------------------------------------------------------
// 3. Missing .ailign.yml -> exit code non-zero, stderr contains "not found"
// ---------------------------------------------------------------------------

func TestValidate_MissingFile_ExitNonZero(t *testing.T) {
	dir := t.TempDir()
	// Do NOT create .ailign.yml

	_, stderr, exitCode := executeCommand([]string{"validate"}, dir)

	assert.NotEqual(t, 0, exitCode, "missing .ailign.yml should exit non-zero")
	assert.Contains(t, stderr, "not found", "stderr should indicate file not found")
}

func TestValidate_MissingFile_StdoutEmpty(t *testing.T) {
	dir := t.TempDir()

	stdout, _, exitCode := executeCommand([]string{"validate"}, dir)

	assert.NotEqual(t, 0, exitCode)
	assert.Empty(t, stdout, "stdout should be empty when file is missing")
}

// ---------------------------------------------------------------------------
// 4. --format json with valid file -> stdout is valid JSON with "valid": true
// ---------------------------------------------------------------------------

func TestValidate_FormatJSON_ValidConfig(t *testing.T) {
	dir := t.TempDir()
	cfgContent := "targets:\n  - claude\n  - cursor\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".ailign.yml"), []byte(cfgContent), 0644))

	stdout, stderr, exitCode := executeCommand([]string{"validate", "--format", "json"}, dir)

	assert.Equal(t, 0, exitCode, "valid config with --format json should exit 0")
	assert.Empty(t, stderr)

	var result map[string]interface{}
	err := json.Unmarshal([]byte(stdout), &result)
	require.NoError(t, err, "stdout must be valid JSON, got: %s", stdout)

	assert.Equal(t, true, result["valid"], "JSON output should have valid: true")
	assert.Equal(t, ".ailign.yml", result["file"], "JSON output should include file name")
}

func TestValidate_FormatJSON_ValidConfig_ErrorsAndWarningsAreArrays(t *testing.T) {
	dir := t.TempDir()
	cfgContent := "targets:\n  - claude\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".ailign.yml"), []byte(cfgContent), 0644))

	stdout, _, exitCode := executeCommand([]string{"validate", "--format", "json"}, dir)

	assert.Equal(t, 0, exitCode)

	var raw map[string]json.RawMessage
	err := json.Unmarshal([]byte(stdout), &raw)
	require.NoError(t, err)

	// errors and warnings must be arrays (not null)
	assert.Equal(t, byte('['), raw["errors"][0], "errors should be a JSON array")
	assert.Equal(t, byte('['), raw["warnings"][0], "warnings should be a JSON array")
}

func TestValidate_FormatJSON_ValidConfig_ShortFlag(t *testing.T) {
	dir := t.TempDir()
	cfgContent := "targets:\n  - windsurf\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".ailign.yml"), []byte(cfgContent), 0644))

	stdout, _, exitCode := executeCommand([]string{"validate", "-f", "json"}, dir)

	assert.Equal(t, 0, exitCode)

	var result map[string]interface{}
	err := json.Unmarshal([]byte(stdout), &result)
	require.NoError(t, err, "stdout must be valid JSON when using -f shorthand")
	assert.Equal(t, true, result["valid"])
}

// ---------------------------------------------------------------------------
// 5. --format json with invalid file -> stderr is valid JSON with "valid": false
// ---------------------------------------------------------------------------

func TestValidate_FormatJSON_InvalidConfig(t *testing.T) {
	dir := t.TempDir()
	cfgContent := "targets:\n  - vscode\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".ailign.yml"), []byte(cfgContent), 0644))

	stdout, stderr, exitCode := executeCommand([]string{"validate", "--format", "json"}, dir)

	assert.NotEqual(t, 0, exitCode, "invalid config with --format json should exit non-zero")
	assert.Empty(t, stdout, "stdout should be empty for invalid config in JSON mode")

	var result map[string]interface{}
	err := json.Unmarshal([]byte(stderr), &result)
	require.NoError(t, err, "stderr must be valid JSON, got: %s", stderr)

	assert.Equal(t, false, result["valid"], "JSON output should have valid: false")
	assert.Equal(t, ".ailign.yml", result["file"])
}

func TestValidate_FormatJSON_InvalidConfig_HasErrors(t *testing.T) {
	dir := t.TempDir()
	cfgContent := "targets:\n  - vscode\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".ailign.yml"), []byte(cfgContent), 0644))

	_, stderr, exitCode := executeCommand([]string{"validate", "--format", "json"}, dir)

	assert.NotEqual(t, 0, exitCode)

	var result struct {
		Valid  bool `json:"valid"`
		Errors []struct {
			FieldPath   string  `json:"field_path"`
			Message     string  `json:"message"`
			Expected    string  `json:"expected"`
			Actual      *string `json:"actual"`
			Remediation string  `json:"remediation"`
		} `json:"errors"`
	}
	err := json.Unmarshal([]byte(stderr), &result)
	require.NoError(t, err, "stderr must be valid JSON")

	assert.False(t, result.Valid)
	assert.NotEmpty(t, result.Errors, "errors array should not be empty for invalid config")

	// At least one error should reference the invalid target
	foundTarget := false
	for _, e := range result.Errors {
		assert.NotEmpty(t, e.FieldPath)
		assert.NotEmpty(t, e.Message)
		if e.Actual != nil && *e.Actual == "vscode" {
			foundTarget = true
		}
	}
	assert.True(t, foundTarget, "expected an error referencing 'vscode'")
}

func TestValidate_FormatJSON_EmptyTargets(t *testing.T) {
	dir := t.TempDir()
	cfgContent := "targets: []\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".ailign.yml"), []byte(cfgContent), 0644))

	_, stderr, exitCode := executeCommand([]string{"validate", "--format", "json"}, dir)

	assert.NotEqual(t, 0, exitCode)

	var result map[string]interface{}
	err := json.Unmarshal([]byte(stderr), &result)
	require.NoError(t, err, "stderr must be valid JSON for empty targets")
	assert.Equal(t, false, result["valid"])
}

func TestValidate_FormatJSON_MissingFile(t *testing.T) {
	dir := t.TempDir()

	_, stderr, exitCode := executeCommand([]string{"validate", "--format", "json"}, dir)

	assert.NotEqual(t, 0, exitCode)
	assert.Contains(t, stderr, "not found", "stderr should mention file not found even in JSON mode")
}

// ---------------------------------------------------------------------------
// 6. Valid file with unknown fields -> exit 0, stdout "valid", stderr warning
// ---------------------------------------------------------------------------

func TestValidate_UnknownFields_ExitZeroWithWarning(t *testing.T) {
	dir := t.TempDir()
	cfgContent := "targets:\n  - claude\ncustom_field: hello\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".ailign.yml"), []byte(cfgContent), 0644))

	stdout, stderr, exitCode := executeCommand([]string{"validate"}, dir)

	assert.Equal(t, 0, exitCode, "unknown fields should not cause validation failure")
	assert.Contains(t, stdout, ".ailign.yml: valid", "stdout should indicate validity")
	assert.NotEmpty(t, stderr, "stderr should contain a warning about unknown fields")
	assert.Contains(t, stderr, "custom_field", "warning should mention the unknown field name")
}

func TestValidate_UnknownFields_StdoutIndicatesWarningCount(t *testing.T) {
	dir := t.TempDir()
	cfgContent := "targets:\n  - claude\nextra_one: foo\nextra_two: bar\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".ailign.yml"), []byte(cfgContent), 0644))

	stdout, stderr, exitCode := executeCommand([]string{"validate"}, dir)

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout, "valid")
	assert.Contains(t, stdout, "warning", "stdout should mention warnings exist")
	assert.Contains(t, stderr, "extra_one")
	assert.Contains(t, stderr, "extra_two")
}

func TestValidate_FormatJSON_UnknownFields(t *testing.T) {
	dir := t.TempDir()
	cfgContent := "targets:\n  - cursor\nunknown_key: value\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".ailign.yml"), []byte(cfgContent), 0644))

	stdout, stderr, exitCode := executeCommand([]string{"validate", "--format", "json"}, dir)

	assert.Equal(t, 0, exitCode, "unknown fields with valid targets should exit 0")

	// For JSON format with warnings, the JSON goes to stdout (valid: true)
	// and warnings are included in the JSON output.
	var result struct {
		Valid    bool `json:"valid"`
		Warnings []struct {
			FieldPath string `json:"field_path"`
			Message   string `json:"message"`
		} `json:"warnings"`
	}

	// The primary JSON output should be on stdout since the config is valid
	jsonOutput := stdout
	if jsonOutput == "" {
		// Fall back to stderr if the implementation puts all JSON there
		jsonOutput = stderr
	}

	err := json.Unmarshal([]byte(jsonOutput), &result)
	require.NoError(t, err, "output must be valid JSON, got stdout=%q stderr=%q", stdout, stderr)

	assert.True(t, result.Valid, "config with unknown fields should still be valid")
	assert.NotEmpty(t, result.Warnings, "warnings should include unknown field info")

	foundUnknown := false
	for _, w := range result.Warnings {
		if w.FieldPath == "unknown_key" {
			foundUnknown = true
		}
	}
	assert.True(t, foundUnknown, "expected a warning for 'unknown_key'")
}

// ---------------------------------------------------------------------------
// Edge cases
// ---------------------------------------------------------------------------

func TestValidate_MalformedYAML_ExitNonZero(t *testing.T) {
	dir := t.TempDir()
	cfgContent := "targets:\n  - claude\n bad: [yaml\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".ailign.yml"), []byte(cfgContent), 0644))

	_, stderr, exitCode := executeCommand([]string{"validate"}, dir)

	assert.NotEqual(t, 0, exitCode, "malformed YAML should exit non-zero")
	assert.NotEmpty(t, stderr, "stderr should contain parse error details")
}

func TestValidate_EmptyFile_ExitNonZero(t *testing.T) {
	dir := t.TempDir()
	// Empty file parses to zero-value Config with nil targets,
	// which is invalid (targets is required and must have at least one item).
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".ailign.yml"), []byte(""), 0644))

	_, stderr, exitCode := executeCommand([]string{"validate"}, dir)

	assert.NotEqual(t, 0, exitCode, "empty config should be invalid (no targets)")
	assert.NotEmpty(t, stderr)
}
