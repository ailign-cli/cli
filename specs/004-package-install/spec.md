# Feature Specification: Package Manifest & Install

**Feature Branch**: `004-package-install`
**Created**: 2026-02-20
**Status**: Draft
**Input**: User description: "Package manifest format and ailign install command with type-scoped registry paths, YAML manifest (non-dotfile), and contract-first stub registry testing. Start with type: instructions only."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Install Packages from Registry (Priority: P1)

**In order to** use org-wide instruction standards in my repository without manually copying files
**As a** developer working in an organization with multiple repositories
**I want to** run `ailign install` and have the declared instruction packages fetched and rendered to my configured tool formats

A developer has a `.ailign.yml` configuration that declares one or more instruction packages (e.g., `instructions/company/security@1.3.0`). They run `ailign install` and the CLI fetches each package from the registry, composes the content with any local overlays, and renders the result to the configured target formats (Claude, Cursor, Copilot). The developer sees a summary of what was installed and can verify the output with `ailign status`.

**Why this priority**: This is the core value proposition — fetching and applying shared instruction packages. Without this, the tool only supports local overlays (feature 002). This enables the central-baseline-plus-repo-overlay model that is the product's reason for being.

**Independent Test**: Run `ailign install` with a config referencing a package, verify the package content is fetched, composed with overlays, and rendered to target files. Can be tested against a stub registry.

**Acceptance Scenarios**: See [`features/package-install.feature`](../../features/package-install.feature)

| Scenario                                       | Description                                                              |
|------------------------------------------------|--------------------------------------------------------------------------|
| Install single instruction package             | Fetches package, renders to configured targets                           |
| Install multiple packages                      | Fetches all declared packages, composes in declared order                |
| Install with local overlays                    | Package content is composed with local overlay files                     |
| Install shows summary                          | Output lists each installed package with version                         |
| Install with JSON output                       | Summary is machine-readable when `--format json` is used                 |
| Install is idempotent                          | Running install twice produces the same output files                     |
| Install creates lock file                      | Records exact versions and checksums of installed packages               |

---

### User Story 2 - Package Manifest Defines Package Identity (Priority: P1)

**In order to** publish and discover instruction packages with clear metadata
**As a** package author (e.g., a security team lead)
**I want to** define my package's identity, type, version, and content in a manifest file

A package author creates an `ailign-pkg.yml` manifest file in their package directory. The manifest declares the package name (scoped: `company/security`), type (`instructions`), version (semver), and the content file(s) that make up the package. The CLI validates the manifest when installing and uses it to determine how to fetch and compose the package.

**Why this priority**: The manifest is the contract between package authors and consumers. Without it, there is no shared understanding of what a package contains. This is co-P1 with install because install depends on manifest validation.

**Independent Test**: Create a valid manifest, verify the CLI accepts it. Create invalid manifests (missing fields, bad version, wrong type), verify the CLI rejects them with clear errors.

**Acceptance Scenarios**: See [`features/package-manifest.feature`](../../features/package-manifest.feature)

| Scenario                                       | Description                                                              |
|------------------------------------------------|--------------------------------------------------------------------------|
| Valid instruction package manifest             | Manifest with all required fields is accepted                            |
| Manifest missing required field                | Clear error identifying the missing field                                |
| Manifest with invalid version format           | Error explaining semver requirement                                      |
| Manifest with unsupported type                 | Error listing supported types                                            |
| Manifest name matches registry path            | Name `company/security` matches `instructions/company/security` path     |

---

### User Story 3 - Type-Scoped Registry Paths (Priority: P2)

**In order to** distinguish between different kinds of packages (instructions, MCP configs, commands) in the registry
**As a** developer declaring dependencies in `.ailign.yml`
**I want to** use type-prefixed package references that make the content type explicit

Package references in `.ailign.yml` include the content type as a path prefix: `instructions/company/security@1.3.0`. This makes it immediately clear what kind of content you are requesting. The type in the reference must match the type declared in the package manifest. Initially only `instructions` is supported; other types (`mcp`, `commands`, `agents`, `packages`) are reserved for future features.

**Why this priority**: Type-scoping is an architectural decision that affects the registry contract and `.ailign.yml` schema. It must be designed now even though only `instructions` is implemented initially. However, the actual multi-type support can come later, making this P2.

**Independent Test**: Declare a package with `instructions/` prefix in config, verify it resolves correctly. Declare a package with an unsupported type prefix, verify a clear error is returned.

**Acceptance Scenarios**: See [`features/package-registry-paths.feature`](../../features/package-registry-paths.feature)

