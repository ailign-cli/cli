# AIlign CLI — Copilot Review Instructions

## Project Overview

AIlign is an instruction governance and distribution CLI for engineering organizations. It manages AI coding assistant instructions across tools (Claude, Cursor, Copilot, Windsurf) and repositories. Single Go binary, zero runtime dependencies.

## Constitution Principles

Every PR must comply with these core principles. Flag violations as blocking comments.

### I. CLI-First with Dual Output
- All features accessible via CLI with `--format human` (default) and `--format json`
- Errors to stderr, success to stdout
- Exit codes: 0=success, 1=drift/outdated, 2=error
- No interactive prompts — automation-friendly
- Single binary, zero runtime dependencies

### II. Transparency Over Magic (NON-NEGOTIABLE)
- Every action must be explainable
- Deterministic, predictable behavior
- Error messages must explain what went wrong AND how to fix it
- No silent failures, no black-box operations, no "trust me" behavior

### III. Fail Safe, Not Silent
- Validate before modifying files
- Atomic operations (all-or-nothing)
- Never swallow errors or leave system in inconsistent state
- Collect ALL validation errors before reporting (never early-exit on first error)

### IV. Test-First Development
- TDD mandatory: tests written and failing before implementation
- Core libraries: >90% coverage
- Renderers: 100% coverage
- CLI commands: integration tests required

### V. Composition Over Monolith
- Each package has single responsibility
- No hidden dependencies between packages
- Adding tool support must not affect existing renderers

### VI. Governance as Foundation
- All packages versioned and immutable (semver)
- Changes traceable, provenance clear
- Lock files prevent unexpected changes

### VII. Size-Aware by Design
- Renderers must respect tool size limits
- Content tiers: critical/recommended/extra
- Never silently truncate content

### VIII. Cross-Tool Parity
- No tool is "primary" — all are equal citizens
- Central content is tool-agnostic
- Tool-specific formats are render targets, not sources

## Prohibited Patterns

Flag these as blocking issues in any PR:

- Silent auto-updates
- Mutable packages
- Unexplainable merges
- Single-tool lock-in
- Interactive prompts in automation paths
- Hidden failures or swallowed errors
- Unbounded content without size checks
- Code shipped without tests

## Error Handling Patterns

### Validation errors must include all fields
- `field_path`: dot notation path (e.g., `targets`, `targets[0]`)
- `expected`: what the schema requires
- `actual`: what was found (empty/null when missing)
- `message`: human-readable description
- `remediation`: concrete action to fix the issue

### Always collect all errors
```go
// CORRECT: collect all errors
var errs []ValidationError
// ... check each rule, append errors
return errs

// WRONG: early return
if err != nil {
    return err
}
```

### Wrap errors with context
```go
// CORRECT
return fmt.Errorf("loading config from %s: %w", path, err)

// WRONG
return fmt.Errorf("failed to load config")
```

## Output Routing

- `stdout`: success messages, valid command output
- `stderr`: all errors, all warnings, diagnostics
- JSON format: pretty-printed, 2-space indent, `null` for missing values, arrays never `null` (use `[]`)

## Go Conventions

- Standard Go project layout: `cmd/` for entry point, `internal/` for all packages
- Tests alongside source files (`*_test.go`)
- Use `testify/assert` for non-fatal checks, `testify/require` for fatal preconditions
- Table-driven tests for multiple scenarios
- Use `t.TempDir()` for filesystem tests
- Return errors, never panic
- Use `errors.Is` / `errors.As` for error checking

## Commit Messages

Follow Conventional Commits: `<type>[!]: <description>`
- Types: feat, fix, chore, docs, test, refactor, perf, build, ci, revert
- Breaking changes: append `!` (e.g., `feat!: remove deprecated variable`)
- Lowercase, imperative mood, no period, under 72 characters

## What to Check in PR Reviews

1. **Constitution compliance**: Does the change violate any principle?
2. **Test coverage**: Are there tests? Do they cover error cases and edge cases?
3. **Error handling**: Are errors collected (not early-returned)? Do they include remediation?
4. **Output routing**: Errors to stderr? Success to stdout? Both formats working?
5. **No over-engineering**: Is the change minimal and focused? No unnecessary abstractions?
6. **Security**: No hardcoded secrets, credentials, or API keys
7. **Breaking changes**: If any, is there a migration guide?
