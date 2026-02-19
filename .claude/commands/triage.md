---
description: Analyze code review comments and CI/CD check failures, decide validity, implement fixes, and commit grouped changes.
---

## User Input

```text
$ARGUMENTS
```

## Workflow

### Step 0: Determine input source

The user input can be one of three things:

1. **Inline review comments** — text containing file references with `#L` line markers and `>` quoted reviewer comments. If the input matches this pattern, use it directly as the review comments and proceed to Step 1.

2. **A PR reference** — one of:
   - A number: `19` or `#19` (assumes the local repo)
   - A cross-repo reference: `OWNER/REPO#NUMBER` (e.g., `acme/web-api#123`)
   - A GitHub PR URL: `https://github.com/OWNER/REPO/pull/NUMBER`

   Extract the PR number (and optionally `OWNER/REPO`) and fetch review comments from it (see "Fetching PR review comments" below).

3. **Empty input** — no arguments provided. Detect the current PR for the active branch and fetch review comments from it (see "Fetching PR review comments" below).

#### Fetching PR review comments

When fetching from a PR:

1. **Determine the PR number and target repository**:
   - If a cross-repo reference (`OWNER/REPO#NUMBER`) or GitHub URL was provided, extract both the PR number and `OWNER/REPO` from the input.
   - If a plain number was provided, use it as the PR number. Determine the repository from the local remote:
     ```bash
     gh repo view --json nameWithOwner -q '.nameWithOwner'
     ```
     Parse the result to extract `OWNER` and `REPO` (split on `/`).
   - If empty, detect the current branch's PR: `gh pr view --json number -q '.number'` and determine the repository from the local remote (as above).
   - If no PR exists for the current branch, stop and report: "No open PR found for the current branch."

2. **Validate local repository match**:
   Get the current working directory's git remote:
   ```bash
   git remote get-url origin
   ```
   Extract `OWNER/REPO` from the remote URL (handles both HTTPS `https://github.com/OWNER/REPO.git` and SSH `git@github.com:OWNER/REPO.git` formats).

   Compare with the target `OWNER/REPO` from step 1. If they don't match:
   - **Still fetch and display** the review comments (continue through sub-steps 3–7 below as read-only) so the developer can see what was flagged.
   - After presenting the comments, provide a **brief summary** — e.g., how many comments, which authors, and a one-line characterization of the themes or patterns (e.g., "7 comments from @datadog-official — all about unpinned GitHub Actions in CI workflows").
   - After the summary, stop and report:
     > "PR #N belongs to `OWNER/REPO`, but your current directory is linked to `LOCAL_OWNER/LOCAL_REPO`. I've listed the review comments above, but cannot read code, fetch CI check logs, or implement fixes from this directory. Navigate to a local clone of `OWNER/REPO` and re-run `/triage` to process them."
   - **Do not** proceed to Step 0b (CI checks) or Step 4 (implementation) — the agent cannot read the referenced files, fetch CI logs, or apply fixes for the correct repository.

3. **Fetch review threads via GraphQL**:
   ```bash
   gh api graphql -F owner="$OWNER" -F repo="$REPO" -F prNumber="$PR_NUMBER" -f query='query($owner: String!, $repo: String!, $prNumber: Int!) { repository(owner: $owner, name: $repo) { pullRequest(number: $prNumber) { reviewThreads(first: 100) { nodes { id isResolved comments(first: 100) { nodes { id databaseId body author { login } path line originalLine } } } } } } }'
   ```
   **Quote safety:** The query MUST be passed as a single-line string to `-f query=` with ASCII single quotes. Markdown renderers, AI tools, and text editors can silently convert quotes to Unicode curly quotes (`'` `'`), causing `UNKNOWN_CHAR` GraphQL parse errors. Keeping the query on one line makes copy-paste safer. If the query needs to be multi-line for readability during editing, use a heredoc alternative:
   ```bash
   QUERY=$(cat <<'GQL'
   query($owner: String!, $repo: String!, $prNumber: Int!) {
     repository(owner: $owner, name: $repo) {
       pullRequest(number: $prNumber) {
         reviewThreads(first: 100) {
           nodes {
             id
             isResolved
             comments(first: 100) {
               nodes { id databaseId body author { login } path line originalLine }
             }
           }
         }
       }
     }
   }
   GQL
   )
   gh api graphql -f query="$QUERY" -F owner="$OWNER" -F repo="$REPO" -F prNumber="$PR_NUMBER"
   ```

