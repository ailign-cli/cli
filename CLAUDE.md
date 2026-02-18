# AIlign CLI - Development Guide

Instruction governance & distribution for engineering organizations.
One source of truth for AI coding assistant instructions across tools
and repositories.

## Project Status

Go CLI with config parsing, schema validation, local instruction
sync, and dry-run preview implemented. BDD + unit test coverage
across all features.

## Key Documents

- `vision.md` - Problem statement, solution approach, business value
- `scope.md` - MVP scope, key decisions, what's in/out
- `.specify/memory/constitution.md` - Design principles (9 core
  principles, governance rules, prohibited patterns)
- `.specify/memory/soul.md` - Agent culture and working style
- `README.md` - Quick navigation and project overview

## Constitution (Non-Negotiable)

Every design and implementation decision MUST align with these
principles. When in conflict, follow the values hierarchy
(Safety > Trust > Working Software > Testing > Governance > Performance).

1. **CLI-First** - Every feature via CLI. stdin/args -> stdout,
   stderr for errors. JSON + human output. No interactive prompts.
   Exit codes: 0=success, 1=drift, 2=error.
2. **Transparency Over Magic** - All actions explainable via
   `ailign explain`. Deterministic behavior. No silent changes.
3. **Fail Safe, Not Silent** - Validate before modifying. Atomic
   operations. Clear rollback. Never swallow errors.
4. **Test-First** - TDD mandatory. Write tests -> approve -> fail ->
   implement. Coverage: >90% core, 100% renderers, E2E for CLI.
5. **Composition Over Monolith** - Central baseline + repo overlays.
   Modular, independently versioned packages. Independent renderers.
6. **Governance as Foundation** - Immutable versioned packages.
   Traceable changes. Built-in drift detection. Lock files.
7. **Size-Aware** - Respect tool limits (e.g., Cursor 8KB). Content
   tiers: critical/recommended/extra. Never silently truncate.
8. **Cross-Tool Parity** - Tool-agnostic source content. Renderers
   handle tool quirks. No tool is primary. No lock-in.
9. **Working Software** - Commits and PRs MUST include both code
   and tests. Every commit MUST validate. Every PR MUST build.

### Prohibited Patterns

- Silent auto-updates
- Mutable packages
- Unexplainable merges
- Single-tool lock-in
- Interactive prompts in automation
- Hidden failures
- Unbounded content
- Shipping without tests
- Code-only commits without tests
- PRs that do not build

## Build & Development Commands

```bash
# Build
go build ./...

# Run all tests (unit + BDD)
go test ./... -count=1

# Run tests with verbose output and JUnit report
gotestsum --junitfile test-results.xml -- ./... -count=1

# Run tests for a specific package
go test ./internal/sync/... -count=1

# Run only BDD feature tests
go test ./features/steps/... -v -run TestFeatures

# Lint
golangci-lint run

# Run the CLI
go run ./cmd/ailign <command> [flags]
```

CI pipeline is defined in `.github/workflows/validate.yml`
(lint + test + goreleaser check).

## Speckit Workflow

This project uses speckit for specification-driven development.
All features follow this command sequence:

### Command Chain

```
/speckit.constitution  -> Project principles (.specify/memory/constitution.md)
/speckit.specify       -> Feature spec (specs/[###-feature]/spec.md)
/speckit.clarify       -> Resolve ambiguities in spec
/speckit.plan          -> Technical plan (plan.md, research.md, data-model.md, contracts/)
/speckit.tasks         -> Task breakdown (tasks.md)
/speckit.checklist     -> Quality checklists (requirement validation)
/speckit.analyze       -> Cross-artifact consistency check
/speckit.implement     -> Execute tasks with TDD workflow
/speckit.taskstoissues -> Convert tasks to GitHub issues (optional)
```

### When to Use Each Command

- **New feature**: Start with `/speckit.specify`, then follow the
  chain through `/speckit.implement`
- **Unclear requirements**: Use `/speckit.clarify` after specify
- **Before implementation**: Always run `/speckit.analyze` to catch
  inconsistencies between spec, plan, and tasks
- **Quality gate**: `/speckit.checklist` validates requirement
  writing quality, not implementation
- **Constitution changes**: `/speckit.constitution` propagates
  changes to dependent templates

### Feature Directory Structure

```
specs/[###-feature-name]/
├── spec.md          # /speckit.specify output
├── plan.md          # /speckit.plan output
├── research.md      # /speckit.plan output
├── data-model.md    # /speckit.plan output
├── contracts/       # /speckit.plan output
├── quickstart.md    # /speckit.plan output
└── tasks.md         # /speckit.tasks output
```

