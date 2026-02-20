# Implementation Plan: Branching Strategy

**Spec**: `.specify/spec/branching-strategy/spec.md`
**Created**: 2026-02-20

## Overview

This plan covers the changes needed to support hierarchical
feature branches (`NNN-FEATURE` → `NNN-FEATURE/spec`,
`NNN-FEATURE/phase-N`). The scope is: shell scripts, CI
workflows, labeler config, release-drafter config, and
documentation.

---

## Phase 1: Script Changes

### 1.1 `create-new-feature.sh` — Push to reserve number

After `git checkout -b "$BRANCH_NAME"` (line 275), add:

```bash
git push -u origin "$BRANCH_NAME"
```

This reserves the feature number on the remote immediately,
closing the race condition window. The push MUST happen before
creating the spec directory to ensure a failed push doesn't
leave orphaned local state.

**Error handling**: If the push fails (e.g., no remote, auth
failure), the script SHOULD warn but NOT exit — the branch is
still useful locally. The warning tells the developer to push
manually.

### 1.2 `create-new-feature.sh` — Create `/spec` sub-branch

After pushing the feature branch, also create and checkout
the spec sub-branch:

```bash
git checkout -b "${BRANCH_NAME}/spec"
git push -u origin "${BRANCH_NAME}/spec"
```

The spec directory and template copy already happen after this
point, so they'll land on the `/spec` branch automatically.

Update the JSON and plain-text output to also include
`SPEC_BRANCH`.

### 1.3 `common.sh` — No changes needed

`check_feature_branch()` and `find_feature_dir_by_prefix()`
already support the hierarchical pattern. Verified in the
current implementation:

- `check_feature_branch()` accepts `NNN-name`, `NNN-name/spec`,
  `NNN-name/slug` (regex: `^[0-9]{3}-`)
- `find_feature_dir_by_prefix()` strips `/suffix` via
  `${branch_name%%/*}`

---

## Phase 2: Speckit Prompt Changes

The speckit command prompts (`.claude/commands/speckit.*.md`)
contain hardcoded branch instructions that conflict with the
hierarchical model. These are the instructions that Claude
agents follow — they MUST be updated to match the new strategy.

### 2.1 `speckit.specify.md` — Branch creation flow

**Current behavior** (step 2e, line 62–66):
1. `create-new-feature.sh` creates and checks out `NNN-NAME`
2. Prompt then renames it: `git branch -m $BRANCH_NAME/spec`

This is wrong for two reasons:
- The feature integration branch (`NNN-NAME`) never exists on
  the remote — it's immediately renamed to `/spec`
- There's no integration branch to merge phase PRs into later

**New behavior**:
1. `create-new-feature.sh` creates `NNN-NAME` and pushes it
   (Phase 1.1 + 1.2 changes above)
2. Script also creates `NNN-NAME/spec` sub-branch
3. Prompt no longer needs the `git branch -m` rename step

**Changes to `speckit.specify.md`**:
- Remove step 2e (the `git branch -m` rename)
- Update the explanation text to describe the new two-branch
  creation: feature integration branch + spec sub-branch
- Update the IMPORTANT notes to mention that the script now
  creates and pushes both branches
- Remove the duplicate branch-checking logic in steps 2a–2c
  (this is already handled by `create-new-feature.sh`'s
  `check_existing_branches` function — the prompt currently
  duplicates this work by running `git ls-remote`, `git branch`,
  and scanning `specs/` before calling the script)

### 2.2 `speckit.implement.md` — Phase branches off feature, not main

**Current behavior** (step 2, lines 35–44):
```
Phase branches: Each implementation phase [...] gets its own
branch off `main`
```
```bash
git checkout main && git pull
git checkout -b <feature>/<phase-slug>
```

This is the core change — phase branches MUST branch off the
feature integration branch, not `main`.

**New behavior**:
```
Phase branches: Each implementation phase [...] gets its own
branch off the feature integration branch
```
```bash
git checkout <feature> && git pull origin <feature>
git merge main  # keep feature branch current
git checkout -b <feature>/<phase-slug>
```

**Changes to `speckit.implement.md`**:
- Step 2: Replace "branch off `main`" with "branch off the
  feature integration branch (`NNN-FEATURE`)"
- Step 2 branch creation: Change from
  `git checkout main && git pull` to
  `git checkout NNN-FEATURE && git pull origin NNN-FEATURE`
