<!--
  === Sync Impact Report ===
  Version change: 1.1.0 -> 1.2.0
  Bump rationale: MINOR — new principle added (X. Subtraction),
    new Deprecation Process section added
  Modified principles: None (existing I-IX unchanged)
  Added sections:
    - Principle X: Subtraction
    - Deprecation Process (under Versioning & Breaking Changes)
  Removed sections: None
  Updated sections:
    - Decision Framework: added question 10 (Subtraction)
    - Prohibited Patterns: added one entry for X violations
  Templates requiring updates:
    - .specify/templates/plan-template.md ✅ compatible
      (Constitution Check section uses dynamic "[Gates determined based on
      constitution file]" — no update needed)
    - .specify/templates/spec-template.md ✅ compatible
      (User stories, requirements, success criteria unaffected)
    - .specify/templates/tasks-template.md ✅ compatible
      (Test-first workflow, parallel markers, checkpoints unaffected)
    - .specify/templates/checklist-template.md ✅ compatible
      (Generic template, dynamically populated)
    - .specify/templates/agent-file-template.md ✅ compatible
    - CLAUDE.md ✅ updated — references updated to "10 core principles"
    - README.md ✅ compatible (links to constitution.md)
    - features/steps/suite_test.go ✅ updated — added @deprecated and
      @pending-deprecation to tag exclusion filter
    - .specify/memory/deprecation.md ✅ created — deprecation tracker
  Follow-up TODOs: None
  === End Sync Impact Report ===
-->
# AIlign Constitution

This document defines the core principles that guide all design and
implementation decisions for AIlign.

## Core Principles

### I. CLI-First with Dual Output

**Every feature MUST be accessible via CLI.**

- Text I/O protocol: stdin/args -> stdout, errors -> stderr
- Support both JSON (for automation) and human-readable formats
- Exit codes designed for CI/CD integration
  (0=success, 1=drift/outdated, 2=error)
- No interactive prompts by default (automation-friendly)
- Single binary distribution, zero runtime dependencies

**Rationale:** Developers live in the terminal. CLI-first ensures
automation, scripting, and CI/CD integration are first-class use
cases, not afterthoughts.

### II. Transparency Over Magic (NON-NEGOTIABLE)

**Developers MUST always understand what AIlign is doing and why.**

- Every action MUST be explainable via `ailign explain`
- Deterministic, predictable behavior
  (no AI-powered "smart" merging in v1)
- Clear diff before any changes (`ailign diff`)
- Status shows current vs. desired state
- Error messages MUST explain what went wrong AND how to fix it

**Anti-patterns:**
- Black-box operations that cannot be explained
- Silent failures or warnings
- Unexpected file modifications
- "Trust me" behavior

**Rationale:** Without transparency, developers will not trust the
tool. Without trust, no adoption.

### III. Fail Safe, Not Silent

**Errors MUST be obvious and recoverable, never hidden.**

- Validate config before modifying files
- Atomic operations (all-or-nothing)
- Clear rollback path (`ailign pull --version <previous>`)
- Drift detection alerts but does not auto-fix
- Lock files prevent unexpected changes

**Anti-patterns:**
- Partially applied changes
- Swallowing errors
- Auto-fixing without confirmation
- Leaving system in inconsistent state

**Rationale:** A tool that breaks things quietly is worse than no
tool at all.

### IV. Test-First Development

**BDD+TDD dual-loop is mandatory for all features.**

- Outer loop (BDD): `.feature` file scenarios drive acceptance criteria
  - Scenarios written during `/speckit.plan` as executable Gherkin
  - Step definitions written before implementation (RED)
  - Step definitions pass after implementation (GREEN)
- Inner loop (TDD): Unit tests drive component design
  - Tests written → Tests fail → Implement → Tests pass → Refactor
  - Red-Green-Refactor cycle strictly enforced
- Integration tests required for:
  - Renderer contracts (new tool support)
  - Package composition logic
  - API client interactions
  - CLI command workflows

**Coverage requirements:**
- Core libraries: >90% unit test coverage
- Renderers: 100% coverage (too critical to skip)
- CLI commands: BDD scenarios + E2E integration tests
- Every user story: At least one passing `.feature` scenario

**Rationale:** BDD ensures we build the right thing (acceptance criteria).
TDD ensures we build it right (code quality). Both are required.

### V. Composition Over Monolith

**Enable flexible combination without entanglement.**