| Scenario                                       | Description                                                              |
|------------------------------------------------|--------------------------------------------------------------------------|
| Instructions type prefix resolves              | `instructions/company/security@1.0.0` fetches correctly                  |
| Unsupported type prefix rejected               | `mcp/company/tools@1.0.0` returns error with supported types list        |
| Missing type prefix rejected                   | `company/security@1.0.0` (no type) returns error explaining the format   |
| Type mismatch between config and manifest      | Error when config says `instructions/` but manifest says different type   |

---

### User Story 4 - Version Resolution and Lock File (Priority: P2)

**In order to** have reproducible builds and controlled updates
**As a** developer or CI/CD pipeline
**I want to** lock installed package versions and verify integrity through checksums

When `ailign install` runs, it records the exact resolved version and content checksum of each installed package in a lock file (`ailign-lock.yml`). Subsequent installs use the lock file to ensure the same versions are used unless the developer explicitly updates. The lock file is committed to version control.

**Why this priority**: Version locking is essential for governance (Principle VI) and reproducibility, but install can work without it initially by always fetching the declared version. The lock file adds safety but is not strictly required for the first working install.

**Independent Test**: Run install, verify lock file is created. Run install again without changing config, verify it uses locked versions. Change a version in config, verify lock file updates.

**Acceptance Scenarios**: See [`features/package-version-lock.feature`](../../features/package-version-lock.feature)

| Scenario                                       | Description                                                              |
|------------------------------------------------|--------------------------------------------------------------------------|
| Lock file created on first install             | Records versions and checksums                                           |
| Subsequent install uses locked versions        | Same output even if registry has newer patch                             |
| Config version change updates lock             | Changing `@1.3.0` to `@1.4.0` fetches new version and updates lock      |
| Lock file checksum mismatch detected           | Error when package content doesn't match recorded checksum               |
| Lock file is human-readable                    | YAML format, sorted deterministically                                    |

---

### User Story 5 - Registry Contract for Testing (Priority: P3)

**In order to** develop the CLI and registry independently
**As a** developer of either the CLI or the registry
**I want to** a well-defined contract between CLI and registry that both sides test against

The registry API contract defines how the CLI discovers and fetches packages. The contract lives in the CLI repository (under `contracts/registry/`) as an OpenAPI spec, JSON schemas for the manifest format, and fixture files with example request/response pairs. The CLI tests against a stub registry that implements this contract. The future registry service fetches the same contract artifacts and validates conformance against them. This consumer-driven approach enables the CLI to develop first while providing a clear specification for the registry to implement against.

When the registry repository is created, the contract can be extracted into a shared repository (e.g., `ailign-cli/contracts`) that both sides depend on. This migration is mechanical — move files, update references.

**Why this priority**: The contract is architecturally important but is a developer concern, not a user-facing feature. The stub registry is needed for testing but is not shipped to users.

**Independent Test**: CLI tests pass against the stub registry. Contract specification exists as a shareable artifact that an independent team could use to build a compatible registry.

**Acceptance Scenarios**: See [`features/package-registry-contract.feature`](../../features/package-registry-contract.feature)

| Scenario                                       | Description                                                              |
|------------------------------------------------|--------------------------------------------------------------------------|
| Fetch package by type, name, and version       | Stub returns correct package content                                     |
| Package not found                              | Stub returns 404, CLI shows clear error                                  |
| Version not found                              | Stub returns 404 for version, CLI suggests available versions            |
| Registry unreachable                           | CLI shows connection error with retry guidance                           |
| Registry returns invalid manifest              | CLI rejects malformed response with details                              |

---

### Edge Cases

