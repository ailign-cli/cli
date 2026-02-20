# Tasks: Package Manifest & Install

**Input**: Design documents from `specs/004-package-install/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, contracts/

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization — new packages, test fixtures, contract artifacts

- [ ] T001 Create `internal/registry/` package directory and `internal/install/` package directory
- [ ] T002 [P] Create test fixture directory `testdata/registry/packages/` with sample package content for `instructions-company-security-1.3.0/` (ailign-pkg.yml + instructions.md) and `instructions-company-typescript-2.1.0/` (ailign-pkg.yml + instructions.md)
- [ ] T003 [P] Copy contract artifacts from `specs/004-package-install/contracts/registry/` to project root `contracts/registry/` (openapi.yaml, schemas/manifest.json, schemas/lock-file.json)
- [ ] T004 [P] Add `contracts/` and `testdata/` to `.gitignore` exclusion list if needed (ensure they are tracked, not ignored)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core types and config schema extension that ALL user stories depend on

**CRITICAL**: No user story work can begin until this phase is complete

- [ ] T005 Define `PackageRef` type with `Type`, `Scope`, `Name`, `Version` fields and `ParsePackageRef(raw string) (PackageRef, error)` function in `internal/registry/types.go` — include type prefix validation (only `instructions` supported; reject `mcp`, `commands`, `agents`, `packages` with supported-types list; reject missing prefix with format explanation)
- [ ] T006 [P] Write unit tests for `ParsePackageRef` covering valid refs, missing type prefix, each unsupported type (`mcp`, `commands`, `agents`, `packages`), invalid version, missing version, malformed references in `internal/registry/types_test.go`
- [ ] T007 [P] Define `Manifest` type with `Name`, `Type`, `Version`, `Description`, `Content.Main` fields in `internal/registry/types.go`
- [ ] T008 [P] Define `LockedPackage` type with `Reference`, `Version`, `Resolved`, `Integrity` fields and `LockFile` type with `LockfileVersion`, `Packages` fields in `internal/registry/types.go`
- [ ] T009 Extend `Config` struct in `internal/config/config.go` to add `Packages []string` field with `yaml:"packages"` tag
- [ ] T010 [P] Update JSON Schema in `internal/config/schema.json` to add `packages` property — array of strings matching pattern `^[a-z]+/[a-z][a-z0-9-]*/[a-z][a-z0-9-]*@\d+\.\d+\.\d+$`
- [ ] T011 [P] Add `"packages"` to `knownSchemaProperties` map in `internal/config/validator.go`
- [ ] T012 Write unit tests for config parsing with packages field — valid packages, empty packages, duplicate detection in `internal/config/loader_test.go`

**Checkpoint**: Core types defined, config schema extended — user story implementation can begin

---

## Phase 3: User Story 2 — Package Manifest Defines Package Identity (Priority: P1) 🎯 MVP

**Goal**: Validate `ailign-pkg.yml` manifest files with clear error messages for missing/invalid fields

**Independent Test**: Create valid and invalid manifests, verify acceptance/rejection with specific error messages

> US2 is implemented before US1 because US1 (install) depends on manifest validation.

### BDD Scenarios

- [ ] T013 [US2] Remove `@wip` tag from `features/package-manifest.feature` and write step definitions in `features/steps/package_manifest_steps_test.go` — register via `registerPackageManifestSteps(ctx, w)` in `features/steps/suite_test.go` (expect RED)

### Implementation

- [ ] T014 [US2] Implement manifest validation using JSON Schema (`contracts/registry/schemas/manifest.json`) in `internal/registry/manifest.go` — `ValidateManifest(data []byte) (*Manifest, error)` and `ValidateManifestAgainstRef(manifest *Manifest, ref PackageRef) error` for type consistency check
- [ ] T015 [P] [US2] Write unit tests for manifest validation — valid manifest, missing each required field, invalid version, unsupported type, name-path matching in `internal/registry/manifest_test.go`
- [ ] T016 [US2] Define domain errors `ErrInvalidManifest`, `ErrTypeMismatch`, `ErrUnsupportedType` with remediation messages in `internal/registry/errors.go`
- [ ] T017 [US2] Verify BDD step definitions pass (GREEN) — run `go test ./features/steps/... -v -run TestFeatures`

**Checkpoint**: Manifest validation works standalone with clear error messages

---

## Phase 4: User Story 3 — Type-Scoped Registry Paths (Priority: P2)

**Goal**: Validate type-prefixed package references in `.ailign.yml` with clear errors for unsupported/missing types

**Independent Test**: Declare packages with `instructions/` prefix (resolves), `mcp/` prefix (rejected), no prefix (rejected)

### BDD Scenarios

- [ ] T018 [US3] Remove `@wip` tag from `features/package-registry-paths.feature` and write step definitions in `features/steps/package_registry_paths_steps_test.go` — register via `registerPackageRegistryPathsSteps(ctx, w)` in `features/steps/suite_test.go` (expect RED)

### Implementation

> Note: Core `ParsePackageRef` validation (type prefix, unsupported types, missing prefix) was implemented in T005/T006. This phase wires it into the BDD scenarios.

- [ ] T019 [US3] Verify BDD step definitions pass (GREEN) — run `go test ./features/steps/... -v -run TestFeatures`

**Checkpoint**: Type-scoped references validated with clear error messages. Type mismatch (config vs manifest) tested in US2.

---

## Phase 5: User Story 5 — Registry Contract for Testing (Priority: P3)

**Goal**: Stub registry serves fixture packages via `httptest.Server`, validating the CLI-registry contract

**Independent Test**: CLI tests pass against the stub registry. Contract artifacts are self-contained for independent registry implementation.

> US5 is implemented before US1 because the stub registry is needed for US1's install integration tests.

### BDD Scenarios

- [ ] T022 [US5] Remove `@wip` tag from `features/package-registry-contract.feature` and write step definitions in `features/steps/package_registry_contract_steps_test.go` — register via `registerPackageRegistryContractSteps(ctx, w)` in `features/steps/suite_test.go` (expect RED)

### Implementation

- [ ] T023 [US5] Implement `RegistryClient` interface in `internal/registry/client.go` — `GetPackageVersion(ctx, ref PackageRef) (*PackageVersionResponse, error)` with `PackageVersionResponse` containing manifest, content URL, and integrity checksum
- [ ] T024 [P] [US5] Implement `HTTPRegistryClient` (concrete) in `internal/registry/client.go` — uses `net/http` with configurable base URL, timeout (30s), user-agent header. Maps HTTP status codes to domain errors (`ErrPackageNotFound` for 404, `ErrRegistryUnreachable` for network errors, `ErrServerError` for 5xx)
- [ ] T025 [P] [US5] Define domain errors `ErrPackageNotFound`, `ErrRegistryUnreachable`, `ErrServerError`, `ErrRateLimited` in `internal/registry/errors.go`
- [ ] T026 [US5] Implement stub registry using `httptest.NewServer` in `internal/registry/client_test.go` — serves fixture packages from `testdata/registry/packages/`, returns 404 for unknown packages, returns invalid manifest for broken package
- [ ] T027 [P] [US5] Write unit tests for `HTTPRegistryClient` against stub registry — fetch by type/name/version, package not found, version not found, unreachable host, invalid manifest response in `internal/registry/client_test.go`
- [ ] T028 [US5] Verify BDD step definitions pass (GREEN) — run `go test ./features/steps/... -v -run TestFeatures`

**Checkpoint**: Registry client works against stub. Contract artifacts sufficient for independent registry implementation.

---

## Phase 6: User Story 4 — Version Resolution and Lock File (Priority: P2)

**Goal**: Create, read, verify, and update `ailign-lock.yml` with deterministic ordering and checksum verification

**Independent Test**: Run install → lock file created. Run again → unchanged. Change version → updated. Tamper checksum → error.

### BDD Scenarios

- [ ] T029 [US4] Remove `@wip` tag from `features/package-version-lock.feature` and write step definitions in `features/steps/package_version_lock_steps_test.go` — register via `registerPackageVersionLockSteps(ctx, w)` in `features/steps/suite_test.go` (expect RED)

### Implementation

- [ ] T030 [US4] Implement lock file read/write in `internal/install/lockfile.go` — `ReadLockFile(path string) (*LockFile, error)`, `WriteLockFile(path string, lf *LockFile) error` with deterministic ordering (sort packages by Reference before writing), managed-content header comment
- [ ] T031 [P] [US4] Implement lock file verification in `internal/install/lockfile.go` — `VerifyIntegrity(locked LockedPackage, content []byte) error` comparing `sha256-{base64}` checksum, `ComputeIntegrity(content []byte) string` producing SRI format
- [ ] T032 [P] [US4] Write unit tests for lock file operations — create, read back, verify deterministic order, checksum match, checksum mismatch, corrupt YAML in `internal/install/lockfile_test.go`
- [ ] T033 [US4] Verify BDD step definitions pass (GREEN) — run `go test ./features/steps/... -v -run TestFeatures`

**Checkpoint**: Lock file operations work standalone. Checksums verified. Deterministic output confirmed.

---

## Phase 7: User Story 1 — Install Packages from Registry (Priority: P1) 🎯 MVP

**Goal**: `ailign install` fetches packages, composes with overlays, renders to targets, creates lock file — end-to-end working command

**Independent Test**: Run `ailign install` with config referencing a package, verify content fetched, composed with overlays, rendered to target files, lock file created

> US1 is implemented last among P1 stories because it integrates all preceding components (manifest validation, registry client, lock file).

### BDD Scenarios

- [ ] T034 [US1] Remove `@wip` tag from `features/package-install.feature` and write step definitions in `features/steps/package_install_steps_test.go` — register via `registerPackageInstallSteps(ctx, w)` in `features/steps/suite_test.go` (expect RED)

### Implementation

- [ ] T035 [US1] Implement install orchestration in `internal/install/install.go` — `Install(cfg *config.Config, client RegistryClient, baseDir string, opts InstallOptions) (*InstallResult, error)` workflow: parse refs → check lock → fetch manifests → validate type match → fetch content → verify checksums → compose with overlays → write hub → create symlinks → write lock file
- [ ] T036 [P] [US1] Write unit tests for install orchestration — single package, multiple packages, with overlays, idempotent re-install, checksum mismatch abort in `internal/install/install_test.go`
- [ ] T037 [US1] Add `InstallFormatter` interface to `internal/output/formatter.go` — `FormatInstallResult(result InstallResult) string`
- [ ] T038 [P] [US1] Implement `FormatInstallResult` for `HumanFormatter` in `internal/output/human.go` — package list with status, hub status, target links, lock file status, summary line
- [ ] T039 [P] [US1] Implement `FormatInstallResult` for `JSONFormatter` in `internal/output/json.go` — packages array, hub object, links array, lock object, summary object
- [ ] T040 [US1] Implement `install` CLI command in `internal/cli/install.go` — register on root command in `internal/cli/root.go`, use `GetConfig()`, create `HTTPRegistryClient`, call `Install()`, format and print result, return `ErrAlreadyReported` on error
- [ ] T041 [P] [US1] Write unit tests for install output formatting — human and JSON, single/multiple packages, dry-run variants in `internal/output/formatter_test.go`
- [ ] T042 [US1] Verify BDD step definitions pass (GREEN) — run `go test ./features/steps/... -v -run TestFeatures`

**Checkpoint**: `ailign install` works end-to-end. All 7 acceptance scenarios from package-install.feature pass.

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Validation, cleanup, and cross-cutting improvements

- [ ] T043 Run full test suite `go test ./... -count=1` and verify all tests pass (unit + BDD)
- [ ] T044 [P] Run `golangci-lint run` and fix any lint issues
- [ ] T045 [P] Validate quickstart.md scenarios against actual CLI output
- [ ] T046 [P] Verify contract artifacts (`contracts/registry/`) are self-contained — an independent team could implement a registry from these files alone
- [ ] T047 Verify existing `ailign sync` command still works unchanged — run sync BDD scenarios
- [ ] T048 Verify `ailign validate` correctly validates the new `packages` field in `.ailign.yml`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion — BLOCKS all user stories
- **US2 Manifest (Phase 3)**: Depends on Phase 2 — provides manifest validation for all downstream stories
- **US3 Registry Paths (Phase 4)**: Depends on Phase 2 — can run in parallel with US2
- **US5 Registry Contract (Phase 5)**: Depends on Phase 2 + US2 (manifest validation) — provides stub registry for US1
- **US4 Lock File (Phase 6)**: Depends on Phase 2 — can run in parallel with US5
- **US1 Install (Phase 7)**: Depends on US2, US3, US5, US4 — integrates all components
- **Polish (Phase 8)**: Depends on all user stories complete

### User Story Dependencies

```text
Phase 1 (Setup)
     │
