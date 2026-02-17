# Learnings

Project-specific conventions and patterns discovered through code
reviews and development. Curated — not a changelog.

## Documentation

- Use generic/fictional examples in documentation (e.g.,
  `acme/web-api#123`) instead of real repository names, even
  internal ones. Avoids implying affiliation, leaking private repo
  names, and confusing access controls.

## Go Error Handling

- **File errors**: Always `errors.Is(err, os.ErrNotExist)` — never treat all errors as "file missing". Permission denied, EIO, etc. must propagate.
- **Collect all errors**: Multi-item validation MUST `errors.Join`, not early-return on first. Constitution principle.
- **Error context must match operation**: Dry-run paths doing read-only checks should not say "writing" in error wraps.

## Go Path Handling

- **Path traversal**: `strings.HasPrefix(rel, "..")` rejects `"..notes.md"`. Use `rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator))`.
- **Symlink escapes**: Lexical validation insufficient. After `filepath.Clean`/`Rel`, also `filepath.EvalSymlinks` and verify resolved path stays within base.

## Go Serialization

- **nil vs empty slice**: Use an empty slice (`make([]T, 0)`) instead of `nil` when you want JSON arrays to serialize as `[]` rather than `null`.
- **CreateTemp permissions**: `os.CreateTemp` uses 0600 — `os.Chmod` to 0644 before rename if file should be world-readable.

## Go Testing

- macOS `/var` → `/private/var`: use `filepath.EvalSymlinks(t.TempDir())` via `resolveDir` helper for path comparisons.

## CI/CD

- **Workflow permissions**: An explicit least-privilege `permissions:` block is required on all GitHub Actions workflows. Document permission rationale alongside each workflow in `.github/workflows/`.

## PR & Commit Workflow

- **Target ~400 lines** per PR (soft limit ~500, reserve 15-20% for review fixes).
- **BDD steps are expensive**: A single `*_steps_test.go` can be 300+ lines. Split into own PR when pushing over limit.
- **Test multiplier**: Estimate 1.5-2x implementation lines for Go tests.
- **Pre-commit stash**: Stashes unstaged but NOT untracked files. Stage all interdependent files together or linter fails on undefined symbols.
- When a PR accumulates features beyond the original description, update the PR body to list all changes. Reviewers (human and bot) judge scope against the description.

## BDD Architecture

- Shared steps (`itReportsErrorContaining`, `itExitsWithCode`) live in `config_parsing_steps_test.go`, reused across features.
- Error checking: `w.stderr` first (CLI), fall back to `w.loadErr` (direct API).
- After hook restores dir permissions before `os.RemoveAll` (read-only dir tests).

## PR hygiene

- When a PR accumulates features beyond the original description,
  update the PR body to list all changes. Reviewers (human and bot)
  judge scope against the description.
