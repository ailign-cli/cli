# Feature Specification: Configuration File Parsing

**Feature Branch**: `001-config-parsing`
**Created**: 2026-02-13
**Status**: Draft
**Input**: User description: "Being able to parse a configuration file which can be stored within the working directory (where the CLI will run). There needs to be schema validation."

## Scope

This feature covers **parsing and validating** the `.ailign.yml`
configuration file only. The schema for this feature is limited to the
`targets` field. It does NOT include:
- Config file generation (`ailign init`) - separate feature
- Schema documentation / discovery commands - separate feature
- `packages` and `local_overlays` schema fields - separate features
- Acting on the parsed config (fetching packages, rendering) -
  separate features
- Package registry interaction - separate feature

## Clarifications

### Session 2026-02-13

- Q: Should `ailign init` (config generation) remain in this feature
  or be moved to a separate feature? → A: Move to separate feature.
  This feature = parse + validate only.
- Q: Should this feature include a standalone `ailign validate`
  command? → A: Yes, add `ailign validate` as an explicit
  validate-only command with no side effects.
- Q: How should schema documentation be exposed to developers?
  → A: Out of scope for this feature. Defer to a separate
  documentation feature.
- Q: What is the minimum valid `.ailign.yml` configuration?
  → A: Only `targets` is required. `packages` and `local_overlays`
  can be empty or omitted.
- Q: Should the CLI walk up parent directories to find config?
  → A: No. Current working directory only. Simple and predictable.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Parse Configuration File (Priority: P1)

In order to declare which AI tools my repository targets, as a
developer working in a repository, I want to be able to create a
`.ailign.yml` config file that specifies target tools, and have the
CLI parse it correctly.

The CLI reads and parses the `.ailign.yml` file from the working
directory, making the configuration available for all subsequent
operations (pull, status, diff, explain). If the file is well-formed
and valid, the CLI proceeds silently. If the file is missing, the CLI
reports a clear error explaining that no configuration was found and
exits with code 2.

**Why this priority**: Without the ability to read and parse the
configuration file, no other CLI feature can function. This is the
foundational capability that every command depends on.

**Independent Test**: Can be fully tested by placing a valid
`.ailign.yml` file in a directory, running the CLI, and verifying the
configuration values are correctly loaded. Delivers the core ability
to configure AIlign per-repository.

**Acceptance Scenarios**:

1. **Given** a repository with a valid `.ailign.yml` in the working
   directory, **When** the CLI parses the configuration, **Then** all
   declared targets are correctly loaded and accessible to other
   commands.
2. **Given** a repository with no `.ailign.yml` file, **When** the CLI
   attempts to load configuration, **Then** it reports an error to
   stderr explaining the file is missing and exits with code 2.
3. **Given** a repository with an `.ailign.yml` file that is empty,
   **When** the CLI attempts to parse it, **Then** it reports a
   validation error to stderr indicating that the required `targets`
   field is missing.

---

### User Story 2 - Schema Validation with Actionable Errors (Priority: P2)

In order to catch mistakes before running `ailign pull`, as a
developer, I want to be able to validate my config file syntax and
schema and receive actionable error messages.

The CLI validates the configuration against a defined schema before
proceeding with any command. Additionally, a dedicated
`ailign validate` command allows developers to explicitly check their
configuration without triggering any other operations. If validation
fails, the CLI reports every violation with the specific field path,
what was expected, what was found, and how to fix it. The developer
can correct all issues in one pass rather than fixing them one at a
time.

**Why this priority**: Schema validation ensures the CLI never operates
on invalid configuration, preventing downstream failures. Actionable
error messages align with the Transparency principle and reduce
developer frustration.

**Independent Test**: Can be tested by providing `.ailign.yml` files
with various violations (wrong types, missing fields, invalid formats)
and verifying each produces a specific, helpful error message.

**Acceptance Scenarios**:

1. **Given** a configuration file with an invalid target name (e.g.,
   `vscode`), **When** the CLI validates the schema, **Then** it
   reports the exact field path, the expected values, and what was
   provided, to stderr.
