# Implementation Plan: Package Manifest & Install

**Branch**: `004-package-install` | **Date**: 2026-02-20 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `specs/004-package-install/spec.md`

## Summary

Implement `ailign install` to fetch instruction packages from a registry, compose them with local overlays, and render to configured tool formats. Includes package manifest validation (`ailign-pkg.yml`), type-scoped registry paths (`instructions/company/security@1.3.0`), lock file for reproducibility (`ailign-lock.yml`), and a consumer-driven registry contract (`contracts/registry/`) for stub-based testing.

## Technical Context

**Language/Version**: Go 1.24+ (targeting Go 1.26)
**Primary Dependencies**: Cobra (CLI), goccy/go-yaml (YAML), santhosh-tekuri/jsonschema v6 (validation), net/http (registry client), crypto/sha256 (checksums)
**Storage**: File system — reads `.ailign.yml`, fetches packages via HTTPS, writes composed instructions + lock file
**Testing**: godog (BDD) + go test + testify (TDD), httptest.Server (stub registry)
**Target Platform**: Cross-platform (Linux, macOS, Windows) — single binary
**Project Type**: Single CLI project (existing structure)
**Performance Goals**: `ailign install` with 3 packages < 30 seconds; `ailign status` < 1 second
**Constraints**: Binary < 50MB, memory < 100MB, graceful offline degradation
**Scale/Scope**: MVP — exact versions only, `instructions` type only, text-only packages

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. CLI-First | PASS | `ailign install` is a CLI command with `--format json` and `--format human` output. Exit codes: 0=success, 2=error. No interactive prompts. |
| II. Transparency | PASS | Install shows summary of what was installed. `ailign status` shows current vs desired state. `ailign explain` can trace package origin. |
| III. Fail Safe | PASS | Atomic file operations (temp+rename). Checksum verification. Lock file prevents unexpected changes. Hard error on mismatch. |
| IV. Test-First | PASS | BDD scenarios per user story. Stub registry for testing. TDD for all domain logic. |
| V. Composition | PASS | Packages composed with local overlays. Independent renderers unaffected. Modular package structure. |
| VI. Governance | PASS | Lock file with checksums. Exact versions only. Immutable packages. Traceable via `resolved` URL. |
| VII. Size-Aware | PASS | Package content is text-only. Size limits enforced by renderers (existing). |
| VIII. Cross-Tool Parity | PASS | Packages are tool-agnostic. Renderers handle tool-specific formats (existing). |
| IX. Working Software | PASS | Each PR includes implementation + tests. All PRs build and validate. |
| X. Subtraction | PASS | Config `packages` field format changes from `company/security@1.3.0` to `instructions/company/security@1.3.0`. Pre-v1.0.0, no formal deprecation needed. Evaluated whether existing sync code can be reused — install composes via the same `ComposeOverlays` path after fetching. |

**Post-Phase 1 re-check**: All gates pass. No constitution violations.

## Project Structure

### Documentation (this feature)

```text
specs/004-package-install/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
│   └── registry/
│       ├── openapi.yaml
│       └── schemas/
│           ├── manifest.json
│           └── lock-file.json
└── tasks.md             # Phase 2 output (/speckit.tasks)

features/                          # BDD feature files (Gherkin) - at project root
├── package-install.feature        # User Story 1
├── package-manifest.feature       # User Story 2
├── package-registry-paths.feature # User Story 3
├── package-version-lock.feature   # User Story 4
├── package-registry-contract.feature # User Story 5
└── steps/
    ├── suite_test.go                          # godog test runner (existing)
    ├── world_test.go                          # shared testWorld struct (extended)
    ├── package_install_steps_test.go          # US1: install from registry
    ├── package_manifest_steps_test.go         # US2: manifest validation
    ├── package_registry_paths_steps_test.go   # US3: type-scoped paths
    ├── package_version_lock_steps_test.go     # US4: lock file
    └── package_registry_contract_steps_test.go # US5: registry contract
```

