# Quickstart: Package Manifest & Install

**Feature**: 004-package-install | **Date**: 2026-02-20

## Scenario 1: Install a Single Package

A developer has a `.ailign.yml` referencing one instruction package.

### Setup
```yaml
# .ailign.yml
packages:
  - instructions/company/security@1.3.0
targets:
  - claude
  - cursor
local_overlays:
  - .ai-instructions/project-context.md
```

### Run
```bash
ailign install
```

### Expected Output (human)
```
Installing packages...
  instructions/company/security@1.3.0   fetched

Composing instructions...
  .ailign/instructions.md   written

Linking targets...
  claude    .claude/instructions.md        created
  cursor    .cursorrules                   created

Lock file: ailign-lock.yml   created

Installed 1 package, linked 2 targets.
```

### Expected Output (JSON)
```bash
ailign install --format json
```
```json
{
  "packages": [
    {
      "reference": "instructions/company/security",
      "version": "1.3.0",
      "status": "fetched"
    }
  ],
  "hub": {
    "path": ".ailign/instructions.md",
    "status": "written"
  },
  "links": [
    {"target": "claude", "link_path": ".claude/instructions.md", "status": "created"},
    {"target": "cursor", "link_path": ".cursorrules", "status": "created"}
  ],
  "lock": {
    "path": "ailign-lock.yml",
    "status": "created"
  },
  "summary": {
    "packages_installed": 1,
    "targets_linked": 2
  }
}
```

## Scenario 2: Install Multiple Packages with Overlays

### Setup
```yaml
# .ailign.yml
packages:
  - instructions/company/security@1.3.0
  - instructions/company/typescript@2.1.0
targets:
  - claude
  - cursor
  - copilot
local_overlays:
  - .ai-instructions/project-context.md
```

### Run
```bash
ailign install
```

### Expected Output
```
Installing packages...
  instructions/company/security@1.3.0      fetched
  instructions/company/typescript@2.1.0     fetched

Composing instructions...
  .ailign/instructions.md   written

Linking targets...
  claude    .claude/instructions.md                created
  cursor    .cursorrules                           created
  copilot   .github/copilot-instructions.md        created

Lock file: ailign-lock.yml   created

Installed 2 packages, linked 3 targets.
```

## Scenario 3: Idempotent Re-install

Running install a second time with no changes produces unchanged status.

```bash
ailign install
```
```
Installing packages...
  instructions/company/security@1.3.0      cached
  instructions/company/typescript@2.1.0     cached

Composing instructions...
  .ailign/instructions.md   unchanged

Linking targets...
  claude    .claude/instructions.md                exists
  cursor    .cursorrules                           exists
  copilot   .github/copilot-instructions.md        exists

Lock file: ailign-lock.yml   unchanged

Installed 2 packages (0 fetched, 2 cached), linked 3 targets.
```

## Scenario 4: Invalid Package Reference

```yaml
# .ailign.yml
packages:
  - company/security@1.3.0  # missing type prefix
```

```bash
ailign install
```
```
Error: invalid package reference "company/security@1.3.0": missing type prefix

Package references must use the format: <type>/<scope>/<name>@<version>
Example: instructions/company/security@1.3.0

Supported types: instructions

Exit code: 2
```

## Scenario 5: Unsupported Type

```yaml
# .ailign.yml
packages:
  - mcp/company/tools@1.0.0
```

```bash
ailign install
```
```
Error: unsupported package type "mcp" in "mcp/company/tools@1.0.0"

Supported types: instructions

Other types (mcp, commands, agents, packages) are reserved for future releases.

Exit code: 2
```

## Scenario 6: Package Not Found

```bash
ailign install
```
```
Error: package not found: instructions/company/nonexistent@1.0.0

The registry does not have a package matching this reference.
Check the package name and version, or contact the package author.

Exit code: 2
```

## Scenario 7: Checksum Mismatch

```bash
ailign install
```
```
Error: integrity check failed for instructions/company/security@1.3.0

Expected: sha256-47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU=
Got:      sha256-abc123def456...

The package content has changed since it was locked. This may indicate
a registry issue or tampering. Delete ailign-lock.yml and re-run
`ailign install` to fetch fresh content.

Exit code: 2
```

## Scenario 8: Package Manifest Validation

### Valid manifest
```yaml
# ailign-pkg.yml
name: "company/security"
type: "instructions"
version: "1.3.0"
description: "Company-wide security instructions for AI coding assistants"
content:
  main: "instructions.md"
```

### Invalid manifest (missing field)
```
Error: invalid package manifest for instructions/company/security@1.3.0

  Field "description" is required but missing.

  A valid ailign-pkg.yml must contain:
    name:         Scoped name (e.g., "company/security")
    type:         Content type (e.g., "instructions")
    version:      Semver version (e.g., "1.3.0")
    description:  Human-readable description
    content.main: Path to primary content file

Exit code: 2
```
