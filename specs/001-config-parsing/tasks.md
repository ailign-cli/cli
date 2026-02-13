# Tasks: Configuration File Parsing

**Input**: Design documents from `/specs/001-config-parsing/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, contracts/

**Tests**: TDD is mandatory per constitution (Principle IV). Test tasks are written FIRST, must FAIL, then implementation follows.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2)
- Include exact file paths in descriptions

## Path Conventions

- **Go CLI project**: `cmd/ailign/` for entry point, `internal/` for all packages
- Tests live alongside source files per Go conventions (`*_test.go`)

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and Go module setup

- [x] T001 Create directory structure: `cmd/ailign/`, `internal/config/`, `internal/cli/`, `internal/output/`, `internal/target/`
- [x] T002 Initialize Go module with `go mod init github.com/ailign/cli` and add dependencies: `github.com/spf13/cobra`, `github.com/goccy/go-yaml`, `github.com/santhosh-tekuri/jsonschema/v6`, `github.com/stretchr/testify`
- [x] T003 [P] Create .gitignore for Go project (binaries, coverage files, IDE files)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core shared infrastructure that MUST be complete before ANY user story can be implemented

**CRITICAL**: No user story work can begin until this phase is complete

- [x] T004 [P] Write tests for target registry (IsValid, KnownTargets) in internal/target/registry_test.go
- [x] T005 [P] Implement target registry with known targets (claude, cursor, copilot, windsurf) and Target interface stub in internal/target/registry.go
- [x] T006 Define Formatter interface (FormatSuccess, FormatErrors, FormatWarnings methods) in internal/output/formatter.go
- [x] T007 [P] Write tests and implement HumanFormatter (indented error output with field path, expected, found, fix) in internal/output/human.go and internal/output/human_test.go
- [x] T008 [P] Write tests and implement JSONFormatter (structured JSON output per contracts/cli-commands.md) in internal/output/json.go and internal/output/json_test.go

**Checkpoint**: Foundation ready - target registry, output formatters, and interfaces available for user stories

---

## Phase 3: User Story 1 - Parse Configuration File (Priority: P1) MVP

**Goal**: Load and parse `.ailign.yml` from the working directory. Return a structured Config object. Handle missing files and YAML parse errors.

**Independent Test**: Place a valid `.ailign.yml` with targets in a temp directory, run the CLI, verify targets are loaded. Place no file, verify exit code 2 with error on stderr.

### Tests for User Story 1

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T009 [P] [US1] Write tests for Config struct (fields, defaults) in internal/config/config_test.go
- [ ] T010 [P] [US1] Write tests for YAML loader (valid file, missing file, empty file, YAML parse error, tabs, BOM, symlinks, permission errors) in internal/config/loader_test.go

### Implementation for User Story 1

- [ ] T011 [US1] Define Config struct with Targets field and YAML tags in internal/config/config.go
- [ ] T012 [US1] Implement LoadFromFile function (read `.ailign.yml` from path, parse YAML into Config struct, handle file-not-found and parse errors) in internal/config/loader.go
- [ ] T013 [US1] Copy JSONSchema definition from contracts/config-schema.json to internal/config/schema.json and embed via go:embed directive in internal/config/config.go
- [ ] T014 [US1] Create root Cobra command with global `--format` flag and PersistentPreRunE that loads config from CWD in internal/cli/root.go
- [ ] T015 [US1] Create CLI entry point that wires root command and executes in cmd/ailign/main.go

**Checkpoint**: At this point, the CLI can read a valid `.ailign.yml` and report errors for missing/malformed files. User Story 1 is fully functional and testable independently.

---

## Phase 4: User Story 2 - Schema Validation with Actionable Errors (Priority: P2)

**Goal**: Validate parsed config against embedded JSONSchema. Report all errors at once with field paths and remediation. Add `ailign validate` command. Detect unknown fields as warnings.

**Independent Test**: Provide `.ailign.yml` files with invalid targets, missing targets field, unknown fields, and duplicates. Verify each produces the correct error with field path and remediation. Run `ailign validate` on valid and invalid files, verify exit codes.

### Tests for User Story 2

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T016 [P] [US2] Write tests for ValidationError and ValidationResult types in internal/config/errors_test.go
- [ ] T017 [P] [US2] Write tests for JSONSchema validator (valid config, missing targets, invalid target name, empty targets array, duplicate targets, unknown fields as warnings, multiple errors at once) in internal/config/validator_test.go
- [ ] T018 [P] [US2] Write tests for `ailign validate` command (valid file → exit 0 + stdout, invalid file → exit 2 + stderr, missing file → exit 2, --format json output, --format human output) in internal/cli/validate_test.go

### Implementation for User Story 2

- [ ] T019 [US2] Define ValidationError struct (field_path, expected, actual, message, remediation, severity) and ValidationResult struct (valid, errors, warnings, config) in internal/config/errors.go
- [ ] T020 [US2] Implement Validate function: marshal Config to JSON, validate against embedded JSONSchema, collect all errors in internal/config/validator.go
- [ ] T021 [US2] Implement unknown field detection: compare parsed YAML keys against schema properties, emit warnings for unrecognized fields in internal/config/validator.go
- [ ] T022 [US2] Implement error transformation: convert raw JSONSchema validation errors into user-friendly ValidationError structs with remediation guidance in internal/config/errors.go
- [ ] T023 [US2] Integrate validation into root command PersistentPreRunE (load → validate → report errors/warnings → exit) in internal/cli/root.go
- [ ] T024 [US2] Implement `ailign validate` command (load config, validate, format output via --format flag, exit 0 on success or 2 on failure) in internal/cli/validate.go
- [ ] T025 [US2] Write integration tests for root command: valid config proceeds, invalid config exits 2 with all errors, warnings are emitted but don't block in internal/cli/root_test.go

**Checkpoint**: At this point, the CLI validates all configs against JSONSchema, reports actionable errors, and `ailign validate` works standalone. User Stories 1 AND 2 are both independently functional.

---

## Phase 5: Polish & Cross-Cutting Concerns

**Purpose**: Final validation, coverage, and edge case verification

- [ ] T026 Run full test suite and verify >90% coverage for internal/config/ package
- [ ] T027 Validate quickstart.md steps work end-to-end (build binary, create config, run validate)
- [ ] T028 Verify all edge cases from spec are covered by tests (tabs in YAML, Unicode BOM, file permission errors, symlinks, duplicate targets)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup (T001, T002) completion - BLOCKS all user stories
- **User Story 1 (Phase 3)**: Depends on Foundational (Phase 2) completion
- **User Story 2 (Phase 4)**: Depends on User Story 1 (Phase 3) completion (needs Config struct and loader)
- **Polish (Phase 5)**: Depends on all user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational. No dependencies on other stories. Delivers core config loading.
- **User Story 2 (P2)**: Depends on US1 (needs Config struct and loader to validate against). Adds validation layer and `ailign validate` command.

### Within Each User Story

- Tests MUST be written and FAIL before implementation
- Types/structs before logic
- Core logic before CLI wiring
- CLI wiring before integration tests

### Parallel Opportunities

- T003 can run in parallel with T001/T002
- T004/T005 can run in parallel with T007/T008 (different packages)
- T009/T010 can run in parallel (different test files)
- T016/T017/T018 can run in parallel (different test files)

---

## Parallel Examples

### Phase 2: Foundational

```bash
# These can run in parallel (different packages):
Task: "Write tests for target registry in internal/target/registry_test.go"
Task: "Write tests and implement HumanFormatter in internal/output/human.go"
Task: "Write tests and implement JSONFormatter in internal/output/json.go"
```

### Phase 3: User Story 1

```bash
# Write all tests in parallel:
Task: "Write tests for Config struct in internal/config/config_test.go"
Task: "Write tests for YAML loader in internal/config/loader_test.go"
```

### Phase 4: User Story 2

```bash
# Write all tests in parallel:
Task: "Write tests for ValidationError types in internal/config/errors_test.go"
Task: "Write tests for JSONSchema validator in internal/config/validator_test.go"
Task: "Write tests for ailign validate command in internal/cli/validate_test.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL - blocks all stories)
3. Complete Phase 3: User Story 1
4. **STOP and VALIDATE**: Build binary, create `.ailign.yml`, verify config loads
5. Binary can load valid configs and reject missing/malformed files

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 → Test independently → MVP (config loading works)
3. Add User Story 2 → Test independently → Full feature (validation + `ailign validate`)
4. Polish → Coverage, quickstart validation, edge case review

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Verify tests fail before implementing
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- US2 depends on US1 (needs Config struct and loader)
