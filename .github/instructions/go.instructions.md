---
applyTo: "**/*.go"
---

# Go Code Review Instructions

## Error Handling

- Functions that can fail MUST return `error` as the last return value
- Never use `panic` for expected error conditions
- Wrap errors with context: `fmt.Errorf("operation context: %w", err)`
- Use `errors.Is` and `errors.As` for error type checking
- Validation functions MUST collect all errors, never early-return on the first

```go
// Flag this pattern:
if err != nil {
    panic(err)
}

// Expect this pattern:
if err != nil {
    return nil, fmt.Errorf("loading config from %s: %w", path, err)
}
```

## Validation Error Completeness

Every `ValidationError` must have all fields populated:
- `FieldPath` — never empty, use dot notation (`targets`, `targets[0]`)
- `Expected` — what the schema requires
- `Actual` — what was found (empty string only when truly missing)
- `Message` — human-readable description
- `Remediation` — always a concrete action the developer can take
- `Severity` — `"error"` or `"warning"`, nothing else

Flag any ValidationError construction that omits `Remediation`.

## Output Routing

Check that error/warning output goes to `os.Stderr` and success output goes to `os.Stdout`:

```go
// CORRECT
fmt.Fprintln(os.Stderr, formatter.FormatErrors(result))
fmt.Fprintln(os.Stdout, formatter.FormatSuccess(result))

// FLAG THIS — errors to stdout
fmt.Println(formatter.FormatErrors(result))
```

## Testing

- Every exported function should have corresponding tests
- Test files must be in the same package (`_test.go` suffix)
- Use `testify/assert` for non-fatal assertions, `testify/require` for preconditions
- Use `t.TempDir()` for any test that touches the filesystem
- Table-driven tests preferred for multiple input scenarios

```go
// Expect this pattern for multiple scenarios:
tests := []struct {
    name   string
    input  string
    expect bool
}{
    {"valid case", "claude", true},
    {"invalid case", "vscode", false},
}
for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        assert.Equal(t, tt.expect, IsValid(tt.input))
    })
}
```

## Interface Design

- Keep interfaces small (1-3 methods)
- Define interfaces where they are consumed, not where they are implemented
- Accept interfaces, return concrete types

## Package Organization

- `internal/` packages must not be imported from outside the module
- Each package should have a single, clear responsibility
- No circular dependencies between packages

## Cobra Commands

- Use `RunE` (not `Run`) so errors propagate properly
- Global flags on root via `PersistentFlags`
- Config loading in `PersistentPreRunE` on root command
- Validate `--format` flag value (only `"human"` or `"json"`)

## JSON Serialization

- Use snake_case JSON tags (`json:"field_path"`)
- Use `*string` for nullable string fields (nil -> JSON `null`)
- Initialize slices with `make([]T, 0)` to serialize as `[]` not `null`
- Pretty-print with `json.MarshalIndent(v, "", "  ")`
- Internal fields (like `Severity`) should not leak into JSON output

## Security

Flag any of these:
- Hardcoded credentials, API keys, or tokens
- Use of `os.Exec` or `exec.Command` with unsanitized input
- File path operations without proper sanitization
- Logging of sensitive data

## Performance

- Config parse + validate should complete in <100ms
- No unnecessary allocations in hot paths
- Use `strings.Builder` for string concatenation in loops