4. **Parse the GraphQL response**: From `data.repository.pullRequest.reviewThreads.nodes[]`, filter to threads where `isResolved` is `false`. For each unresolved thread, extract:
   - `thread_id` — the thread's `id` field (format: `PRRT_kwDO...`). **Store this for later use in reply/resolve operations.**
   - From `comments.nodes[0]` (the root comment): `path`, `line` (fall back to `originalLine` if `line` is null — this happens on outdated diffs), `body`, `author.login`, `databaseId`
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

Proceed to Step 0b (CI checks) and then Step 1 with the resulting review comments (either inline or fetched).

### Step 0b: Fetch CI/CD check statuses

**Skip this step** if the input was inline review comments (no PR reference exists). Also skip if the PR belongs to a different repository (cross-repo case detected in Step 0).

#### 1. Fetch all checks for the PR

```bash
gh pr checks "$PR_NUMBER" --repo "$OWNER/$REPO" \
  --json name,state,bucket,link,workflow
```

If the command succeeds, parse the JSON array and filter to entries where `bucket` is `"fail"`. If no checks have `bucket == "fail"`, note "All CI checks pass" and skip the remaining sub-steps of Step 0b.

If the `gh pr checks` command itself fails (non-zero exit status) for any reason (e.g., network issue, authentication/permission error, or PR not found), treat this as a non-fatal condition: emit a clear warning that CI status could not be determined (including the error message and concrete remediation steps such as checking GitHub login, repo access, PR number, and network connectivity), skip the remaining sub-steps of Step 0b, and proceed with the rest of the triage workflow without CI analysis.

#### 2. Extract run IDs from failed check links

For each failed check, the `link` field contains a URL of the form:
```
https://github.com/OWNER/REPO/actions/runs/{run_id}/job/{job_id}
```
Extract `run_id` using the pattern `/actions/runs/(\d+)/`. Extract `job_id` from the trailing `/job/(\d+)` segment if present. Deduplicate run IDs — multiple failing jobs can share the same workflow run.

#### 3. For each unique failed run, get job breakdown

```bash
gh run view "$RUN_ID" --repo "$OWNER/$REPO" --json jobs
```

From the `jobs` array, filter to jobs where `conclusion` is `"failure"`, `"cancelled"` (cancelled mid-step), `"timed_out"`, or `"action_required"` (manual intervention needed). Treat all of these as CI failures for analysis. For each such job, record:
- `job_id` (the `databaseId` field)
- `job_name`
- `workflow_name` (from the outer run context or `gh pr checks` `workflow` field)
- The list of steps — filter to steps where `conclusion` is `"failure"`; for each, record `name` and `number`

#### 4. Fetch failed step logs

For each failing run, fetch the failed-step log output:

```bash
gh run view "$RUN_ID" --repo "$OWNER/$REPO" --log-failed
```

This returns interleaved log lines for all failing steps in the run. Each log line is tab-separated:
```
<job_name>\t<step_name>\t<log_line>
```

For each line, parse out `job_name`, `step_name`, and `log_line`, and group lines by the `(job_name, step_name)` pair. Treat each unique `(job_name, step_name)` as a distinct step and, for that group, retain only the last **50** `log_line` entries (per group, not globally) as the "relevant error output" for that failing step. These final 50 lines typically contain the actual error rather than setup noise; if a step produced more than 50 lines, record that its output was truncated.

**Fallback:** If `--log-failed` produces no output or fails, fetch per-job:
```bash
gh run view "$RUN_ID" --repo "$OWNER/$REPO" --job "$JOB_ID" --log-failed
```

#### 5. Optionally fetch annotations

For additional structured error context (especially useful for linters and test reporters that emit annotations):

First, obtain the `CHECK_RUN_ID`. For checks whose `link` field points to a GitHub Actions job URL (`/actions/runs/{run_id}/job/{job_id}`), the `job_id` extracted in sub-step 2 is also the check run ID. For other checks (non-Actions status checks), fetch check runs for the PR's head commit:
```bash
gh api "/repos/$OWNER/$REPO/commits/$HEAD_SHA/check-runs" --jq '.check_runs[] | {id, name}'
```
Match by `name` against the failing check's `name` to find the corresponding `id`.

