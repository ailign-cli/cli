# Data Model: Configuration File Parsing

**Feature Branch**: `001-config-parsing`
**Date**: 2026-02-13

## Entities

### Config

The root configuration object parsed from `.ailign.yml`.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| targets | []TargetName | Yes | AI tools to render for (min 1 item) |

**Validation rules**:
- `targets` MUST be present and contain at least one item
- Unknown top-level fields produce warnings (not errors)

**Future fields** (will be added by subsequent features):
- `packages` ([]PackageRef) - versioned instruction packages
- `local_overlays` ([]string) - relative paths to overlay files

### TargetName

A known AI tool identifier.

| Value | Tool | Output File |
|-------|------|-------------|
| `claude` | Claude Code | `.claude/instructions.md` |
| `cursor` | Cursor | `.cursorrules` |
| `copilot` | GitHub Copilot | `.github/copilot-instructions.md` |
| `windsurf` | Windsurf | `.windsurfrules` |

**Validation rules**:
- MUST be one of the known values listed above
- Case-sensitive (lowercase only)
- Duplicates are not allowed (`uniqueItems: true`)
- The set of known targets is defined in a registry, making it easy
  to add new targets without changing validation logic

**Note**: The output file paths are documented here for context but
are NOT used by this feature. Rendering is a separate feature.

### ValidationError

An error produced during schema validation.

| Field | Type | Description |
|-------|------|-------------|
| field_path | string | Dot-notation path (e.g., `targets`, `targets[0]`) |
| expected | string | What the schema expects |
| actual | string | What was found |
| message | string | Human-readable error description |
| remediation | string | How to fix the issue |
| severity | string | `error` or `warning` |

**Behavior**:
- All errors are collected before reporting (no early exit)
- Unknown fields produce `warning` severity
- All other violations produce `error` severity
- Presence of any `error` severity item causes exit code 2

### ValidationResult

The outcome of validating a config file.

| Field | Type | Description |
|-------|------|-------------|
| valid | bool | Whether config passed validation |
| errors | []ValidationError | List of errors (severity=error) |
| warnings | []ValidationError | List of warnings (severity=warning) |
| config | *Config | Parsed config (nil if invalid) |

## Relationships

```
Config
└── has many TargetName (via targets field)

ValidationResult
├── references Config (if valid)
└── has many ValidationError (errors + warnings)
```

## State Transitions

Config loading follows a linear pipeline with no branching state:

```
File not found ──→ Error (exit 2)
File found ──→ YAML parse
  ├── YAML parse error ──→ Error (exit 2)
  └── YAML parse OK ──→ Schema validation
      ├── Validation errors ──→ Report all + Error (exit 2)
      ├── Validation warnings only ──→ Report warnings + Continue
      └── Validation clean ──→ Continue (config available)
```

No persistent state. Config is read-only and loaded fresh on every
CLI invocation.

## JSONSchema Definition

The schema is defined in `internal/config/schema.json` and embedded
into the binary via `go:embed`. See `contracts/config-schema.json`
for the full schema.