## Development Workflow

1. **Spec first** - Write feature spec in `specs/[name]/spec.md`
2. **Constitution check** - Verify alignment with 9 principles
3. **Task breakdown** - Create `tasks.md` for complex features
4. **TDD cycle** - Write tests -> get approval -> implement
5. **Integration test** - Verify feature works end-to-end
6. **Documentation** - Update user docs and examples

### Pull Request Size Limits

PRs are enforced by CI to stay small and independently
deployable:

- **Hard limit: 750 lines** (additions + deletions). CI fails.
- **Soft limit: 500 lines**. CI warns.
- **Ideal: <333 lines** (size:m label or smaller).

When planning work (`/speckit.plan`, `/speckit.tasks`), split
tasks into PRs that each deliver an independently releasable
increment. Each PR must:

- Pass all tests on its own (no broken intermediate states)
- Be deployable/releasable without depending on future PRs
- Have a clear, single purpose described by a conventional
  commit title

If a feature exceeds 500 lines, split it across multiple PRs
by user story, layer, or component boundary.

### CI/CD

Pull requests are validated automatically by GitHub Actions:

- **Lint** (`golangci-lint`) and **Test** (`go test ./...`)
  run via a reusable validate workflow
- **PR size check** enforces the limits above
- **Auto-labeling** (srvaroa/labeler) applies type, version,
  and size labels based on changed files and PR title

### Code Review Gates

- Constitution compliance verified
- Complexity justified (YAGNI)
- Breaking changes have migration guide
- Performance impact measured
- All tests pass
- Coverage thresholds met
- CLI help text updated
- Examples work

## Project Architecture

### Core Concept

```
Central Registry (API)
  ├─ company/security@1.3.0
  ├─ company/typescript@2.1.0
  └─ team/platform@0.4.0
       ↓
  CLI: ailign pull
       ↓
  Composition (baseline + overlay)
       ↓
  Rendered formats
  ├─ .claude/instructions.md
  ├─ .cursorrules
  └─ .github/copilot-instructions.md
```

### MVP Commands

- `ailign init` - Bootstrap config + generate initial output files
- `ailign pull` - Fetch packages + render to tool formats
- `ailign status` - Show installed versions + drift detection
- `ailign diff` - Preview changes before update
- `ailign explain` - Show origin of each instruction section

### Config Format

```yaml
# .ailign.yml
packages:
  - company/security@1.3.0
  - company/typescript@2.1.0
  - team/platform@0.4.0
targets:
  - claude
  - cursor
local_overlays:
  - .ai-instructions/project-context.md
```

### Performance Targets

- `ailign status`: <1 second
- `ailign pull`: <30 seconds
- `ailign diff`: <2 seconds
- `ailign explain`: <1 second
- Binary: <50MB, <100MB memory

## Versioning

Semver (MAJOR.MINOR.PATCH). Breaking changes require:
- Migration guide
- Deprecation warnings in previous minor
- 2-week notice minimum

Exception: Security vulnerabilities get immediate fixes.

## File Locations

| Purpose | Path |
|---------|------|
| Constitution | `.specify/memory/constitution.md` |
| Templates | `.specify/templates/` |
| Speckit commands | `.claude/commands/speckit.*.md` |
| Bash helpers | `.specify/scripts/bash/` |
| Feature specs | `specs/[###-feature-name]/` |
| CI workflows | `.github/workflows/` |
| Labeler config | `.github/labeler.yml` |
| Vision | `vision.md` |
| Scope | `scope.md` |

## Active Technologies
- Go 1.24+ (targeting Go 1.26) + Cobra (CLI), goccy/go-yaml (YAML), santhosh-tekuri/jsonschema v6 (validation), testify (TDD), godog (BDD) (001-config-parsing)
- N/A (file system read-only, single `.ailign.yml` file) (001-config-parsing)
- File system — reads overlay files, writes `.ailign/instructions.md`, creates symlinks (002-local-instruction-sync)
- Go 1.24+ (existing), POSIX shell (install script), Node.js (NPM wrapper) + GoReleaser v2.13+ (existing), GitHub Actions (003-install-distribution)
- N/A (distribution only — no data persistence) (003-install-distribution)

## Recent Changes
- 001-config-parsing: Added godog (BDD) for executable Gherkin feature files alongside existing TDD
- 001-config-parsing: Added Go 1.24+ (targeting Go 1.26) + Cobra (CLI), goccy/go-yaml (YAML), santhosh-tekuri/jsonschema v6 (validation), testify (testing)