### Source Code (repository root)

```text
internal/
├── config/
│   ├── config.go          # MODIFY: add Packages field to Config struct
│   ├── schema.json        # MODIFY: add packages schema with type-prefixed format
│   └── validator.go       # MODIFY: add "packages" to knownSchemaProperties
├── registry/              # NEW: registry client package
│   ├── client.go          # HTTP client for registry API
│   ├── client_test.go     # Tests with httptest.Server stub
│   ├── errors.go          # Domain errors (ErrPackageNotFound, etc.)
│   ├── manifest.go        # Manifest parsing and validation
│   ├── manifest_test.go   # Manifest validation tests
│   └── types.go           # PackageRef, Manifest, LockedPackage types
├── install/               # NEW: install orchestration
│   ├── install.go         # Install workflow (fetch, compose, render, lock)
│   ├── install_test.go    # Install logic tests
│   ├── lockfile.go        # Lock file read/write/verify
│   └── lockfile_test.go   # Lock file tests
├── sync/                  # EXISTING: local instruction sync (unchanged)
│   ├── compose.go         # Reused for content composition
│   └── hub.go             # Reused for atomic file writes
├── cli/
│   ├── root.go            # MODIFY: register install command
│   └── install.go         # NEW: install command implementation
├── output/
│   ├── formatter.go       # MODIFY: add InstallFormatter interface
│   ├── human.go           # MODIFY: add FormatInstallResult
│   └── json.go            # MODIFY: add FormatInstallResult
└── target/                # EXISTING: unchanged

contracts/                 # NEW: registry API contract (consumer-driven)
└── registry/
    ├── openapi.yaml       # OpenAPI 3.1 spec
    └── schemas/
        ├── manifest.json  # Package manifest JSON Schema
        └── lock-file.json # Lock file JSON Schema

testdata/                  # NEW: test fixtures
└── registry/
    └── packages/
        ├── instructions-company-security-1.3.0/
        │   ├── manifest.yml
        │   └── instructions.md
        └── instructions-company-typescript-2.1.0/
            ├── manifest.yml
            └── instructions.md
```

**Structure Decision**: Extends the existing single-project Go structure. New packages `internal/registry` (API client) and `internal/install` (orchestration) follow the established pattern of `internal/sync` and `internal/config`. Contract artifacts live at project root in `contracts/registry/` per the consumer-driven contract decision (Option B from spec).

## PR Decomposition

> **Each PR MUST be independently deployable and within CI size limits.**

- **Hard limit**: 750 lines (additions + deletions). CI fails above this.
- **Soft limit**: 500 lines. CI warns above this.
- **Target**: <333 lines per PR (size:m or smaller).

| PR | Scope | Est. Lines | Independently Deployable? |
|----|-------|-----------|---------------------------|
| PR 1 | Setup + foundational types + config schema extension + tests | ~250 | Yes — validates type-prefixed package refs, existing commands unaffected |
| PR 2 | Package manifest validation + BDD (ailign-pkg.yml) | ~250 | Yes — manifest parsing works standalone |
| PR 3 | Registry paths + client + stub + contract + BDD | ~300 | Yes — client testable against stub, no CLI integration yet |
| PR 4 | Lock file read/write/verify + BDD | ~200 | Yes — lock file operations work standalone |
| PR 5 | Install orchestration + CLI command + output formatters + BDD | ~300 | Yes — full `ailign install` working end-to-end |
| PR 6 | Polish + validation + cross-cutting | ~150 | Yes — lint, test suite, quickstart validation |

**Splitting strategies**:
1. By layer: types/validation (PR 1-2), infrastructure (PR 3-4), integration (PR 5-6)
2. BDD step definitions are included in each PR alongside their story implementation
3. Each PR is independently deployable — no broken intermediate states

## Complexity Tracking

No constitution violations requiring justification.
