---
name: tdd-workflow
description: 'TDD workflow and Go testing patterns for AIlign CLI. Use when writing tests, implementing features, or reviewing test coverage. Enforces Constitution Principle IV (Test-First Development) with red-green-refactor cycle, table-driven tests, and coverage requirements.'
---

# TDD Workflow

Test-First Development patterns for AIlign CLI, enforcing Constitution Principle IV.

## The Cycle: Red-Green-Refactor

TDD is mandatory. Every feature follows this strict order:

1. **Red**: Write a test that fails (function/type doesn't exist yet)
2. **Green**: Write the minimum code to make the test pass
3. **Refactor**: Clean up while keeping tests green

NEVER write implementation before tests. Verify tests fail before implementing.

```bash
# 1. Write test → verify it FAILS
go test ./internal/config/ -run TestLoadFromFile
# Expected: compilation error or test failure

# 2. Implement → verify it PASSES
go test ./internal/config/ -run TestLoadFromFile
# Expected: PASS

# 3. Refactor → verify still PASSES
go test ./internal/config/ -v
# Expected: all PASS
```

## Test Naming Convention

Format: `Test<Function>_<Scenario>`

```go
// Unit tests
func TestIsValid_KnownTargets(t *testing.T)
func TestIsValid_UnknownTargets(t *testing.T)
func TestLoadFromFile_ValidConfig(t *testing.T)
func TestLoadFromFile_MissingFile(t *testing.T)
func TestLoadFromFile_EmptyFile(t *testing.T)
func TestValidate_InvalidTargetName(t *testing.T)

// Integration tests
func TestRootCommand_ValidConfig(t *testing.T)
func TestValidateCommand_ExitCode2OnError(t *testing.T)
```

## Table-Driven Tests

Use table-driven tests for multiple scenarios of the same function:

```go
func TestIsValid(t *testing.T) {
    tests := []struct {
        name   string
        input  string
        expect bool
    }{
        {"known target claude", "claude", true},
        {"known target cursor", "cursor", true},
        {"unknown target", "vscode", false},
        {"empty string", "", false},
        {"case sensitive", "Claude", false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            assert.Equal(t, tt.expect, IsValid(tt.input))
        })
    }
}
```

## Testify Usage

Use `assert` for non-fatal checks, `require` for fatal preconditions:

```go
import (
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestLoadFromFile_ValidConfig(t *testing.T) {
    // require: test cannot continue if this fails
    cfg, err := LoadFromFile(path)
    require.NoError(t, err)
    require.NotNil(t, cfg)

    // assert: check individual fields, continue on failure
    assert.Len(t, cfg.Targets, 2)
    assert.Contains(t, cfg.Targets, "claude")
    assert.Contains(t, cfg.Targets, "cursor")
}
```

## Test File Organization

Tests live alongside source files per Go conventions:

```
internal/config/
├── config.go
├── config_test.go      # unit tests for config types
├── loader.go
├── loader_test.go      # unit tests for loader
├── validator.go
├── validator_test.go   # unit tests for validator
├── errors.go
└── errors_test.go      # unit tests for error types
```

## Temp Directories for File Tests

Use `t.TempDir()` for tests that need filesystem access:

```go
func TestLoadFromFile_MissingFile(t *testing.T) {
    dir := t.TempDir()
    _, err := LoadFromFile(filepath.Join(dir, ".ailign.yml"))
    assert.Error(t, err)
}

func TestLoadFromFile_ValidConfig(t *testing.T) {
    dir := t.TempDir()
    configPath := filepath.Join(dir, ".ailign.yml")
    os.WriteFile(configPath, []byte("targets:\n  - claude\n"), 0644)

    cfg, err := LoadFromFile(configPath)
    require.NoError(t, err)
    assert.Equal(t, []string{"claude"}, cfg.Targets)
}
```

## Coverage Requirements

From the constitution:

| Package | Minimum Coverage |
|---------|-----------------|
| `internal/config/` | >90% |
| `internal/target/` | >90% |
| `internal/output/` | >90% |
| Renderers (future) | 100% |
| `internal/cli/` | Integration tests required |

Check coverage:

```bash
go test -coverprofile=coverage.out ./internal/...
go tool cover -func=coverage.out
```

## What to Test

### Always test

- Happy path (valid input, expected output)
- Error cases (missing file, invalid input, parse errors)
- Edge cases (empty input, boundary values, duplicates)
- Return values AND side effects

### For validation specifically

- Each validation rule independently
- Multiple errors at once (collect-all pattern)
- Warnings vs errors (severity distinction)
- Error field completeness (field_path, expected, actual, remediation)

### For CLI commands

- Exit codes (0 for success, 2 for error)
- Stdout vs stderr routing
- `--format human` and `--format json` output
- Missing config file behavior

## Execution Order in Implementation

When implementing a feature with multiple components:

1. Write tests for types/structs (they'll fail — types don't exist)
2. Implement types/structs (tests pass)
3. Write tests for core logic (they'll fail — functions don't exist)
4. Implement core logic (tests pass)
5. Write tests for CLI wiring (they'll fail — commands don't exist)
6. Implement CLI wiring (tests pass)
7. Run full suite, check coverage
