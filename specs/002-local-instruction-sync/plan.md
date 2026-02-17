# Implementation Plan: Local Instruction Sync

**Branch**: `002-local-instruction-sync` | **Date**: 2026-02-17 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/002-local-instruction-sync/spec.md`

## Summary

Add an `ailign sync` command that composes local overlay files into
a single central hub file (`.ailign/instructions.md`) and creates
symlinks from each target's expected instruction path to the hub.
This centralizes instruction storage while making content available
at all tool-specific paths via symlinks.

The feature extends the config schema with `local_overlays`, refactors
the target registry into a modular interface+registry pattern (like
Terraform providers), and implements atomic file writes with symlink
management. Windows symlink support is out of scope (separate feature).

## Technical Context

**Language/Version**: Go 1.24+ (targeting Go 1.26)
**Primary Dependencies**: Cobra (CLI), goccy/go-yaml (YAML), santhosh-tekuri/jsonschema v6 (validation), testify (TDD), godog (BDD)
**Storage**: File system — reads overlay files, writes `.ailign/instructions.md`, creates symlinks
**Testing**: godog (BDD) + `go test` + testify (TDD)
**Target Platform**: Cross-platform (Linux, macOS, Windows). Single static binary.
**Project Type**: Single CLI application
**Performance Goals**: Sync 3 overlays to 4 targets <1 second
**Constraints**: Single binary <50MB, <100MB memory, zero runtime dependencies
**Scale/Scope**: 1-10 overlay files, 1-4 targets, <1MB total content

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Evidence |
|-----------|--------|----------|
| I. CLI-First | PASS | `ailign sync` command, `--dry-run` flag, `--format` for JSON/human, exit codes 0/2, stderr for errors |
| II. Transparency | PASS | `--dry-run` previews all changes. Symlinks are visible (`ls -la`). Managed-content header shows origin. |
| III. Fail Safe | PASS | Atomic writes (write-temp-rename). Validate config before writing. All errors reported with file paths and remediation. |
| IV. Test-First | PASS | BDD scenarios for both user stories. TDD for sync logic, target implementations, symlink management. |
| V. Composition | PASS | Overlays composed in order. Targets are modular (interface+registry). Each target is independent. |
| VI. Governance | PASS | Schema versioned (`$id` v2). Hub files are deterministic. Managed header tracks source files. |
| VII. Size-Aware | N/A | Size limits not enforced in this feature (future enhancement). |
| VIII. Cross-Tool Parity | PASS | All targets receive equal treatment. No target is primary. Same compose logic for all. |

**Gate result**: PASS — all applicable principles satisfied.

*Post-design re-check (Phase 1 complete)*: All gates still pass.
The hub-spoke model with symlinks adds a level of indirection but
improves transparency (symlinks are visible) and prepares for future
registry integration. Target modularity follows YAGNI — each target
implementation is minimal (~20 lines) with a clear extension point.

## Project Structure

### Documentation (this feature)

```text
specs/002-local-instruction-sync/
├── plan.md              # This file
├── research.md          # Phase 0: architecture and technology decisions
├── data-model.md        # Phase 1: entity definitions
├── quickstart.md        # Phase 1: developer usage guide
├── contracts/
│   ├── cli-commands.md      # CLI command contracts
│   └── config-schema-v2.json # Updated JSONSchema
└── tasks.md             # /speckit.tasks output

features/                                    # BDD feature files (Gherkin)
├── sync-local-instructions.feature          # US1: core sync
├── preview-sync-changes.feature             # US2: dry-run
└── steps/
    ├── suite_test.go                        # godog test runner
    ├── world_test.go                        # shared testWorld struct
    ├── config_parsing_steps_test.go         # (existing from 001)
    ├── schema_validation_steps_test.go      # (existing from 001)
    ├── sync_steps_test.go                   # US1 step definitions
    └── preview_steps_test.go               # US2 step definitions
