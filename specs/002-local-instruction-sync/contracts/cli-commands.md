# CLI Command Contracts: Local Instruction Sync

**Feature**: 002-local-instruction-sync
**Date**: 2026-02-17

## ailign sync

### Synopsis

```
ailign sync [--dry-run] [--format human|json]
```

### Description

Composes local overlay files and syncs the result to all configured
targets. Writes composed content to `.ailign/instructions.md` and
creates symlinks from each target's instruction path to the hub file.

### Flags

| Flag        | Short | Default | Description                        |
|-------------|-------|---------|------------------------------------|
| `--dry-run` | `-n`  | false   | Preview changes without modifying files |
| `--format`  | `-f`  | human   | Output format: `human` or `json`   |

### Behavior

1. Load and validate `.ailign.yml`
2. Validate `local_overlays` is present and non-empty
3. Read each overlay file in order
4. Compose overlay content (concatenate with newline separator)
5. Prepend managed-content header to composed content
6. Write to `.ailign/instructions.md` (atomic: write-temp-rename)
7. For each configured target:
   a. Get target from registry
   b. Ensure symlink exists from `target.InstructionPath()` to
      `.ailign/instructions.md` (create directories as needed)
   c. Record result (created/exists/error)
8. Report results to stdout

### Exit Codes

| Code | Meaning                                              |
|------|------------------------------------------------------|
| 0    | All targets synced successfully                      |
| 2    | Error (missing overlays, invalid config, write fail) |

### Output — Human Format

```
Syncing instructions to 3 targets...

  .ailign/instructions.md              written (1.2 KB)
  .claude/instructions.md              symlink created
  .cursorrules                         symlink created
  .github/copilot-instructions.md      symlink created

Synced 3 targets from 2 overlays.
```

Subsequent run (content unchanged):
```
Syncing instructions to 3 targets...

  .ailign/instructions.md              unchanged
  .claude/instructions.md              symlink ok
  .cursorrules                         symlink ok
  .github/copilot-instructions.md      symlink ok

All 3 targets up to date.
```

### Output — JSON Format

```json
{
  "dry_run": false,
  "hub": {
    "path": ".ailign/instructions.md",
    "status": "written"
  },
  "links": [
    {
      "target": "claude",
      "link_path": ".claude/instructions.md",
      "status": "created"
    },
    {
      "target": "cursor",
      "link_path": ".cursorrules",
      "status": "created"
    }
  ],
  "summary": {
    "total": 2,
    "created": 2,
    "existing": 0,
    "errors": 0
  }
}
```

### Error Output (stderr)

Missing overlay:
```
Error: overlay file not found: .ai-instructions/missing.md

No files were modified. Add the file or remove it from local_overlays
in .ailign.yml.
```

No overlays configured:
```
Error: no local_overlays configured in .ailign.yml

Add a local_overlays section to your config:

  local_overlays:
    - .ai-instructions/base.md
```

Path traversal:
```
Error: overlay path escapes repository root: ../../secrets/instructions.md

Overlay paths must be relative to the repository root and cannot
traverse above it.
```

### Dry-Run Output — Human Format

```
Dry run — no files will be modified.

  .ailign/instructions.md              would be written (1.2 KB)
  .claude/instructions.md              would create symlink
  .cursorrules                         would create symlink

Would sync 2 targets from 1 overlay.
```

### Dry-Run Output — JSON Format

Same structure as normal output with `"dry_run": true`.

## ailign validate (updated)

### Changes from Feature 001

The validate command now also validates `local_overlays` paths in
the schema (format only — existence is NOT checked by validate).

No changes to the command interface. The JSONSchema is extended to
include `local_overlays` as an optional field.
