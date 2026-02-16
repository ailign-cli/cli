# Research: Configuration File Parsing

**Feature Branch**: `001-config-parsing`
**Date**: 2026-02-13

## Decision 1: Programming Language

**Decision**: Go 1.24+ (minimum), targeting Go 1.26
**Rationale**: User requirement. Go produces single static binaries
with zero runtime dependencies (aligns with Constitution I: CLI-First).
Cross-compilation built-in for all target platforms.
**Alternatives considered**: None (explicit user requirement).

## Decision 2: CLI Framework

**Decision**: Cobra (`github.com/spf13/cobra`)
**Rationale**: De facto standard for Go CLI applications. Used by
Kubernetes, Docker, GitHub CLI, Hugo, Helm. Provides nested
subcommands, auto-generated help text, shell completions, and Viper
integration for config. Largest ecosystem and community support.
**Alternatives considered**:
- Kong (`github.com/alecthomas/kong`): Struct-based, less
  boilerplate, but smaller community. Good alternative if team
  prefers declarative style.
- urfave/cli (`github.com/urfave/cli/v2`): Solid but no clear
  advantage over Cobra or Kong.

## Decision 3: YAML Parsing

**Decision**: goccy/go-yaml (`github.com/goccy/go-yaml`)
**Rationale**: The previous standard `gopkg.in/yaml.v3` has been
archived by its maintainer. goccy/go-yaml is the recommended
successor with better YAML spec compliance (355/402 test suite cases
vs 295 for yaml.v3). Major projects (GitHub CLI, Viper, Prometheus)
are migrating to it. Drop-in replacement API.
**Alternatives considered**:
- gopkg.in/yaml.v3: Archived/unmaintained. Not suitable for new
  projects.

## Decision 4: JSONSchema Validation

**Decision**: santhosh-tekuri/jsonschema v6
(`github.com/santhosh-tekuri/jsonschema/v6`)
**Rationale**: Supports Draft 2020-12 (latest), pure validation
focus, zero external dependencies, well-maintained (latest release
May 2025). User explicitly requested JSONSchema for schema definition
and validation.
**Alternatives considered**:
- xeipuuv/gojsonschema: No Draft 2020-12 support, maintenance-only
  mode. Not recommended for new projects.
- google/jsonschema-go: Newer (Jan 2026), Google-backed, includes
  schema construction + validation. Consider if we need schema
  generation from Go types later. For pure validation,
  santhosh-tekuri is simpler and more focused.

## Decision 5: Testing Framework

**Decision**: Standard `testing` package + testify
(`github.com/stretchr/testify`)
**Rationale**: testify is the de facto standard assertion/mocking
library (27% of Go developers). Provides `assert`, `require`, and
`mock` packages. Table-driven tests using stdlib `testing` as
foundation.
**Alternatives considered**:
- Stdlib only: Viable but verbose for assertion-heavy tests.
- Ginkgo/Gomega: BDD-style, unnecessary complexity for this project.

## Decision 6: Config Schema Approach

**Decision**: YAML config file validated via embedded JSONSchema.
Two-pass validation approach.
**Rationale**: User specified "schema can either be YAML or JSON,
validation and definition through JSONSchema." The config file is
YAML (developer-friendly), but the schema definition is JSONSchema
(industry standard, tooling ecosystem).

**Validation strategy**:
1. Parse YAML into Go structs (goccy/go-yaml)
2. Marshal Go structs to JSON in-memory
3. Validate JSON against embedded JSONSchema (santhosh-tekuri)
4. Separate pass: detect unknown fields and emit warnings

This approach allows:
- YAML for human authoring (readable, comments supported)
- JSONSchema for validation (standard, shareable, tooling support)
- Unknown field detection as warnings (not errors per spec FR-006)

## Decision 7: Error Formatting

**Decision**: Custom error transformer that converts JSONSchema
validation errors into user-friendly messages with field path,
expected value, actual value, and remediation guidance.
**Rationale**: Raw JSONSchema errors are technical and lack
remediation guidance. Constitution II (Transparency) requires errors
that explain what went wrong AND how to fix it. Dual output format
(JSON + human-readable) per FR-014.

## Decision 8: Target Interface Design

**Decision**: Define a `Target` interface in Go with a registry of
known target names. Implementation per target is out of scope.
**Rationale**: User specified "instructions for each of the targets
need to be clearly separated from each other, but the interface will
be the same for them all." This feature only validates target names
against a known set. Future features will implement the interface.

## Decision 9: Project Structure

**Decision**: Standard Go project layout with `cmd/` and `internal/`.

```
cmd/ailign/main.go              # Entry point (thin)
internal/
├── config/                     # Config loading + validation
│   ├── config.go               # Config types
│   ├── loader.go               # File loading (YAML parse)
│   ├── validator.go            # JSONSchema validation
│   ├── errors.go               # Error types + formatting
│   └── schema.json             # Embedded JSONSchema (via go:embed)
├── cli/                        # Cobra commands
│   ├── root.go                 # Root command + global flags
│   └── validate.go             # `ailign validate` command
├── output/                     # Output formatting
│   ├── formatter.go            # Formatter interface
│   ├── human.go                # Human-readable output
│   └── json.go                 # JSON output
└── target/                     # Target registry (stub)
    └── registry.go             # Known target names + interface
```

**Rationale**: `internal/` enforces encapsulation (Go compiler
prevents external imports). `cmd/` keeps entry point thin. Each
package has a single responsibility. Aligns with Constitution V
(Composition Over Monolith).

## Dependency Summary

| Dependency | Import Path | Purpose |
|------------|-------------|---------|
| Cobra | `github.com/spf13/cobra` | CLI framework |
| goccy/go-yaml | `github.com/goccy/go-yaml` | YAML parsing |
| jsonschema v6 | `github.com/santhosh-tekuri/jsonschema/v6` | JSONSchema validation |
| testify | `github.com/stretchr/testify` | Test assertions |

Total: 4 external dependencies (minimal, aligns with single binary
goal).
