---
description: Analyze code review comments, decide validity, implement fixes, and commit grouped changes.
---

## User Input

```text
$ARGUMENTS
```

## Workflow

### Step 0: Determine input source

The user input can be one of three things:

1. **Inline review comments** — text containing file references with `#L` line markers and `>` quoted reviewer comments. If the input matches this pattern, use it directly as the review comments and proceed to Step 1.

2. **A PR reference** — a number (e.g., `19`), a `#`-prefixed number (e.g., `#19`), or a GitHub PR URL. Extract the PR number and fetch review comments from it (see "Fetching PR review comments" below).

3. **Empty input** — no arguments provided. Detect the current PR for the active branch and fetch review comments from it (see "Fetching PR review comments" below).

#### Fetching PR review comments

When fetching from a PR:

1. **Determine the PR number**:
   - If a number was provided, use it directly
   - If empty, detect the current branch's PR: `gh pr view --json number -q '.number'`
   - If no PR exists for the current branch, stop and report: "No open PR found for the current branch."

2. **Determine the repository**: `gh repo view --json nameWithOwner -q '.nameWithOwner'`

3. **Ensure the `gh pr-review` extension is installed**:
   - Run `gh extension list` and check for `agynio/gh-pr-review`
   - If not found, install it: `gh extension install agynio/gh-pr-review`

4. **Fetch unresolved review comments**:
   ```bash
   gh pr-review review view <pr-number> -R <owner>/<repo> --unresolved --format json
   ```

5. **Parse the JSON output**: Extract all comments from the `reviews[].comments[]` array (and nested `thread_comments[]` for reply context). For each comment, extract:
   - `path` — the file path
   - `line` — the line number
   - `body` — the reviewer's comment text
   - Ignore resolved threads and review-level summaries (reviews that have a `body` but no `comments[]`)

6. **Format as inline review comments**: Convert each extracted comment into the standard format:
   ```
   `<path>`#L<line>:
   > <body>
   ```
   Concatenate all formatted comments with blank lines between them.

7. **If no unresolved comments found**, stop and report: "No unresolved review comments found on PR #&lt;number&gt;."

8. **Present the fetched comments** to the user before proceeding, showing which PR they came from and the number of comments found.

Proceed to Step 1 with the resulting review comments (either inline or fetched).

### Step 1: Parse review comments

Extract each review comment into a structured list:

| # | File | Lines | Comment Summary |
|---|------|-------|-----------------|

If the input is empty or contains no actionable comments, stop and report: "No review comments provided."

### Step 2: Read and analyze each comment

For **every** comment, in order:

1. **Read the referenced file** at the specified lines (include surrounding context, typically ±20 lines)
2. **Understand the reviewer's concern** — what problem are they pointing out?
3. **Evaluate validity** by considering:
   - Is the concern technically correct?
   - Does it identify a real bug, code smell, inconsistency, or missed edge case?
   - Does it improve readability, maintainability, or correctness?
   - Is it consistent with project conventions and the constitution?
   - Or is it subjective/stylistic with no material impact?
4. **Decide**: Accept (will fix) or Reject (explain why)

### Step 3: Present the analysis

Present a verdict table summarizing all comments before making any changes:

```
## Review Analysis

| # | Verdict | Summary | Rationale |
|---|---------|---------|-----------|
| 1 | ✅ Accept | [what will be fixed] | [why it's valid] |
| 2 | ✅ Accept | [what will be fixed] | [why it's valid] |
| 3 | ❌ Reject | [no change needed] | [why it's not valid or not worth changing] |
```

**Wait for user approval** before proceeding with implementation. The user may:
- Approve all verdicts
- Override specific verdicts (accept a rejection, reject an acceptance)
- Ask for more detail on a specific comment

### Step 4: Implement fixes

For each accepted comment:

1. **Read the file** (full context around the change site)
2. **Implement the fix** — prefer minimal, targeted changes that address the reviewer's concern
3. **Verify correctness** — if the change touches Go code, run `go build ./...` to check compilation
4. **Run tests** if the change could affect behavior: `go test ./... -count=1`

**Implementation principles:**
- Fix the root cause, not just the symptom
- If a comment reveals a pattern problem (same issue in multiple places), fix all occurrences
- Keep changes minimal — don't refactor beyond what the comment requires
- Update tests if the fix changes behavior or adds new error paths

### Step 5: Group and commit

After all fixes are implemented:

1. **Group changes** by logical unit — related comments that touch the same concern or file area belong together
2. **Present commit plan** and wait for user approval:

```
Proposed commits (in order):

1. <type>: <description> (addresses comments #X, #Y)
   Files: <file list>

2. <type>: <description> (addresses comment #Z)
   Files: <file list>
```

3. For each approved group, create a commit:
   - Stage only the relevant files: `git add <file1> <file2> ...`
   - Never use `git add -A` or `git add .`
   - Commit using HEREDOC format:

```bash
git commit -m "$(cat <<'EOF'
<type>: <description>

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>
EOF
)"
```

### Step 6: Summary

After all commits, present a final summary:

```
## Review Complete

**Comments**: N total — X accepted, Y rejected
**Commits**: M created

| Commit | Description | Comments Addressed |
|--------|-------------|--------------------|
| <hash> | <type>: <desc> | #1, #2 |
| <hash> | <type>: <desc> | #3 |

**Rejected comments:**
- #Y: [brief reason]
```

## Conventional Commit Format

Use the `conventional-commits` skill for formatting. Most review fixes will be:

| Change | Type |
|--------|------|
| Bug fix identified by reviewer | `fix:` |
| Code improvement / cleanup | `refactor:` |
| Test fix or missing test | `test:` |
| Documentation fix | `docs:` |

## Rules

- **Always** read the actual code before judging a comment — never assume
- **Always** present the analysis and wait for approval before implementing
- **Always** present the commit plan and wait for approval before committing
- **Always** run tests after implementing fixes to Go code
- **Never** use `git add -A` or `git add .`
- **Never** amend existing commits unless explicitly asked
- **Never** skip hooks (no `--no-verify`)
- **Never** push unless explicitly asked
- **Never** make changes beyond what the review comment requires
- **Never** dismiss a comment without a concrete technical justification
