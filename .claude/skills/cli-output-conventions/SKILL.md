---
name: cli-output-conventions
description: 'CLI output conventions for AIlign CLI. Use when writing CLI commands, formatting output, handling exit codes, or routing messages to stdout/stderr. Ensures consistency with Constitution Principle I (CLI-First) and Principle II (Transparency).'
---

# CLI Output Conventions

Output routing, exit codes, and formatting rules for AIlign CLI, aligned with Constitution Principle I (CLI-First with Dual Output) and Principle II (Transparency).

## Exit Codes

| Code | Meaning | When |
|------|---------|------|
| 0 | Success | Valid config, command completed, warnings present but no errors |
| 2 | Error | Validation failure, missing file, parse error, any fatal condition |

Note: Exit code 1 is reserved for future use (drift/outdated detection).

Warnings NEVER cause a non-zero exit code.

## Output Routing

### stdout (success, results)

- Success confirmation messages
- Valid command output
- JSON results when `--format json` and operation succeeded

### stderr (errors, warnings, diagnostics)

- All validation errors
- All warnings (unknown fields, deprecations)
- File-not-found messages
- YAML parse errors
- Any diagnostic information

```go
// CORRECT
fmt.Fprintln(os.Stdout, formatter.FormatSuccess(result))
fmt.Fprintln(os.Stderr, formatter.FormatErrors(result))
fmt.Fprintln(os.Stderr, formatter.FormatWarnings(result))

// WRONG: errors to stdout
fmt.Println(formatter.FormatErrors(result))
```

## The --format Flag

Global flag inherited by all commands. Two values:

| Value | Description | Default |
|-------|-------------|---------|
| `human` | Indented, readable text output | Yes |
| `json` | Machine-parseable JSON output | No |

### Registration (Cobra)

```go
rootCmd.PersistentFlags().StringVarP(&formatFlag, "format", "f", "human",
    "Output format: human or json")
```

### Usage pattern

```go
func getFormatter(format string) output.Formatter {
    switch format {
    case "json":
        return &output.JSONFormatter{}
    default:
        return &output.HumanFormatter{}
    }
}
```

## Human Format

### Success (stdout)

```
.ailign.yml: valid
```

### Success with warnings (stdout + stderr)

```
# stderr:
Warning: .ailign.yml has warnings

  custom_field: unrecognized field
    This field is not part of the AIlign config schema
    Fix: Remove it or check for typos

# stdout:
.ailign.yml: valid (1 warning)
```

### Errors (stderr)

```
Error: .ailign.yml validation failed

  targets: required field missing
    Expected: array of target names (claude, cursor, copilot, windsurf)
    Fix: Add a "targets" field with at least one target

  targets[0]: invalid target name
    Expected: one of claude, cursor, copilot, windsurf
    Found: "vscode"
    Fix: Use a supported target name

2 errors found
```

### Error entry format

```
  {field_path}: {message}
    Expected: {expected}
    Found: {actual}          ← omit line if actual is empty
    Fix: {remediation}
```

## JSON Format

### Success (stdout)

```json
{
  "valid": true,
  "errors": [],
  "warnings": [],
  "file": ".ailign.yml"
}
```

### Errors (stderr)

```json
{
  "valid": false,
  "errors": [
    {
      "field_path": "targets",
      "expected": "array of target names",
      "actual": null,
      "message": "required field missing",
      "remediation": "Add a \"targets\" field with at least one target"
    }
  ],
  "warnings": [],
  "file": ".ailign.yml"
}
```

### JSON rules

- Pretty-printed with 2-space indentation
- `actual` is JSON `null` when field is missing, string when present
- `errors` and `warnings` are always arrays (empty `[]`, never `null`)
- `severity` field is NOT included in JSON output
- Snake_case keys (`field_path`, not `fieldPath`)

## Command Output Pattern

Standard pattern for any command that uses validation:

```go
func runCommand(cmd *cobra.Command, args []string) error {
    format, _ := cmd.Flags().GetString("format")
    formatter := getFormatter(format)

    // Load and validate config
    result := validate(cfg)

    // Route output correctly
    if len(result.Warnings) > 0 {
        fmt.Fprint(os.Stderr, formatter.FormatWarnings(result))
    }

    if !result.Valid {
        fmt.Fprint(os.Stderr, formatter.FormatErrors(result))
        os.Exit(2)
    }

    fmt.Fprint(os.Stdout, formatter.FormatSuccess(result))
    return nil
}
```

## No Interactive Prompts

Per Constitution Principle I, the CLI MUST be automation-friendly:

- No interactive prompts by default
- No "Are you sure?" confirmations
- No spinners or progress bars that break pipe usage
- All input via args, flags, or stdin
- All output parseable when using `--format json`

## Cobra Command Conventions

### Command structure

```go
var validateCmd = &cobra.Command{
    Use:   "validate",
    Short: "Validate the .ailign.yml configuration file",
    RunE:  runValidate,
}
```

### Global flags via PersistentFlags on root

```go
// root.go
rootCmd.PersistentFlags().StringVarP(&formatFlag, "format", "f", "human",
    "Output format: human or json")
```

### Config loading in PersistentPreRunE

```go
// root.go — runs before every subcommand
rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
    cfg, err := config.LoadFromFile(filepath.Join(".", ".ailign.yml"))
    if err != nil {
        // format and print error to stderr
        os.Exit(2)
    }
    // store cfg for subcommands
    return nil
}
```