Phase 2 (Foundational)
     │
     ├──────────────┬──────────────┐
     │              │              │
Phase 3 (US2)  Phase 4 (US3)  Phase 6 (US4)
  Manifest     Registry Paths   Lock File
     │              │              │
     └──────┬───────┘              │
            │                      │
      Phase 5 (US5)                │
      Registry Contract            │
            │                      │
            └──────────┬───────────┘
                       │
                 Phase 7 (US1)
                 Install E2E
                       │
                 Phase 8 (Polish)
```

### PR Mapping

| PR | Phase(s) | Scope | Est. Lines |
|----|----------|-------|-----------|
| PR 1 | Phase 1 + Phase 2 | Config schema + package reference types + tests | ~250 |
| PR 2 | Phase 3 | Manifest types + validation + BDD | ~250 |
| PR 3 | Phase 4 + Phase 5 | Registry paths + client + stub + contract + BDD | ~300 |
| PR 4 | Phase 6 | Lock file operations + BDD | ~200 |
| PR 5 | Phase 7 | Install orchestration + CLI command + output formatters + BDD | ~300 |
| PR 6 | Phase 8 | Polish + validation + cross-cutting | ~150 |

### Within Each User Story

1. Remove `@wip` tag from `.feature` file
2. Write step definitions → RED (expect failures)
3. Write unit tests → RED
4. Implement → unit tests GREEN
5. Step definitions → GREEN (run godog)
6. Story checkpoint

### Parallel Opportunities

- **Phase 2**: T006, T007, T008, T010, T011 can run in parallel (different files)
- **Phase 3**: T015 can run in parallel with T014 (test file vs implementation file)
- **Phase 5**: T024, T025, T027 can run in parallel
- **Phase 6**: T031, T032 can run in parallel
- **Phase 7**: T036, T038, T039, T041 can run in parallel (different files)

---

## Implementation Strategy

### MVP First (User Stories 2 + 1)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational types + config schema
3. Complete Phase 3: Manifest validation (US2)
4. Complete Phases 4-6: Registry paths, contract, lock file
5. Complete Phase 7: Install command (US1)
6. **STOP and VALIDATE**: All BDD scenarios pass end-to-end

### Incremental Delivery

1. PR 1: Config schema + types → foundation ready
2. PR 2: Manifest validation → can validate package manifests standalone
3. PR 3: Registry client + stub → can fetch packages from stub registry
4. PR 4: Lock file → can create/verify lock files
5. PR 5: Install command → full `ailign install` working
6. PR 6: Polish → production-ready

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- US2 (manifest) and US5 (registry contract) are implemented before US1 (install) because install integrates them
- Each PR maps to one or two phases and stays within the 333-line target
- Existing `ailign sync` must remain unchanged — verify in polish phase
- Stub registry uses `httptest.NewServer` — no external dependencies
- Contract artifacts live in `contracts/registry/` per consumer-driven decision (Option B)