- Clear separation: central baseline + repo overlays
- Packages are modular and independently versioned
- Multiple packages can be combined without conflicts
- Content tiers allow priority-based composition
- Each renderer is independent (adding tools MUST NOT affect
  existing ones)

**Anti-patterns:**
- Tightly coupled packages
- Hidden dependencies between packages
- Forcing single-package solutions
- Mixing concerns in one package

**Rationale:** Organizations have different needs. Monolithic
packages force one-size-fits-all.

### VI. Governance as Foundation

**Compliance and auditability are first-class concerns.**

- All packages are versioned and immutable (semver)
- Changes are traceable (who, when, which version)
- Drift detection is built-in, not optional
- Package provenance MUST always be clear
- Lock files prevent unexpected changes

**Anti-patterns:**
- Mutable packages (overwrites)
- No audit trail
- Allowing unversioned content
- Auto-updates without approval

**Rationale:** Security teams need compliance. Enterprises need
auditability. Build it in from day one.

### VII. Size-Aware by Design

**Respect tool constraints and bandwidth limits.**

- Every renderer MUST know its tool's limits (e.g., Cursor 8KB)
- Content tiers enable priority-based inclusion
  (critical/recommended/extra)
- Size budgets prevent silent truncation
- Warnings when content exceeds limits
- Critical content MUST fit or error (never silent exclusion)

**Anti-patterns:**
- Ignoring size limits
- Silent truncation
- Treating all content as equal priority
- Unbounded package sizes

**Rationale:** Different tools have different limits. Ignoring this
breaks tools or loses critical content.

### VIII. Cross-Tool Parity

**One source of truth serves all AI tools equally.**

- Tool-specific formats are render targets, not sources
- Central content is tool-agnostic
- Renderers handle tool quirks (size limits, format, etc.)
- Adding new tool support MUST NOT require changing packages
- No tool is "primary" - all are equal citizens

**Anti-patterns:**
- Writing content in one tool's format
- Tool-specific packages
- Assuming single tool usage
- Ignoring tool constraints

**Rationale:** Teams use multiple tools. Lock-in kills adoption.

### IX. Working Software

**Functional, validated software is the primary measure of progress.**

- Every commit that changes code or tests MUST include both
  implementation and corresponding tests
- Every commit MUST validate (lint, format, static checks pass)
- Every commit SHOULD be able to build
- Every Pull Request MUST include code implementation and test
  implementation
- Every Pull Request MUST validate (all checks pass)
- Every Pull Request MUST build successfully
- No "test-only" or "code-only" commits when both are needed
  for the change to be meaningful

**Anti-patterns:**
- Commits with implementation but no tests
- Commits with tests but no implementation (except RED-phase
  TDD commits that are immediately followed by GREEN-phase)
- PRs that break the build
- PRs that skip validation

**Rationale:** Working software is the primary measure of progress.
A commit or PR that does not validate or build is not progress —
it is technical debt. Keeping code and tests together ensures
every increment is a verified, deliverable unit.

### X. Subtraction

**Every change MUST consider what can be removed, not only
what needs to be added.**

People have a natural bias toward addition. When designing,
implementing, or specifying changes, removal MUST always be
evaluated as an option alongside addition. This applies to:

- **Code:** Previously written code that can be rewritten more
  simply in light of new development SHOULD be replaced, not
  preserved alongside the new approach
- **Features:** Capabilities that are superseded by new ones
  SHOULD follow the Deprecation Process (see below) rather
  than accumulating indefinitely
- **Specifications:** Existing specs in `specs/` and existing
  BDD scenarios are immutable records. However, `@deprecate`
  and `@deprecated` tags MAY be added to existing scenarios.
  New replacement scenarios MAY be added to the same feature
  file, tagged `@pending-deprecation` (excluded from test
  execution) and referencing the scenario they replace in the
  Gherkin free-form description below the scenario title
- **Dependencies:** Unused or redundant dependencies SHOULD
  be removed when new work makes them unnecessary

This principle does not mandate removal in every change. It
mandates that removal is *considered* as an option during every
specification and implementation decision.

**Anti-patterns:**
- Adding new code alongside old code that serves the same
  purpose without evaluating removal
- Accumulating features without ever deprecating superseded ones
- Keeping dead code, unused imports, or orphaned files out of
  habit
- Treating existing code as untouchable when new development
  offers a simpler path