Then fetch annotations:
```bash
gh api "/repos/$OWNER/$REPO/check-runs/$CHECK_RUN_ID/annotations"
```

This returns structured annotation objects with `path`, `start_line`, `end_line`, `annotation_level`, and `message`. If annotations exist, they supplement the log output and can pinpoint specific files and lines to read.

**When to perform this sub-step:**
- The captured error log for a failing step does not already contain clear file and line references, and you need more precise locations to investigate
- The failing check is a linter, test reporter, or similar tool known to emit GitHub annotations
- You want to cross-check log messages against structured locations to build a more accurate `annotations` list

**When to skip this sub-step:**
- The error log already contains sufficient, unambiguous file/line references (e.g., `path/to/file.go:123` style messages) and additional structure would not change the fix you apply
- The check type is known not to emit annotations (based on previous runs or documentation)
- The API returns an empty array

In all skip cases, proceed with just the captured log output; leave `annotations` empty for that failing step.

#### 6. Build the CI failure inventory

Produce an internal data structure (used in Step 1 and Step 2) with one entry per failing step:

```
CI-N:
  check_name:    <name field from gh pr checks>
  workflow_name: <workflow name>
  job_name:      <job name>
  step_name:     <failing step name>
  run_id:        <run_id>
  job_id:        <job_id>
  error_log:     <captured log lines, up to 50>
  annotations:   <list of {path, line, message} if available>
```

Number the entries sequentially as `CI-1`, `CI-2`, etc. — distinct from the review comment numbering (`#1`, `#2`, ...) used throughout the workflow.

If no failures are found after all fetching, note "All CI checks pass" and omit CI sections in subsequent steps.

### Step 1: Parse review comments

Extract each review comment into a structured list:

| # | File | Lines | Author | Comment Summary | Thread |
|---|------|-------|--------|-----------------|--------|