2. **Given** a configuration file with multiple validation errors,
   **When** the CLI validates it, **Then** all errors are reported
   at once (not just the first one), each with field path, expected
   value, and remediation guidance.
3. **Given** a configuration file with an unknown field not in the
   schema, **When** the CLI validates it, **Then** it reports a
   warning about the unrecognized field (to stderr) but does not
   treat it as a fatal error.
4. **Given** a configuration file where all fields are valid, **When**
   the CLI validates the schema, **Then** no warnings or errors are
   emitted and the CLI proceeds normally.
5. **Given** a valid `.ailign.yml`, **When** the developer runs
   `ailign validate`, **Then** the CLI reports success to stdout
   and exits with code 0, without performing any other operations.
6. **Given** an invalid `.ailign.yml`, **When** the developer runs
   `ailign validate`, **Then** the CLI reports all validation errors
   to stderr and exits with code 2.

---

### Edge Cases

- What happens when the configuration file contains valid YAML but
  uses tabs instead of spaces? The CLI MUST parse it normally (tabs
  are valid YAML).
- What happens when the configuration file has a Unicode BOM? The CLI
  MUST handle BOM-prefixed files gracefully.
- What happens when the file permissions prevent reading? The CLI MUST
  report a clear permission error with the file path.
- What happens when `.ailign.yml` is a symlink? The CLI MUST follow
  symlinks and parse the target file.
- What happens when the targets list contains duplicates? The CLI
  MUST report a validation error (duplicates not allowed).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The CLI MUST look for `.ailign.yml` in the current
  working directory only (no parent directory traversal) when any
  command is executed.
- **FR-002**: The CLI MUST parse the configuration file as YAML and
  make its contents available to all commands.
- **FR-003**: The CLI MUST validate the parsed configuration against
  a defined schema before any command proceeds.
- **FR-004**: The CLI MUST report all validation errors at once, each
  including the field path, expected value/format, actual value, and
  remediation guidance.
- **FR-005**: Validation errors MUST be written to stderr and cause
  the CLI to exit with code 2.
- **FR-006**: The CLI MUST report unrecognized fields as warnings to
  stderr without treating them as fatal errors.
- **FR-007**: The `targets` field MUST be required and contain at
  least one item.
- **FR-008**: The CLI MUST validate that target names are from a
  known set of supported AI tools (claude, cursor, copilot,
  windsurf).
- **FR-009**: The CLI MUST reject duplicate target names.
- **FR-011**: When no configuration file is found, the CLI MUST
  report the absence to stderr and exit with code 2.
- **FR-012**: The `ailign validate` command MUST validate the config
  file and report results without triggering any other operations.
- **FR-013**: The `ailign validate` command MUST exit with code 0 on
  success and code 2 on validation failure.
- **FR-014**: The CLI MUST support both JSON and human-readable
  error output for validation errors (controlled by a `--format`
  flag or equivalent).

### Key Entities

- **Configuration File**: The `.ailign.yml` file in the working
  directory. Contains target tool names (required). One per
  repository. Future features will add `packages` and
  `local_overlays` fields.
- **Target**: A named AI tool that AIlign renders output for (e.g.,
  `claude`, `cursor`, `copilot`, `windsurf`). Appears in the
  targets list.
- **Schema**: The set of rules defining valid configuration
  structure, field types, required fields, and value constraints.
  Defined as JSONSchema, embedded in the binary.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A valid configuration file is parsed and loaded in
  under 100 milliseconds.
- **SC-002**: 100% of schema violations produce an error message
  that includes the field path and remediation guidance.
- **SC-003**: Developers can resolve all configuration errors in a
  single edit cycle (all errors reported at once, not one at a time).
- **SC-004**: The CLI correctly rejects 100% of invalid
  configurations and accepts 100% of valid configurations (zero
  false positives or negatives in schema validation).

## Assumptions

- The configuration file name is `.ailign.yml` as defined in the
  project scope document.
- The configuration format is YAML, chosen for readability and
  developer familiarity.
- Supported target tool names: `claude`, `cursor`, `copilot`,
  `windsurf`.
- Schema validation happens synchronously as part of CLI startup,
  before any network calls or file modifications.
- Config file generation (`ailign init`) will be a separate feature.