```

### Source Code (repository root)

```text
cmd/
└── ailign/
    └── main.go                    # Entry point (unchanged)

internal/
├── config/
│   ├── config.go                  # Config struct (add LocalOverlays field)
│   ├── loader.go                  # LoadAndValidate (unchanged logic)
│   ├── validator.go               # JSONSchema validation (unchanged logic)
│   ├── errors.go                  # Error types (unchanged)
│   ├── schema.json                # Updated schema (add local_overlays)
│   ├── config_test.go             # Updated tests
│   ├── loader_test.go             # Updated tests
│   └── validator_test.go          # Updated tests
├── cli/
│   ├── root.go                    # Root command (updated: skip config for sync)
│   ├── validate.go                # Validate command (unchanged)
│   ├── sync.go                    # NEW: ailign sync command
│   ├── root_test.go               # Updated tests
│   ├── validate_test.go           # (unchanged)
│   └── sync_test.go               # NEW: sync command tests
├── output/
│   ├── formatter.go               # Formatter interface (add SyncResult types)
│   ├── human.go                   # Human formatter (add sync output)
│   ├── json.go                    # JSON formatter (add sync output)
│   ├── human_test.go              # Updated tests
│   └── json_test.go               # Updated tests
├── target/
│   ├── target.go                  # REFACTORED: Target interface + Registry
│   ├── claude.go                  # NEW: Claude target implementation
│   ├── cursor.go                  # NEW: Cursor target implementation
│   ├── copilot.go                 # NEW: Copilot target implementation
│   ├── windsurf.go                # NEW: Windsurf target implementation
│   ├── target_test.go             # NEW: interface + registry tests
│   └── targets_test.go            # NEW: per-target implementation tests
└── sync/
    ├── compose.go                 # NEW: overlay composition
    ├── hub.go                     # NEW: hub file writing (atomic)
    ├── symlink.go                 # NEW: symlink creation + management
    ├── sync.go                    # NEW: orchestration (compose → write → link)
    ├── compose_test.go            # NEW
    ├── hub_test.go                # NEW
    ├── symlink_test.go            # NEW
    └── sync_test.go               # NEW
