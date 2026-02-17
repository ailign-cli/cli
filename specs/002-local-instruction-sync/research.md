# Research: Local Instruction Sync

**Feature**: 002-local-instruction-sync
**Date**: 2026-02-17

## R1: Sync Architecture — Hub-Spoke vs Discover-Everywhere

### Decision: Hub-Spoke with Central Storage

Use a central `.ailign/rendered/` directory as the hub. Each target
gets a rendered file in the hub. Target-specific paths (e.g.,
`.cursorrules`) become symlinks to the hub files.

### Rationale

- **Clear single source**: All rendered content lives in one place
- **Registry-ready**: When remote packages are added, the hub is the
  natural download destination
- **Symlink benefit**: Changes to hub files are immediately visible
  at all target paths
- **Familiar pattern**: Similar to package managers (npm, composer)
  that install to a central location

### Alternatives Considered

**Discover-Everywhere**: Scan the filesystem for existing instruction
files and propagate changes. Rejected because:
- Requires tracking file locations (more state)
- Harder to determine source of truth
- Doesn't scale to registry-based packages
- Ambiguous: which file is the source?

**Direct Write (no hub)**: Compose overlays and write directly to
each target path. Rejected because:
- Duplicates content across target files
- Harder to manage changes (edit each file?)
- No centralized storage for future registry integration
- Violates the user's stated preference for centralized storage

## R2: Sync Mechanism — Symlinks vs File Copies

### Decision: Relative Symlinks (with future --copy fallback)

Create relative symlinks from target paths to hub files. Relative
paths ensure portability when the repository is moved or cloned.

### Rationale

- **Single write**: Only the hub file changes; symlinks auto-update
- **Transparency**: `ls -la` shows exactly where content comes from
- **Portability**: Relative symlinks survive repo relocation
- **Precedent**: GNU Stow, chezmoi, and Nix all use symlinks
- **Git support**: Git stores symlinks as text files; cloning
  recreates them correctly on POSIX systems

### Hub Structure

```
.ailign/
└── instructions.md     # Single hub file for all targets
```

All targets symlink to this one file. Per-target content
differentiation is a future enhancement; when needed, the hub
can evolve to per-target files.

### Symlink Mappings (relative paths)

| Target Path                        | Symlink Target                  |
|------------------------------------|---------------------------------|
| `.claude/instructions.md`          | `../.ailign/instructions.md`    |
| `.cursorrules`                     | `.ailign/instructions.md`       |
| `.github/copilot-instructions.md`  | `../../.ailign/instructions.md` |
| `.windsurfrules`                   | `.ailign/instructions.md`       |

### Cross-Platform Considerations

- **macOS/Linux**: Symlinks work without restrictions
- **Windows**: Requires Developer Mode (Win 10 1703+) or admin
  privileges. Go's `os.Symlink()` handles the OS call.
- **Windows**: Out of scope for this feature. Git on Windows
  defaults to `core.symlinks=false`, which checks out symlinks as
  plain text files containing the target path. Windows users need
  `core.symlinks=true` and Developer Mode or admin privileges. This
  is a significant friction point and will be addressed as a
  separate feature (likely with a `--copy` fallback mode).

### Alternatives Considered

**File Copies**: Write full content to each target path. Rejected
because:
- Requires hash comparison to detect changes
- More I/O (write N files vs 1 hub file + N symlinks)
- Doesn't centralize storage as user requested

**Hard Links**: Share inodes between hub and target files. Rejected
because:
- Don't work across filesystems
- Can't tell a hard link from a regular file (confusing)
- Not supported on all platforms

## R3: Target Modularity — Provider-Style Architecture

### Decision: Interface + Registry Pattern

Each target implements a `Target` interface. A registry maps target
names to implementations. New targets are added by implementing the
interface and registering in the default registry constructor.

### Rationale

- **Clear boundaries**: Each target is self-contained
- **Easy extension**: Add a target = 1 new file + 1 registry line
- **Testable**: Mock targets for sync logic tests
- **Terraform-inspired**: Familiar pattern for infrastructure tooling
- **Future-ready**: Can evolve toward plugin-based loading

### Target Interface

```go
type Target interface {
    Name() string
    InstructionPath() string
}
```

- `Name()`: Target identifier (e.g., "claude")
- `InstructionPath()`: Relative path to the target's instruction
  file (e.g., ".claude/instructions.md")

Since all targets symlink to the same hub file, there is no
per-target rendering. Content (including the managed-content
header) is written once to the hub file. When per-target content
differentiation is needed, a `Render()` method can be added.

### Registration

Explicit registration in a constructor (not `init()` magic):

```go
func NewDefaultRegistry() *Registry {
    r := NewRegistry()
    r.Register(&Claude{})
    r.Register(&Cursor{})
    r.Register(&Copilot{})
    r.Register(&Windsurf{})
    return r
}
```

### Alternatives Considered

**init()-based auto-registration**: Rejected because:
- Import side effects make testing harder
- Harder to create registries with specific target subsets
- Less explicit

**Plugin binaries (like Terraform)**: Rejected because:
- Massive over-engineering for 4 built-in targets
- Adds RPC complexity and binary distribution overhead
- No current need for third-party targets

## R4: Atomic Write Strategy

### Decision: Write-Temp-Rename Pattern (No External Dependency)

Write hub files to temporary files in the same directory, then
atomically rename. No external package needed.

### Rationale

- **Simple**: ~15 lines of Go code
- **No new dependency**: Existing stdlib is sufficient
- **POSIX atomic**: `os.Rename()` is atomic on Linux/macOS
- **Windows**: Not fully atomic, but acceptable for a developer tool
  writing small config files
- **Symlinks auto-update**: Since symlinks point to the hub file
  path (not inode), the renamed file is immediately visible

### Implementation Pattern

```go
func atomicWrite(path string, content []byte) error {
    dir := filepath.Dir(path)
    tmp, err := os.CreateTemp(dir, ".ailign-*.tmp")
    if err != nil { return err }
    defer os.Remove(tmp.Name()) // cleanup on error
    if _, err := tmp.Write(content); err != nil {
        tmp.Close()
        return err
    }
    if err := tmp.Sync(); err != nil {
        tmp.Close()
        return err
    }
    if err := tmp.Close(); err != nil { return err }
    return os.Rename(tmp.Name(), path)
}
```

### Cross-Target Atomicity

Full cross-target atomicity (all-or-nothing across all targets) is
not implemented. Each hub file is written atomically individually.
If a failure occurs mid-sync, some targets may have updated content
while others retain old content. Re-running `ailign sync` resolves
this. The risk is extremely low for small file operations.

### Alternatives Considered

**google/renameio**: Provides cross-platform atomic writes and
atomic symlink replacement. Rejected because:
- Adds a dependency for ~15 lines of code
- Our use case is simple (small files, same filesystem)
- Can revisit if Windows atomicity becomes an issue

**Directory swap**: Write all hub files to `.ailign/rendered.new/`,
then swap directories. Rejected because:
- Atomic directory swap is not possible on most OSes
- Over-complex for the problem at hand