**Rationale:** Codebases grow in complexity over time. Without
a deliberate counter-pressure toward simplicity, accidental
complexity accumulates and slows development. Subtraction is
the antidote to entropy.

## Performance Standards

### Speed Requirements

- `ailign status`: <1 second (offline mode)
- `ailign pull`: <30 seconds (typical 3-package fetch + render)
- `ailign diff`: <2 seconds
- `ailign explain`: <1 second

### Resource Constraints

- Binary size: <50MB
- Memory usage: <100MB during normal operations
- Network: Graceful offline degradation (use cached packages)

**Rationale:** Slow tools do not get used. Developer time is
expensive.

## Security Requirements

### Supply Chain Security

**v1 (MVP):**
- Text-only packages (no executable skills)
- HTTPS for all API communication
- Package checksums in lock file

**v2 (Post-MVP):**
- Executable skills require signature verification
- Allowlist mechanism for script execution
- Audit trail for all executions

**Rationale:** Distributing executable code is a supply-chain risk.
Start text-only, add security later.

### Sensitive Data

- MUST never log API keys, tokens, or credentials
- MUST redact sensitive data in error messages
- MUST support environment variable injection
  (never hardcode secrets)

## Development Workflow

### Feature Development

1. **Spec first:** Write feature spec in `specs/[name]/spec.md`
2. **Constitution check:** Verify alignment with principles
3. **Task breakdown:** Create `tasks.md` for complex features
4. **BDD outer loop:** Write/verify `.feature` files and step definitions (RED)
5. **TDD inner loop:** Write unit tests -> Get approval -> Implement
6. **BDD verification:** Step definitions pass (GREEN)
7. **Integration test:** Verify feature works end-to-end
8. **Documentation:** Update user docs and examples

### Code Review Requirements

- All PRs MUST verify constitution compliance
- Complexity MUST be justified (YAGNI principle)
- Breaking changes require migration guide
- Performance impact MUST be measured

### Quality Gates

- All tests MUST pass (no exceptions)
- Coverage MUST meet thresholds
- CLI help text MUST be updated
- Examples MUST work
- Every commit MUST validate and build (Principle IX)
- Every PR MUST include implementation + tests (Principle IX)

## Versioning & Breaking Changes

### Versioning Scheme

**Format:** MAJOR.MINOR.PATCH (semver)

- **MAJOR:** Breaking changes
  (config format, API contract, CLI interface)
- **MINOR:** New features (backward compatible)
- **PATCH:** Bug fixes (backward compatible)

### Breaking Change Policy

- Breaking changes require:
  - Migration guide
  - Deprecation warnings in previous MINOR version
  - Announcement with 2-week notice minimum
- Exception: Security vulnerabilities (immediate fix)

**Rationale:** Developers depend on stable interfaces. Breaking
changes without warning kill trust.

### Deprecation Process

This process governs the removal of capabilities that were
previously possible. It MUST be followed after v1.0.0 is
released. Below v1.0.0, breaking changes can be expected
without formal deprecation.

Existing specs in `specs/` are immutable records and MUST NOT
be modified. Existing BDD scenarios are also immutable — their
Given/When/Then steps MUST NOT be changed. However:

- `@deprecate` and `@deprecated` tags MAY be added to
  existing scenarios
- New replacement scenarios MAY be added to the same feature
  file, tagged `@pending-deprecation` (excluded from test
  execution until the deprecation is enacted), referencing the
  scenario they replace in the Gherkin free-form description
  below the scenario title

#### Identifying Deprecations

When the specification process (`/speckit.specify`,
`/speckit.plan`, `/speckit.tasks`) determines that an existing
user-facing capability (command, flag, config option, or any
interface through which the user interacts with the CLI) is
being superseded:

1. The `spec.md` of the new feature MUST reference the
   existing specs and feature files (including specific
   scenarios if applicable) that are superseded
2. The superseded feature file(s) and/or specific scenario(s)
   MUST receive the `@deprecate` tag
3. If the new feature includes replacement behavior for the
   deprecated capability, a new scenario MUST be added to the
   same feature file tagged `@pending-deprecation`. This
   scenario MUST reference the scenario it replaces in the
   Gherkin free-form description below the scenario title
