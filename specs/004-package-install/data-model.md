# Data Model: Package Manifest & Install

**Feature**: 004-package-install | **Date**: 2026-02-20

## Entities

### PackageRef

A parsed package reference from `.ailign.yml`.

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| Type | string | Content type prefix | Must be `"instructions"` (others reserved) |
| Scope | string | Organization/team scope | Non-empty, lowercase alphanumeric + hyphens |
| Name | string | Package name within scope | Non-empty, lowercase alphanumeric + hyphens |
| Version | string | Exact semver version | Must match `MAJOR.MINOR.PATCH` pattern |

**String format**: `{type}/{scope}/{name}@{version}` (e.g., `instructions/company/security@1.3.0`)

**Parsing rules**:
- Split on first `@` to separate reference from version
- Split reference on `/` — must have exactly 3 segments (type, scope, name)
- Missing type prefix is an error (not inferred)
- Unsupported type prefix is an error listing supported types

### Manifest

The parsed `ailign-pkg.yml` file from a package.

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| Name | string | Scoped name (`scope/name`) | Must match `{scope}/{name}` pattern |
| Type | string | Content type | Must be `"instructions"` |
| Version | string | Semver version | Must match `MAJOR.MINOR.PATCH` |
| Description | string | Human-readable description | Non-empty |
| Content.Main | string | Path to primary content file | Non-empty, relative path |

**Type consistency rule**: The type in the manifest must match the type prefix in the package reference. `instructions/company/security@1.3.0` must have `type: instructions` in its manifest.

### LockedPackage

An entry in `ailign-lock.yml` recording an installed package.

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| Reference | string | Type-scoped reference (without version) | `{type}/{scope}/{name}` |
| Version | string | Exact resolved version | Semver |
| Resolved | string | Full URL fetched | Valid HTTPS URL |
| Integrity | string | Content checksum | `sha256-{base64}` format |

### LockFile

The top-level `ailign-lock.yml` structure.

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| LockfileVersion | int | Schema version | Must be `1` |
| Packages | []LockedPackage | Installed packages | Sorted by Reference ascending |

### InstallResult

The result of an `ailign install` operation.

| Field | Type | Description |
|-------|------|-------------|
| Packages | []InstalledPackage | Packages that were processed |
| HubPath | string | Path to composed instructions file |
| HubStatus | string | `"written"` or `"unchanged"` |
| Links | []LinkResult | Symlink results per target (reused from sync) |
| LockPath | string | Path to lock file |
| LockStatus | string | `"created"`, `"updated"`, or `"unchanged"` |
| Warnings | []string | Non-fatal warnings |

### InstalledPackage

Status of a single package in the install result.

| Field | Type | Description |
|-------|------|-------------|
| Reference | string | Type-scoped reference |
| Version | string | Installed version |
| Status | string | `"fetched"`, `"cached"`, or `"unchanged"` |

## Relationships

```text
.ailign.yml
  └── packages: []PackageRef ──┐
                                │ fetch each
                                ▼
                          Registry API
                                │
                                ▼
                          Manifest (ailign-pkg.yml)
                                │ validate type match
                                │ read content.main
                                ▼
                          Content (instructions.md)
                                │
                    ┌───────────┴───────────┐
                    ▼                       ▼
              Lock File               Composition
          (ailign-lock.yml)      (content + overlays)
                                        │
                                        ▼
                                   Hub File
                              (.ailign/instructions.md)
                                        │
                                        ▼
                                    Symlinks
                              (per target format)
```

## State Transitions

### Install Workflow States

```text
START ──► Parse Config ──► Validate PackageRefs ──► Fetch Manifests
                                                         │
                    ┌──── Type Mismatch Error ◄──────────┤
                    │                                     │
                    │                             Fetch Content
                    │                                     │
                    │              Checksum Mismatch ◄────┤
                    │                                     │
                    │                             Compose Content
                    │                                     │
                    │                             Write Hub File
                    │                                     │
                    │                             Create Symlinks
                    │                                     │
                    │                             Write Lock File
                    │                                     │
                    ▼                                     ▼
                  ERROR (exit 2)                    SUCCESS (exit 0)
```

### Lock File States

| Current State | Event | Next State |
|--------------|-------|------------|
| Does not exist | `ailign install` succeeds | Created |
| Exists, versions match | `ailign install` | Unchanged (verify checksums) |
| Exists, config version changed | `ailign install` | Updated (re-fetch changed package) |
| Exists, checksum mismatch | `ailign install` | Error — manual intervention required |

## Error Types

| Error | Trigger | Exit Code | Remediation |
|-------|---------|-----------|-------------|
| `ErrPackageNotFound` | Registry returns 404 | 2 | Check package name and version |
| `ErrRegistryUnreachable` | Network/connection failure | 2 | Check network, retry |
| `ErrInvalidManifest` | Manifest missing required fields | 2 | Fix package manifest |
| `ErrTypeMismatch` | Config type ≠ manifest type | 2 | Fix package reference or manifest |
| `ErrUnsupportedType` | Type not `instructions` | 2 | List supported types |
| `ErrMissingTypePrefix` | No type prefix in reference | 2 | Explain required format |
| `ErrChecksumMismatch` | Lock integrity ≠ fetched content | 2 | Investigate, re-install |
| `ErrDuplicatePackage` | Same package declared twice | 2 | Remove duplicate |
| `ErrInvalidVersion` | Version not valid semver | 2 | Use MAJOR.MINOR.PATCH format |
