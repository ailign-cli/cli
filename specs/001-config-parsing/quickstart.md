# Quickstart: Configuration File Parsing

**Feature Branch**: `001-config-parsing`
**Date**: 2026-02-13

## Prerequisites

- Go 1.24 or later installed
- Git repository initialized

## Setup

```bash
# Clone and enter the repository
git clone <repo-url> && cd ailign

# Install dependencies
go mod download

# Build the CLI
go build -o ailign ./cmd/ailign

# Verify it works
./ailign --help
```

## Create a Configuration File

Create `.ailign.yml` in your repository root:

```yaml
# Which AI tools to render instructions for (required)
targets:
  - claude
  - cursor
```

## Validate Your Configuration

```bash
# Validate the config file
./ailign validate

# Output on success:
# .ailign.yml: valid

# Validate with JSON output (for CI/CD)
./ailign validate --format json
```

## Common Validation Errors

**Missing targets field**:
```
Error: .ailign.yml validation failed

  targets: required field missing
    Expected: array of target names (claude, cursor, copilot, windsurf)
    Fix: Add a "targets" field with at least one target
```

**Invalid target name**:
```
Error: .ailign.yml validation failed

  targets[0]: invalid target name
    Expected: one of claude, cursor, copilot, windsurf
    Found: "vscode"
    Fix: Use a supported target name
```

**Unknown field** (warning, not error):
```
Warning: .ailign.yml has warnings

  custom_field: unrecognized field
    This field is not part of the AIlign config schema
    Fix: Remove it or check for typos

.ailign.yml: valid (1 warning)
```

## Run Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out

# Run only config package tests
go test ./internal/config/...

# Run only CLI command tests
go test ./internal/cli/...
```

## Project Structure

```
cmd/ailign/main.go              # Entry point
internal/
├── config/                     # Config loading + validation
│   ├── config.go               # Config types
│   ├── loader.go               # File loading (YAML parse)
│   ├── validator.go            # JSONSchema validation
│   ├── errors.go               # Error types + formatting
│   └── schema.json             # Embedded JSONSchema
├── cli/                        # Cobra commands
│   ├── root.go                 # Root command + global flags
│   └── validate.go             # ailign validate command
├── output/                     # Output formatting
│   ├── formatter.go            # Formatter interface
│   ├── human.go                # Human-readable output
│   └── json.go                 # JSON output
└── target/                     # Target registry
    └── registry.go             # Known target names + interface
```