4. A deprecation entry MUST be added to
   `.specify/memory/deprecation.md` under the heading:
   `### <YYYY-MM-DD> - <Deprecation title> (<current version>)`.
   The date MUST use the ISO 8601 format `YYYY-MM-DD`. The entry MUST use
   the checklist format defined in `.specify/memory/deprecation.md` and list any
   `@pending-deprecation` scenarios that need to be activated
   upon removal

#### Deprecation Warnings

All BDD scenarios tagged `@deprecate` MUST cause a standard
deprecation notice to be emitted to the user. The test
harness MUST provide a shared assertion — bound to the
`@deprecate` tag via step definitions or a `Before` hook —
that verifies this automatically without modifying the
scenario's Given/When/Then steps. This notice MUST:

- Clearly identify the capability being deprecated
- Explain what replaces it (the new capability)
- Reference the version in which deprecation was announced

A dedicated BDD scenario SHOULD exist to test the
deprecation notice behavior itself (i.e., that the harness
correctly emits and verifies the notice for tagged
scenarios).

#### Deprecation Timeline

- The operator determines when actual removal is warranted
- The operator SHOULD be notified after 5-10 minor/patch
  releases (depending on scope of the deprecation)
- If any specification results in a MAJOR (breaking change)
  release, all open deprecations MUST be implemented in that
  release as a dedicated phase

#### Executing Removal

When the actual removal takes place:

1. The `@deprecate` tag on affected scenarios MUST be
   replaced with the `@deprecated` tag
2. The `@pending-deprecation` tag on replacement scenarios
   MUST be removed, activating them for test execution
3. Scenarios tagged `@deprecated` are excluded from test
   execution (alongside `@wip` and `@ci`)
4. The deprecation entry in `.specify/memory/deprecation.md`
   MUST be updated with the removal version and date

## Decision Framework

When evaluating new features or design choices, ask:

1. **Transparency:** Can we explain exactly what this does and why?
2. **Friction:** Does this add or remove developer effort?
3. **Safety:** What happens if this fails? Can we recover?
4. **Composition:** Does this work with other features or create
   coupling?
5. **Governance:** Can we audit and trace this decision?
6. **Parity:** Does this work equally well across all supported
   tools?
7. **Testing:** Can we test this reliably?
8. **Performance:** Does this meet speed requirements?
9. **Working Software:** Does every increment validate and build?
10. **Subtraction:** Have we considered what can be removed or
    simplified instead of only adding?

**If the answer to any question is "no" or "unclear," revisit the
design.**

## Values Hierarchy

When principles conflict, prioritize in this order:

1. **Safety** (Fail Safe, Not Silent) -
   Never leave system broken
2. **Trust** (Transparency Over Magic) -
   Without trust, no adoption
3. **Working Software** (Working Software) -
   Non-validated code is not progress
4. **Testing** (Test-First Development) -
   Quality is non-negotiable
5. **Governance** (Governance as Foundation) -
   Required for enterprise
6. **Performance** (Speed Requirements) -
   Slow tools do not get used
7. **Other principles** -
   Important but can be optimized later

## Prohibited Patterns

These are explicitly forbidden as they violate core principles:

- **Silent auto-updates** - Violates trust and safety
- **Mutable packages** - Violates governance
- **Unexplainable merges** - Violates transparency
- **Single-tool lock-in** - Violates cross-tool parity
- **Interactive prompts in automation** - Violates CLI-first
- **Hidden failures** - Violates fail safe
- **Unbounded content** - Violates size-aware design
- **Shipping without tests** - Violates test-first
- **Shipping without BDD scenarios** - Violates test-first
  (every user story needs executable acceptance criteria)
- **Code-only commits without tests** - Violates working software
  (implementation and tests MUST ship together)
- **PRs that do not build** - Violates working software
  (every PR MUST be a validated, buildable increment)
- **Adding without considering removal** - Violates subtraction
  (every change MUST evaluate whether existing code, features,
  or dependencies can be removed or simplified)

## Governance

### Constitution Authority

- This constitution supersedes all other practices
- Amendments require:
  - Documented justification
  - Team approval
  - Migration plan (if breaking existing features)
  - Version update

### Compliance Verification

- All PRs MUST verify constitution compliance
- Code reviews MUST check against principles
- Feature specs MUST reference relevant principles
- Violations block merging

### Evolution Process

- Constitution is a living document (can be amended)
- Changes follow same rigor as code changes
- Track amendments with version history

**Version**: 1.2.0 | **Ratified**: 2025-02-13 | **Last Amended**: 2026-02-20
