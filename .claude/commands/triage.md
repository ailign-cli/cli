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

2. **Determine the repository**:
   ```bash
   gh repo view --json nameWithOwner -q '.nameWithOwner'
   ```
   Parse the result to extract `OWNER` and `REPO` (split on `/`).

3. **Fetch review threads via GraphQL**:
   ```bash
   gh api graphql -f query='
   query($owner: String!, $repo: String!, $prNumber: Int!) {
     repository(owner: $owner, name: $repo) {
       pullRequest(number: $prNumber) {
         reviewThreads(first: 100) {
           nodes {
             id
             isResolved
             comments(first: 100) {
               nodes { id databaseId body author { login } path line }
             }
           }
         }
       }
     }
   }' -F owner="$OWNER" -F repo="$REPO" -F prNumber="$PR_NUMBER"
   ```

4. **Parse the GraphQL response**: From `data.repository.pullRequest.reviewThreads.nodes[]`, filter to threads where `isResolved` is `false`. For each unresolved thread, extract:
   - `thread_id` — the thread's `id` field (format: `PRRT_kwDO...`). **Store this for later use in reply/resolve operations.**
   - From `comments.nodes[0]` (the root comment): `path`, `line`, `body`, `author.login`, `databaseId`
   - From `comments.nodes[1..]` (reply comments): additional context for understanding the full thread discussion
   - Construct a **GitHub link** for each root comment: `https://github.com/OWNER/REPO/pull/PR_NUMBER#discussion_r<databaseId>`

5. **Format as inline review comments**: Convert each extracted thread into the standard format:
   ```
   `<path>`#L<line>:
   > <body>
   ```
   Concatenate all formatted comments with blank lines between them. Maintain a mapping of **comment number → thread_id** for use in later steps.

6. **If no unresolved threads found**, stop and report: "No unresolved review comments found on PR #<number>."

7. **Present the fetched comments** to the user before proceeding, showing which PR they came from and the number of unresolved threads found.

Proceed to Step 1 with the resulting review comments (either inline or fetched).

### Step 1: Parse review comments

Extract each review comment into a structured list:

| # | File | Lines | Author | Comment Summary | Thread |
|---|------|-------|--------|-----------------|--------|

- **Author**: the `author.login` from the review comment (e.g., `copilot-pull-request-reviewer`, a colleague's GitHub handle, or a bot name). Helps distinguish human vs. automated feedback.
- **Thread**: a clickable link to the review comment on GitHub (e.g., `[view](url)`). Only populated when fetched from a PR; leave empty for inline input. The underlying `thread_id` is stored internally for GraphQL reply/resolve operations and does not need to be displayed in the table.

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
4. **Decide**:
   - **Accept** — the concern is valid and will be fixed
   - **Reject** — the concern is not valid or not worth changing (explain why)
   - **Unclear** — the comment is ambiguous, incomplete, or could be interpreted multiple ways. Formulate specific clarification questions.

### Step 3: Present the analysis

Present a verdict table summarizing all comments before making any changes:

```
## Review Analysis

| # | Verdict | Author | Summary | Rationale | Thread |
|---|---------|--------|---------|-----------|--------|
| 1 | ✅ Accept | @user | [what will be fixed] | [why it's valid] | [link](url) |
| 2 | ❌ Reject | @bot | [no change needed] | [why it's not valid] | [link](url) |
| 3 | ❓ Unclear | @user | [what's ambiguous] | [clarification questions] | [link](url) |
```

**Wait for user approval** before proceeding with implementation. The user may:
- Approve all verdicts
- Override specific verdicts (accept a rejection, reject an acceptance)
- Ask for more detail on a specific comment

### Step 3a: Reply and resolve rejected comments

After the user approves the analysis, for each **rejected** comment (only when input was fetched from a PR, not inline):

1. **Reply** to the review thread with the rejection rationale:
   ```bash
   gh api graphql -f query='
   mutation($threadId: ID!, $body: String!) {
     addPullRequestReviewThreadReply(input: {
       pullRequestReviewThreadId: $threadId
       body: $body
     }) {
       comment { id }
     }
   }' -F threadId="$THREAD_ID" -F body="$REPLY_BODY"
   ```
   The `REPLY_BODY` should be a concise explanation of why the comment was rejected, derived from the Rationale column in the verdict table.

2. **Resolve** the thread:
   ```bash
   gh api graphql -f query='
   mutation($threadId: ID!) {
     resolveReviewThread(input: {threadId: $threadId}) {
       thread { id isResolved }
     }
   }' -F threadId="$THREAD_ID"
   ```

**Important:** Use the `thread_id` (format `PRRT_kwDO...`) from Step 0 sub-step 4, **not** individual comment IDs.

If the input was inline review comments (not fetched from a PR), skip this step entirely — there are no thread IDs to reply to.

### Step 3b: Post clarification questions for unclear comments

After the user approves the analysis, for each **unclear** comment (only when input was fetched from a PR, not inline):

1. **Reply** to the review thread with the clarification questions:
   ```bash
   gh api graphql -f query='
   mutation($threadId: ID!, $body: String!) {
     addPullRequestReviewThreadReply(input: {
       pullRequestReviewThreadId: $threadId
       body: $body
     }) {
       comment { id }
     }
   }' -F threadId="$THREAD_ID" -F body="$REPLY_BODY"
   ```
   The `REPLY_BODY` should tag the comment author (e.g., `@username`) and list the specific clarification questions. Be concise and direct.

2. **Do NOT resolve** the thread — it stays open so the author can respond.

If the input was inline review comments, skip this step.

**Re-triage after clarification:** When `/triage` is run again on the same PR, previously unclear threads that now have replies will reappear as unresolved. Read the full thread (including new replies) and re-evaluate — the comment may now be clear enough to accept or reject. If it's still unclear, post follow-up questions. This cycle repeats until the comment can be resolved.

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
EOF
)"
```

### Step 5a: Push and resolve accepted comments

After all commits are created:

1. **Push** the commits to the remote branch:
   ```bash
   git push
   ```
   If the push fails (e.g., remote has new commits), report the error and stop. Do not force-push.

2. **Collect commit hashes** recorded immediately after each commit in Step 5 (captured with `git rev-parse HEAD`), rather than inferring them later from `git log`.

3. **Reply and resolve** each accepted comment's thread (only when input was fetched from a PR, not inline):

   For each accepted comment:

   a. **Reply** to the review thread referencing the commit:
      ```bash
      gh api graphql -f query='
      mutation($threadId: ID!, $body: String!) {
        addPullRequestReviewThreadReply(input: {
          pullRequestReviewThreadId: $threadId
          body: $body
        }) {
          comment { id }
        }
      }' -F threadId="$THREAD_ID" -F body="$REPLY_BODY"
      ```
      The `REPLY_BODY` should briefly describe the fix and reference the commit hash, e.g.:
      `"Fixed in <commit_hash> — <brief description of what was changed>."`

   b. **Resolve** the thread:
      ```bash
      gh api graphql -f query='
      mutation($threadId: ID!) {
        resolveReviewThread(input: {threadId: $threadId}) {
          thread { id isResolved }
        }
      }' -F threadId="$THREAD_ID"
      ```

**Important:** Use the `thread_id` (format `PRRT_kwDO...`) from Step 0 sub-step 4, **not** individual comment IDs.

If the input was inline review comments (not fetched from a PR), skip the reply/resolve part and do not push unless explicitly requested.

### Step 6: Distill learnings

Extract actionable learnings from the review comments — patterns, mistakes, or conventions worth remembering for future work.

#### 6.1 Identify new learnings

From the accepted and rejected comments, distill what went wrong, what convention was missed, or what project-specific decisions were clarified. Accepted comments reveal gaps; rejected comments can reveal important conventions or architectural decisions worth documenting:

```
## Learnings

