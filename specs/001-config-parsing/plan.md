# Implementation Plan: Configuration File Parsing

**Branch**: `001-config-parsing` | **Date**: 2026-02-13 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/001-config-parsing/spec.md`

## Summary

Parse and validate `.ailign.yml` configuration files from the working
directory using Go. The config file uses YAML format; validation is
performed via an embedded JSONSchema definition. The schema for this
feature is limited to the `targets` field only (`packages` and
`local_overlays` will be added by future features). Includes an
`ailign validate` command for explicit standalone validation. All
validation errors are reported at once with field paths and
remediation guidance. Per-target rendering implementation is out of
scope; only the target registry and interface are defined here.

## Technical Context

**Language/Version**: Go 1.24+ (targeting Go 1.26)
**Primary Dependencies**: Cobra (CLI), goccy/go-yaml (YAML), santhosh-tekuri/jsonschema v6 (validation), testify (testing)
**Storage**: N/A (file system read-only, single `.ailign.yml` file)
**Testing**: `go test` + testify (assert, require)
**Target Platform**: Cross-platform (Linux, macOS, Windows). Single static binary.
**Project Type**: Single CLI application
**Performance Goals**: Config parse + validate <100ms
**Constraints**: Single binary <50MB, <100MB memory, zero runtime dependencies
**Scale/Scope**: Single config file per repository, ~10-50 lines typical

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Evidence |
|-----------|--------|----------|
| I. CLI-First | PASS | `ailign validate` command, `--format` flag for JSON/human output, exit codes 0/2, stderr for errors |
| II. Transparency | PASS | All errors include field path + expected + actual + remediation. No hidden behavior. |
| III. Fail Safe | PASS | Validate before any operations. All errors reported at once. Exit code 2 on failure. No partial state changes. |
| IV. Test-First | PASS | TDD workflow planned. >90% coverage target for config package. Integration tests for CLI commands. |
| V. Composition | PASS | Target interface separated from config parsing. Each package has single responsibility. |
| VI. Governance | PASS | Schema is versioned (`$id` field). Config format is deterministic. |
| VII. Size-Aware | N/A | Config files are small (not a rendering feature). |
| VIII. Cross-Tool Parity | PASS | Targets are equal citizens in registry. No tool is primary. Config is tool-agnostic. |

**Gate result**: PASS - all applicable principles satisfied.

*Post-design re-check (Phase 1 complete)*: All gates still pass.
Schema is minimal (targets only). Design is simple with no
unnecessary abstractions. YAGNI respected.

## Project Structure

### Documentation (this feature)

```text
specs/001-config-parsing/
├── plan.md              # This file
├── research.md          # Phase 0: technology decisions
├── data-model.md        # Phase 1: entity definitions
├── quickstart.md        # Phase 1: developer quickstart
├── contracts/
│   ├── config-schema.json   # JSONSchema for .ailign.yml
│   └── cli-commands.md      # CLI command contracts
└── tasks.md             # Phase 2 output (/speckit.tasks)
```

### Source Code (repository root)

```text
cmd/
└── ailign/
    └── main.go                 # Entry point (thin, wires Cobra)

internal/
├── config/
│   ├── config.go               # Config struct + PackageRef type
│   ├── loader.go               # Load + parse YAML from file
│   ├── validator.go            # JSONSchema validation logic
│   ├── errors.go               # ValidationError + ValidationResult types
│   ├── schema.json             # Embedded JSONSchema (go:embed)
│   ├── config_test.go          # Unit tests for config types
│   ├── loader_test.go          # Unit tests for loader
│   ├── validator_test.go       # Unit tests for validator
│   └── errors_test.go          # Unit tests for error formatting
├── cli/
│   ├── root.go                 # Root command, global --format flag
│   ├── validate.go             # `ailign validate` command
│   ├── root_test.go            # Integration tests for root
│   └── validate_test.go        # Integration tests for validate
├── output/
│   ├── formatter.go            # Formatter interface
│   ├── human.go                # Human-readable error/success output
│   ├── json.go                 # JSON error/success output
│   ├── human_test.go           # Unit tests
│   └── json_test.go            # Unit tests
└── target/
    ├── registry.go             # Known targets + Target interface
    └── registry_test.go        # Unit tests

go.mod
go.sum
```

**Structure Decision**: Standard Go CLI layout using `cmd/` for the
entry point and `internal/` for all application code. Go's `internal`
directory enforces that nothing outside the module can import these
packages, providing encapsulation. Tests live alongside source files
following Go conventions.

## Design Decisions

### YAML Config + JSONSchema Validation

The config file is authored in YAML (developer-friendly, supports
comments) but validated using JSONSchema (industry standard, tooling
ecosystem). The validation pipeline:

1. Read `.ailign.yml` from working directory
2. Parse YAML into `Config` struct via goccy/go-yaml
3. Marshal struct to JSON in-memory
4. Validate JSON against embedded JSONSchema
5. Separate pass: detect unknown fields, emit as warnings
6. Transform validation errors into user-friendly messages

The JSONSchema is embedded in the binary via `go:embed`, ensuring
the schema is always available without external files.

### Target Interface

The `target` package defines:
- A `Target` interface (methods TBD by future features)
- A registry of known target names (`claude`, `cursor`, `copilot`,
  `windsurf`)
- A `IsValid(name string) bool` function for validation

Per-target implementation is explicitly out of scope. The interface
exists to establish the contract for future features.

### Output Formatting

The `output` package provides a `Formatter` interface with two
implementations:
- `HumanFormatter`: Indented, colored (if terminal), readable output
- `JSONFormatter`: Machine-parseable JSON output

Selected via `--format` flag on all commands. Default is `human`.

### Error Handling Strategy

Validation errors are collected (never early-exit) and transformed
into `ValidationError` structs with:
- `field_path`: JSONPath-style (e.g., `packages[0]`, `targets`)
- `expected`: What the schema requires
- `actual`: What was found (null if missing)
- `remediation`: Concrete action to fix

This aligns with Constitution II (Transparency) and III (Fail Safe).

## Complexity Tracking

No constitution violations to justify. Design is minimal and
follows all principles.

## Artifacts Generated

| Artifact | Path | Description |
|----------|------|-------------|
| Research | [research.md](research.md) | Technology decisions and rationale |
| Data Model | [data-model.md](data-model.md) | Entity definitions and relationships |
| Config Schema | [contracts/config-schema.json](contracts/config-schema.json) | JSONSchema for `.ailign.yml` |
| CLI Contracts | [contracts/cli-commands.md](contracts/cli-commands.md) | Command signatures, flags, output formats |
| Quickstart | [quickstart.md](quickstart.md) | Developer setup and usage guide |
