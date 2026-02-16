---
name: smart-commit
description: 'Analyze staged and unstaged changes, group them into logical commits, and create each commit separately using conventional commit format. Use when asked to "commit", "commit changes", "smart commit", or "group and commit".'
---

# Smart Commit

Analyze all uncommitted changes, group them into logical units, and create separate conventional commits for each group.

## When to Use This Skill

- User asks to commit current changes
- User asks to group and commit
- User says "smart commit" or "commit changes"

## Workflow

### Step 1: Assess the current state

Run these commands in parallel to understand what needs to be committed:

1. `git status` — see all modified, added, and untracked files
2. `git diff` — see unstaged changes (content)
3. `git diff --cached` — see already-staged changes (content)
4. `git log --oneline -10` — see recent commit style for context

### Step 2: Group changes logically

Analyze all changed files and their diffs. Group them by **logical unit of work**, not by file type or directory. A logical group is a set of changes that together accomplish one coherent purpose.

**Grouping principles:**

- **Feature code + its tests** belong in the same commit (e.g., `validator.go` + `validator_test.go`)
- **Interface/type definitions** that multiple features depend on may warrant their own commit if they are a distinct foundational change
- **Configuration/build changes** (go.mod, .gitignore, CI files) group together when related
- **Documentation-only changes** (README, comments, docs/) group separately
- **Refactors with no behavior change** group separately from feature work
- If a single file contains changes for multiple logical purposes, note this and assign it to the primary group

**Group ordering:**

Commit groups in dependency order — foundational changes first, then features that build on them:

1. Build/dependency/config changes
2. Type definitions and interfaces
3. Core library implementations (with their tests)
4. CLI/integration layer (with their tests)
5. Documentation
6. Chores (formatting, cleanup)

### Step 3: Present the plan

Before committing anything, present a clear summary to the user:

```
Proposed commits (in order):

1. <type>: <description>
   Files: <file list>

2. <type>: <description>
   Files: <file list>

...
```

Wait for the user to approve, adjust, or override the grouping.

### Step 4: Create commits

For each group, in order:

1. Stage only the files belonging to that group: `git add <file1> <file2> ...`
2. Verify staging is correct: `git diff --cached --stat`
3. Create the commit using the conventional commits format:

```bash
git commit -m "$(cat <<'EOF'
<type>: <description>

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>
EOF
)"
```

4. Verify the commit was created: `git log --oneline -1`

### Step 5: Confirm completion

After all commits are created, show:

```
Created N commits:

<short hash> <type>: <description>
<short hash> <type>: <description>
...
```

## Conventional Commit Format

Follow the `conventional-commits` skill for message formatting:

- **Format:** `<type>[!]: <description>`
- **Types:** feat, fix, chore, docs, style, refactor, perf, test, build, ci, revert
- **Description:** lowercase, imperative mood, no period, under 72 chars, specific
- **Breaking changes:** append `!` after type

### Type Selection Guide

| Change | Type |
|--------|------|
| New feature or capability | `feat:` |
| Bug fix | `fix:` |
| New or updated tests only | `test:` |
| Documentation only | `docs:` |
| Dependencies, build config | `build:` |
| CI/CD pipeline changes | `ci:` |
| Refactor (no behavior change) | `refactor:` |
| Code formatting | `style:` |
| Performance improvement | `perf:` |
| Everything else (maintenance) | `chore:` |

## Rules

- **Never** use `git add -A` or `git add .` — always add specific files
- **Never** amend existing commits unless explicitly asked
- **Never** skip pre-commit hooks (no `--no-verify`)
- **Never** push unless explicitly asked
- **Always** present the plan before committing
- **Always** include the `Co-Authored-By` trailer
- If only one logical group exists, create a single commit (no need to over-split)
- If unsure about grouping, prefer fewer larger commits over many tiny ones