| Learning | Scope | Comments |
| -------- | ----- | -------- |
| <learning> | project | #1, #2 |
| <learning> | user | #3 |
```

**Scope** determines where the learning is stored:

- **project** — specific to this codebase (conventions, architecture decisions, project-specific patterns). Stored in the project's learnings file (e.g., `.specify/memory/learnings.md`, `LEARNINGS.md`, or wherever this project keeps its memory).
- **user** — general coding practices applicable across any codebase (language idioms, review patterns, universal best practices). Stored in the user's personal memory (e.g., `~/.claude/CLAUDE.md` or equivalent user-level config).

#### 6.2 Integrate with existing learnings

For each learning, **read the target file first** and compare against what's already there:

1. **Duplicate** — the learning already exists in substance. Skip it.
2. **Refinement** — the learning strengthens or clarifies an existing one. Update the existing entry in place (don't add a second entry).
3. **Contradiction** — the new learning supersedes an older one. Replace the old entry.
4. **Novel** — the learning is genuinely new. Add it.

**Never blindly append.** The goal is a curated, non-redundant set of learnings that stays useful as it grows. Remove learnings that have been fully absorbed into project conventions or are no longer relevant.

#### 6.3 Write changes

Apply the additions, updates, and removals to the appropriate files. If a target file doesn't exist yet, create it with a brief header explaining its purpose.

### Step 7: Summary

After all commits and thread resolutions, present a final summary:

```
## Review Complete

**Comments**: N total — X accepted, Y rejected, Z unclear
**Commits**: M created
**Threads resolved**: R (X accepted + Y rejected)

| Commit | Description | Comments Addressed |
|--------|-------------|--------------------|
| <hash> | <type>: <desc> | #1, #2 |
| <hash> | <type>: <desc> | #3 |

**Rejected comments** (replied + resolved):
- #Y: [brief reason]

**Unclear comments** (questions posted, awaiting response):
- #Z: @author — [summary of what was asked]
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
- **Always** push after committing when processing PR review threads (required for commit hash references in replies)
- **Never** push when processing inline review comments unless explicitly asked
- **Never** make changes beyond what the review comment requires
- **Never** dismiss a comment without a concrete technical justification
- **Always** use `-F` flags for GraphQL variables — never inline `$variable` references in the query string, as the shell will interpret `$` before `gh` sees it. The query declares variables (e.g., `$threadId: ID!`) inside single quotes (safe from shell expansion), and `-F threadId="value"` passes them at the GraphQL level.

## Thread Management Reference

For manual thread management outside the triage workflow:

**Unresolve a thread** (if a resolved thread needs to be reopened):
```bash
gh api graphql -f query='
mutation($threadId: ID!) {
  unresolveReviewThread(input: {threadId: $threadId}) {
    thread { id isResolved }
  }
}' -F threadId="PRRT_kwDO..."
```

This is not used by the triage workflow itself but is documented here for manual recovery if a thread was resolved prematurely.
