---
name: go-error-handling
description: 'Go error handling patterns for AIlign CLI. Use when writing error types, validation logic, error formatting, or any code that produces or handles errors. Ensures consistency with the collect-all-errors, field-path, and remediation patterns defined in the project constitution and data model.'
---

# Go Error Handling Patterns

Error handling patterns specific to AIlign CLI, aligned with Constitution Principle II (Transparency) and Principle III (Fail Safe).

## Core Rule: Collect All Errors

NEVER return on the first error. Always collect all errors and report them together so developers can fix everything in one pass.

```go
// CORRECT: collect all errors
var errs []ValidationError
if missing targets {
    errs = append(errs, ValidationError{...})
}
if invalid target name {
    errs = append(errs, ValidationError{...})
}
return errs

// WRONG: early return on first error
if missing targets {
    return error
}
```

## ValidationError Structure

Every validation error MUST include all fields:

```go
type ValidationError struct {
    FieldPath   string // dot notation + array indices: "targets", "targets[0]"
    Expected    string // what the schema requires
    Actual      string // what was found (empty string if missing)
    Message     string // human-readable description
    Remediation string // concrete action to fix the issue
    Severity    string // "error" or "warning"
}
```

### Field Rules

- **FieldPath**: Use dot notation and array indices (`targets`, `targets[0]`, `packages[1].version`)
- **Expected**: Describe what the schema expects (`"array of target names (claude, cursor, copilot, windsurf)"`)
- **Actual**: The value found. Leave empty (not "null" string) when the field is missing entirely
- **Message**: Short description (`"required field missing"`, `"invalid target name"`)
- **Remediation**: Always a concrete action (`"Add a \"targets\" field with at least one target"`)
- **Severity**: `"error"` for schema violations, `"warning"` for unknown fields

## Severity Rules

| Condition | Severity | Exit Code |
|-----------|----------|-----------|
| Schema violation (missing required, wrong type, invalid value) | `error` | 2 |
| Unknown/unrecognized field | `warning` | 0 (proceed) |
| File not found | `error` | 2 |
| YAML parse error | `error` | 2 |
| Permission error | `error` | 2 |

Warnings NEVER block execution. Only errors cause exit code 2.

## Standard Go Error Patterns

### Return errors, don't panic

```go
// CORRECT
func LoadFromFile(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("reading config: %w", err)
    }
    return &cfg, nil
}

// WRONG
func LoadFromFile(path string) *Config {
    data, err := os.ReadFile(path)
    if err != nil {
        panic(err) // never panic
    }
    return &cfg
}
```

### Wrap errors with context

```go
// CORRECT: adds context while preserving the original error
return nil, fmt.Errorf("loading config from %s: %w", path, err)

// WRONG: loses the original error
return nil, fmt.Errorf("failed to load config")
```

### Use errors.Is and errors.As for checking

```go
if errors.Is(err, os.ErrNotExist) {
    // file not found handling
}
```

## Error Output Routing

- Validation errors and warnings go to **stderr**
- Success messages go to **stdout**
- JSON output follows the same routing (valid result to stdout, errors to stderr)

## Remediation Guidelines

Every error message MUST include actionable remediation. Examples:

| Error | Remediation |
|-------|-------------|
| Missing `targets` field | `Add a "targets" field with at least one target` |
| Invalid target name | `Use a supported target name: claude, cursor, copilot, windsurf` |
| Empty targets array | `Add at least one target to the "targets" array` |
| Duplicate target | `Remove the duplicate target entry` |
| Unknown field | `Remove it or check for typos` |
| File not found | `Run "ailign init" to create a configuration file` |
| YAML parse error | `Check YAML syntax at line N` |
