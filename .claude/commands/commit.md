---
description: Analyze all uncommitted changes, group them into logical commits, and create each commit separately using conventional commit format.
---

## User Input

```text
$ARGUMENTS
```

Consider the user input for any additional instructions (e.g., "only staged", "single commit", "include X in the message").

## Workflow

### Step 1: Assess the current state

Run these commands **in parallel** to understand what needs to be committed:

1. `git status` — all modified, added, and untracked files (never use `-uall`)
2. `git diff` — unstaged changes
3. `git diff --cached` — already-staged changes
4. `git log --oneline -10` — recent commits for context

If there are no changes (nothing modified, nothing untracked), stop and report: "Nothing to commit."

### Step 2: Group changes logically

Analyze all changed files and their diffs. Group them by **logical unit of work** — a set of changes that together accomplish one coherent purpose.

**Grouping principles:**

- **Feature code + its tests** belong in the same commit (e.g., `validator.go` + `validator_test.go`)
- **Interface/type definitions** that multiple features depend on may warrant their own commit if they represent a distinct foundational change
- **Configuration/build changes** (go.mod, .gitignore, CI files) group together when related
- **Documentation-only changes** (README, specs, docs/) group separately
- **Refactors with no behavior change** group separately from feature work
- If a single file contains changes serving multiple logical purposes, assign it to the primary group

**Ordering — commit in dependency order:**

1. Build/dependency/config changes
2. Type definitions and interfaces
3. Core library implementations (with their tests)
4. CLI/integration layer (with their tests)
5. Documentation and specs
6. Chores (formatting, cleanup)

If only one logical group exists, create a single commit — do not over-split.

### Step 3: Present the plan

Before committing anything, present a summary and **wait for user approval**:

```
Proposed commits (in order):

1. <type>: <description>
   Files: <file list>

2. <type>: <description>
   Files: <file list>
```

Do **not** proceed until the user approves, adjusts, or overrides the grouping.

### Step 4: Create commits

For each approved group, in order:

1. **Stage** only the files for that group: `git add <file1> <file2> ...`
   - Never use `git add -A` or `git add .`
2. **Verify** staging: `git diff --cached --stat`
3. **Commit** using a HEREDOC for proper formatting:

```bash
git commit -m "$(cat <<'EOF'
<type>: <description>

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>
EOF
)"
```

4. **Verify** the commit: `git log --oneline -1`

### Step 5: Confirm completion

After all commits are created, show a summary:

```
Created N commits:

<hash> <type>: <description>
<hash> <type>: <description>
```

## Conventional Commit Format

Use the `conventional-commits` skill for formatting rules:

- **Format:** `<type>[!]: <description>`
- **Types:** feat, fix, chore, docs, style, refactor, perf, test, build, ci, revert
- **Description:** lowercase, imperative mood, no period, under 72 chars, specific
- **Breaking changes:** append `!` after type

| Change | Type |
|--------|------|
| New feature or capability | `feat:` |
| Bug fix | `fix:` |
| Tests only | `test:` |
| Documentation only | `docs:` |
| Dependencies, build config | `build:` |
| CI/CD changes | `ci:` |
| Refactor (no behavior change) | `refactor:` |
| Formatting | `style:` |
| Performance | `perf:` |
| Maintenance | `chore:` |

## Rules

- **Never** use `git add -A` or `git add .`
- **Never** amend existing commits unless explicitly asked
- **Never** skip hooks (no `--no-verify`)
- **Never** push unless explicitly asked
- **Always** present the plan and wait for approval before committing
- **Always** include the `Co-Authored-By` trailer
- **Always** use HEREDOC syntax for commit messages