- Add `git merge main` before creating phase branch (keep
  feature branch current per spec requirement)
- Step 2 PR target: Add explicit note that phase PRs target
  the feature integration branch, not `main`
- Step 2 completion: After all phases merged into feature
  branch, note that a final PR from `NNN-FEATURE` → `main`
  completes the feature

### 2.3 `speckit.implement.md` — Feature detection from branch

**Current behavior** (step 1, lines 16–20):
- From `main`: detect feature from recent specs directories
- From feature branch: auto-detect from branch name

**New behavior**: Implementation should typically start from the
feature integration branch (`NNN-FEATURE`), not from `main`.
The prompt should reflect this:
- From `NNN-FEATURE`: auto-detect feature, create phase branch
- From `NNN-FEATURE/phase`: already on a phase branch, continue
- From `main`: still supported but less common — detect feature
  from context

### 2.4 `speckit.plan.md`, `speckit.tasks.md`, etc. — No changes needed

These commands run on the `/spec` branch and don't create
branches. They use `check-prerequisites.sh` which already
resolves the feature dir correctly via `find_feature_dir_by_prefix()`.

No prompt changes needed for: `speckit.plan`, `speckit.tasks`,
`speckit.clarify`, `speckit.analyze`, `speckit.checklist`,
`speckit.constitution`, `speckit.taskstoissues`.

---

## Phase 3: CI Workflow Changes

### 3.1 `pull-request.yml` — PR-size conditional

**Problem**: The PR-size check (750 error / 333 warning) makes
sense for phase-to-feature PRs (small increments), but NOT for
the final feature-to-main PR which aggregates all phases.

**Solution**: Add a condition that skips the size check when the
PR targets `main` AND the source branch matches the feature
integration pattern (`NNN-*` without a `/`).

```yaml
pr-size:
  name: PR Size Check
  runs-on: ubuntu-latest
  if: >-
    !(
      github.event.pull_request.base.ref == 'main' &&
      contains(github.event.pull_request.head.ref, '/') == false
    )
  permissions:
    pull-requests: write
  steps:
    - uses: ookami-kb/gh-pr-size-watcher@v1.5.0
      ...
```

**Logic**: If the base is `main` AND the head branch has no `/`
(meaning it's a feature integration branch like `004-auth`, not
a phase branch like `004-auth/spec`), skip the size check.

**Alternative considered**: Using the `allow-large-pr` label
exclusion that's already configured. This works but requires
manual labeling. The conditional approach is automatic and
convention-based.

**Decision**: Use BOTH — the automatic conditional for the
common case, and keep the `allow-large-pr` label as an escape
hatch for edge cases.

### 3.2 `pull-request.yml` — Validate runs on all PRs ✅

The validate job already triggers on all `pull_request` events
regardless of target branch. No changes needed.

### 3.3 `pull-request.yml` — Labeler runs on all PRs ✅

The labeler already triggers on all PRs. No changes needed.

### 3.4 `main.yml` — Release-drafter only on main pushes ✅

Already scoped to `push: branches: [main]`. No changes needed.

---

## Phase 4: Label & Release-Drafter Strategy

### 4.1 Current label categories (no changes needed)

The existing labeler categories work correctly with the
hierarchical model:

| Label | Applies to |
|-------|-----------|
| `type:code` | Any PR with Go source changes |
| `type:test` | Any PR with `*_test.go` |
| `type:bdd` | Any PR with `features/*` |
| `type:docs` | Any PR with `*.md` |
| `type:specification` | Any PR with `specs/*` or `.specify/*` |
| `type:tooling` | Any PR with `.claude/*` etc. |
| `type:build-definition` | Any PR with CI/config files |
| `version:major/minor/patch` | Based on PR title prefix |
| `size:s/m/l` | Based on additions + deletions |

Phase PRs (into feature branch) get labeled normally. The
feature-to-main PR also gets labeled — and this is the one
that matters for release-drafter since release-drafter only
picks up PRs merged into the default branch.

### 4.2 Release-drafter — Specification category

**Current state**: The `release-drafter.yml` has categories for
Breaking Changes, Features, Bug Fixes, Documentation, and
Build & CI.

**Gap**: No category for `type:specification` or `type:bdd`.
These labels are applied by the labeler but don't appear in
release notes.

**Solution**: Add a "Specification" category:

