# Tasks: Local Instruction Sync

**Input**: Design documents from `/specs/002-local-instruction-sync/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: TDD+BDD mandatory per constitution (Principle IV). BDD outer loop drives acceptance criteria, TDD inner loop drives component design.

**Organization**: Tasks grouped by PR decomposition. Each phase is independently deployable and within the soft 500-line PR limit.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story (US1, US2)
- Exact file paths included

## PR Mapping

| Phase | PR | Scope | Est. Lines |
|-------|----|-------|-----------|
| Phase 2 | PR 1 | Schema extension + target refactor | ~295 |
| Phase 3 | PR 2 | Sync engine + unit tests | ~475 |
| Phase 4 | PR 3 | CLI command + output formatting + BDD (US1) | ~470 |
| Phase 5 | PR 4 | Dry-run (US2) | ~237 |

---

## Phase 1: Setup

**Purpose**: Verify baseline before starting feature work

- [X] T001 Verify all existing tests pass by running `go test ./...` from repository root

---

## Phase 2: Foundational — Schema Extension + Target Refactor

**Purpose**: Extend config schema with `local_overlays` and refactor target package to interface+registry pattern. All subsequent phases depend on this.

**Maps to**: PR 1 (~295 lines)

**CRITICAL**: No user story work can begin until this phase is complete

### Schema Extension

- [X] T002 [P] Update embedded schema from v1 to v2: add optional `local_overlays` array field in internal/config/schema.json (match contracts/config-schema-v2.json)
- [X] T003 [P] Add `LocalOverlays []string` field with yaml/json tags to Config struct in internal/config/config.go
- [X] T004 Add `"local_overlays"` to `knownSchemaProperties` map in internal/config/validator.go
- [X] T005 [P] Write unit tests for Config parsing with local_overlays (single, multiple, absent) in internal/config/config_test.go
- [X] T006 [P] Write unit tests for schema validation of local_overlays (valid paths, empty strings, absolute paths) in internal/config/validator_test.go
- [X] T007 [P] Update loader integration tests for configs including local_overlays in internal/config/loader_test.go

### Target Refactor

- [X] T008 Refactor internal/target/registry.go to internal/target/target.go: add `InstructionPath()` to Target interface, create Registry struct with `Register`/`Get`/`IsValid`/`KnownTargets` methods, add `NewDefaultRegistry()` constructor
- [X] T009 [P] Create Claude target (Name=`claude`, InstructionPath=`.claude/instructions.md`) in internal/target/claude.go
- [X] T010 [P] Create Cursor target (Name=`cursor`, InstructionPath=`.cursorrules`) in internal/target/cursor.go
- [X] T011 [P] Create Copilot target (Name=`copilot`, InstructionPath=`.github/copilot-instructions.md`) in internal/target/copilot.go
- [X] T012 [P] Create Windsurf target (Name=`windsurf`, InstructionPath=`.windsurfrules`) in internal/target/windsurf.go
- [X] T013 Write registry tests (Register, Get, IsValid, KnownTargets, duplicate registration, unknown target) and schema-registry invariant test in internal/target/target_test.go
- [X] T014 [P] Write per-target tests (verify Name and InstructionPath for each of the 4 targets) in internal/target/targets_test.go

**Checkpoint**: `go test ./...` passes. `ailign validate` works with configs containing `local_overlays`. Target registry returns correct instruction paths.

---

## Phase 3: Sync Engine — Core Library (US1 internals)

**Purpose**: Implement the `internal/sync/` package with overlay composition, atomic hub writing, symlink management, and orchestration. Proven working via unit tests, no CLI integration yet.

**Maps to**: PR 2 (~475 lines)

### TDD Unit Tests

> **NOTE: Write unit tests, ensure they FAIL before implementation**

- [X] T015 [P] [US1] Write unit tests for overlay composition (read files, compose in order, prepend managed header, path traversal rejection, non-UTF-8 rejection, empty file warning) in internal/sync/compose_test.go
- [X] T016 [P] [US1] Write unit tests for atomic hub file writing (write-temp-rename, fsync, directory creation) in internal/sync/hub_test.go
- [X] T017 [P] [US1] Write unit tests for symlink management (create new, replace file, existing correct symlink, existing wrong symlink, directory creation, permission error) in internal/sync/symlink_test.go
- [X] T018 [P] [US1] Write unit tests for sync orchestration (full flow, missing overlay error, no overlays error, partial failure) in internal/sync/sync_test.go

### Implementation

- [X] T019 [P] [US1] Implement overlay composition (read files, validate UTF-8, validate paths, compose in order, prepend managed-content header) in internal/sync/compose.go
- [X] T020 [P] [US1] Implement atomic hub file writing (write to temp file, fsync, rename; create .ailign/ dir if needed) in internal/sync/hub.go
- [X] T021 [P] [US1] Implement symlink creation and management (create relative symlink, replace existing file/symlink, create target directories, detect existing correct symlink) in internal/sync/symlink.go
- [X] T022 [US1] Implement sync orchestration: compose overlays → write hub → create symlinks per target, returning SyncResult in internal/sync/sync.go

**Checkpoint**: `go test ./internal/sync/...` passes. All sync logic works via unit tests. No CLI command yet.

---

## Phase 4: CLI Integration + Output + BDD (US1 complete)

**Purpose**: Wire sync engine to `ailign sync` CLI command, extend output formatters, write BDD step definitions. After this phase, US1 is fully complete and user-facing.

**Maps to**: PR 3 (~470 lines)

### BDD Scenarios for User Story 1

> **NOTE: Feature file exists at features/sync-local-instructions.feature (14 scenarios including 4 edge cases)**

- [ ] T023 [US1] Review features/sync-local-instructions.feature and verify scenarios are concrete and aligned with data-model.md
- [ ] T024 [US1] Extend testWorld in features/steps/world_test.go with sync-related fields and helper methods (writeOverlayFile, writeConfigWithOverlays, runSync, etc.)
- [ ] T025 [US1] Write step definitions for US1 scenarios in features/steps/sync_steps_test.go and register via `registerSyncSteps(ctx, w)` in features/steps/suite_test.go (expect RED)

### Output Formatting

- [ ] T026 [US1] Add SyncResult and LinkResult types to output package and extend Formatter interface with `FormatSyncResult(result SyncResult) string` in internal/output/formatter.go
- [ ] T027 [P] [US1] Implement sync result formatting in HumanFormatter (target list, status per target, summary line) in internal/output/human.go
- [ ] T028 [P] [US1] Implement sync result formatting in JSONFormatter (hub status, links array, summary counts) in internal/output/json.go
- [ ] T029 [P] [US1] Write unit tests for sync output formatting (human and JSON, success, errors, multiple targets) in internal/output/human_test.go and internal/output/json_test.go

### CLI Command

- [ ] T030 [US1] Implement `ailign sync` command: load config, validate local_overlays present, call sync engine, format and print result in internal/cli/sync.go
- [ ] T031 [US1] Register sync command in root and update PersistentPreRunE if needed in internal/cli/root.go
- [ ] T032 [P] [US1] Write CLI integration tests for sync command (valid sync, missing overlays, no overlays, format flag, exit codes) in internal/cli/sync_test.go
- [ ] T033 [US1] Verify US1 BDD step definitions pass (GREEN) by running `go test ./features/steps/ -v -run TestFeatures`

**Checkpoint**: `ailign sync` works end-to-end. All 14 US1 BDD scenarios pass (including edge cases: empty overlay, path traversal, non-UTF-8, read-only directory). `go test ./...` passes.

---

## Phase 5: User Story 2 — Preview Sync Changes (Priority: P2)

**Goal**: `ailign sync --dry-run` shows what files would be created/updated without modifying any files.

**Independent Test**: Run `ailign sync --dry-run`, verify output lists expected changes, verify no files on disk were modified.

**Maps to**: PR 4 (~237 lines)

### BDD Scenarios for User Story 2

> **NOTE: Feature file exists at features/preview-sync-changes.feature (4 scenarios)**

- [ ] T034 [US2] Review features/preview-sync-changes.feature and verify scenarios are concrete and aligned with CLI contracts
- [ ] T035 [US2] Write step definitions for US2 scenarios in features/steps/preview_steps_test.go and register via `registerPreviewSteps(ctx, w)` in features/steps/suite_test.go (expect RED)

### TDD Unit Tests for User Story 2

> **NOTE: Write unit tests, ensure they FAIL before implementation**

- [ ] T036 [P] [US2] Write unit tests for dry-run mode (no files written, correct result status, "would" prefix in human output) in internal/sync/sync_test.go
- [ ] T037 [P] [US2] Write unit tests for dry-run output formatting (human: "would" language, JSON: dry_run=true) in internal/output/human_test.go and internal/output/json_test.go

### Implementation for User Story 2

- [ ] T038 [US2] Add `--dry-run`/`-n` flag to sync command and pass DryRun option to sync engine in internal/cli/sync.go
- [ ] T039 [US2] Implement dry-run mode in sync orchestration: compose and compute symlink status without writing files in internal/sync/sync.go
- [ ] T040 [P] [US2] Add dry-run output formatting to HumanFormatter ("would be written", "would create symlink") in internal/output/human.go
- [ ] T041 [P] [US2] Add dry-run output formatting to JSONFormatter (dry_run field set to true) in internal/output/json.go
- [ ] T042 [US2] Verify US2 BDD step definitions pass (GREEN) by running `go test ./features/steps/ -v -run TestFeatures`

**Checkpoint**: `ailign sync --dry-run` works. All 4 US2 BDD scenarios pass. No files are modified during dry-run. `go test ./...` passes.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Final validation across all stories

- [ ] T043 Run full test suite `go test ./...` and verify all tests pass (unit + BDD)
- [ ] T044 [P] Validate quickstart.md workflow: execute documented commands and verify expected file structure and output
- [ ] T045 Mark all tasks complete in specs/002-local-instruction-sync/tasks.md

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — verify baseline
- **Foundational (Phase 2)**: Depends on Phase 1 — BLOCKS all user stories
- **Sync Engine (Phase 3)**: Depends on Phase 2 — core sync library
- **CLI + BDD (Phase 4)**: Depends on Phase 3 — wires engine to CLI, completes US1
- **Dry-Run (Phase 5)**: Depends on Phase 4 — extends sync command with --dry-run
- **Polish (Phase 6)**: Depends on Phase 5

### User Story Dependencies

- **US1 (P1)**: Spans Phase 3 + Phase 4 — sync engine then CLI integration
- **US2 (P2)**: Phase 5 — extends the sync command with --dry-run flag

### Within Each Phase

1. TDD unit tests written → RED (for each component)
2. Implementation → unit tests GREEN
3. BDD step definitions written → RED (Phase 4 and 5 only)
4. CLI integration → step definitions GREEN
5. Phase checkpoint

### Parallel Opportunities

**Phase 2** (within phase):
- T002, T003 can run in parallel (schema.json and config.go are different files)
- T005, T006, T007 can run in parallel (different test files)
- T009, T010, T011, T012 can run in parallel (one file per target)
- T013, T014 can run in parallel (different test files)

**Phase 3** (within phase):
- T015, T016, T017, T018 can run in parallel (different test files)
- T019, T020, T021 can run in parallel (different source files)

**Phase 4** (within phase):
- T027, T028 can run in parallel (human vs JSON formatter)
- T029, T032 can run in parallel (output tests vs CLI tests)

**Phase 5** (within phase):
- T036, T037 can run in parallel (different test files)
- T040, T041 can run in parallel (human vs JSON formatter)

---

## Parallel Example: Phase 3 Sync Engine

```
# Launch all unit tests in parallel (different files):
T015: Compose tests in internal/sync/compose_test.go
T016: Hub tests in internal/sync/hub_test.go
T017: Symlink tests in internal/sync/symlink_test.go
T018: Orchestration tests in internal/sync/sync_test.go