- What happens when the registry is down during install? CLI shows a clear connection error and exits with code 2. If a lock file exists, suggest using `--offline` mode (future feature).
- What happens when a package version is yanked/deleted from the registry after being locked? Checksum mismatch error with clear explanation that the package content has changed.
- What happens when `.ailign.yml` declares the same package twice? Error listing the duplicate with line numbers.
- What happens when the lock file is corrupted or has invalid YAML? Error with remediation: delete lock file and re-run install.
- What happens when the lock file references a version that no longer matches the config? Install detects the drift and fetches the version declared in config, updating the lock file.
- What happens when install is run without any packages declared? No-op with informational message ("No packages declared in .ailign.yml").
- What happens when the network drops mid-download? Atomic operation: no partial state, no corrupted lock file. Error with retry guidance.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: `ailign install` MUST fetch all packages declared in `.ailign.yml` from the registry
- **FR-002**: `ailign install` MUST compose fetched package content with local overlays in declared order
- **FR-003**: `ailign install` MUST render composed content to all configured target formats
- **FR-004**: `ailign install` MUST produce both human-readable and JSON output (`--format json`)
- **FR-005**: `ailign install` MUST be idempotent — running it twice with no config changes produces identical output files
- **FR-006**: `ailign install` MUST create or update a lock file (`ailign-lock.yml`) recording exact versions and checksums
- **FR-007**: `ailign install` MUST exit with code 0 on success, code 2 on error
- **FR-008**: Package references in `.ailign.yml` MUST use the format `<type>/<scope>/<name>@<version>` (e.g., `instructions/company/security@1.3.0`)
- **FR-009**: The CLI MUST validate that the type prefix in the package reference matches the type declared in the fetched package manifest
- **FR-010**: Only the `instructions` type MUST be supported initially; other type prefixes (`mcp`, `commands`, `agents`, `packages`) MUST be rejected with an error listing supported types
- **FR-011**: The package manifest file MUST be named `ailign-pkg.yml` (not a dotfile)
- **FR-012**: The package manifest MUST contain: `name` (scoped: `<scope>/<name>`), `type` (content type), `version` (semver), `description`, and `content.main` (path to primary content file)
- **FR-013**: The CLI MUST validate the package manifest schema and report clear errors for missing or invalid fields
- **FR-014**: The lock file MUST be deterministically ordered (sorted by package reference) so version control diffs are minimal
- **FR-015**: The CLI MUST detect checksum mismatches between the lock file and fetched content, and report them as errors
- **FR-016**: The existing `ailign sync` command (feature 002, local-only overlays) MUST continue to work unchanged
- **FR-017**: `ailign install` MUST perform all file operations atomically — no partial state on failure
- **FR-018**: The registry API contract MUST be defined as a testable specification (OpenAPI spec, JSON schemas, fixture files) in the CLI repository under `contracts/registry/`, enabling both CLI and a future registry to validate against it independently

### Key Entities

- **Package**: A versioned, typed unit of content identified by `<type>/<scope>/<name>@<version>`. Contains a manifest (`ailign-pkg.yml`) and content files.
- **Package Manifest**: The `ailign-pkg.yml` file that declares a package's identity (name, type, version), metadata (description), and content structure (main file path).
- **Package Reference**: A string in `.ailign.yml` that identifies a specific package version: `instructions/company/security@1.3.0`. Composed of type prefix, scoped name, and version.
- **Lock File**: `ailign-lock.yml` at repository root. Records the exact resolved version and content checksum of each installed package for reproducibility.
- **Registry**: External service that stores and serves packages. Accessed via a defined API contract. Not part of this repository — the CLI implements the client side only.
- **Stub Registry**: A test-only implementation of the registry contract used for CLI testing. Serves predefined package content from fixture files.
- **Contract Artifacts**: OpenAPI spec, JSON schemas, and fixture files living in `contracts/registry/` that define the registry API. Consumer-driven: owned by the CLI repo initially, extractable to a shared repo when the registry is built.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A developer can go from an empty repository to a working `ailign install` with one instruction package in under 3 minutes
- **SC-002**: `ailign install` with 3 packages completes in under 30 seconds (per constitution performance target)
- **SC-003**: Running `ailign install` twice with no config changes produces byte-identical output files and an unchanged lock file
- **SC-004**: 100% of manifest validation errors include the specific field name and expected format in the error message
- **SC-005**: The registry contract specification is sufficient for an independent team to build a compatible registry without access to CLI source code
- **SC-006**: All CLI package-install functionality is testable without a live registry (stub registry covers all scenarios)

## Assumptions

- The registry API will use HTTPS for all communication (per constitution security requirements)
- Package content is text-only for v1 (per scope.md — no executable skills)
- The `instructions` content type produces a single rendered file per target (same as current overlay sync behavior)
- Version ranges (e.g., `^1.3.0`, `~1.3`) are out of scope for this feature — exact versions only
- Package authentication/authorization is out of scope — the registry is assumed to be accessible (authentication can be layered later)
- The `.ailign.yml` `packages:` field will replace the current bare package references (e.g., `company/security@1.3.0`) with type-prefixed references (e.g., `instructions/company/security@1.3.0`). This is a breaking change to the config schema.
- The registry contract lives in the CLI repository initially (consumer-driven, Option B). When the registry repo is created, the contract artifacts can be extracted to a shared repo (e.g., `ailign-cli/contracts`) — this migration is mechanical and does not affect the contract itself.

## Supersedes

This feature extends the `.ailign.yml` schema introduced in feature 001 (config parsing) by adding a new `packages` field with type-prefixed package references (e.g., `instructions/company/security@1.3.0`). The current schema has no `packages` field, so this is a purely additive change — no migration or breaking change is needed.

This feature builds on feature 002 (local instruction sync) by adding remote package content as an additional composition source. The `ailign sync` command continues to work for local-only workflows. `ailign install` adds the fetch-from-registry step before composition.