```

**Structure Decision**: Extends the existing Go CLI layout from
feature 001. New `internal/sync/` package handles all sync logic.
`internal/target/` is refactored from a simple string registry to
a full interface+registry pattern with per-target implementations.

## Design Decisions

### Single Hub File Architecture

All instruction content is stored in a single central file:
`.ailign/instructions.md`. Each target's instruction file path
becomes a relative symlink to this hub file.

**Flow**:
```
local_overlays → compose → .ailign/instructions.md → symlinks → target paths
```

**Benefits**:
- One file to manage — easy to find, edit, inspect
- Symlinks auto-reflect hub changes
- Ready for future registry integration (packages write to hub)
- When per-target content is needed, hub can evolve to per-target
  files without changing the symlink approach

See [research.md](research.md) R1 for full rationale.

### Symlink Strategy

- **Relative symlinks** for portability across repo locations
- **`filepath.Rel()`** computes relative path from target dir to hub
- **Replace existing files**: If a regular file exists at the target
  path, it is replaced with a symlink. `--dry-run` shows this before
  it happens.
- **Detect existing symlinks**: If symlink already points to our hub,
  only the hub content is updated (no symlink recreation).
- **POSIX only (v1)**: macOS/Linux work natively. Windows symlink
  support (Developer Mode, `core.symlinks=true`) is out of scope
  and will be addressed as a separate feature.

See [research.md](research.md) R2 for platform details.

### Target Modularity (Provider Pattern)

Each target is a self-contained implementation of the `Target`
interface. The registry holds all available targets. The sync engine
uses targets via the interface — it never needs to know target
specifics.

```go
type Target interface {
    Name() string
    InstructionPath() string
}
```

Adding a new target requires:
1. Create `internal/target/<name>.go` implementing `Target`
2. Add it to `NewDefaultRegistry()` in `target.go`
3. Add to enum in `schema.json`
4. No changes to sync logic

See [research.md](research.md) R3 for design details.

### Atomic Writes

The hub file is written atomically using write-temp-rename:
1. Write content to `.ailign/.ailign-instructions-<random>.tmp`
2. `fsync()` the temp file
3. `os.Rename()` temp to final path (atomic on POSIX)

Symlinks pick up the change automatically. No need to recreate them.

See [research.md](research.md) R4 for implementation pattern.

### Config Schema Extension

The JSONSchema is extended from v1 to v2:
- `local_overlays` added as optional array of strings
- Path format validated (relative, non-empty)
- Existence of overlay files validated at sync time, not schema time

The `ailign validate` command validates format only. `ailign sync`
validates both format and file existence.

### Output Formatting

The existing `output.Formatter` interface is extended with sync
result formatting. Both human and JSON formats are supported.
The human format shows a clear summary of per-target actions.
The JSON format provides machine-parseable results for CI/CD.

## PR Decomposition

> **Each PR MUST be independently deployable and within CI size limits.**

| PR | Scope | Est. Lines | Independently Deployable? |
|----|-------|-----------|---------------------------|
| PR 1 | Schema extension + target refactor | ~295 | Yes — validate still works, new target interface compiles and tests pass |
| PR 2 | Sync engine + unit tests | ~475 | Yes — tested library, no CLI yet |
| PR 3 | CLI command + output formatting + BDD (US1) | ~470 | Yes — `ailign sync` works end-to-end |
| PR 4 | Dry-run (US2) | ~237 | Yes — additive `--dry-run` flag, no breaking changes |

### PR 1: Schema Extension + Target Refactor

- Add `local_overlays` to `Config` struct and `schema.json`
- Refactor `internal/target/` to interface + registry pattern
- Add per-target implementations (claude, cursor, copilot, windsurf)
- Update existing tests, add new target tests
- Validate: `go test ./...` passes, `ailign validate` works

### PR 2: Sync Engine

- Add `internal/sync/` package (compose, hub, symlink, orchestration)
- Unit tests for all sync components
- Validate: `go test ./internal/sync/...` passes

### PR 3: CLI Integration + BDD (US1)

- Add `ailign sync` command in `internal/cli/sync.go`
- Extend output formatter with sync result types
- BDD step definitions for US1 (14 scenarios including edge cases)
- Validate: `ailign sync` works end-to-end, all US1 BDD scenarios pass

### PR 4: Dry-Run (US2)

- Add `--dry-run` flag to sync command
- Dry-run output formatting (human + JSON)
- BDD step definitions for US2
- Validate: `ailign sync --dry-run` works, all US2 BDD scenarios pass

**Splitting strategy**: PR 1 is foundational. PR 2 is the sync
library. PR 3 wires it to users (completes US1). PR 4 delivers US2.
Split between PR 2 and PR 3 is for PR size management — the original
US1 scope exceeded the soft 500-line PR limit.

## Complexity Tracking

No constitution violations to justify. Design follows all
principles. Symlinks add a level of indirection but this is
explicitly requested and improves future extensibility.

**Deferred to separate feature**:
- Windows symlink support (`core.symlinks=true`, Developer Mode,
  `--copy` fallback). Git on Windows defaults to
  `core.symlinks=false`, checking out symlinks as plain text files.
  This is a significant friction point for Windows teams and
  warrants its own feature spec.

## Artifacts Generated

| Artifact | Path | Description |
|----------|------|-------------|
| Research | [research.md](research.md) | Architecture and technology decisions |
| Data Model | [data-model.md](data-model.md) | Entity definitions and relationships |
| Config Schema v2 | [contracts/config-schema-v2.json](contracts/config-schema-v2.json) | Updated JSONSchema |
| CLI Contracts | [contracts/cli-commands.md](contracts/cli-commands.md) | Command signatures, flags, output formats |
| Quickstart | [quickstart.md](quickstart.md) | Developer setup and usage guide |