# Launch independent implementations in parallel:
T019: Compose in internal/sync/compose.go
T020: Hub in internal/sync/hub.go
T021: Symlink in internal/sync/symlink.go

# Then orchestration (depends on T019-T021):
T022: Sync orchestration in internal/sync/sync.go
```

## Parallel Example: Phase 2 Target Refactor

```
# After T008 (target.go refactor), launch all target implementations:
T009: Create Claude target in internal/target/claude.go
T010: Create Cursor target in internal/target/cursor.go
T011: Create Copilot target in internal/target/copilot.go
T012: Create Windsurf target in internal/target/windsurf.go

# Then launch tests in parallel:
T013: Registry tests in internal/target/target_test.go
T014: Per-target tests in internal/target/targets_test.go
```

---

## Implementation Strategy

### MVP First (Phase 2 + Phase 3 + Phase 4)

1. Complete Phase 1: Verify baseline
2. Complete Phase 2: Schema + target refactor (PR 1)
3. Complete Phase 3: Sync engine + tests (PR 2)
4. Complete Phase 4: CLI + output + BDD (PR 3)
5. **STOP and VALIDATE**: `ailign sync` works end-to-end with all 14 BDD scenarios passing
6. Ship PR 1 → PR 2 → PR 3

### Incremental Delivery

1. PR 1 (Phase 2): Schema + targets — `ailign validate` works with new config format
2. PR 2 (Phase 3): Sync engine — tested library, no CLI yet
3. PR 3 (Phase 4): CLI + BDD — `ailign sync` works end-to-end (MVP!)
4. PR 4 (Phase 5): Dry-run — `ailign sync --dry-run` adds transparency

Each PR is independently deployable and within the soft 500-line limit.

---

## Notes

- [P] tasks = different files, no dependencies on incomplete tasks
- [Story] label maps task to specific user story for traceability
- US1 spans Phase 3 (engine) + Phase 4 (CLI) — split for PR size management
- US2 is entirely in Phase 5 — extends sync command
- Edge cases (empty overlay, path traversal, non-UTF-8, permissions) are part of US1
- All symlinks point to single hub file `.ailign/instructions.md` (not per-target files)
- Windows symlink support is out of scope (separate feature)
- Target implementations are flat files in internal/target/ (not subdirectories)
