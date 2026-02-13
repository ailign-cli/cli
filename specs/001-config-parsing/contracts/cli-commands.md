# CLI Command Contracts: Configuration File Parsing

**Feature Branch**: `001-config-parsing`
**Date**: 2026-02-13

## Commands

### `ailign validate`

Validates the `.ailign.yml` configuration file in the current
working directory against the schema. Reports all errors and
warnings. Does not trigger any other operations.

**Usage**:
```
ailign validate [flags]
```

**Flags**:

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| --format | -f | string | human | Output format: `human` or `json` |

**Exit Codes**:

| Code | Meaning |
|------|---------|
| 0 | Config is valid (may have warnings) |
| 2 | Config is invalid or file not found |

**Stdout** (on success, human format):
```
.ailign.yml: valid
```

**Stdout** (on success, JSON format):
```json
{
  "valid": true,
  "errors": [],
  "warnings": [],
  "file": ".ailign.yml"
}
```

**Stderr** (on validation errors, human format):
```
Error: .ailign.yml validation failed

  targets: required field missing
    Expected: array of target names (claude, cursor, copilot, windsurf)
    Fix: Add a "targets" field with at least one target

1 error found
```

**Stderr** (on validation errors, JSON format):
```json
{
  "valid": false,
  "errors": [
    {
      "field_path": "targets",
      "expected": "array of target names",
      "actual": null,
      "message": "required field missing",
      "remediation": "Add a \"targets\" field with at least one target"
    }
  ],
  "warnings": [],
  "file": ".ailign.yml"
}
```

**Stderr** (warnings, human format):
```
Warning: .ailign.yml has warnings

  custom_field: unrecognized field
    This field is not part of the AIlign config schema
    Fix: Remove it or check for typos

.ailign.yml: valid (1 warning)
```

**Stderr** (file not found, human format):
```
Error: .ailign.yml not found in current directory

  No AIlign configuration file exists at: /path/to/cwd/.ailign.yml

Run "ailign init" to create a configuration file.
```

---

### `ailign` (root command / implicit validation)

All subcommands (future: pull, status, diff, explain) load and
validate the config as a prerequisite. On validation failure, the
command does not proceed.

**Behavior**:
1. Look for `.ailign.yml` in current working directory
2. Parse YAML
3. Validate against schema
4. If valid: proceed with subcommand
5. If invalid: print errors to stderr, exit with code 2

**Flags inherited by all subcommands**:

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| --format | -f | string | human | Output format: `human` or `json` |

---

## Error Message Contract

All validation errors follow this structure:

**Human format** (one per error):
```
  {field_path}: {message}
    Expected: {expected}
    Found: {actual}
    Fix: {remediation}
```

**JSON format** (per error object):
```json
{
  "field_path": "string",
  "expected": "string",
  "actual": "string | null",
  "message": "string",
  "remediation": "string"
}
```

**Rules**:
- `field_path` uses dot notation and array indices:
  `targets`, `targets[0]`
- `actual` is null when the field is missing (not present)
- `remediation` always includes a concrete action the developer
  can take
- Errors go to stderr, success confirmation goes to stdout