```yaml
categories:
  - title: "Breaking Changes"
    labels:
      - "version:major"
  - title: "Features"
    labels:
      - "version:minor"
  - title: "Bug Fixes"
    labels:
      - "version:patch"
  - title: "Specification"
    labels:
      - "type:specification"
  - title: "Testing"
    labels:
      - "type:bdd"
      - "type:test"
  - title: "Documentation"
    labels:
      - "type:docs"
  - title: "Build & CI"
    labels:
      - "type:build-definition"
```

### 4.3 Release-drafter — Feature integration PR labeling

**Consideration**: When a feature integration branch merges into
`main`, it aggregates all phase work. The PR title should follow
conventional commits (e.g., `feat: add user authentication`),
which triggers `version:minor`. The type labels are applied based
on files changed, so a feature PR touching Go code, tests, and
specs will get `type:code`, `type:test`, `type:specification` etc.

Release-drafter assigns a PR to the **first matching category**,
so the version label (`version:minor`) takes priority. This is
correct behavior — the feature PR appears under "Features" in
the release notes.

**No changes needed** for this flow.

### 4.4 Label consideration — `type:tooling`

The `type:tooling` label has no release-drafter category. This
is intentional — tooling changes (`.claude/`, `.cursorrules`,
etc.) are internal and don't belong in user-facing release notes.
If a tooling PR also touches code, the version label handles it.

---

## Phase 5: Documentation

### 5.1 CLAUDE.md — Branching strategy section

Add a "Branching Strategy" subsection under "Development
Workflow" in `CLAUDE.md`:

```markdown
### Branching Strategy

Feature development uses hierarchical branches:

- `NNN-FEATURENAME` — integration branch (merges to `main`)
- `NNN-FEATURENAME/spec` — specification work
- `NNN-FEATURENAME/PHASE` — implementation phases

Phase PRs target the feature branch. Only the feature
branch targets `main`. See
`.specify/spec/branching-strategy/spec.md` for details.
```

### 5.2 Constitution — No changes needed

The constitution already references Governance as Foundation
(Principle VI) which covers traceability. The branching
strategy is a process implementation of that principle, not a
new principle.

---

## Phase 6: Speckit Command Validation (Future)

### 6.1 Branch-type validation in speckit commands

**Scope**: This is a nice-to-have, not blocking. The speckit
commands (bash scripts) could validate they're running on the
correct branch type:

| Command | Expected branch |
|---------|----------------|
| `speckit.specify` | `main` (creates feature + spec branch) |
| `speckit.plan`, `speckit.tasks`, etc. | `NNN-*/spec` |
| `speckit.implement` | `NNN-*/PHASE` (not `/spec`) |

**Decision**: Defer to a separate task. The current scripts work
without this validation — it's a developer convenience guard
rail, not a functional requirement for the branching strategy
itself.

---

## Implementation Order

1. **Phase 1** (Scripts) — `create-new-feature.sh` push +
   `/spec` sub-branch creation
2. **Phase 2** (Prompts) — `speckit.specify.md` and
   `speckit.implement.md` branch flow updates
3. **Phase 3** (CI) — `pull-request.yml` PR-size conditional
4. **Phase 4** (Labels) — `release-drafter.yml` categories
5. **Phase 5** (Docs) — `CLAUDE.md` branching section

Phases 1–2 are tightly coupled (script behavior must match
prompt instructions). Phases 3–5 are independent.

All phases can be delivered in a single PR — estimated ~80
lines across 6 files:
- `.specify/scripts/bash/create-new-feature.sh`
- `.claude/commands/speckit.specify.md`
- `.claude/commands/speckit.implement.md`
- `.github/workflows/pull-request.yml`
- `.github/release-drafter.yml`
- `CLAUDE.md`

Phase 6 is deferred.

---

## Risk Assessment

| Risk | Mitigation |
|------|-----------|
| Push fails in `create-new-feature.sh` | Warn, don't exit; developer pushes manually |
| PR-size conditional regex mismatch | `allow-large-pr` label as fallback |
| Release-drafter category ordering | Version labels listed first, take priority |
| Speckit commands run on wrong branch | Deferred; current behavior is permissive but functional |
| Prompt changes not picked up by agents | Agents always re-read prompts; no caching |
| Duplicate branch-checking in specify prompt | Removing it simplifies the prompt and avoids divergence with the script |

## Open Questions

None — all decisions resolved in spec and this plan.