- **Author**: the `author.login` from the review comment (e.g., `copilot-pull-request-reviewer`, a colleague's GitHub handle, or a bot name). Helps distinguish human vs. automated feedback.
- **Thread**: a clickable link to the review comment on GitHub (e.g., `[view](url)`). Only populated when fetched from a PR; leave empty for inline input. The underlying `thread_id` is stored internally for GraphQL reply/resolve operations and does not need to be displayed in the table.

If the input is empty or contains no actionable comments, stop and report: "No review comments or CI failures found." (unless CI failures were collected in Step 0b, in which case proceed with CI analysis only).

If CI failures were collected in Step 0b, also present a **CI Failures** table:

| # | Check | Workflow | Job | Step | Error Summary |
|---|-------|----------|-----|------|---------------|

- **#**: The `CI-N` identifier assigned in Step 0b
- **Check**: The check name from `gh pr checks`
- **Workflow**: The workflow name
- **Job**: The job name within the workflow
- **Step**: The specific failing step name
- **Error Summary**: One-line characterization of the error (e.g., "exported function Foo lacks comment", "FAIL internal/sync: panic in TestSyncDry")

If there are no CI failures, omit this table entirely.

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

#### CI Failure Analysis

For each CI failure (`CI-N`) from Step 0b:

1. **Read the error log** captured in Step 0b
2. **Identify the root cause** — what is the failing check actually reporting? Common cases:
   - Lint violation: specific file, line, and rule
   - Compilation error: file and line where it fails
   - Test failure: test name, failure message, stack trace
   - Missing file or dependency: build system error
3. **Read relevant source files** — if the error references specific files and lines (from log parsing or annotations), read those files with ±20 lines of context, the same as for review comments
4. **Propose a fix** — what change would make this check pass?
5. **Classify**:
   - **Fix** — the failure is clearly caused by something in this PR and can be addressed
   - **Blocked** — the failure requires operator input: external flaky test, missing secret, infrastructure issue, or a situation where multiple valid fix approaches exist (list the options and ask the operator to choose)
   - **Unclear** — the failure is ambiguous; formulate specific questions for the operator

For **Blocked** items, be explicit: describe why you cannot proceed unilaterally and what the operator must decide or provide.

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

If CI failures exist, also present a **CI Check Analysis** table:

```
## CI Check Analysis

| # | Verdict | Check | Job / Step | Root Cause Summary | Proposed Fix |
|---|---------|-------|------------|--------------------|--------------|
| CI-1 | 🔧 Fix | lint | golangci-lint / Run golangci-lint | Exported func Foo missing godoc | Add godoc comment to Foo |
| CI-2 | 🚫 Blocked | test | unit-tests / Run tests | Panic in TestSyncDry — two valid approaches | Option A: ..., Option B: ... |
| CI-3 | ❓ Unclear | build | compile / Build binary | Unrecognized import path | [clarification questions] |
```

**Verdicts for CI failures:**
- **🔧 Fix** — will be addressed in Step 4 alongside review comment fixes
- **🚫 Blocked** — cannot proceed without operator input. Clearly state what decision is needed. Do NOT implement anything for this item until the operator responds.
- **❓ Unclear** — ask specific questions before proceeding

**Key constraint:** Failed checks should NOT be dismissed or worked around (e.g., disabling a lint rule, skipping a test, adding `//nolint` directives) unless there is truly no other option. If dismissal is the only path, flag it explicitly to the operator and wait for approval before doing it.

When a failure appears **unrelated to the PR** (e.g., flaky test, known base-branch breakage):
- Compare against the base branch CI status. If the same check is already failing on the base branch, document it explicitly in the analysis (including links to the failing runs) and treat it as **🚫 Blocked** or **❓ Unclear** rather than silently ignoring it.
- For flaky checks, call out the flakiness pattern in the CI Check Analysis and ask the operator whether to re-run the check in CI. Do **not** locally skip, mute, or weaken the check to get a green run.
- If the operator decides to proceed despite a pre-existing or flaky failure, record that decision in the summary so the behavior remains explainable and auditable.

**Wait for user approval** before proceeding with implementation. The user may:
- Approve all verdicts (review comments and CI failures)
- Override specific verdicts (accept a rejection, reject an acceptance, resolve a blocked item)
- Ask for more detail on a specific comment or CI failure
- Provide direction for blocked or unclear CI items

### Step 3a: Reply and resolve rejected comments

After the user approves the analysis, for all **rejected** comments (only when input was fetched from a PR, not inline):

**Batch all replies and resolves using GraphQL aliases** — this makes two API calls total instead of 2N. **Batch size limit:** GitHub GraphQL hits `RESOURCE_LIMITS_EXCEEDED` at around 8–9 aliased mutations per request. When there are more than 8 threads to process, split into chunks of at most 8 aliases per call.

1. **Reply to all threads** in batched mutation(s):
   Build a mutation using aliases (`r0`, `r1`, ...) for each rejected thread (max 8 per call):
   ```bash
   QUERY=$(cat <<'GQL'
   mutation {
     r0: addPullRequestReviewThreadReply(input: {pullRequestReviewThreadId: "THREAD_ID_0", body: "REPLY_0"}) { comment { id } }
     r1: addPullRequestReviewThreadReply(input: {pullRequestReviewThreadId: "THREAD_ID_1", body: "REPLY_1"}) { comment { id } }
   }
   GQL
   )
   gh api graphql -f query="$QUERY"
   ```
   Replace `THREAD_ID_N` and `REPLY_N` with actual values. The `REPLY_BODY` should be a concise explanation of why the comment was rejected, derived from the Rationale column in the verdict table. **Escape double quotes and newlines** in reply bodies since they're embedded in the query string.

2. **Resolve all threads** in batched mutation(s) (max 8 per call):
   ```bash
   QUERY=$(cat <<'GQL'
   mutation {
     t0: resolveReviewThread(input: {threadId: "THREAD_ID_0"}) { thread { id } }
     t1: resolveReviewThread(input: {threadId: "THREAD_ID_1"}) { thread { id } }
   }
   GQL
   )
   gh api graphql -f query="$QUERY"
   ```

**Why two calls instead of one?** Replies must be posted before resolving, and GraphQL does not guarantee execution order of mutations within a single request. Batching replies first, then resolves, ensures correct ordering.

**Important:** Use the `thread_id` (format `PRRT_kwDO...`) from Step 0 sub-step 4, **not** individual comment IDs.

**Fallback:** If a batched mutation fails (e.g., `RESOURCE_LIMITS_EXCEEDED` or one thread ID is invalid), fall back to smaller chunks or individual calls for the failing threads so that valid threads are still processed.

If the input was inline review comments (not fetched from a PR), skip this step entirely — there are no thread IDs to reply to.

### Step 3b: Post clarification questions for unclear comments

After the user approves the analysis, for all **unclear** comments (only when input was fetched from a PR, not inline):

**Batch all replies using GraphQL aliases** — one API call for all unclear threads (max 8 aliases per call; split into chunks if more):

1. **Reply to all unclear threads** in batched mutation(s):
   ```bash
   QUERY=$(cat <<'GQL'
   mutation {
     c0: addPullRequestReviewThreadReply(input: {pullRequestReviewThreadId: "THREAD_ID_0", body: "REPLY_0"}) { comment { id } }
     c1: addPullRequestReviewThreadReply(input: {pullRequestReviewThreadId: "THREAD_ID_1", body: "REPLY_1"}) { comment { id } }
   }
   GQL
   )
   gh api graphql -f query="$QUERY"
   ```
   The `REPLY_BODY` should tag the comment author (e.g., `@username`) and list the specific clarification questions. Be concise and direct. **Escape double quotes and newlines** in reply bodies.

2. **Do NOT resolve** the threads — they stay open so the authors can respond.

**Fallback:** If a batched mutation fails (e.g., `RESOURCE_LIMITS_EXCEEDED`), fall back to smaller chunks or individual calls for each thread.

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

#### CI Fix Implementation

After implementing review comment fixes, implement fixes for each CI failure marked as **Fix** (in order of `CI-N`):

1. **Re-read the error log** and the relevant source files
2. **Implement the targeted fix** — the same principles apply: fix the root cause, not just the symptom; keep changes minimal; if the same issue appears in multiple places, fix all occurrences
3. **Verify** by re-running the failing check locally if possible:
   - Lint: `golangci-lint run`
   - Build: `go build ./...`
   - Tests: `go test ./... -count=1`
4. **For Blocked items**: do not implement anything. Record the item as pending operator input.

CI fixes and review comment fixes may be grouped together in Step 5 if they are logically related (e.g., a review comment and a lint failure both touch the same function).

### Step 5: Group and commit

After all fixes are implemented:

1. **Group changes** by logical unit — related comments that touch the same concern or file area belong together
2. **Present commit plan** and wait for user approval:

```
Proposed commits (in order):

1. <type>: <description> (addresses #X, #Y, CI-1)
   Files: <file list>

2. <type>: <description> (addresses #Z)
   Files: <file list>

3. <type>: <description> (addresses CI-2)
   Files: <file list>
```

CI fixes and review comment fixes may be combined in a single commit when they are logically cohesive (e.g., both address the same function). Keep them separate when they address distinct concerns.

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
   - **Immediately after each commit**, record the hash and map it to the addressed comment numbers:
     ```bash
     git rev-parse HEAD
     ```
     Store this mapping (commit hash → comment numbers) for use in Step 5a when replying to threads.

### Step 5a: Push and resolve accepted comments

After all commits are created:

1. **Push** the commits to the remote branch:
   ```bash
   git push
   ```
   If the push fails (e.g., remote has new commits), report the error and stop. Do not force-push.

2. **Collect commit hashes** recorded immediately after each commit in Step 5 (captured with `git rev-parse HEAD`), rather than inferring them later from `git log`.

3. **Reply and resolve** all accepted comments' threads using **batched GraphQL mutations** (only when input was fetched from a PR, not inline):

   a. **Reply to all threads** in batched mutation(s) (max 8 aliases per call):
      ```bash
      QUERY=$(cat <<'GQL'
      mutation {
        r0: addPullRequestReviewThreadReply(input: {pullRequestReviewThreadId: "THREAD_ID_0", body: "Fixed in abc123 — description of fix."}) { comment { id } }
        r1: addPullRequestReviewThreadReply(input: {pullRequestReviewThreadId: "THREAD_ID_1", body: "Fixed in def456 — description of fix."}) { comment { id } }
      }
      GQL
      )
      gh api graphql -f query="$QUERY"
      ```
      Each `REPLY_BODY` should briefly describe the fix and reference the commit hash. **Escape double quotes and newlines** in reply bodies. If there are more than 8 threads, split into multiple calls of at most 8 aliases each.

   b. **Resolve all threads** in batched mutation(s) (max 8 aliases per call):
      ```bash
      QUERY=$(cat <<'GQL'
      mutation {
        t0: resolveReviewThread(input: {threadId: "THREAD_ID_0"}) { thread { id } }
        t1: resolveReviewThread(input: {threadId: "THREAD_ID_1"}) { thread { id } }
      }
      GQL
      )
      gh api graphql -f query="$QUERY"
      ```

   **Why two calls instead of one?** Replies must be posted before resolving, and GraphQL does not guarantee execution order of mutations within a single request.

   **Fallback:** If a batched mutation fails (e.g., `RESOURCE_LIMITS_EXCEEDED`), fall back to smaller chunks or individual calls for the failing threads.

**Important:** Use the `thread_id` (format `PRRT_kwDO...`) from Step 0 sub-step 4, **not** individual comment IDs.

If the input was inline review comments (not fetched from a PR), skip the reply/resolve part and do not push unless explicitly requested.

CI check failures resolved by the committed fixes will re-run automatically after the push — no explicit thread management is needed for CI items. After the push, verify the status of the re-run checks; if any still fail or new failures appear, investigate them and, if appropriate, run `/triage` again to process the new CI feedback.

### Step 6: Distill learnings

Extract actionable learnings from the review comments — patterns, mistakes, or conventions worth remembering for future work.

#### 6.1 Identify new learnings

From the accepted and rejected comments **and from CI failures**, distill what went wrong, what convention was missed, or what project-specific decisions were clarified. Accepted comments reveal gaps; rejected comments can reveal important conventions or architectural decisions worth documenting. CI failures often reveal:
- Missing or incorrect test coverage
- Lint rules the developer wasn't aware of
- Build constraints (e.g., build tags, Go version requirements)
- Workflow-specific requirements (e.g., generated file checks, formatting rules)
- Process issues (e.g., tests or linters weren't run locally before pushing)
- Tooling gaps (e.g., local `golangci-lint` version or config differs from CI)

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
**CI failures**: C total — F fixed, B blocked, U unclear
**Commits**: M created
**Threads resolved**: R (X accepted + Y rejected)

| Commit | Description | Comments Addressed | CI Failures Addressed |
|--------|-------------|--------------------|-----------------------|
| <hash> | <type>: <desc> | #1, #2 | CI-1 |
| <hash> | <type>: <desc> | #3 | — |
| <hash> | <type>: <desc> | — | CI-2 |

**Rejected comments** (replied + resolved):
- #Y: [brief reason]

**Unclear comments** (questions posted, awaiting response):
- #Z: @author — [summary of what was asked]

**Blocked CI failures** (awaiting operator input):
- CI-N: [what decision is needed]
```

If there are no CI failures, omit the CI-related sections (the `**CI failures**` line and the `**Blocked CI failures**` section). Keep the `CI Failures Addressed` column in the commits table and use `—` in each cell for consistency.

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
- **Always** use `-F` flags for GraphQL variables — never inline *shell* variable references (e.g., `$THREAD_ID`) in the query string, as the shell will interpret `$` before `gh` sees it.
- **Always** use heredoc (`cat <<'GQL' ... GQL`) for multi-line GraphQL queries to avoid Unicode curly quote corruption. Markdown renderers and AI tools silently convert ASCII `'` to `'`/`'`, causing `UNKNOWN_CHAR` parse errors. Heredoc quotes are immune to this.
- **Always** batch multiple thread operations using GraphQL aliases (e.g., `r0: mutation(...)`, `r1: mutation(...)`) to minimize API calls. Batch replies first, then resolves, in two separate calls to ensure ordering. **Limit each batched call to at most 8 aliases** — GitHub GraphQL hits `RESOURCE_LIMITS_EXCEEDED` at ~9 mutations per request. When processing more than 8 threads, split into chunks.
- **Always** fetch CI check statuses for PR-sourced triage runs before analysis (Step 0b)
- **Never** dismiss or disable a failing check (lint rule, test skip, `//nolint` directive, etc.) without explicit operator approval and a concrete justification
- **Always** label CI failure references as `CI-N` (distinct from review comment numbers `#N`) to avoid ambiguity throughout the workflow

## Thread Management Reference

For manual thread management outside the triage workflow:

**Unresolve a thread** (if a resolved thread needs to be reopened):
```bash
QUERY=$(cat <<'GQL'
mutation($threadId: ID!) {
  unresolveReviewThread(input: {threadId: $threadId}) {
    thread { id isResolved }
  }
}
GQL
)
gh api graphql -f query="$QUERY" -F threadId="$THREAD_ID"
```

This is not used by the triage workflow itself but is documented here for manual recovery if a thread was resolved prematurely.
